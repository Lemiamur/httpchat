package repositoryerr

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRepositoryError(t *testing.T) {
	// Test basic error creation
	err := &RepositoryError{
		Code: ErrorCodeMessageNotFound,
		Op:   "GetMessageByID",
		Err:  ErrMessageNotFound,
	}

	assert.Equal(t, "repository error [MESSAGE_NOT_FOUND] in GetMessageByID: message not found", err.Error())
	assert.Equal(t, ErrorCodeMessageNotFound, err.ErrorCode())
	assert.Equal(t, ErrMessageNotFound, err.Unwrap())

	// Test error without code
	errWithoutCode := &RepositoryError{
		Op:  "CreateMessage",
		Err: errors.New("database error"),
	}

	assert.Equal(t, "repository error in CreateMessage: database error", errWithoutCode.Error())
	assert.Equal(t, "", errWithoutCode.ErrorCode())
}

func TestRepositoryErrorIs(t *testing.T) {
	// Test error matching with code
	errWithCode := &RepositoryError{
		Code: ErrorCodeMessageNotFound,
		Op:   "GetMessageByID",
		Err:  ErrMessageNotFound,
	}

	targetWithSameCode := &RepositoryError{
		Code: ErrorCodeMessageNotFound,
		Op:   "UpdateMessageStatus",
		Err:  errors.New("different error"),
	}

	assert.True(t, errWithCode.Is(targetWithSameCode))

	// Test error matching without code
	errWithoutCode := &RepositoryError{
		Op:  "CreateMessage",
		Err: ErrMessageNotFound,
	}

	targetWithoutCode := &RepositoryError{
		Op:  "CreateMessage",
		Err: ErrMessageNotFound,
	}

	assert.True(t, errWithoutCode.Is(targetWithoutCode))

	// Test non-matching errors
	differentErr := &RepositoryError{
		Code: ErrorCodeDatabaseConnection,
		Op:   "GetMessageByID",
		Err:  ErrDatabaseConnection,
	}

	assert.False(t, errWithCode.Is(differentErr))
}

func TestNewRepositoryError(t *testing.T) {
	err := New(ErrorCodeDuplicateEntry, "CreateMessage", errors.New("duplicate key value violates unique constraint"))
	
	assert.Equal(t, ErrorCodeDuplicateEntry, err.Code)
	assert.Equal(t, "CreateMessage", err.Op)
	assert.Equal(t, "duplicate key value violates unique constraint", err.Err.Error())
}