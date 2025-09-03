// Package validation provides input validation functionality.
package validation

import (
	"regexp"
	"unicode/utf8"
)

// MessageValidator validates message-related inputs
type MessageValidator struct {
	maxContentLength int
}

// NewMessageValidator creates a new MessageValidator instance
func NewMessageValidator(maxContentLength int) *MessageValidator {
	return &MessageValidator{
		maxContentLength: maxContentLength,
	}
}

// ValidateMessageContent validates message content
func (v *MessageValidator) ValidateMessageContent(content string) error {
	// Check if content is empty
	if content == "" {
		return &Error{
			Code:    ValidationErrorCodeEmptyContent,
			Message: "message content cannot be empty",
		}
	}

	// Check content length doesn't exceed maximum
	if utf8.RuneCountInString(content) > v.maxContentLength {
		return &Error{
			Code:    ValidationErrorCodeContentTooLong,
			Message: "message content too long",
		}
	}

	// Check for prohibited characters (HTML tags)
	if match, _ := regexp.MatchString(`[<>]`, content); match {
		return &Error{
			Code:    ValidationErrorCodeInvalidCharacters,
			Message: "message content contains invalid characters",
		}
	}

	return nil
}

// ValidateMessageID validates message ID
func (v *MessageValidator) ValidateMessageID(id int64) error {
	// Check if ID is positive (valid IDs start from 1)
	if id <= 0 {
		return &Error{
			Code:    ValidationErrorCodeInvalidID,
			Message: "message ID must be positive",
		}
	}

	return nil
}

// Error represents a validation error
type Error struct {
	Code    string
	Message string
}

func (e *Error) Error() string {
	return e.Message
}

// Validation error codes
const (
	ValidationErrorCodeEmptyContent      = "EMPTY_CONTENT"
	ValidationErrorCodeContentTooLong    = "CONTENT_TOO_LONG"
	ValidationErrorCodeInvalidCharacters = "INVALID_CHARACTERS"
	ValidationErrorCodeInvalidID         = "INVALID_ID"
)