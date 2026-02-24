package errors_test

import (
	"errors"
	"fmt"
	"net/http"
	"testing"

	apperrors "github.com/omni/backend/internal/errors"
)

// ---------------------------------------------------------------------------
// Type constant tests
// ---------------------------------------------------------------------------

func TestTypeConstants(t *testing.T) {
	tests := []struct {
		name     string
		constant string
		want     string
	}{
		{"TypeValidation", apperrors.TypeValidation, "VALIDATION_ERROR"},
		{"TypeNotFound", apperrors.TypeNotFound, "NOT_FOUND"},
		{"TypeUnauthorized", apperrors.TypeUnauthorized, "UNAUTHORIZED"},
		{"TypeForbidden", apperrors.TypeForbidden, "FORBIDDEN"},
		{"TypeConflict", apperrors.TypeConflict, "CONFLICT"},
		{"TypeRateLimit", apperrors.TypeRateLimit, "RATE_LIMIT_EXCEEDED"},
		{"TypeInternal", apperrors.TypeInternal, "INTERNAL_ERROR"},
		{"TypeBadRequest", apperrors.TypeBadRequest, "BAD_REQUEST"},
		{"TypeServiceUnavail", apperrors.TypeServiceUnavail, "SERVICE_UNAVAILABLE"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.constant != tt.want {
				t.Errorf("got %q, want %q", tt.constant, tt.want)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// AppError.Error() method
// ---------------------------------------------------------------------------

func TestAppError_Error_WithoutInternalErr(t *testing.T) {
	err := &apperrors.AppError{
		Code:    http.StatusBadRequest,
		Type:    apperrors.TypeValidation,
		Message: "field is required",
	}
	if got := err.Error(); got != "field is required" {
		t.Errorf("Error() = %q, want %q", got, "field is required")
	}
}

func TestAppError_Error_WithInternalErr(t *testing.T) {
	internal := fmt.Errorf("database connection refused")
	err := &apperrors.AppError{
		Code:        http.StatusInternalServerError,
		Type:        apperrors.TypeInternal,
		Message:     "An internal error occurred",
		InternalErr: internal,
	}
	got := err.Error()
	want := "An internal error occurred: database connection refused"
	if got != want {
		t.Errorf("Error() = %q, want %q", got, want)
	}
}

// ---------------------------------------------------------------------------
// AppError.Unwrap() method
// ---------------------------------------------------------------------------

func TestAppError_Unwrap_ReturnsInternalErr(t *testing.T) {
	sentinel := fmt.Errorf("sentinel error")
	appErr := &apperrors.AppError{
		Code:        http.StatusInternalServerError,
		Type:        apperrors.TypeInternal,
		Message:     "wrapped",
		InternalErr: sentinel,
	}
	if unwrapped := appErr.Unwrap(); unwrapped != sentinel {
		t.Errorf("Unwrap() = %v, want %v", unwrapped, sentinel)
	}
	// errors.Is should work through Unwrap
	if !errors.Is(appErr, sentinel) {
		t.Error("errors.Is should find the sentinel via Unwrap")
	}
}

func TestAppError_Unwrap_NilWhenNoInternalErr(t *testing.T) {
	appErr := &apperrors.AppError{
		Code:    http.StatusBadRequest,
		Type:    apperrors.TypeBadRequest,
		Message: "no inner",
	}
	if unwrapped := appErr.Unwrap(); unwrapped != nil {
		t.Errorf("Unwrap() = %v, want nil", unwrapped)
	}
}

// ---------------------------------------------------------------------------
// Constructor tests
// ---------------------------------------------------------------------------

func TestNewValidationError(t *testing.T) {
	details := map[string]string{"field": "name", "reason": "required"}
	err := apperrors.NewValidationError("validation failed", details)

	if err.Code != http.StatusBadRequest {
		t.Errorf("Code = %d, want %d", err.Code, http.StatusBadRequest)
	}
	if err.Type != apperrors.TypeValidation {
		t.Errorf("Type = %q, want %q", err.Type, apperrors.TypeValidation)
	}
	if err.Message != "validation failed" {
		t.Errorf("Message = %q, want %q", err.Message, "validation failed")
	}
	if err.Details == nil {
		t.Error("Details should not be nil")
	}
	if err.InternalErr != nil {
		t.Errorf("InternalErr should be nil, got %v", err.InternalErr)
	}
}

func TestNewValidationError_NilDetails(t *testing.T) {
	err := apperrors.NewValidationError("bad input", nil)
	if err.Details != nil {
		t.Errorf("Details = %v, want nil", err.Details)
	}
}

func TestNewNotFoundError(t *testing.T) {
	err := apperrors.NewNotFoundError("Product")

	if err.Code != http.StatusNotFound {
		t.Errorf("Code = %d, want %d", err.Code, http.StatusNotFound)
	}
	if err.Type != apperrors.TypeNotFound {
		t.Errorf("Type = %q, want %q", err.Type, apperrors.TypeNotFound)
	}
	want := "Product not found"
	if err.Message != want {
		t.Errorf("Message = %q, want %q", err.Message, want)
	}
}

func TestNewUnauthorizedError_WithMessage(t *testing.T) {
	err := apperrors.NewUnauthorizedError("invalid token")

	if err.Code != http.StatusUnauthorized {
		t.Errorf("Code = %d, want %d", err.Code, http.StatusUnauthorized)
	}
	if err.Type != apperrors.TypeUnauthorized {
		t.Errorf("Type = %q, want %q", err.Type, apperrors.TypeUnauthorized)
	}
	if err.Message != "invalid token" {
		t.Errorf("Message = %q, want %q", err.Message, "invalid token")
	}
}

func TestNewUnauthorizedError_EmptyMessage_DefaultsToAuthRequired(t *testing.T) {
	err := apperrors.NewUnauthorizedError("")

	want := "Authentication required"
	if err.Message != want {
		t.Errorf("Message = %q, want %q", err.Message, want)
	}
}

func TestNewForbiddenError_WithMessage(t *testing.T) {
	err := apperrors.NewForbiddenError("insufficient permissions")

	if err.Code != http.StatusForbidden {
		t.Errorf("Code = %d, want %d", err.Code, http.StatusForbidden)
	}
	if err.Type != apperrors.TypeForbidden {
		t.Errorf("Type = %q, want %q", err.Type, apperrors.TypeForbidden)
	}
	if err.Message != "insufficient permissions" {
		t.Errorf("Message = %q, want %q", err.Message, "insufficient permissions")
	}
}

func TestNewForbiddenError_EmptyMessage_DefaultsToAccessDenied(t *testing.T) {
	err := apperrors.NewForbiddenError("")

	want := "Access denied"
	if err.Message != want {
		t.Errorf("Message = %q, want %q", err.Message, want)
	}
}

func TestNewConflictError(t *testing.T) {
	err := apperrors.NewConflictError("resource already exists")

	if err.Code != http.StatusConflict {
		t.Errorf("Code = %d, want %d", err.Code, http.StatusConflict)
	}
	if err.Type != apperrors.TypeConflict {
		t.Errorf("Type = %q, want %q", err.Type, apperrors.TypeConflict)
	}
	if err.Message != "resource already exists" {
		t.Errorf("Message = %q, want %q", err.Message, "resource already exists")
	}
}

func TestNewRateLimitError(t *testing.T) {
	err := apperrors.NewRateLimitError()

	if err.Code != http.StatusTooManyRequests {
		t.Errorf("Code = %d, want %d", err.Code, http.StatusTooManyRequests)
	}
	if err.Type != apperrors.TypeRateLimit {
		t.Errorf("Type = %q, want %q", err.Type, apperrors.TypeRateLimit)
	}
	if err.Message == "" {
		t.Error("Message should not be empty")
	}
}

func TestNewInternalError(t *testing.T) {
	cause := fmt.Errorf("db query failed")
	err := apperrors.NewInternalError(cause)

	if err.Code != http.StatusInternalServerError {
		t.Errorf("Code = %d, want %d", err.Code, http.StatusInternalServerError)
	}
	if err.Type != apperrors.TypeInternal {
		t.Errorf("Type = %q, want %q", err.Type, apperrors.TypeInternal)
	}
	if err.InternalErr != cause {
		t.Errorf("InternalErr = %v, want %v", err.InternalErr, cause)
	}
	// Error() must embed the internal message
	if got := err.Error(); got == "" {
		t.Error("Error() should not be empty")
	}
}

func TestNewBadRequestError(t *testing.T) {
	err := apperrors.NewBadRequestError("missing required fields")

	if err.Code != http.StatusBadRequest {
		t.Errorf("Code = %d, want %d", err.Code, http.StatusBadRequest)
	}
	if err.Type != apperrors.TypeBadRequest {
		t.Errorf("Type = %q, want %q", err.Type, apperrors.TypeBadRequest)
	}
	if err.Message != "missing required fields" {
		t.Errorf("Message = %q, want %q", err.Message, "missing required fields")
	}
}

func TestNewServiceUnavailableError(t *testing.T) {
	err := apperrors.NewServiceUnavailableError("Shopee API")

	if err.Code != http.StatusServiceUnavailable {
		t.Errorf("Code = %d, want %d", err.Code, http.StatusServiceUnavailable)
	}
	if err.Type != apperrors.TypeServiceUnavail {
		t.Errorf("Type = %q, want %q", err.Type, apperrors.TypeServiceUnavail)
	}
	want := "Shopee API is temporarily unavailable"
	if err.Message != want {
		t.Errorf("Message = %q, want %q", err.Message, want)
	}
}

// ---------------------------------------------------------------------------
// Wrap tests
// ---------------------------------------------------------------------------

func TestWrap_WithPlainError_CreatesInternalError(t *testing.T) {
	cause := fmt.Errorf("original error")
	wrapped := apperrors.Wrap(cause, "operation failed")

	if wrapped.Code != http.StatusInternalServerError {
		t.Errorf("Code = %d, want %d", wrapped.Code, http.StatusInternalServerError)
	}
	if wrapped.Type != apperrors.TypeInternal {
		t.Errorf("Type = %q, want %q", wrapped.Type, apperrors.TypeInternal)
	}
	if wrapped.Message != "operation failed" {
		t.Errorf("Message = %q, want %q", wrapped.Message, "operation failed")
	}
	if wrapped.InternalErr != cause {
		t.Errorf("InternalErr = %v, want %v", wrapped.InternalErr, cause)
	}
}

func TestWrap_WithAppError_PrependsMessageAndReturnsSameError(t *testing.T) {
	original := apperrors.NewNotFoundError("Order")
	originalCode := original.Code

	wrapped := apperrors.Wrap(original, "context")

	// Must return the SAME *AppError (mutated)
	if wrapped != original {
		t.Error("Wrap should return the same *AppError pointer when wrapping AppError")
	}
	if wrapped.Code != originalCode {
		t.Errorf("Code changed from %d to %d", originalCode, wrapped.Code)
	}
	want := "context: Order not found"
	if wrapped.Message != want {
		t.Errorf("Message = %q, want %q", wrapped.Message, want)
	}
}

func TestWrap_ErrorMethodIncludesInternalMessage(t *testing.T) {
	cause := fmt.Errorf("tcp dial timeout")
	wrapped := apperrors.Wrap(cause, "connect")

	got := wrapped.Error()
	if got != "connect: tcp dial timeout" {
		t.Errorf("Error() = %q, want %q", got, "connect: tcp dial timeout")
	}
}

// ---------------------------------------------------------------------------
// AppError implements error interface
// ---------------------------------------------------------------------------

func TestAppError_ImplementsErrorInterface(t *testing.T) {
	var _ error = (*apperrors.AppError)(nil)
}

func TestAppError_IsUsableWithErrorsPackage(t *testing.T) {
	sentinel := fmt.Errorf("sentinel")
	appErr := apperrors.NewInternalError(sentinel)
	if !errors.Is(appErr, sentinel) {
		t.Error("errors.Is should find sentinel via Unwrap chain")
	}
}
