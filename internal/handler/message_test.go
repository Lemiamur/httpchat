package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"httpchat/internal/interfaces"
	"httpchat/internal/logger"
	"httpchat/internal/model"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

// mockMessageService implements interfaces.MessageService for testing
type mockMessageService struct {
	createMessageFunc  func(ctx context.Context, content string) (int64, error)
	processMessageFunc func(ctx context.Context, id int64) error
	getStatisticsFunc  func(ctx context.Context) (*model.Statistics, error)
}

func (m *mockMessageService) CreateMessage(ctx context.Context, content string) (int64, error) {
	if m.createMessageFunc != nil {
		return m.createMessageFunc(ctx, content)
	}
	return 0, nil
}

func (m *mockMessageService) ProcessMessage(ctx context.Context, id int64) error {
	if m.processMessageFunc != nil {
		return m.processMessageFunc(ctx, id)
	}
	return nil
}

func (m *mockMessageService) GetStatistics(ctx context.Context) (*model.Statistics, error) {
	if m.getStatisticsFunc != nil {
		return m.getStatisticsFunc(ctx)
	}
	return &model.Statistics{}, nil
}

// Ensure mockMessageService implements interfaces.MessageService
var _ interfaces.MessageService = (*mockMessageService)(nil)

func setupTestRouter(handler *MessageHandler) *gin.Engine {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.POST("/messages", handler.CreateMessageHandler)
	router.GET("/statistics", handler.GetStatisticsHandler)
	router.PUT("/messages/:id/process", handler.ProcessMessageHandler)
	return router
}

func TestCreateMessageHandler(t *testing.T) {
	// Create a mock service
	mockService := &mockMessageService{
		createMessageFunc: func(_ context.Context, _ string) (int64, error) {
			return 1, nil
		},
	}

	// Create logger for testing
	testLogger, _ := logger.New()

	// Create handler with mock service
	handler := NewMessageHandler(mockService, testLogger)

	// Setup router
	router := setupTestRouter(handler)

	// Test successful message creation
	t.Run("SuccessfulCreation", func(t *testing.T) {
		requestBody := `{"content": "Test message"}`
		req, _ := http.NewRequest("POST", "/messages", bytes.NewBufferString(requestBody))
		req.Header.Set("Content-Type", "application/json")

		rr := httptest.NewRecorder()
		router.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code)

		var response CreateMessageResponse
		err := json.Unmarshal(rr.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, int64(1), response.ID)
	})

	// Test empty content
	t.Run("EmptyContent", func(t *testing.T) {
		requestBody := `{"content": ""}`
		req, _ := http.NewRequest("POST", "/messages", bytes.NewBufferString(requestBody))
		req.Header.Set("Content-Type", "application/json")

		rr := httptest.NewRecorder()
		router.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusBadRequest, rr.Code)

		var response ErrorResponse
		err := json.Unmarshal(rr.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, "Message content cannot be empty", response.Error)
	})

	// Test content too long
	t.Run("ContentTooLong", func(t *testing.T) {
		longContent := ""
		for i := 0; i < 1001; i++ {
			longContent += "a"
		}
		requestBody := `{"content": "` + longContent + `"}`
		req, _ := http.NewRequest("POST", "/messages", bytes.NewBufferString(requestBody))
		req.Header.Set("Content-Type", "application/json")

		rr := httptest.NewRecorder()
		router.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusBadRequest, rr.Code)

		var response ErrorResponse
		err := json.Unmarshal(rr.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, "Message content too long (max 1000 characters)", response.Error)
	})

	// Test invalid characters
	t.Run("InvalidCharacters", func(t *testing.T) {
		requestBody := `{"content": "<script>alert('xss')</script>"}`
		req, _ := http.NewRequest("POST", "/messages", bytes.NewBufferString(requestBody))
		req.Header.Set("Content-Type", "application/json")

		rr := httptest.NewRecorder()
		router.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusBadRequest, rr.Code)

		var response ErrorResponse
		err := json.Unmarshal(rr.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, "Message content contains invalid characters", response.Error)
	})

	// Test invalid JSON
	t.Run("InvalidJSON", func(t *testing.T) {
		requestBody := `{"content": "Test message"`
		req, _ := http.NewRequest("POST", "/messages", bytes.NewBufferString(requestBody))
		req.Header.Set("Content-Type", "application/json")

		rr := httptest.NewRecorder()
		router.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusBadRequest, rr.Code)

		var response ErrorResponse
		err := json.Unmarshal(rr.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, "Invalid JSON", response.Error)
	})
}

func TestGetStatisticsHandler(t *testing.T) {
	// Create a mock service
	stats := &model.Statistics{
		TotalMessages:       10,
		ProcessedMessages:   7,
		UnprocessedMessages: 3,
	}

	mockService := &mockMessageService{
		getStatisticsFunc: func(_ context.Context) (*model.Statistics, error) {
			return stats, nil
		},
	}

	// Create logger for testing
	testLogger, _ := logger.New()

	// Create handler with mock service
	handler := NewMessageHandler(mockService, testLogger)

	// Setup router
	router := setupTestRouter(handler)

	// Test successful retrieval
	t.Run("SuccessfulRetrieval", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/statistics", nil)

		rr := httptest.NewRecorder()
		router.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code)

		var response model.Statistics
		err := json.Unmarshal(rr.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, stats.TotalMessages, response.TotalMessages)
		assert.Equal(t, stats.ProcessedMessages, response.ProcessedMessages)
		assert.Equal(t, stats.UnprocessedMessages, response.UnprocessedMessages)
	})
}

func TestProcessMessageHandler(t *testing.T) {
	// Create a mock service
	mockService := &mockMessageService{
		processMessageFunc: func(_ context.Context, _ int64) error {
			return nil
		},
	}

	// Create logger for testing
	testLogger, _ := logger.New()

	// Create handler with mock service
	handler := NewMessageHandler(mockService, testLogger)

	// Setup router
	router := setupTestRouter(handler)

	// Test successful processing
	t.Run("SuccessfulProcessing", func(t *testing.T) {
		req, _ := http.NewRequest("PUT", "/messages/1/process", nil)

		rr := httptest.NewRecorder()
		router.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code)
	})

	// Test invalid ID format
	t.Run("InvalidIDFormat", func(t *testing.T) {
		req, _ := http.NewRequest("PUT", "/messages/abc/process", nil)

		rr := httptest.NewRecorder()
		router.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusBadRequest, rr.Code)

		var response ErrorResponse
		err := json.Unmarshal(rr.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, "Invalid message ID", response.Error)
	})

	// Test zero ID
	t.Run("ZeroID", func(t *testing.T) {
		req, _ := http.NewRequest("PUT", "/messages/0/process", nil)

		rr := httptest.NewRecorder()
		router.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusBadRequest, rr.Code)

		var response ErrorResponse
		err := json.Unmarshal(rr.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, "Invalid message ID", response.Error)
	})

	// Test negative ID
	t.Run("NegativeID", func(t *testing.T) {
		req, _ := http.NewRequest("PUT", "/messages/-1/process", nil)

		rr := httptest.NewRecorder()
		router.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusBadRequest, rr.Code)

		var response ErrorResponse
		err := json.Unmarshal(rr.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, "Invalid message ID", response.Error)
	})
}
