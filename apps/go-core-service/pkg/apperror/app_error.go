package apperror

import (
	"fmt"
	"net/http"
)

// ErrorCode represents a machine-readable error code for client consumption
type ErrorCode string

const (
	CodeBadRequest   ErrorCode = "BAD_REQUEST"
	CodeUnauthorized ErrorCode = "UNAUTHORIZED"
	CodeForbidden    ErrorCode = "FORBIDDEN"
	CodeNotFound     ErrorCode = "NOT_FOUND"
	CodeConflict     ErrorCode = "CONFLICT"
	CodeValidation   ErrorCode = "VALIDATION_ERROR"
	CodeInternal     ErrorCode = "INTERNAL_ERROR"
)

// AppError is the standard error type used across all layers (repository → service → handler).
// It carries an HTTP status code, a machine-readable ErrorCode, a human-readable message,
// and optionally the original (internal) error for logging purposes.
type AppError struct {
	Code    ErrorCode `json:"code"`
	Message string    `json:"message"`
	Status  int       `json:"-"` // HTTP status code, not serialized to JSON
	Err     error     `json:"-"` // Original error for internal logging, not exposed to client
}

// Error implements the error interface.
func (e *AppError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("[%s] %s: %v", e.Code, e.Message, e.Err)
	}
	return fmt.Sprintf("[%s] %s", e.Code, e.Message)
}

// Unwrap supports errors.Is / errors.As unwrapping.
func (e *AppError) Unwrap() error {
	return e.Err
}

// HTTPStatus returns the HTTP status code associated with this error.
func (e *AppError) HTTPStatus() int {
	return e.Status
}

// --- Constructor functions ---

// NewBadRequest creates a 400 Bad Request error.
func NewBadRequest(message string) *AppError {
	return &AppError{
		Code:    CodeBadRequest,
		Message: message,
		Status:  http.StatusBadRequest,
	}
}

// NewValidation creates a 400 Validation error with a specific VALIDATION_ERROR code.
func NewValidation(message string) *AppError {
	return &AppError{
		Code:    CodeValidation,
		Message: message,
		Status:  http.StatusBadRequest,
	}
}

// NewUnauthorized creates a 401 Unauthorized error.
func NewUnauthorized(message string) *AppError {
	return &AppError{
		Code:    CodeUnauthorized,
		Message: message,
		Status:  http.StatusUnauthorized,
	}
}

// NewForbidden creates a 403 Forbidden error.
func NewForbidden(message string) *AppError {
	return &AppError{
		Code:    CodeForbidden,
		Message: message,
		Status:  http.StatusForbidden,
	}
}

// NewNotFound creates a 404 Not Found error.
// resourceName is the name of the resource that was not found (e.g. "task", "workspace").
func NewNotFound(resourceName string) *AppError {
	return &AppError{
		Code:    CodeNotFound,
		Message: fmt.Sprintf("%s not found", resourceName),
		Status:  http.StatusNotFound,
	}
}

// NewConflict creates a 409 Conflict error.
func NewConflict(message string) *AppError {
	return &AppError{
		Code:    CodeConflict,
		Message: message,
		Status:  http.StatusConflict,
	}
}

// NewInternal creates a 500 Internal Server Error.
// The message should be safe for client consumption (no SQL details).
func NewInternal(message string) *AppError {
	return &AppError{
		Code:    CodeInternal,
		Message: message,
		Status:  http.StatusInternalServerError,
	}
}

// Wrap preserves the original error (for logging) while returning an AppError to the caller.
func Wrap(err error, appErr *AppError) *AppError {
	appErr.Err = err
	return appErr
}
