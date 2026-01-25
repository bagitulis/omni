package middleware

import (
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/omni/backend/internal/utils/logger"
)

const RequestIDKey = "requestId"

// RequestLogger logs incoming requests and responses with structured logging
func RequestLogger() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Generate or get request ID
		requestID := c.GetHeader("X-Request-ID")
		if requestID == "" {
			requestID = uuid.New().String()
		}
		c.Set(RequestIDKey, requestID)
		c.Header("X-Request-ID", requestID)

		// Get tenant ID if available
		tenantID := GetTenantID(c)

		// Create logger with context
		log := logger.WithRequestID(requestID)
		if tenantID != "" {
			log = log.WithTenantID(tenantID)
		}

		// Log request start
		start := time.Now()
		path := c.Request.URL.Path
		raw := c.Request.URL.RawQuery

		if raw != "" {
			path = path + "?" + raw
		}

		log.WithFields(map[string]interface{}{
			"method":   c.Request.Method,
			"path":     path,
			"clientIP": c.ClientIP(),
		}).Info("Request started")

		// Process request
		c.Next()

		// Log response
		duration := time.Since(start)
		statusCode := c.Writer.Status()

		logFields := map[string]interface{}{
			"method":     c.Request.Method,
			"path":       path,
			"statusCode": statusCode,
			"duration":   duration.String(),
			"durationMs": duration.Milliseconds(),
		}

		if len(c.Errors) > 0 {
			logFields["errors"] = c.Errors.Errors()
		}

		logLevel := logger.LevelInfo
		if statusCode >= 500 {
			logLevel = logger.LevelError
		} else if statusCode >= 400 {
			logLevel = logger.LevelWarn
		}

		switch logLevel {
		case logger.LevelError:
			log.WithFields(logFields).Error("Request completed with error")
		case logger.LevelWarn:
			log.WithFields(logFields).Warn("Request completed with warning")
		default:
			log.WithFields(logFields).Info("Request completed")
		}
	}
}

// GetRequestID returns the request ID from context
func GetRequestID(c *gin.Context) string {
	if id, exists := c.Get(RequestIDKey); exists {
		return id.(string)
	}
	return ""
}
