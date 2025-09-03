package repository

import (
	"context"

	"httpchat/internal/model"
)

// MessageRepository defines the interface for working with messages in the storage
type MessageRepository interface {
	CreateMessage(ctx context.Context, content string) (int64, error)
	GetMessageByID(ctx context.Context, id int64) (*model.Message, error)
	UpdateMessageStatus(ctx context.Context, id int64, processed bool) error
	GetAllMessages(ctx context.Context) ([]*model.Message, error)
	GetStatistics(ctx context.Context) (*model.Statistics, error)
}