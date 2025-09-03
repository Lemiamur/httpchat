package repository

import (
	"context"
	"database/sql"
	"log"
	"os"
	"testing"
	"time"

	"httpchat/internal/interfaces"

	_ "github.com/lib/pq"
	"github.com/stretchr/testify/assert"
)

var testDB *sql.DB

func TestMain(m *testing.M) {
	// Initialize test database connection
	// This assumes you have a test database running
	databaseURL := os.Getenv("TEST_DATABASE_URL")
	if databaseURL == "" {
		databaseURL = "postgres://user:password@localhost:5432/messages_test_db?sslmode=disable"
	}

	var err error
	testDB, err = sql.Open("postgres", databaseURL)
	if err != nil {
		log.Fatal("Failed to open database:", err)
	}

	defer func() {
		_ = testDB.Close()
	}()

	// Create test table
	_, err = testDB.Exec(`
		CREATE TABLE IF NOT EXISTS messages (
			id SERIAL PRIMARY KEY,
			content TEXT NOT NULL,
			processed BOOLEAN NOT NULL DEFAULT FALSE,
			created_at TIMESTAMP NOT NULL DEFAULT NOW(),
			updated_at TIMESTAMP NOT NULL DEFAULT NOW()
		)
	`)
	if err != nil {
		log.Fatal("Failed to create test table:", err)
	}

	// Run tests
	code := m.Run()

	// Clean up test table
	_, err = testDB.Exec(`DROP TABLE IF EXISTS messages`)
	if err != nil {
		log.Println("Failed to drop test table:", err)
	}

	os.Exit(code)
}

func setupTestRepository() interfaces.MessageRepository {
	return &PostgreSQLMessageRepository{db: testDB}
}

func cleanupTestData(t *testing.T) {
	_, err := testDB.Exec(`DELETE FROM messages`)
	if err != nil {
		t.Fatal("Failed to clean up test data:", err)
	}
}

func TestPostgreSQLMessageRepository_Integration(t *testing.T) {
	repo := setupTestRepository()

	// Verify that PostgreSQLMessageRepository implements interfaces.MessageRepository
	var _ interfaces.MessageRepository = &PostgreSQLMessageRepository{}

	// Clean up before test
	cleanupTestData(t)

	t.Run("CreateMessage", func(t *testing.T) {
		// Clean up before test
		cleanupTestData(t)

		message, err := repo.CreateMessage(context.Background(), "Test message")
		assert.NoError(t, err)
		assert.True(t, message.ID > 0)

		// Verify the message was created
		retrievedMessage, err := repo.GetMessageByID(context.Background(), message.ID)
		assert.NoError(t, err)
		assert.Equal(t, message.ID, retrievedMessage.ID)
		assert.Equal(t, "Test message", retrievedMessage.Content)
		assert.False(t, retrievedMessage.Processed)
		assert.WithinDuration(t, retrievedMessage.CreatedAt, retrievedMessage.UpdatedAt, 1*time.Second)
	})

	t.Run("GetMessageByID", func(t *testing.T) {
		// Clean up before test
		cleanupTestData(t)

		// Create a test message
		message, err := repo.CreateMessage(context.Background(), "Test message")
		assert.NoError(t, err)

		// Get the message
		retrievedMessage, err := repo.GetMessageByID(context.Background(), message.ID)
		assert.NoError(t, err)
		assert.Equal(t, message.ID, retrievedMessage.ID)
		assert.Equal(t, "Test message", retrievedMessage.Content)
		assert.False(t, retrievedMessage.Processed)

		// Try to get a non-existent message
		_, err = repo.GetMessageByID(context.Background(), 999999)
		assert.Error(t, err)
	})

	t.Run("UpdateMessageStatus", func(t *testing.T) {
		// Clean up before test
		cleanupTestData(t)

		// Create a test message
		message, err := repo.CreateMessage(context.Background(), "Test message")
		assert.NoError(t, err)

		// Update message status
		err = repo.UpdateMessageStatus(context.Background(), message.ID, true)
		assert.NoError(t, err)

		// Get the message to verify
		retrievedMessage, err := repo.GetMessageByID(context.Background(), message.ID)
		assert.NoError(t, err)
		assert.True(t, retrievedMessage.Processed)
		assert.True(t, retrievedMessage.UpdatedAt.After(retrievedMessage.CreatedAt))

		// Try to update a non-existent message
		err = repo.UpdateMessageStatus(context.Background(), 999999, true)
		assert.Error(t, err)
	})

	t.Run("GetAllMessages", func(t *testing.T) {
		// Clean up before test
		cleanupTestData(t)

		// Create test messages
		message1, err := repo.CreateMessage(context.Background(), "Test message 1")
		assert.NoError(t, err)

		message2, err := repo.CreateMessage(context.Background(), "Test message 2")
		assert.NoError(t, err)

		// Get all messages
		messages, err := repo.GetAllMessages(context.Background())
		assert.NoError(t, err)
		assert.Len(t, messages, 2)

		// Messages should be ordered by created_at DESC
		assert.Equal(t, message2.ID, messages[0].ID)
		assert.Equal(t, message1.ID, messages[1].ID)
	})

	t.Run("GetStatistics", func(t *testing.T) {
		// Clean up before test
		cleanupTestData(t)

		// Create test messages
		message1, err := repo.CreateMessage(context.Background(), "Test message 1")
		assert.NoError(t, err)

		_, err = repo.CreateMessage(context.Background(), "Test message 2")
		assert.NoError(t, err)

		// Update one message to processed
		err = repo.UpdateMessageStatus(context.Background(), message1.ID, true)
		assert.NoError(t, err)

		// Get statistics
		stats, err := repo.GetStatistics(context.Background())
		assert.NoError(t, err)
		assert.Equal(t, int64(2), stats.TotalMessages)
		assert.Equal(t, int64(1), stats.ProcessedMessages)
		assert.Equal(t, int64(1), stats.UnprocessedMessages)
	})
}

func TestPostgreSQLMessageRepository_ErrorHandling(t *testing.T) {
	repo := setupTestRepository()

	// Clean up before test
	cleanupTestData(t)

	t.Run("GetMessageByID_NotFound", func(t *testing.T) {
		_, err := repo.GetMessageByID(context.Background(), 999999)
		assert.Error(t, err)
		
		// Check if it's the expected error type
		// Note: We can't directly check the error type because it's wrapped
		assert.Contains(t, err.Error(), "message not found")
	})

	t.Run("UpdateMessageStatus_NotFound", func(t *testing.T) {
		err := repo.UpdateMessageStatus(context.Background(), 999999, true)
		assert.Error(t, err)
		
		// Check if it's the expected error type
		// Note: We can't directly check the error type because it's wrapped
		assert.Contains(t, err.Error(), "message not found")
	})
}