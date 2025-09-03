// Package kafka provides Kafka producer and consumer implementations.
package kafka

import (
	"context"
	"fmt"

	"httpchat/internal/interfaces"

	"github.com/segmentio/kafka-go"
)

// ConsumerImpl implements the interfaces.KafkaConsumer interface for Kafka
type ConsumerImpl struct {
	reader interfaces.KafkaReader
	topic  string
}

// NewConsumer creates a new Kafka consumer
func NewConsumer(brokers []string, topic string, groupID string) interfaces.KafkaConsumer {
	// Create Kafka reader configuration
	config := kafka.ReaderConfig{
		Brokers: brokers,
		Topic:   topic,
		GroupID: groupID,
	}

	// Create Kafka reader
	reader := kafka.NewReader(config)

	// Return consumer implementation
	return &ConsumerImpl{
		reader: reader,
		topic:  topic,
	}
}

// Ensure ConsumerImpl implements interfaces.KafkaConsumer
var _ interfaces.KafkaConsumer = (*ConsumerImpl)(nil)

// ReadMessage reads a message from Kafka
func (c *ConsumerImpl) ReadMessage(ctx context.Context, _ string) ([]byte, error) {
	// Read a message from Kafka
	message, err := c.reader.ReadMessage(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to read message from Kafka: %w", err)
	}

	// Return just the message value (payload)
	return message.Value, nil
}

// Close closes the connection to Kafka
func (c *ConsumerImpl) Close() error {
	// Close the Kafka reader connection
	if err := c.reader.Close(); err != nil {
		return fmt.Errorf("failed to close Kafka reader: %w", err)
	}
	return nil
}
