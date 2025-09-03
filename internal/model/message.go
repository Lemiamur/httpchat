// Package model provides data models for the application.
package model

import (
	"time"
)

// Message represents a message in the system
type Message struct {
	ID        int64     `json:"id" db:"id"`
	Content   string    `json:"content" db:"content"`
	Processed bool      `json:"processed" db:"processed"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
}

// Statistics represents message statistics
type Statistics struct {
	TotalMessages       int64 `json:"total_messages" db:"total_messages"`
	ProcessedMessages   int64 `json:"processed_messages" db:"processed_messages"`
	UnprocessedMessages int64 `json:"unprocessed_messages" db:"unprocessed_messages"`
}
