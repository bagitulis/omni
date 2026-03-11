package config

import (
	"context"
	"errors"
	"fmt"
	"github.com/rs/zerolog/log"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// SilentRecordNotFoundLogger is a custom GORM logger that suppresses "record not found" logs
// These are expected behaviors when checking if settings exist before creating defaults
type SilentRecordNotFoundLogger struct {
	logger.Interface
	LogLevel logger.LogLevel
}

// NewSilentRecordNotFoundLogger creates a new logger that ignores ErrRecordNotFound
func NewSilentRecordNotFoundLogger(level logger.LogLevel) *SilentRecordNotFoundLogger {
	return &SilentRecordNotFoundLogger{
		Interface: logger.Default.LogMode(level),
		LogLevel:  level,
	}
}

// LogMode implements logger.Interface
func (l *SilentRecordNotFoundLogger) LogMode(level logger.LogLevel) logger.Interface {
	return &SilentRecordNotFoundLogger{
		Interface: l.Interface.LogMode(level),
		LogLevel:  level,
	}
}

// Info implements logger.Interface
func (l *SilentRecordNotFoundLogger) Info(ctx context.Context, msg string, data ...interface{}) {
	if l.LogLevel >= logger.Info {
		log.Info().Msgf(msg, data...)
	}
}

// Warn implements logger.Interface
func (l *SilentRecordNotFoundLogger) Warn(ctx context.Context, msg string, data ...interface{}) {
	if l.LogLevel >= logger.Warn {
		log.Info().Msgf("[WARN] "+msg, data...)
	}
}

// Error implements logger.Interface
func (l *SilentRecordNotFoundLogger) Error(ctx context.Context, msg string, data ...interface{}) {
	if l.LogLevel >= logger.Error {
		log.Info().Msgf("[ERROR] "+msg, data...)
	}
}

// Trace implements logger.Interface - this is where SQL queries are logged
func (l *SilentRecordNotFoundLogger) Trace(ctx context.Context, begin time.Time, fc func() (sql string, rowsAffected int64), err error) {
	// Skip logging for ErrRecordNotFound - this is expected behavior
	if err != nil && errors.Is(err, gorm.ErrRecordNotFound) {
		return
	}

	// For other cases, use default behavior based on log level
	if l.LogLevel <= logger.Silent {
		return
	}

	elapsed := time.Since(begin)
	sql, rows := fc()

	// Only log errors (excluding record not found)
	if err != nil {
		if l.LogLevel >= logger.Error {
			log.Info().Msgf("[ERROR] %s [%.3fms] [rows:%d] %s", err.Error(), float64(elapsed.Nanoseconds())/1e6, rows, sql)
		}
		return
	}

	// Log slow queries (> 200ms) as warnings
	if elapsed > 200*time.Millisecond && l.LogLevel >= logger.Warn {
		log.Info().Msgf("[SLOW SQL] [%.3fms] [rows:%d] %s", float64(elapsed.Nanoseconds())/1e6, rows, sql)
		return
	}

	// Log all queries in Info mode (development)
	if l.LogLevel >= logger.Info {
		log.Info().Msgf("[SQL] [%.3fms] [rows:%d] %s", float64(elapsed.Nanoseconds())/1e6, rows, sql)
	}
}

// ParamsFilter implements logger.Interface (optional, for parameter filtering)
func (l *SilentRecordNotFoundLogger) ParamsFilter(ctx context.Context, sql string, params ...interface{}) (string, []interface{}) {
	return sql, params
}

// String returns logger description
func (l *SilentRecordNotFoundLogger) String() string {
	return fmt.Sprintf("SilentRecordNotFoundLogger(level=%d)", l.LogLevel)
}
