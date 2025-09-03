// Package repository provides database repository implementations.
package repository

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"httpchat/internal/interfaces"
	"httpchat/internal/model"
	"httpchat/internal/repositoryerr"

	"github.com/lib/pq"
)

// PostgreSQLMessageRepository implements interfaces.MessageRepository for PostgreSQL
type PostgreSQLMessageRepository struct {
	db *sql.DB
}

// NewPostgreSQLMessageRepository creates a new PostgreSQLMessageRepository
func NewPostgreSQLMessageRepository(databaseURL string) (interfaces.MessageRepository, error) {
	// Open database connection
	db, err := sql.Open("postgres", databaseURL)
	if err != nil {
		return nil, &repositoryerr.RepositoryError{
			Op:  "NewPostgreSQLMessageRepository",
			Err: fmt.Errorf("failed to open database connection: %w", err),
		}
	}

	// Configure connection pool
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(25)
	db.SetConnMaxLifetime(5 * time.Minute)

	// Test connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := db.PingContext(ctx); err != nil {
		return nil, &repositoryerr.RepositoryError{
			Op:  "NewPostgreSQLMessageRepository",
			Err: fmt.Errorf("failed to ping database: %w", err),
		}
	}

	// Create messages table if it doesn't exist
	if err := createMessagesTable(db); err != nil {
		return nil, &repositoryerr.RepositoryError{
			Op:  "NewPostgreSQLMessageRepository",
			Err: fmt.Errorf("failed to create messages table: %w", err),
		}
	}

	// Return repository implementation
	return &PostgreSQLMessageRepository{
		db: db,
	}, nil
}

// Ensure PostgreSQLMessageRepository implements interfaces.MessageRepository
var _ interfaces.MessageRepository = (*PostgreSQLMessageRepository)(nil)

// createMessagesTable creates the messages table if it doesn't exist
func createMessagesTable(db *sql.DB) error {
	// Create the messages table with all required fields
	query := `
	CREATE TABLE IF NOT EXISTS messages (
		id SERIAL PRIMARY KEY,
		content TEXT NOT NULL,
		processed BOOLEAN NOT NULL DEFAULT FALSE,
		created_at TIMESTAMP NOT NULL DEFAULT NOW(),
		updated_at TIMESTAMP NOT NULL DEFAULT NOW()
	)`

	_, err := db.Exec(query)
	if err != nil {
		return repositoryerr.New(
			"", // No specific code
			"createMessagesTable",
			fmt.Errorf("failed to create messages table: %w", err),
		)
	}

	// Create indexes for better query performance
	indexes := []string{
		`CREATE INDEX IF NOT EXISTS idx_messages_processed ON messages(processed)`,
		`CREATE INDEX IF NOT EXISTS idx_messages_created_at ON messages(created_at)`,
	}

	// Create each index
	for _, idxQuery := range indexes {
		_, err = db.Exec(idxQuery)
		if err != nil {
			return repositoryerr.New(
				"", // No specific code
				"createMessagesTable",
				fmt.Errorf("failed to create index: %w", err),
			)
		}
	}

	return nil
}

// CreateMessage creates a new message in the database
func (r *PostgreSQLMessageRepository) CreateMessage(ctx context.Context, content string) (*model.Message, error) {
	// SQL query to insert a new message and return the created record
	query := `
	INSERT INTO messages (content, created_at, updated_at)
	VALUES ($1, $2, $3)
	RETURNING id, content, processed, created_at, updated_at`

	var message model.Message
	now := time.Now()
	
	// Execute the query and scan the result into our message struct
	err := r.db.QueryRowContext(ctx, query, content, now, now).Scan(
		&message.ID,
		&message.Content,
		&message.Processed,
		&message.CreatedAt,
		&message.UpdatedAt,
	)
	
	if err != nil {
		// Handle specific PostgreSQL error codes for better error reporting
		if pqErr, ok := err.(*pq.Error); ok {
			switch pqErr.Code {
			case "23505": // unique_violation
				return nil, repositoryerr.New(
					repositoryerr.ErrorCodeDuplicateEntry,
					"CreateMessage",
					fmt.Errorf("duplicate message entry: %w", err),
				)
			case "23502": // not_null_violation
				return nil, repositoryerr.New(
					repositoryerr.ErrorCodeInvalidInput,
					"CreateMessage",
					fmt.Errorf("missing required field: %w", err),
				)
			case "23503": // foreign_key_violation
				return nil, repositoryerr.New(
					repositoryerr.ErrorCodeInvalidInput,
					"CreateMessage",
					fmt.Errorf("foreign key violation: %w", err),
				)
			case "25P02": // in_failed_sql_transaction
				return nil, repositoryerr.New(
					repositoryerr.ErrorCodeTransactionFailed,
					"CreateMessage",
					fmt.Errorf("transaction failed: %w", err),
				)
			}
		}
		return nil, repositoryerr.New(
			"", // No specific code
			"CreateMessage",
			fmt.Errorf("failed to insert message: %w", err),
		)
	}

	return &message, nil
}

// GetMessageByID retrieves a message by ID from the database
func (r *PostgreSQLMessageRepository) GetMessageByID(ctx context.Context, id int64) (*model.Message, error) {
	// SQL query to get a message by its ID
	query := `
	SELECT id, content, processed, created_at, updated_at
	FROM messages
	WHERE id = $1`

	var message model.Message
	
	// Execute the query and scan the result into our message struct
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&message.ID,
		&message.Content,
		&message.Processed,
		&message.CreatedAt,
		&message.UpdatedAt,
	)
	
	if err != nil {
		// Handle case when no message is found
		if err == sql.ErrNoRows {
			return nil, repositoryerr.New(
				repositoryerr.ErrorCodeMessageNotFound,
				"GetMessageByID",
				repositoryerr.ErrMessageNotFound,
			)
		}

		// Handle specific PostgreSQL error codes
		if pqErr, ok := err.(*pq.Error); ok {
			switch pqErr.Code {
			case "22P02": // invalid_text_representation
				return nil, repositoryerr.New(
					repositoryerr.ErrorCodeInvalidInput,
					"GetMessageByID",
					fmt.Errorf("invalid message ID format: %w", err),
				)
			case "25P02": // in_failed_sql_transaction
				return nil, repositoryerr.New(
					repositoryerr.ErrorCodeTransactionFailed,
					"GetMessageByID",
					fmt.Errorf("transaction failed: %w", err),
				)
			}
		}

		return nil, repositoryerr.New(
			"", // No specific code
			"GetMessageByID",
			fmt.Errorf("failed to get message: %w", err),
		)
	}

	return &message, nil
}

// UpdateMessageStatus updates a message's status in the database
func (r *PostgreSQLMessageRepository) UpdateMessageStatus(ctx context.Context, id int64, processed bool) error {
	// SQL query to update the processed status of a message
	query := `
	UPDATE messages
	SET processed = $1, updated_at = $2
	WHERE id = $3`

	now := time.Now()
	
	// Execute the update query
	result, err := r.db.ExecContext(ctx, query, processed, now, id)
	if err != nil {
		// Handle specific PostgreSQL error codes
		if pqErr, ok := err.(*pq.Error); ok {
			switch pqErr.Code {
			case "22P02": // invalid_text_representation
				return repositoryerr.New(
					repositoryerr.ErrorCodeInvalidInput,
					"UpdateMessageStatus",
					fmt.Errorf("invalid message ID format: %w", err),
				)
			case "25P02": // in_failed_sql_transaction
				return repositoryerr.New(
					repositoryerr.ErrorCodeTransactionFailed,
					"UpdateMessageStatus",
					fmt.Errorf("transaction failed: %w", err),
				)
			}
		}

		return repositoryerr.New(
			"", // No specific code
			"UpdateMessageStatus",
			fmt.Errorf("failed to update message: %w", err),
		)
	}

	// Check how many rows were affected to ensure the message existed
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return repositoryerr.New(
			repositoryerr.ErrorCodeSerializationFailed,
			"UpdateMessageStatus",
			fmt.Errorf("failed to get rows affected: %w", err),
		)
	}

	// If no rows were affected, the message didn't exist
	if rowsAffected == 0 {
		return repositoryerr.New(
			repositoryerr.ErrorCodeMessageNotFound,
			"UpdateMessageStatus",
			repositoryerr.ErrMessageNotFound,
		)
	}

	return nil
}

// GetAllMessages retrieves all messages from the database
func (r *PostgreSQLMessageRepository) GetAllMessages(ctx context.Context) ([]*model.Message, error) {
	// SQL query to get all messages ordered by creation time
	query := `
	SELECT id, content, processed, created_at, updated_at
	FROM messages
	ORDER BY created_at DESC`

	// Execute the query
	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		// Handle specific PostgreSQL error codes
		if pqErr, ok := err.(*pq.Error); ok {
			switch pqErr.Code {
			case "25P02": // in_failed_sql_transaction
				return nil, repositoryerr.New(
					repositoryerr.ErrorCodeTransactionFailed,
					"GetAllMessages",
					fmt.Errorf("transaction failed: %w", err),
				)
			}
		}

		return nil, repositoryerr.New(
			"", // No specific code
			"GetAllMessages",
			fmt.Errorf("failed to query messages: %w", err),
		)
	}
	
	// Ensure rows are closed when function returns
	defer func() {
		_ = rows.Close()
	}()

	// Process each row and build our messages slice
	var messages []*model.Message
	for rows.Next() {
		var message model.Message
		err := rows.Scan(
			&message.ID,
			&message.Content,
			&message.Processed,
			&message.CreatedAt,
			&message.UpdatedAt,
		)
		if err != nil {
			return nil, repositoryerr.New(
				repositoryerr.ErrorCodeSerializationFailed,
				"GetAllMessages",
				fmt.Errorf("failed to scan message: %w", err),
			)
		}
		messages = append(messages, &message)
	}

	// Check for errors during iteration
	if err := rows.Err(); err != nil {
		return nil, repositoryerr.New(
			repositoryerr.ErrorCodeSerializationFailed,
			"GetAllMessages",
			fmt.Errorf("error iterating rows: %w", err),
		)
	}

	return messages, nil
}

// GetStatistics retrieves message statistics from the database
func (r *PostgreSQLMessageRepository) GetStatistics(ctx context.Context) (*model.Statistics, error) {
	// SQL query to get message statistics using COUNT with CASE conditions
	query := `
	SELECT 
		COUNT(*) as total_messages,
		COUNT(CASE WHEN processed = TRUE THEN 1 END) as processed_messages,
		COUNT(CASE WHEN processed = FALSE THEN 1 END) as unprocessed_messages
	FROM messages`

	var stats model.Statistics
	
	// Execute the query and scan results into our statistics struct
	err := r.db.QueryRowContext(ctx, query).Scan(
		&stats.TotalMessages,
		&stats.ProcessedMessages,
		&stats.UnprocessedMessages,
	)
	
	if err != nil {
		// Handle specific PostgreSQL error codes
		if pqErr, ok := err.(*pq.Error); ok {
			switch pqErr.Code {
			case "25P02": // in_failed_sql_transaction
				return nil, repositoryerr.New(
					repositoryerr.ErrorCodeTransactionFailed,
					"GetStatistics",
					fmt.Errorf("transaction failed: %w", err),
				)
			}
		}

		return nil, repositoryerr.New(
			"", // No specific code
			"GetStatistics",
			fmt.Errorf("failed to get statistics: %w", err),
		)
	}

	return &stats, nil
}
