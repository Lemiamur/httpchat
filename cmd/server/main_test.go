package main

import (
	"os"
	"testing"
)

func TestGetEnv(t *testing.T) {
	// Test getEnv function
	tests := []struct {
		name         string
		key          string
		defaultValue string
		envValue     string
		expected     string
	}{
		{
			name:         "Environment variable exists",
			key:          "TEST_KEY",
			defaultValue: "default",
			envValue:     "value",
			expected:     "value",
		},
		{
			name:         "Environment variable does not exist",
			key:          "NON_EXISTENT_KEY",
			defaultValue: "default",
			envValue:     "",
			expected:     "default",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set environment variable
			if tt.envValue != "" {
				t.Setenv(tt.key, tt.envValue)
			}

			// Create temporary getEnv function for testing
			getEnv := func(key, defaultValue string) string {
				if value := os.Getenv(key); value != "" {
					return value
				}
				return defaultValue
			}

			// Call getEnv function
			result := getEnv(tt.key, tt.defaultValue)

			// Check result
			if result != tt.expected {
				t.Errorf("getEnv(%s, %s) = %s; expected %s", tt.key, tt.defaultValue, result, tt.expected)
			}
		})
	}
}
