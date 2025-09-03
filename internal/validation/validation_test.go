package validation

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMessageValidator_ValidateMessageContent(t *testing.T) {
	validator := NewMessageValidator(1000)

	// Test empty content
	t.Run("EmptyContent", func(t *testing.T) {
		err := validator.ValidateMessageContent("")
		assert.Error(t, err)
		validationErr, ok := err.(*Error)
		assert.True(t, ok)
		assert.Equal(t, ValidationErrorCodeEmptyContent, validationErr.Code)
	})

	// Test content too long
	t.Run("ContentTooLong", func(t *testing.T) {
		longContent := ""
		for i := 0; i < 1001; i++ {
			longContent += "a"
		}
		err := validator.ValidateMessageContent(longContent)
		assert.Error(t, err)
		validationErr, ok := err.(*Error)
		assert.True(t, ok)
		assert.Equal(t, ValidationErrorCodeContentTooLong, validationErr.Code)
	})

	// Test invalid characters
	t.Run("InvalidCharacters", func(t *testing.T) {
		err := validator.ValidateMessageContent("<script>alert('xss')</script>")
		assert.Error(t, err)
		validationErr, ok := err.(*Error)
		assert.True(t, ok)
		assert.Equal(t, ValidationErrorCodeInvalidCharacters, validationErr.Code)
	})

	// Test valid content
	t.Run("ValidContent", func(t *testing.T) {
		err := validator.ValidateMessageContent("Hello, world!")
		assert.NoError(t, err)
	})
}

func TestMessageValidator_ValidateMessageID(t *testing.T) {
	validator := NewMessageValidator(1000)

	// Test invalid ID (zero)
	t.Run("ZeroID", func(t *testing.T) {
		err := validator.ValidateMessageID(0)
		assert.Error(t, err)
		validationErr, ok := err.(*Error)
		assert.True(t, ok)
		assert.Equal(t, ValidationErrorCodeInvalidID, validationErr.Code)
	})

	// Test invalid ID (negative)
	t.Run("NegativeID", func(t *testing.T) {
		err := validator.ValidateMessageID(-1)
		assert.Error(t, err)
		validationErr, ok := err.(*Error)
		assert.True(t, ok)
		assert.Equal(t, ValidationErrorCodeInvalidID, validationErr.Code)
	})

	// Test valid ID
	t.Run("ValidID", func(t *testing.T) {
		err := validator.ValidateMessageID(1)
		assert.NoError(t, err)
	})
}
