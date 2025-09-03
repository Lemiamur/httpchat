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
	"httpchat/internal/logger"
	"httpchat/internal/model"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

// Mock implementations for integration testing
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

func setupTestRouter() *gin.Engine {
	// Create a mock service
	mockService := &mockMessageService{
		createMessageFunc: func(_ context.Context, _ string) (int64, error) {
			return 1, nil
		},
		processMessageFunc: func(_ context.Context, _ int64) error {
			return nil
		},
		getStatisticsFunc: func(_ context.Context) (*model.Statistics, error) {
			return &model.Statistics{
				TotalMessages:       10,
				ProcessedMessages:   7,
				UnprocessedMessages: 3,
			}, nil
		},
	}

	// Create logger for testing
	testLogger, _ := logger.New()

	// Create handler with mock service
	messageHandler := handler.NewMessageHandler(mockService, testLogger)

	// Setup routes with Gin
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.POST("/messages", messageHandler.CreateMessageHandler)
	router.GET("/statistics", messageHandler.GetStatisticsHandler)
	router.PUT("/messages/:id/process", messageHandler.ProcessMessageHandler)

	return router
}

func TestServerIntegration(t *testing.T) {
	router := setupTestRouter()

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
		assert.Equal(t, int64(10), response.TotalMessages)
		assert.Equal(t, int64(7), response.ProcessedMessages)
		assert.Equal(t, int64(3), response.UnprocessedMessages)
	})

	// Test processing a message
	t.Run("ProcessMessage", func(t *testing.T) {
		req, _ := http.NewRequest("PUT", "/messages/1/process", nil)

		rr := httptest.NewRecorder()
		router.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code)
	})
}

func TestServerStartup(t *testing.T) {
	// This test verifies that the server can start without errors
	// It starts the server and immediately shuts it down

	// Create a test server
	server := &http.Server{
		Addr:              ":0", // Use port 0 to get an available port
		Handler:           setupTestRouter(),
		ReadHeaderTimeout: 5 * time.Second, // Add timeout to prevent Slowloris attack
	}

	// Start server in a separate goroutine
	go func() {
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			t.Errorf("Server failed to start: %v", err)
		}
	}()

	// Give the server a moment to start
	time.Sleep(100 * time.Millisecond)

	// Shutdown server
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		t.Errorf("Server forced to shutdown: %v", err)
	}
}
