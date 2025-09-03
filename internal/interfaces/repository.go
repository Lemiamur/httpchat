// Package interfaces provides interface definitions for the application.
package interfaces

import (
	"context"

	"httpchat/internal/model"
)

// MessageRepository defines the interface for message repository operations
type MessageRepository interface {
	CreateMessage(ctx context.Context, content string) (*model.Message, error)
	GetMessageByID(ctx context.Context, id int64) (*model.Message, error)
	UpdateMessageStatus(ctx context.Context, id int64, processed bool) error
	GetAllMessages(ctx context.Context) ([]*model.Message, error)
	GetStatistics(ctx context.Context) (*model.Statistics, error)
}