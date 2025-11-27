package errors

import (
	"fmt"
)

// PolykeysError represents a structured error with a code and optional details
type PolykeysError struct {
	Code    ErrorCode
	Message string
	Details map[string]interface{}
	Err     error // underlying error for wrapping
}

// Error implements the error interface
func (e *PolykeysError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("[%s] %s: %v", e.Code, e.Message, e.Err)
	}
	return fmt.Sprintf("[%s] %s", e.Code, e.Message)
}

// Unwrap returns the underlying error for errors.Is and errors.As support
func (e *PolykeysError) Unwrap() error {
	return e.Err
}

// New creates a new PolykeysError with the given code and message
func New(code ErrorCode, message string) *PolykeysError {
	return &PolykeysError{
		Code:    code,
		Message: message,
	}
}

// Wrap wraps an existing error with a Polykeys error code and message
func Wrap(code ErrorCode, message string, err error) *PolykeysError {
	return &PolykeysError{
		Code:    code,
		Message: message,
		Err:     err,
	}
}

// WithDetails adds contextual details to a PolykeysError
func WithDetails(err *PolykeysError, details map[string]interface{}) *PolykeysError {
	err.Details = details
	return err
}

// GetCode extracts the error code from a PolykeysError, or returns ErrCodeUnknown
func GetCode(err error) ErrorCode {
	if pkErr, ok := err.(*PolykeysError); ok {
		return pkErr.Code
	}
	return ErrCodeUnknown
}
