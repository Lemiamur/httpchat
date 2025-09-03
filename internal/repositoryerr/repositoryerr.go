// Package repositoryerr provides custom error types for the repository layer.
package repositoryerr

import (
	"errors"
	"fmt"
)

// Custom error types for repository operations
var (
	ErrMessageNotFound     = errors.New("message not found")
	ErrDatabaseConnection  = errors.New("database connection error")
	ErrDuplicateEntry      = errors.New("duplicate entry")
	ErrInvalidInput        = errors.New("invalid input")
	ErrTransactionFailed   = errors.New("transaction failed")
	ErrSerializationFailed = errors.New("serialization failed")
)

// Error codes for programmatic error handling
const (
	ErrorCodeMessageNotFound     = "MESSAGE_NOT_FOUND"
	ErrorCodeDatabaseConnection  = "DATABASE_CONNECTION_ERROR"
	ErrorCodeDuplicateEntry      = "DUPLICATE_ENTRY"
	ErrorCodeInvalidInput        = "INVALID_INPUT"
	ErrorCodeTransactionFailed   = "TRANSACTION_FAILED"
	ErrorCodeSerializationFailed = "SERIALIZATION_FAILED"
)

// RepositoryError wraps repository errors with additional context
type RepositoryError struct {
	Code string
	Op   string
	Err  error
}

func (e *RepositoryError) Error() string {
	if e.Code != "" {
		return fmt.Sprintf("repository error [%s] in %s: %v", e.Code, e.Op, e.Err)
	}
	return fmt.Sprintf("repository error in %s: %v", e.Op, e.Err)
}

func (e *RepositoryError) Unwrap() error {
	return e.Err
}

// Is checks if the error is of the specified type
func (e *RepositoryError) Is(target error) bool {
	t, ok := target.(*RepositoryError)
	if !ok {
		return false
	}
	// If target has a code, check for code match
	if t.Code != "" {
		return e.Code == t.Code
	}
	// Otherwise check for error match
	return errors.Is(e.Err, t.Err)
}

// ErrorCode returns the error code
func (e *RepositoryError) ErrorCode() string {
	return e.Code
}

// New creates a new RepositoryError
func New(code, op string, err error) *RepositoryError {
	return &RepositoryError{
		Code: code,
		Op:   op,
		Err:  err,
	}
}