package kafka

import "context"

// Producer defines the interface for sending messages to Kafka
type Producer interface {
	SendMessage(ctx context.Context, topic string, message []byte) error
	Close() error
}

// Consumer defines the interface for reading messages from Kafka
type Consumer interface {
	ReadMessage(ctx context.Context, topic string) ([]byte, error)
	Close() error
}