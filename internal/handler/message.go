// Package handler provides HTTP handlers for the message service.
package handler

import (
	"errors"
	"net/http"
	"strconv"

	"httpchat/internal/interfaces"
	"httpchat/internal/logger"
	"httpchat/internal/repositoryerr"
	"httpchat/internal/validation"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// MessageHandler handles HTTP requests for messages
type MessageHandler struct {
	service   interfaces.MessageService
	logger    *logger.Logger
	validator validation.MessageValidator
}

// CreateMessageRequest represents the request body for creating a message
type CreateMessageRequest struct {
	Content string `json:"content" example:"Hello, world!"`
}

// CreateMessageResponse represents the response body for creating a message
type CreateMessageResponse struct {
	ID int64 `json:"id" example:"1"`
}

// ErrorResponse represents an error response
type ErrorResponse struct {
	Error string `json:"error" example:"Something went wrong"`
}

// httpError represents an HTTP error with status code
type httpError struct {
	statusCode int
	message    string
}

// NewMessageHandler creates a new MessageHandler instance
func NewMessageHandler(service interfaces.MessageService, logger *logger.Logger) *MessageHandler {
	return &MessageHandler{
		service:   service,
		logger:    logger,
		validator: *validation.NewMessageValidator(1000), // Max 1000 characters
	}
}

// handleServiceError converts service errors to appropriate HTTP responses
func (h *MessageHandler) handleServiceError(err error) *httpError {
	var repoErr *repositoryerr.RepositoryError
	if errors.As(err, &repoErr) {
		switch repoErr.ErrorCode() {
		case repositoryerr.ErrorCodeInvalidInput:
			h.logger.Warn("Invalid input", zap.Error(err))
			return &httpError{http.StatusBadRequest, "Invalid input"}
		case repositoryerr.ErrorCodeDuplicateEntry:
			h.logger.Warn("Duplicate entry", zap.Error(err))
			return &httpError{http.StatusBadRequest, "Duplicate entry"}
		case repositoryerr.ErrorCodeMessageNotFound:
			h.logger.Warn("Message not found", zap.Error(err))
			return &httpError{http.StatusNotFound, "Message not found"}
		case repositoryerr.ErrorCodeDatabaseConnection:
			h.logger.Error("Database connection error", zap.Error(err))
			return &httpError{http.StatusServiceUnavailable, "Service temporarily unavailable"}
		default:
			h.logger.Error("Service error", zap.Error(err))
			return &httpError{http.StatusInternalServerError, "Internal server error"}
		}
	}
	h.logger.Error("Unexpected service error", zap.Error(err))
	return &httpError{http.StatusInternalServerError, "Internal server error"}
}

// CreateMessageHandler creates a new message and sends it to Kafka
// @Summary Create a new message
// @Description Creates a new message and sends it to Kafka
// @Tags messages
// @Accept  json
// @Produce  json
// @Param content body handler.CreateMessageRequest true "Message content"
// @Success 200 {object} handler.CreateMessageResponse
// @Failure 400 {object} handler.ErrorResponse
// @Failure 500 {object} handler.ErrorResponse
// @Router /messages [post]
func (h *MessageHandler) CreateMessageHandler(c *gin.Context) {
	// Step 1: Parse the JSON request body
	var req CreateMessageRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Warn("Invalid JSON in create message request", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid JSON"})
		return
	}

	// Step 2: Validate the message content
	if err := h.validator.ValidateMessageContent(req.Content); err != nil {
		validationErr, ok := err.(*validation.Error)
		if ok {
			switch validationErr.Code {
			case validation.ValidationErrorCodeEmptyContent:
				h.logger.Warn("Empty message content")
				c.JSON(http.StatusBadRequest, gin.H{"error": "Message content cannot be empty"})
				return
			case validation.ValidationErrorCodeContentTooLong:
				h.logger.Warn("Message content too long", zap.Int("length", len(req.Content)))
				c.JSON(http.StatusBadRequest, gin.H{"error": "Message content too long (max 1000 characters)"})
				return
			case validation.ValidationErrorCodeInvalidCharacters:
				h.logger.Warn("Message content contains invalid characters", zap.String("content", req.Content))
				c.JSON(http.StatusBadRequest, gin.H{"error": "Message content contains invalid characters"})
				return
			}
		}
		h.logger.Warn("Validation error", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid message content"})
		return
	}

	h.logger.Info("Creating new message", zap.String("content", req.Content))

	// Step 3: Process the message through the service layer
	id, err := h.service.CreateMessage(c.Request.Context(), req.Content)
	if err != nil {
		httpErr := h.handleServiceError(err)
		c.JSON(httpErr.statusCode, gin.H{"error": httpErr.message})
		return
	}

	h.logger.Info("Successfully created message", zap.Int64("id", id))

	// Step 4: Return the new message ID to confirm successful creation
	response := CreateMessageResponse{ID: id}
	c.JSON(http.StatusOK, response)
}

// GetStatisticsHandler returns statistics on processed and unprocessed messages
// @Summary Get message statistics
// @Description Returns statistics on processed and unprocessed messages
// @Tags statistics
// @Produce  json
// @Success 200 {object} model.Statistics
// @Failure 500 {object} handler.ErrorResponse
// @Failure 503 {object} handler.ErrorResponse
// @Router /statistics [get]
func (h *MessageHandler) GetStatisticsHandler(c *gin.Context) {
	h.logger.Info("Fetching message statistics")

	// Get message statistics from the service layer
	stats, err := h.service.GetStatistics(c.Request.Context())
	if err != nil {
		httpErr := h.handleServiceError(err)
		c.JSON(httpErr.statusCode, gin.H{"error": httpErr.message})
		return
	}

	h.logger.Info("Successfully fetched statistics", 
		zap.Int64("total", stats.TotalMessages),
		zap.Int64("processed", stats.ProcessedMessages),
		zap.Int64("unprocessed", stats.UnprocessedMessages))

	// Return statistics as JSON response
	c.JSON(http.StatusOK, stats)
}

// ProcessMessageHandler marks a message as processed
// @Summary Process a message
// @Description Marks a message as processed
// @Tags messages
// @Param id path int true "Message ID"
// @Success 200
// @Failure 400 {object} handler.ErrorResponse
// @Failure 404 {object} handler.ErrorResponse
// @Failure 500 {object} handler.ErrorResponse
// @Router /messages/{id}/process [put]
func (h *MessageHandler) ProcessMessageHandler(c *gin.Context) {
	// Step 1: Extract and parse the message ID from URL parameters
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		h.logger.Warn("Invalid message ID format", zap.String("id", idStr), zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid message ID"})
		return
	}

	// Step 2: Validate the message ID
	if err := h.validator.ValidateMessageID(id); err != nil {
		validationErr, ok := err.(*validation.Error)
		if ok && validationErr.Code == validation.ValidationErrorCodeInvalidID {
			h.logger.Warn("Invalid message ID value", zap.Int64("id", id))
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid message ID"})
			return
		}
		h.logger.Warn("Validation error", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid message ID"})
		return
	}

	h.logger.Info("Processing message", zap.Int64("id", id))

	// Step 3: Mark the specified message as processed through the service layer
	if err := h.service.ProcessMessage(c.Request.Context(), id); err != nil {
		httpErr := h.handleServiceError(err)
		c.JSON(httpErr.statusCode, gin.H{"error": httpErr.message})
		return
	}

	h.logger.Info("Successfully processed message", zap.Int64("id", id))

	// Step 4: Return success response (200 OK)
	c.Status(http.StatusOK)
}
