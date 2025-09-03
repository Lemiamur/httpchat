// Package config provides application configuration functionality.
package config

import "github.com/kelseyhightower/envconfig"

// Config contains the application configuration
type Config struct {
	ServerPort        string `envconfig:"SERVER_PORT" default:"8080"`
	DatabaseURL       string `envconfig:"DATABASE_URL" default:"postgres://user:password@localhost:5432/messages_db?sslmode=disable"`
	KafkaBrokers      string `envconfig:"KAFKA_BROKERS" default:"localhost:9092"`
	KafkaTopic        string `envconfig:"KAFKA_TOPIC" default:"messages"`
	KafkaMaxRetries   int    `envconfig:"KAFKA_MAX_RETRIES" default:"3"`
	KafkaRetryDelayMs int    `envconfig:"KAFKA_RETRY_DELAY_MS" default:"5000"`
}

// Load loads configuration from environment variables
func Load() (*Config, error) {
	var cfg Config
	if err := envconfig.Process("", &cfg); err != nil {
		return nil, err
	}
	return &cfg, nil
}
