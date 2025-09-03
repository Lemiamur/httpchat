// Package kafka provides Kafka producer and consumer implementations.
package kafka

import (
	"context"
	"fmt"

	"httpchat/internal/interfaces"

	"github.com/segmentio/kafka-go"
)

// ProducerImpl implements the interfaces.KafkaProducer interface for Kafka
type ProducerImpl struct {
	writer interfaces.KafkaWriter
}

// NewProducer creates a new ProducerImpl instance
func NewProducer(brokers []string) interfaces.KafkaProducer {
	return &ProducerImpl{
		writer: &kafka.Writer{
			Addr: kafka.TCP(brokers...),
		},
	}
}

// Ensure ProducerImpl implements interfaces.KafkaProducer
var _ interfaces.KafkaProducer = (*ProducerImpl)(nil)

// SendMessage sends a message to Kafka
func (p *ProducerImpl) SendMessage(ctx context.Context, topic string, message []byte) error {
	// Send the message to the specified Kafka topic
	err := p.writer.WriteMessages(ctx,
		kafka.Message{
			Topic: topic,
			Value: message,
		},
	)

	if err != nil {
		return fmt.Errorf("failed to write message to Kafka: %w", err)
	}

	return nil
}

// Close closes the connection to Kafka
func (p *ProducerImpl) Close() error {
	// Close the Kafka writer connection
	if err := p.writer.Close(); err != nil {
		return fmt.Errorf("failed to close Kafka writer: %w", err)
	}
	return nil
}