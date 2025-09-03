// Package interfaces provides interface definitions for the application components.
package interfaces

import (
	"context"

	"httpchat/internal/model"
)

// MessageService defines the interface for message-related business logic
type MessageService interface {
	// CreateMessage creates a new message and sends it to Kafka
	CreateMessage(ctx context.Context, content string) (int64, error)

	// ProcessMessage marks a message as processed
	ProcessMessage(ctx context.Context, id int64) error

	// GetStatistics returns message statistics
	GetStatistics(ctx context.Context) (*model.Statistics, error)
}
