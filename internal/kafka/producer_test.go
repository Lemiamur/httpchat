package kafka

import (
	"context"
	"errors"
	"testing"

	"github.com/segmentio/kafka-go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockKafkaWriter is a mock implementation of interfaces.KafkaWriter for testing
type MockKafkaWriter struct {
	mock.Mock
}

func (m *MockKafkaWriter) WriteMessages(ctx context.Context, msgs ...kafka.Message) error {
	args := m.Called(ctx, msgs)
	return args.Error(0)
}

func (m *MockKafkaWriter) Close() error {
	args := m.Called()
	return args.Error(0)
}

// TestProducerSendMessage tests the SendMessage method of ProducerImpl
func TestProducerSendMessage(t *testing.T) {
	// Create a mock Kafka writer
	mockWriter := new(MockKafkaWriter)

	// Create a producer with the mock writer
	producer := &ProducerImpl{
		writer: mockWriter,
	}

	// Set up expectations
	mockWriter.On("WriteMessages", mock.Anything, mock.AnythingOfType("[]kafka.Message")).Return(nil)

	// Test successful message sending
	err := producer.SendMessage(context.Background(), "test-topic", []byte("test message"))
	assert.NoError(t, err)

	// Verify expectations
	mockWriter.AssertExpectations(t)
}

// TestProducerSendMessageError tests the SendMessage method error handling
func TestProducerSendMessageError(t *testing.T) {
	// Create a mock Kafka writer
	mockWriter := new(MockKafkaWriter)

	// Create a producer with the mock writer
	producer := &ProducerImpl{
		writer: mockWriter,
	}

	// Set up expectations
	expectedError := errors.New("kafka write error")
	mockWriter.On("WriteMessages", mock.Anything, mock.AnythingOfType("[]kafka.Message")).Return(expectedError)

	// Test message sending with error
	err := producer.SendMessage(context.Background(), "test-topic", []byte("test message"))
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to write message to Kafka")

	// Verify expectations
	mockWriter.AssertExpectations(t)
}

// TestProducerClose tests the Close method of ProducerImpl
func TestProducerClose(t *testing.T) {
	// Create a mock Kafka writer
	mockWriter := new(MockKafkaWriter)

	// Create a producer with the mock writer
	producer := &ProducerImpl{
		writer: mockWriter,
	}

	// Set up expectations
	mockWriter.On("Close").Return(nil)

	// Test closing the producer
	err := producer.Close()
	assert.NoError(t, err)

	// Verify expectations
	mockWriter.AssertExpectations(t)
}

// TestProducerCloseError tests the Close method error handling
func TestProducerCloseError(t *testing.T) {
	// Create a mock Kafka writer
	mockWriter := new(MockKafkaWriter)

	// Create a producer with the mock writer
	producer := &ProducerImpl{
		writer: mockWriter,
	}

	// Set up expectations
	expectedError := errors.New("kafka close error")
	mockWriter.On("Close").Return(expectedError)

	// Test closing the producer with error
	err := producer.Close()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to close Kafka writer")

	// Verify expectations
	mockWriter.AssertExpectations(t)
}