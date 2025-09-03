// Package service provides business logic implementations.
package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"httpchat/internal/interfaces"
	"httpchat/internal/logger"
	"httpchat/internal/model"
	"httpchat/internal/repositoryerr"

	"go.uber.org/zap"
)

// messageService implements interfaces.MessageService
type messageService struct {
	repo     interfaces.MessageRepository
	producer interfaces.KafkaProducer
	consumer interfaces.KafkaConsumer
	topic    string
	logger   *logger.Logger
}

// NewMessageService creates a new messageService instance
func NewMessageService(
	repo interfaces.MessageRepository,
	producer interfaces.KafkaProducer,
	consumer interfaces.KafkaConsumer,
	topic string,
	logger *logger.Logger,
) interfaces.MessageService {
	return &messageService{
		repo:     repo,
		producer: producer,
		consumer: consumer,
		topic:    topic,
		logger:   logger,
	}
}

// Ensure messageService implements interfaces.MessageService
var _ interfaces.MessageService = (*messageService)(nil)

// handleError is a helper function to handle repository errors consistently
func (s *messageService) handleError(op string, err error, id int64) error {
	var repoErr *repositoryerr.RepositoryError
	if errors.As(err, &repoErr) {
		switch repoErr.ErrorCode() {
		case repositoryerr.ErrorCodeMessageNotFound:
			s.logger.Warn(fmt.Sprintf("Message not found during %s", op), zap.Int64("id", id), zap.Error(err))
			return fmt.Errorf("message not found: %w", err)
		case repositoryerr.ErrorCodeInvalidInput:
			s.logger.Warn(fmt.Sprintf("Invalid input during %s", op), zap.Int64("id", id), zap.Error(err))
			return fmt.Errorf("invalid input: %w", err)
		case repositoryerr.ErrorCodeDatabaseConnection:
			s.logger.Error(fmt.Sprintf("Database connection error during %s", op), zap.Error(err))
			return fmt.Errorf("database unavailable: %w", err)
		case repositoryerr.ErrorCodeDuplicateEntry:
			s.logger.Warn(fmt.Sprintf("Duplicate entry during %s", op), zap.Int64("id", id), zap.Error(err))
			return fmt.Errorf("duplicate entry: %w", err)
		default:
			s.logger.Error(fmt.Sprintf("Failed during %s", op), zap.Int64("id", id), zap.Error(err))
			return fmt.Errorf("operation failed: %w", err)
		}
	}
	s.logger.Error(fmt.Sprintf("Failed during %s", op), zap.Int64("id", id), zap.Error(err))
	return fmt.Errorf("operation failed: %w", err)
}

// CreateMessage creates a new message and sends it to Kafka
func (s *messageService) CreateMessage(ctx context.Context, content string) (int64, error) {
	s.logger.Info("Creating message in repository", zap.String("content", content))

	// Step 1: Save the message to the database
	message, err := s.repo.CreateMessage(ctx, content)
	if err != nil {
		return 0, s.handleError("message creation", err, 0)
	}

	s.logger.Info("Successfully created message in repository", zap.Int64("id", message.ID))

	// Step 2: Convert the message to JSON for sending to Kafka
	messageBytes, err := json.Marshal(message)
	if err != nil {
		s.logger.Error("Failed to marshal message", zap.Int64("id", message.ID), zap.Error(err))
		return 0, fmt.Errorf("failed to marshal message: %w", err)
	}

	s.logger.Debug("Sending message to Kafka", zap.Int64("id", message.ID))

	// Step 3: Send the message to Kafka for processing
	if err := s.producer.SendMessage(ctx, s.topic, messageBytes); err != nil {
		s.logger.Error("Failed to send message to Kafka", zap.Int64("id", message.ID), zap.Error(err))
		return 0, fmt.Errorf("failed to send message to Kafka: %w", err)
	}

	s.logger.Info("Successfully sent message to Kafka", zap.Int64("id", message.ID))

	return message.ID, nil
}

// ProcessMessage marks a message as processed
func (s *messageService) ProcessMessage(ctx context.Context, id int64) error {
	s.logger.Info("Processing message", zap.Int64("id", id))

	// Update the message status in the database to mark it as processed
	if err := s.repo.UpdateMessageStatus(ctx, id, true); err != nil {
		return s.handleError("message processing", err, id)
	}

	s.logger.Info("Successfully processed message", zap.Int64("id", id))

	return nil
}

// GetStatistics returns message statistics
func (s *messageService) GetStatistics(ctx context.Context) (*model.Statistics, error) {
	s.logger.Info("Fetching statistics from repository")

	// Get message statistics from the database
	stats, err := s.repo.GetStatistics(ctx)
	if err != nil {
		var repoErr *repositoryerr.RepositoryError
		if errors.As(err, &repoErr) {
			switch repoErr.ErrorCode() {
			case repositoryerr.ErrorCodeDatabaseConnection:
				s.logger.Error("Database connection error during statistics retrieval", zap.Error(err))
				return nil, fmt.Errorf("database unavailable: %w", err)
			default:
				s.logger.Error("Failed to get statistics from repository", zap.Error(err))
				return nil, fmt.Errorf("failed to get statistics from repository: %w", err)
			}
		}
		s.logger.Error("Failed to get statistics from repository", zap.Error(err))
		return nil, fmt.Errorf("failed to get statistics from repository: %w", err)
	}

	s.logger.Info("Successfully fetched statistics", 
		zap.Int64("total", stats.TotalMessages),
		zap.Int64("processed", stats.ProcessedMessages),
		zap.Int64("unprocessed", stats.UnprocessedMessages))

	return stats, nil
}
