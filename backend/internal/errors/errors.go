package errors

import (
	"fmt"
	"net/http"
)

// AppError represents an application-specific error
type AppError struct {
	Code       int         `json:"-"`
	Type       string      `json:"type"`
	Message    string      `json:"message"`
	Details    interface{} `json:"details,omitempty"`
	InternalErr error       `json:"-"`
}

// Error implements the error interface
func (e *AppError) Error() string {
	if e.InternalErr != nil {
		return fmt.Sprintf("%s: %v", e.Message, e.InternalErr)
	}
	return e.Message
}

// Unwrap returns the internal error for errors.Is/As
func (e *AppError) Unwrap() error {
	return e.InternalErr
}

// Common error types
const (
	TypeValidation     = "VALIDATION_ERROR"
	TypeNotFound       = "NOT_FOUND"
	TypeUnauthorized   = "UNAUTHORIZED"
	TypeForbidden      = "FORBIDDEN"
	TypeConflict       = "CONFLICT"
	TypeRateLimit      = "RATE_LIMIT_EXCEEDED"
	TypeInternal       = "INTERNAL_ERROR"
	TypeBadRequest     = "BAD_REQUEST"
	TypeServiceUnavail = "SERVICE_UNAVAILABLE"
)

// NewValidationError creates a validation error
func NewValidationError(message string, details interface{}) *AppError {
	return &AppError{
		Code:    http.StatusBadRequest,
		Type:    TypeValidation,
		Message: message,
		Details: details,
	}
}

// NewNotFoundError creates a not found error
func NewNotFoundError(resource string) *AppError {
	return &AppError{
		Code:    http.StatusNotFound,
		Type:    TypeNotFound,
		Message: fmt.Sprintf("%s not found", resource),
	}
}

// NewUnauthorizedError creates an unauthorized error
func NewUnauthorizedError(message string) *AppError {
	if message == "" {
		message = "Authentication required"
	}
	return &AppError{
		Code:    http.StatusUnauthorized,
		Type:    TypeUnauthorized,
		Message: message,
	}
}

// NewForbiddenError creates a forbidden error
func NewForbiddenError(message string) *AppError {
	if message == "" {
		message = "Access denied"
	}
	return &AppError{
		Code:    http.StatusForbidden,
		Type:    TypeForbidden,
		Message: message,
	}
}

// NewConflictError creates a conflict error
func NewConflictError(message string) *AppError {
	return &AppError{
		Code:    http.StatusConflict,
		Type:    TypeConflict,
		Message: message,
	}
}

// NewRateLimitError creates a rate limit error
func NewRateLimitError() *AppError {
	return &AppError{
		Code:    http.StatusTooManyRequests,
		Type:    TypeRateLimit,
		Message: "Rate limit exceeded, please try again later",
	}
}

// NewInternalError creates an internal server error
func NewInternalError(err error) *AppError {
	return &AppError{
		Code:        http.StatusInternalServerError,
		Type:        TypeInternal,
		Message:     "An internal error occurred",
		InternalErr: err,
	}
}

// NewBadRequestError creates a bad request error
func NewBadRequestError(message string) *AppError {
	return &AppError{
		Code:    http.StatusBadRequest,
		Type:    TypeBadRequest,
		Message: message,
	}
}

// NewServiceUnavailableError creates a service unavailable error
func NewServiceUnavailableError(service string) *AppError {
	return &AppError{
		Code:    http.StatusServiceUnavailable,
		Type:    TypeServiceUnavail,
		Message: fmt.Sprintf("%s is temporarily unavailable", service),
	}
}

// Wrap wraps an error with additional context
func Wrap(err error, message string) *AppError {
	if appErr, ok := err.(*AppError); ok {
		appErr.Message = message + ": " + appErr.Message
		return appErr
	}
	return &AppError{
		Code:        http.StatusInternalServerError,
		Type:        TypeInternal,
		Message:     message,
		InternalErr: err,
	}
}
