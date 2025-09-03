package kafka

import (
	"context"
	"errors"
	"testing"

	"github.com/segmentio/kafka-go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockKafkaReader is a mock implementation of interfaces.KafkaReader for testing
type MockKafkaReader struct {
	mock.Mock
}

func (m *MockKafkaReader) ReadMessage(ctx context.Context) (kafka.Message, error) {
	args := m.Called(ctx)
	return args.Get(0).(kafka.Message), args.Error(1)
}

func (m *MockKafkaReader) Close() error {
	args := m.Called()
	return args.Error(0)
}

// TestConsumerReadMessage tests the ReadMessage method of ConsumerImpl
func TestConsumerReadMessage(t *testing.T) {
	// Create a mock Kafka reader
	mockReader := new(MockKafkaReader)

	// Create a consumer with the mock reader
	consumer := &ConsumerImpl{
		reader: mockReader,
	}

	// Set up expectations
	expectedMessage := kafka.Message{
		Value: []byte("test message"),
	}
	mockReader.On("ReadMessage", mock.Anything).Return(expectedMessage, nil)

	// Test reading a message
	messageBytes, err := consumer.ReadMessage(context.Background(), "test-topic")
	assert.NoError(t, err)
	assert.Equal(t, []byte("test message"), messageBytes)

	// Verify expectations
	mockReader.AssertExpectations(t)
}

// TestConsumerReadMessageError tests the ReadMessage method error handling
func TestConsumerReadMessageError(t *testing.T) {
	// Create a mock Kafka reader
	mockReader := new(MockKafkaReader)

	// Create a consumer with the mock reader
	consumer := &ConsumerImpl{
		reader: mockReader,
	}

	// Set up expectations
	expectedError := errors.New("kafka read error")
	mockReader.On("ReadMessage", mock.Anything).Return(kafka.Message{}, expectedError)

	// Test reading a message with error
	_, err := consumer.ReadMessage(context.Background(), "test-topic")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to read message from Kafka")

	// Verify expectations
	mockReader.AssertExpectations(t)
}

// TestConsumerClose tests the Close method of ConsumerImpl
func TestConsumerClose(t *testing.T) {
	// Create a mock Kafka reader
	mockReader := new(MockKafkaReader)

	// Create a consumer with the mock reader
	consumer := &ConsumerImpl{
		reader: mockReader,
	}

	// Set up expectations
	mockReader.On("Close").Return(nil)

	// Test closing the consumer
	err := consumer.Close()
	assert.NoError(t, err)

	// Verify expectations
	mockReader.AssertExpectations(t)
}

// TestConsumerCloseError tests the Close method error handling
func TestConsumerCloseError(t *testing.T) {
	// Create a mock Kafka reader
	mockReader := new(MockKafkaReader)

	// Create a consumer with the mock reader
	consumer := &ConsumerImpl{
		reader: mockReader,
	}

	// Set up expectations
	expectedError := errors.New("kafka close error")
	mockReader.On("Close").Return(expectedError)

	// Test closing the consumer with error
	err := consumer.Close()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to close Kafka reader")

	// Verify expectations
	mockReader.AssertExpectations(t)
}