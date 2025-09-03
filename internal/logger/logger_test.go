package logger

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLoggerCreation(t *testing.T) {
	// Test production logger creation
	logger, err := New()
	assert.NoError(t, err)
	assert.NotNil(t, logger)

	// Close the logger
	err = logger.Close()
	assert.NoError(t, err)
}