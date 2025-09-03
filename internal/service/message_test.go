package service

import (
	"context"
	"errors"
	"testing"

	"httpchat/internal/interfaces"
	"httpchat/internal/logger"
	"httpchat/internal/model"
)

// mockMessageRepository implements interfaces.MessageRepository for testing
type mockMessageRepository struct {
	createMessageFunc  func(ctx context.Context, content string) (*model.Message, error)
	getMessageByIDFunc func(ctx context.Context, id int64) (*model.Message, error)
	updateMessageStatusFunc func(ctx context.Context, id int64, processed bool) error
	getAllMessagesFunc func(ctx context.Context) ([]*model.Message, error)
	getStatisticsFunc  func(ctx context.Context) (*model.Statistics, error)
}

func (m *mockMessageRepository) CreateMessage(ctx context.Context, content string) (*model.Message, error) {
	if m.createMessageFunc != nil {
		return m.createMessageFunc(ctx, content)
	}
	return nil, nil
}

func (m *mockMessageRepository) GetMessageByID(ctx context.Context, id int64) (*model.Message, error) {
	if m.getMessageByIDFunc != nil {
		return m.getMessageByIDFunc(ctx, id)
	}
	return nil, nil
}

func (m *mockMessageRepository) UpdateMessageStatus(ctx context.Context, id int64, processed bool) error {
	if m.updateMessageStatusFunc != nil {
		return m.updateMessageStatusFunc(ctx, id, processed)
	}
	return nil
}

func (m *mockMessageRepository) GetAllMessages(ctx context.Context) ([]*model.Message, error) {
	if m.getAllMessagesFunc != nil {
		return m.getAllMessagesFunc(ctx)
	}
	return nil, nil
}

func (m *mockMessageRepository) GetStatistics(ctx context.Context) (*model.Statistics, error) {
	if m.getStatisticsFunc != nil {
		return m.getStatisticsFunc(ctx)
	}
	return nil, nil
}

// Ensure mockMessageRepository implements interfaces.MessageRepository
var _ interfaces.MessageRepository = (*mockMessageRepository)(nil)

type mockKafkaProducer struct {
	sendMessageFunc func(ctx context.Context, topic string, message []byte) error
	closeFunc       func() error
}

func (m *mockKafkaProducer) SendMessage(ctx context.Context, topic string, message []byte) error {
	if m.sendMessageFunc != nil {
		return m.sendMessageFunc(ctx, topic, message)
	}
	return nil
}

func (m *mockKafkaProducer) Close() error {
	if m.closeFunc != nil {
		return m.closeFunc()
	}
	return nil
}

// Ensure mockKafkaProducer implements interfaces.KafkaProducer
var _ interfaces.KafkaProducer = (*mockKafkaProducer)(nil)

type mockKafkaConsumer struct {
	readMessageFunc func(ctx context.Context, topic string) ([]byte, error)
	closeFunc       func() error
}

func (m *mockKafkaConsumer) ReadMessage(ctx context.Context, topic string) ([]byte, error) {
	if m.readMessageFunc != nil {
		return m.readMessageFunc(ctx, topic)
	}
	return []byte{}, nil
}

func (m *mockKafkaConsumer) Close() error {
	if m.closeFunc != nil {
		return m.closeFunc()
	}
	return nil
}

// Ensure mockKafkaConsumer implements interfaces.KafkaConsumer
var _ interfaces.KafkaConsumer = (*mockKafkaConsumer)(nil)

func TestCreateMessage(t *testing.T) {
	ctx := context.Background()
	
	// Create logger for testing
	testLogger, _ := logger.New()

	// Test successful message creation
	t.Run("Successful creation", func(t *testing.T) {
		repo := &mockMessageRepository{
			createMessageFunc: func(_ context.Context, _ string) (*model.Message, error) {
				return &model.Message{
					ID:        1,
					Content:   "Test message",
					Processed: false,
				}, nil
			},
		}
		
		producer := &mockKafkaProducer{
			sendMessageFunc: func(_ context.Context, _ string, _ []byte) error {
				return nil
			},
		}
		
		consumer := &mockKafkaConsumer{}
		
		service := NewMessageService(
			repo,
			producer,
			consumer,
			"test-topic",
			testLogger,
		)
		
		id, err := service.CreateMessage(ctx, "Test message")
		
		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}
		
		if id != 1 {
			t.Errorf("Expected ID 1, got %d", id)
		}
	})
	
	// Test repository error
	t.Run("Repository error", func(t *testing.T) {
		repo := &mockMessageRepository{
			createMessageFunc: func(_ context.Context, _ string) (*model.Message, error) {
				return nil, errors.New("database error")
			},
		}
		
		producer := &mockKafkaProducer{}
		consumer := &mockKafkaConsumer{}
		
		service := NewMessageService(
			repo,
			producer,
			consumer,
			"test-topic",
			testLogger,
		)
		
		_, err := service.CreateMessage(ctx, "Test message")
		
		if err == nil {
			t.Error("Expected error, got none")
		}
	})
	
	// Test Kafka error
	t.Run("Kafka error", func(t *testing.T) {
		repo := &mockMessageRepository{
			createMessageFunc: func(_ context.Context, _ string) (*model.Message, error) {
				return &model.Message{
					ID:        1,
					Content:   "Test message",
					Processed: false,
				}, nil
			},
		}
		
		producer := &mockKafkaProducer{
			sendMessageFunc: func(_ context.Context, _ string, _ []byte) error {
				return errors.New("kafka error")
			},
		}
		
		consumer := &mockKafkaConsumer{}
		
		service := NewMessageService(
			repo,
			producer,
			consumer,
			"test-topic",
			testLogger,
		)
		
		_, err := service.CreateMessage(ctx, "Test message")
		
		if err == nil {
			t.Error("Expected error, got none")
		}
	})
}

func TestProcessMessage(t *testing.T) {
	ctx := context.Background()
	
	// Create logger for testing
	testLogger, _ := logger.New()

	// Test successful message processing
	t.Run("Successful processing", func(t *testing.T) {
		repo := &mockMessageRepository{
			updateMessageStatusFunc: func(_ context.Context, _ int64, _ bool) error {
				return nil
			},
		}
		
		producer := &mockKafkaProducer{}
		consumer := &mockKafkaConsumer{}
		
		service := NewMessageService(
			repo,
			producer,
			consumer,
			"test-topic",
			testLogger,
		)
		
		err := service.ProcessMessage(ctx, 1)
		
		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}
	})
	
	// Test repository error
	t.Run("Repository error", func(t *testing.T) {
		repo := &mockMessageRepository{
			updateMessageStatusFunc: func(_ context.Context, _ int64, _ bool) error {
				return errors.New("database error")
			},
		}
		
		producer := &mockKafkaProducer{}
		consumer := &mockKafkaConsumer{}
		
		service := NewMessageService(
			repo,
			producer,
			consumer,
			"test-topic",
			testLogger,
		)
		
		err := service.ProcessMessage(ctx, 1)
		
		if err == nil {
			t.Error("Expected error, got none")
		}
	})
}

func TestGetStatistics(t *testing.T) {
	ctx := context.Background()
	
	// Create logger for testing
	testLogger, _ := logger.New()

	// Test successful statistics retrieval
	t.Run("Successful retrieval", func(t *testing.T) {
		stats := &model.Statistics{
			TotalMessages:       10,
			ProcessedMessages:   7,
			UnprocessedMessages: 3,
		}
		
		repo := &mockMessageRepository{
			getStatisticsFunc: func(_ context.Context) (*model.Statistics, error) {
				return stats, nil
			},
		}
		
		producer := &mockKafkaProducer{}
		consumer := &mockKafkaConsumer{}
		
		service := NewMessageService(
			repo,
			producer,
			consumer,
			"test-topic",
			testLogger,
		)
		
		result, err := service.GetStatistics(ctx)
		
		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}
		
		if result.TotalMessages != stats.TotalMessages {
			t.Errorf("Expected TotalMessages %d, got %d", stats.TotalMessages, result.TotalMessages)
		}
		
		if result.ProcessedMessages != stats.ProcessedMessages {
			t.Errorf("Expected ProcessedMessages %d, got %d", stats.ProcessedMessages, result.ProcessedMessages)
		}
		
		if result.UnprocessedMessages != stats.UnprocessedMessages {
			t.Errorf("Expected UnprocessedMessages %d, got %d", stats.UnprocessedMessages, result.UnprocessedMessages)
		}
	})
	
	// Test repository error
	t.Run("Repository error", func(t *testing.T) {
		repo := &mockMessageRepository{
			getStatisticsFunc: func(_ context.Context) (*model.Statistics, error) {
				return nil, errors.New("database error")
			},
		}
		
		producer := &mockKafkaProducer{}
		consumer := &mockKafkaConsumer{}
		
		service := NewMessageService(
			repo,
			producer,
			consumer,
			"test-topic",
			testLogger,
		)
		
		_, err := service.GetStatistics(ctx)
		
		if err == nil {
			t.Error("Expected error, got none")
		}
	})
}
