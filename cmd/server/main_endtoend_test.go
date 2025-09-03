package main

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"httpchat/internal/handler"
	"httpchat/internal/interfaces"
	"httpchat/internal/logger"
	"httpchat/internal/model"
	"httpchat/internal/repositoryerr"
	"httpchat/internal/service"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

// Mock implementations for end-to-end testing
type mockMessageRepository struct {
	messages map[int64]*model.Message
	nextID   int64
}

func newMockMessageRepository() *mockMessageRepository {
	return &mockMessageRepository{
		messages: make(map[int64]*model.Message),
		nextID:   1,
	}
}

func (m *mockMessageRepository) CreateMessage(_ context.Context, content string) (*model.Message, error) {
	id := m.nextID
	m.nextID++
	now := time.Now()
	message := &model.Message{
		ID:        id,
		Content:   content,
		Processed: false,
		CreatedAt: now,
		UpdatedAt: now,
	}
	m.messages[id] = message
	return message, nil
}

func (m *mockMessageRepository) GetMessageByID(_ context.Context, id int64) (*model.Message, error) {
	message, exists := m.messages[id]
	if !exists {
		return nil, &repositoryerr.RepositoryError{
			Op:  "GetMessageByID",
			Err: repositoryerr.ErrMessageNotFound,
		}
	}
	return message, nil
}

func (m *mockMessageRepository) UpdateMessageStatus(_ context.Context, id int64, processed bool) error {
	message, exists := m.messages[id]
	if !exists {
		return &repositoryerr.RepositoryError{
			Op:  "UpdateMessageStatus",
			Err: repositoryerr.ErrMessageNotFound,
		}
	}
	message.Processed = processed
	message.UpdatedAt = time.Now()
	return nil
}

func (m *mockMessageRepository) GetAllMessages(_ context.Context) ([]*model.Message, error) {
	messages := make([]*model.Message, 0, len(m.messages))
	for _, message := range m.messages {
		messages = append(messages, message)
	}
	return messages, nil
}

func (m *mockMessageRepository) GetStatistics(_ context.Context) (*model.Statistics, error) {
	total := int64(len(m.messages))
	processed := int64(0)
	unprocessed := int64(0)

	for _, message := range m.messages {
		if message.Processed {
			processed++
		} else {
			unprocessed++
		}
	}

	return &model.Statistics{
		TotalMessages:       total,
		ProcessedMessages:   processed,
		UnprocessedMessages: unprocessed,
	}, nil
}

	// Ensure mockMessageRepository implements interfaces.MessageRepository
	var _ interfaces.MessageRepository = (*mockMessageRepository)(nil)

type mockKafkaProducer struct {
	messages [][]byte
}

func newMockKafkaProducer() *mockKafkaProducer {
	return &mockKafkaProducer{
		messages: make([][]byte, 0),
	}
}

func (m *mockKafkaProducer) SendMessage(_ context.Context, _ string, message []byte) error {
	m.messages = append(m.messages, message)
	return nil
}

func (m *mockKafkaProducer) Close() error {
	return nil
}

// Ensure mockKafkaProducer implements interfaces.KafkaProducer
var _ interfaces.KafkaProducer = (*mockKafkaProducer)(nil)

type mockKafkaConsumer struct {
	messages [][]byte
	index    int
}

func newMockKafkaConsumer(messages [][]byte) *mockKafkaConsumer {
	return &mockKafkaConsumer{
		messages: messages,
		index:    0,
	}
}

func (m *mockKafkaConsumer) ReadMessage(_ context.Context, _ string) ([]byte, error) {
	if m.index >= len(m.messages) {
		// Simulate no more messages
		time.Sleep(100 * time.Millisecond)
		return nil, context.DeadlineExceeded
	}
	message := m.messages[m.index]
	m.index++
	return message, nil
}

func (m *mockKafkaConsumer) Close() error {
	return nil
}

// Ensure mockKafkaConsumer implements interfaces.KafkaConsumer
var _ interfaces.KafkaConsumer = (*mockKafkaConsumer)(nil)

func setupEndToEndTestRouter() (*gin.Engine, *mockMessageRepository, *mockKafkaProducer, *mockKafkaConsumer) {
	// Create mock components
	mockRepo := newMockMessageRepository()
	mockProducer := newMockKafkaProducer()
	mockConsumer := newMockKafkaConsumer(nil)

	// Create logger for testing
	testLogger, _ := logger.New()

	// Create service
	service := service.NewMessageService(mockRepo, mockProducer, mockConsumer, "test-topic", testLogger)

	// Create handler
	messageHandler := handler.NewMessageHandler(service, testLogger)

	// Setup routes with Gin
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.POST("/messages", messageHandler.CreateMessageHandler)
	router.GET("/statistics", messageHandler.GetStatisticsHandler)
	router.PUT("/messages/:id/process", messageHandler.ProcessMessageHandler)

	return router, mockRepo, mockProducer, mockConsumer
}

func TestEndToEndAPIScenario(t *testing.T) {
	router, mockRepo, mockProducer, _ := setupEndToEndTestRouter()

	// Test creating a message
	t.Run("CreateMessage", func(t *testing.T) {
		requestBody := `{"content": "Test message"}`
		req, _ := http.NewRequest("POST", "/messages", bytes.NewBufferString(requestBody))
		req.Header.Set("Content-Type", "application/json")

		rr := httptest.NewRecorder()
		router.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code)

		var response handler.CreateMessageResponse
		err := json.Unmarshal(rr.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, int64(1), response.ID)

		// Verify message was created in repository
		message, err := mockRepo.GetMessageByID(context.Background(), 1)
		assert.NoError(t, err)
		assert.Equal(t, "Test message", message.Content)
		assert.False(t, message.Processed)

		// Verify message was sent to Kafka
		assert.Equal(t, 1, len(mockProducer.messages))
	})

	// Test getting statistics
	t.Run("GetStatistics", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/statistics", nil)

		rr := httptest.NewRecorder()
		router.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code)

		var response model.Statistics
		err := json.Unmarshal(rr.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, int64(1), response.TotalMessages)
		assert.Equal(t, int64(0), response.ProcessedMessages)
		assert.Equal(t, int64(1), response.UnprocessedMessages)
	})

	// Test processing a message
	t.Run("ProcessMessage", func(t *testing.T) {
		req, _ := http.NewRequest("PUT", "/messages/1/process", nil)

		rr := httptest.NewRecorder()
		router.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code)

		// Verify message was processed in repository
		message, err := mockRepo.GetMessageByID(context.Background(), 1)
		assert.NoError(t, err)
		assert.True(t, message.Processed)
	})

	// Test getting updated statistics
	t.Run("GetUpdatedStatistics", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/statistics", nil)

		rr := httptest.NewRecorder()
		router.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code)

		var response model.Statistics
		err := json.Unmarshal(rr.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, int64(1), response.TotalMessages)
		assert.Equal(t, int64(1), response.ProcessedMessages)
		assert.Equal(t, int64(0), response.UnprocessedMessages)
	})
}

func TestEndToEndKafkaProcessingScenario(t *testing.T) {
	// Create mock components
	mockRepo := newMockMessageRepository()
	mockProducer := newMockKafkaProducer()

	// Create a message in the repository first
	createdMessage, err := mockRepo.CreateMessage(context.Background(), "Test message")
	assert.NoError(t, err)
	assert.Equal(t, int64(1), createdMessage.ID)

	// Create the same message to be processed by Kafka
	message := &model.Message{
		ID:        1,
		Content:   "Test message",
		Processed: false,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	messageBytes, _ := json.Marshal(message)

	// Create mock consumer with the message
	mockConsumer := newMockKafkaConsumer([][]byte{messageBytes})

	// Create logger for testing
	testLogger, _ := logger.New()

	// Create service (using the same repository)
	service := service.NewMessageService(mockRepo, mockProducer, mockConsumer, "test-topic", testLogger)

	// Create handler
	messageHandler := handler.NewMessageHandler(service, testLogger)

	// Setup routes with Gin
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.POST("/messages", messageHandler.CreateMessageHandler)
	router.GET("/statistics", messageHandler.GetStatisticsHandler)
	router.PUT("/messages/:id/process", messageHandler.ProcessMessageHandler)

	// Simulate Kafka processing in a separate goroutine
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	go func() {
		// Process messages from Kafka
		for {
			select {
			case <-ctx.Done():
				return
			default:
				messageBytes, err := mockConsumer.ReadMessage(ctx, "test-topic")
				if err != nil {
					if ctx.Err() != nil {
						return
					}
					continue
				}

				// Decode message
				var message model.Message
				if err := json.Unmarshal(messageBytes, &message); err != nil {
					continue
				}

				// Process message (mark as processed)
				if err := service.ProcessMessage(ctx, message.ID); err != nil {
					continue
				}
			}
		}
	}()

	// Give Kafka processing some time
	time.Sleep(100 * time.Millisecond)

	// Test getting statistics after Kafka processing
	t.Run("GetStatisticsAfterKafkaProcessing", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/statistics", nil)

		rr := httptest.NewRecorder()
		router.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code)

		var response model.Statistics
		err := json.Unmarshal(rr.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, int64(1), response.TotalMessages)
		assert.Equal(t, int64(1), response.ProcessedMessages)
		assert.Equal(t, int64(0), response.UnprocessedMessages)
	})
}