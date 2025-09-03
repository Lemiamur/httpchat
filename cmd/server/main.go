// Package main provides the main entry point for the HTTP Chat Service.
package main

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"httpchat/internal/config"
	"httpchat/internal/handler"
	"httpchat/internal/interfaces"
	"httpchat/internal/kafka"
	"httpchat/internal/logger"
	"httpchat/internal/model"
	"httpchat/internal/repository"
	"httpchat/internal/repositoryerr"
	"httpchat/internal/service"

	_ "httpchat/docs/swagger"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"go.uber.org/zap"
)

// @title HTTP Chat Service API
// @version 1.0
// @description Микросервис для обработки сообщений через HTTP API с сохранением в PostgreSQL и отправкой в Kafka.
// @termsOfService http://swagger.io/terms/

// @contact.name API Support
// @contact.url http://www.swagger.io/support
// @contact.email support@swagger.io

// @license.name Apache 2.0
// @license.url http://www.apache.org/licenses/LICENSE-2.0.html

// @host localhost:8080
// @BasePath /
func main() {
	// Initialize logger - this will be used throughout the application for structured logging
	appLogger, err := logger.New()
	if err != nil {
		// If we can't initialize our structured logger, panic since we can't proceed safely
		panic("Failed to initialize logger: " + err.Error())
	}
	// Ensure all log entries are flushed when the application exits
	defer func() {
		_ = appLogger.Close()
	}()

	// Load environment variables from .env file (useful for local development)
	if err := godotenv.Load("configs/.env.local"); err != nil {
		// It's not critical if the .env file is missing, just log a warning
		appLogger.Warn("Warning: Error loading .env file", zap.Error(err))
	}

	// Load configuration from environment variables
	cfg, err := config.Load()
	if err != nil {
		appLogger.Fatal("Failed to load configuration", zap.Error(err))
	}

	// Parse Kafka brokers from configuration (comma-separated list)
	kafkaBrokers := strings.Split(cfg.KafkaBrokers, ",")

	// Initialize dependencies for our application

	// Initialize PostgreSQL repository for storing messages
	var repo interfaces.MessageRepository
	repo, err = repository.NewPostgreSQLMessageRepository(cfg.DatabaseURL)
	if err != nil {
		// Check for specific repository errors to provide better error messages
		var repoErr *repositoryerr.RepositoryError
		if errors.As(err, &repoErr) {
			switch repoErr.ErrorCode() {
			case repositoryerr.ErrorCodeDatabaseConnection:
				appLogger.Fatal("Failed to connect to PostgreSQL database", zap.Error(err))
			default:
				appLogger.Fatal("Failed to initialize PostgreSQL repository", zap.Error(err))
			}
		}
		appLogger.Fatal("Failed to initialize PostgreSQL repository", zap.Error(err))
	}

	// Initialize Kafka producer for sending messages
	producer := kafka.NewProducer(kafkaBrokers)

	// Initialize Kafka consumer for reading messages
	consumer := kafka.NewConsumer(kafkaBrokers, cfg.KafkaTopic, "message-processor-group")

	// Create service that implements our business logic
	messageService := service.NewMessageService(repo, producer, consumer, cfg.KafkaTopic, appLogger)

	// Create handlers that connect HTTP requests to our service
	messageHandler := handler.NewMessageHandler(messageService, appLogger)

	// Setup HTTP routes using Gin framework
	router := gin.Default()
	router.POST("/messages", messageHandler.CreateMessageHandler)
	router.GET("/statistics", messageHandler.GetStatisticsHandler)
	router.PUT("/messages/:id/process", messageHandler.ProcessMessageHandler)

	// Swagger endpoint for API documentation
	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// Create HTTP server with security timeouts
	server := &http.Server{
		Addr:              ":" + cfg.ServerPort,
		Handler:           router,
		ReadHeaderTimeout: 5 * time.Second, // Prevent Slowloris attacks
	}

	// Start HTTP server in a separate goroutine so we can handle shutdown signals
	go func() {
		appLogger.Info("Server starting", zap.String("port", cfg.ServerPort))
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			appLogger.Fatal("Server failed to start", zap.Error(err))
		}
	}()

	// Create a context for graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Start background message processing from Kafka in a separate goroutine
	go func() {
		appLogger.Info("Starting Kafka message processor")
		processKafkaMessages(ctx, messageService, consumer, cfg.KafkaTopic, cfg.KafkaMaxRetries, time.Duration(cfg.KafkaRetryDelayMs)*time.Millisecond, appLogger)
	}()

	// Wait for shutdown signal (Ctrl+C or SIGTERM)
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	appLogger.Info("Shutting down server...")

	// Close connections to external services
	if err := consumer.Close(); err != nil {
		appLogger.Error("Error closing Kafka consumer", zap.Error(err))
	}

	if err := producer.Close(); err != nil {
		appLogger.Error("Error closing Kafka producer", zap.Error(err))
	}

	// Shutdown HTTP server with timeout to allow ongoing requests to complete
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		appLogger.Fatal("Server forced to shutdown", zap.Error(err))
	}

	appLogger.Info("Server exited")
}

// processKafkaMessages processes messages from Kafka
func processKafkaMessages(ctx context.Context, service interfaces.MessageService, consumer interfaces.KafkaConsumer, topic string, maxRetries int, retryDelay time.Duration, appLogger *logger.Logger) {
	for {
		select {
		case <-ctx.Done():
			// Context was cancelled, time to shut down
			appLogger.Info("Kafka message processor shutting down")
			return
		default:
			// Read message from Kafka
			messageBytes, err := consumer.ReadMessage(ctx, topic)
			if err != nil {
				// Check if context was cancelled
				if ctx.Err() != nil {
					appLogger.Info("Kafka message processor shutting down")
					return
				}
				appLogger.Error("Error reading message from Kafka", zap.Error(err))
				continue
			}

			// Decode message from JSON
			var message model.Message
			if err := json.Unmarshal(messageBytes, &message); err != nil {
				appLogger.Error("Error unmarshaling message", zap.Error(err))
				continue
			}

			appLogger.Info("Processing Kafka message", zap.Int64("id", message.ID))

			// Process message with retry mechanism
			shouldSkipMessage := false
			var processErr error
			for attempt := 0; attempt <= maxRetries; attempt++ {
				processErr = service.ProcessMessage(ctx, message.ID)
				if processErr == nil {
					// Success - message processed
					break
				}

				// Check if it's a specific error type that shouldn't be retried
				var repoErr *repositoryerr.RepositoryError
				if errors.As(processErr, &repoErr) {
					switch repoErr.ErrorCode() {
					case repositoryerr.ErrorCodeMessageNotFound:
						// Don't retry if message not found
						appLogger.Warn("Message not found for processing", zap.Int64("id", message.ID))
						shouldSkipMessage = true
						processErr = nil // Clear error to avoid logging failure
					case repositoryerr.ErrorCodeInvalidInput:
						// Don't retry if invalid input
						appLogger.Warn("Invalid message ID format", zap.Int64("id", message.ID), zap.Error(processErr))
						shouldSkipMessage = true
						processErr = nil // Clear error to avoid logging failure
					}
				}

				// If we should skip the message, break out of retry loop
				if shouldSkipMessage {
					break
				}

				// Retry for other errors (including database connection errors)
				if attempt < maxRetries {
					appLogger.Warn("Error processing message, retrying",
						zap.Int64("id", message.ID),
						zap.Int("attempt", attempt+1),
						zap.Error(processErr))
					time.Sleep(retryDelay)
				}
			}

			// If we should skip the message, continue to the next message
			if shouldSkipMessage {
				appLogger.Info("Successfully processed Kafka message", zap.Int64("id", message.ID))
				continue
			}

			if processErr != nil {
				appLogger.Error("Failed to process message after retries",
					zap.Int64("id", message.ID),
					zap.Error(processErr))
				continue
			}

			appLogger.Info("Successfully processed Kafka message", zap.Int64("id", message.ID))
		}
	}
}
