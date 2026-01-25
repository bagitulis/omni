package logger

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"sync"
	"time"
)

// Level represents log severity level
type Level int

const (
	LevelDebug Level = iota
	LevelInfo
	LevelWarn
	LevelError
)

// String returns the string representation of log level
func (l Level) String() string {
	switch l {
	case LevelDebug:
		return "DEBUG"
	case LevelInfo:
		return "INFO"
	case LevelWarn:
		return "WARN"
	case LevelError:
		return "ERROR"
	default:
		return "UNKNOWN"
	}
}

// Entry represents a single log entry
type Entry struct {
	Timestamp string                 `json:"timestamp"`
	Level     string                 `json:"level"`
	Message   string                 `json:"message"`
	RequestID string                 `json:"requestId,omitempty"`
	TenantID  string                 `json:"tenantId,omitempty"`
	UserID    string                 `json:"userId,omitempty"`
	Fields    map[string]interface{} `json:"fields,omitempty"`
}

// Logger is a structured JSON logger
type Logger struct {
	mu        sync.Mutex
	out       io.Writer
	level     Level
	requestID string
	tenantID  string
	userID    string
	fields    map[string]interface{}
}

var defaultLogger *Logger

func init() {
	defaultLogger = New(os.Stdout, LevelInfo)
}

// New creates a new logger
func New(out io.Writer, level Level) *Logger {
	return &Logger{
		out:    out,
		level:  level,
		fields: make(map[string]interface{}),
	}
}

// SetLevel sets the minimum log level
func (l *Logger) SetLevel(level Level) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.level = level
}

// WithRequestID returns a new logger with request ID
func (l *Logger) WithRequestID(requestID string) *Logger {
	return &Logger{
		out:       l.out,
		level:     l.level,
		requestID: requestID,
		tenantID:  l.tenantID,
		userID:    l.userID,
		fields:    copyFields(l.fields),
	}
}

// WithTenantID returns a new logger with tenant ID
func (l *Logger) WithTenantID(tenantID string) *Logger {
	return &Logger{
		out:       l.out,
		level:     l.level,
		requestID: l.requestID,
		tenantID:  tenantID,
		userID:    l.userID,
		fields:    copyFields(l.fields),
	}
}

// WithUserID returns a new logger with user ID
func (l *Logger) WithUserID(userID string) *Logger {
	return &Logger{
		out:       l.out,
		level:     l.level,
		requestID: l.requestID,
		tenantID:  l.tenantID,
		userID:    userID,
		fields:    copyFields(l.fields),
	}
}

// WithField returns a new logger with an additional field
func (l *Logger) WithField(key string, value interface{}) *Logger {
	fields := copyFields(l.fields)
	fields[key] = value
	return &Logger{
		out:       l.out,
		level:     l.level,
		requestID: l.requestID,
		tenantID:  l.tenantID,
		userID:    l.userID,
		fields:    fields,
	}
}

// WithFields returns a new logger with additional fields
func (l *Logger) WithFields(fields map[string]interface{}) *Logger {
	merged := copyFields(l.fields)
	for k, v := range fields {
		merged[k] = v
	}
	return &Logger{
		out:       l.out,
		level:     l.level,
		requestID: l.requestID,
		tenantID:  l.tenantID,
		userID:    l.userID,
		fields:    merged,
	}
}

// Debug logs a debug message
func (l *Logger) Debug(msg string) {
	l.log(LevelDebug, msg)
}

// Info logs an info message
func (l *Logger) Info(msg string) {
	l.log(LevelInfo, msg)
}

// Warn logs a warning message
func (l *Logger) Warn(msg string) {
	l.log(LevelWarn, msg)
}

// Error logs an error message
func (l *Logger) Error(msg string) {
	l.log(LevelError, msg)
}

// Debugf logs a formatted debug message
func (l *Logger) Debugf(format string, args ...interface{}) {
	l.log(LevelDebug, fmt.Sprintf(format, args...))
}

// Infof logs a formatted info message
func (l *Logger) Infof(format string, args ...interface{}) {
	l.log(LevelInfo, fmt.Sprintf(format, args...))
}

// Warnf logs a formatted warning message
func (l *Logger) Warnf(format string, args ...interface{}) {
	l.log(LevelWarn, fmt.Sprintf(format, args...))
}

// Errorf logs a formatted error message
func (l *Logger) Errorf(format string, args ...interface{}) {
	l.log(LevelError, fmt.Sprintf(format, args...))
}

func (l *Logger) log(level Level, msg string) {
	if level < l.level {
		return
	}

	entry := Entry{
		Timestamp: time.Now().UTC().Format(time.RFC3339),
		Level:     level.String(),
		Message:   msg,
		RequestID: l.requestID,
		TenantID:  l.tenantID,
		UserID:    l.userID,
	}

	if len(l.fields) > 0 {
		entry.Fields = l.fields
	}

	l.mu.Lock()
	defer l.mu.Unlock()

	data, _ := json.Marshal(entry)
	fmt.Fprintln(l.out, string(data))
}

func copyFields(fields map[string]interface{}) map[string]interface{} {
	copy := make(map[string]interface{}, len(fields))
	for k, v := range fields {
		copy[k] = v
	}
	return copy
}

// Package-level functions using default logger

// Debug logs a debug message
func Debug(msg string) { defaultLogger.Debug(msg) }

// Info logs an info message
func Info(msg string) { defaultLogger.Info(msg) }

// Warn logs a warning message
func Warn(msg string) { defaultLogger.Warn(msg) }

// Error logs an error message
func Error(msg string) { defaultLogger.Error(msg) }

// Debugf logs a formatted debug message
func Debugf(format string, args ...interface{}) { defaultLogger.Debugf(format, args...) }

// Infof logs a formatted info message
func Infof(format string, args ...interface{}) { defaultLogger.Infof(format, args...) }

// Warnf logs a formatted warning message
func Warnf(format string, args ...interface{}) { defaultLogger.Warnf(format, args...) }

// Errorf logs a formatted error message
func Errorf(format string, args ...interface{}) { defaultLogger.Errorf(format, args...) }

// WithRequestID returns a logger with request ID
func WithRequestID(requestID string) *Logger { return defaultLogger.WithRequestID(requestID) }

// WithTenantID returns a logger with tenant ID
func WithTenantID(tenantID string) *Logger { return defaultLogger.WithTenantID(tenantID) }

// WithField returns a logger with an additional field
func WithField(key string, value interface{}) *Logger { return defaultLogger.WithField(key, value) }

// WithFields returns a logger with multiple fields
func WithFields(fields map[string]interface{}) *Logger { return defaultLogger.WithFields(fields) }

// Named returns a logger with a component name field
func Named(name string) *Logger { return defaultLogger.WithField("component", name) }

// SetLevel sets the default logger level
func SetLevel(level Level) { defaultLogger.SetLevel(level) }
