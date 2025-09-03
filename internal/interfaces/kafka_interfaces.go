// Package interfaces provides interface definitions for the application components.
package interfaces

import (
	"context"

	"github.com/segmentio/kafka-go"
)

// KafkaProducer defines the interface for Kafka message production
type KafkaProducer interface {
	SendMessage(ctx context.Context, topic string, message []byte) error
	Close() error
}

// KafkaConsumer defines the interface for Kafka message consumption
type KafkaConsumer interface {
	ReadMessage(ctx context.Context, topic string) ([]byte, error)
	Close() error
}

// KafkaWriter is an interface that wraps kafka-go Writer for testing
type KafkaWriter interface {
	WriteMessages(ctx context.Context, msgs ...kafka.Message) error
	Close() error
}

// KafkaReader is an interface that wraps kafka-go Reader for testing
type KafkaReader interface {
	ReadMessage(ctx context.Context) (kafka.Message, error)
	Close() error
}
