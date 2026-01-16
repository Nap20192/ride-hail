package errors

import (
	"fmt"
	"net/http"
)

// AppError is the base error type for application errors
type AppError struct {
	Type       ErrorType
	Message    string
	StatusCode int
	Err        error // underlying error
}

// ErrorType represents the category of error
type ErrorType string

const (
	ErrorTypeInvalidInput       ErrorType = "INVALID_INPUT"
	ErrorTypeNotFound           ErrorType = "NOT_FOUND"
	ErrorTypeUnauthorized       ErrorType = "UNAUTHORIZED"
	ErrorTypeForbidden          ErrorType = "FORBIDDEN"
	ErrorTypeConflict           ErrorType = "CONFLICT"
	ErrorTypeInternalError      ErrorType = "INTERNAL_ERROR"
	ErrorTypeServiceUnavailable ErrorType = "SERVICE_UNAVAILABLE"
)

// Error implements the error interface
func (e *AppError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("%s: %s: %v", e.Type, e.Message, e.Err)
	}
	return fmt.Sprintf("%s: %s", e.Type, e.Message)
}

// Unwrap returns the underlying error
func (e *AppError) Unwrap() error {
	return e.Err
}

// NewInvalidInputError creates an error for invalid user input
func NewInvalidInputError(message string) *AppError {
	return &AppError{
		Type:       ErrorTypeInvalidInput,
		Message:    message,
		StatusCode: http.StatusBadRequest,
	}
}

// NewNotFoundError creates an error for resources that don't exist
func NewNotFoundError(resource string) *AppError {
	return &AppError{
		Type:       ErrorTypeNotFound,
		Message:    fmt.Sprintf("%s not found", resource),
		StatusCode: http.StatusNotFound,
	}
}

// NewUnauthorizedError creates an error for authentication failures
func NewUnauthorizedError(message string) *AppError {
	if message == "" {
		message = "authentication required"
	}
	return &AppError{
		Type:       ErrorTypeUnauthorized,
		Message:    message,
		StatusCode: http.StatusUnauthorized,
	}
}

// NewForbiddenError creates an error for authorization failures
func NewForbiddenError(message string) *AppError {
	if message == "" {
		message = "access denied"
	}
	return &AppError{
		Type:       ErrorTypeForbidden,
		Message:    message,
		StatusCode: http.StatusForbidden,
	}
}

// NewConflictError creates an error for resource conflicts
func NewConflictError(message string) *AppError {
	return &AppError{
		Type:       ErrorTypeConflict,
		Message:    message,
		StatusCode: http.StatusConflict,
	}
}

// NewInternalError creates an error for internal server errors
func NewInternalError(message string, err error) *AppError {
	return &AppError{
		Type:       ErrorTypeInternalError,
		Message:    message,
		StatusCode: http.StatusInternalServerError,
		Err:        err,
	}
}

// NewServiceUnavailableError creates an error for service unavailability
func NewServiceUnavailableError(service string) *AppError {
	return &AppError{
		Type:       ErrorTypeServiceUnavailable,
		Message:    fmt.Sprintf("%s service unavailable", service),
		StatusCode: http.StatusServiceUnavailable,
	}
}

// ErrorResponse is the JSON structure for error responses
type ErrorResponse struct {
	Type      string `json:"type"`
	Message   string `json:"message"`
	RequestID string `json:"request_id,omitempty"`
}

// ToHTTPResponse converts an AppError to HTTP status code and error response
func (e *AppError) ToHTTPResponse(requestID string) (int, ErrorResponse) {
	return e.StatusCode, ErrorResponse{
		Type:      string(e.Type),
		Message:   e.Message,
		RequestID: requestID,
	}
}

// GetStatusCode extracts HTTP status code from any error
// Returns 500 for unknown errors
func GetStatusCode(err error) int {
	if appErr, ok := err.(*AppError); ok {
		return appErr.StatusCode
	}
	return http.StatusInternalServerError
}
