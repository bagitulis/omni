package middleware

import (
	"fmt"
	"net/http"
	"os"
	"runtime/debug"

	"github.com/gin-gonic/gin"
	"github.com/omni/backend/internal/errors"
	zlog "github.com/rs/zerolog/log"
)

// ErrorResponse is the standard error response format
type ErrorResponse struct {
	Success bool        `json:"success"`
	Error   string      `json:"error"`
	Type    string      `json:"type,omitempty"`
	Details interface{} `json:"details,omitempty"`
}

// ErrorHandler is middleware that recovers from panics and handles errors
func ErrorHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if r := recover(); r != nil {
				// Log panic with stack trace in development
				if os.Getenv("GO_ENV") != "production" {
					zlog.Error().Interface("panic", r).Str("stack", string(debug.Stack())).Msg("Panic recovered")
				} else {
					zlog.Error().Interface("panic", r).Msg("Panic recovered")
				}

				errMsg := "An unexpected error occurred"
				var details interface{}
				if os.Getenv("GO_ENV") != "production" {
					details = fmt.Sprintf("%v", r)
				}
				c.AbortWithStatusJSON(http.StatusInternalServerError, ErrorResponse{
					Success: false,
					Error:   errMsg,
					Type:    errors.TypeInternal,
					Details: details,
				})
			}
		}()

		c.Next()

		// Handle errors set during request processing
		if len(c.Errors) > 0 {
			err := c.Errors.Last().Err
			handleError(c, err)
		}
	}
}

// handleError converts error to appropriate HTTP response
func handleError(c *gin.Context, err error) {
	if appErr, ok := err.(*errors.AppError); ok {
		response := ErrorResponse{
			Success: false,
			Error:   appErr.Message,
			Type:    appErr.Type,
			Details: appErr.Details,
		}

		// Don't expose internal error details in production
		if appErr.InternalErr != nil && os.Getenv("GO_ENV") != "production" {
			zlog.Error().Err(appErr.InternalErr).Msg("Internal error")
		}

		c.AbortWithStatusJSON(appErr.Code, response)
		return
	}

	// Log the raw error (was previously silently lost)
	zlog.Warn().Err(err).Msg("Unhandled error")

	errMsg := "An unexpected error occurred"
	var details interface{}
	if os.Getenv("GO_ENV") != "production" {
		details = err.Error()
	}
	c.AbortWithStatusJSON(http.StatusInternalServerError, ErrorResponse{
		Success: false,
		Error:   errMsg,
		Type:    errors.TypeInternal,
		Details: details,
	})
}

// AbortWithError is a helper to abort with an AppError
func AbortWithError(c *gin.Context, err *errors.AppError) {
	response := ErrorResponse{
		Success: false,
		Error:   err.Message,
		Type:    err.Type,
		Details: err.Details,
	}
	c.AbortWithStatusJSON(err.Code, response)
}
