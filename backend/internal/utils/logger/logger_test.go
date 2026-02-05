package logger

import (
	"bytes"
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLevelString(t *testing.T) {
	tests := []struct {
		level    Level
		expected string
	}{
		{LevelDebug, "DEBUG"},
		{LevelInfo, "INFO"},
		{LevelWarn, "WARN"},
		{LevelError, "ERROR"},
		{Level(999), "UNKNOWN"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.level.String())
		})
	}
}

func TestNewLogger(t *testing.T) {
	buf := &bytes.Buffer{}
	logger := New(buf, LevelInfo)

	assert.NotNil(t, logger)
	assert.Equal(t, LevelInfo, logger.level)
}

func TestLogger_SetLevel(t *testing.T) {
	buf := &bytes.Buffer{}
	logger := New(buf, LevelInfo)

	logger.SetLevel(LevelDebug)
	assert.Equal(t, LevelDebug, logger.level)

	logger.SetLevel(LevelError)
	assert.Equal(t, LevelError, logger.level)
}

func TestLogger_Debug(t *testing.T) {
	buf := &bytes.Buffer{}
	logger := New(buf, LevelDebug)

	logger.Debug("debug message")

	var entry Entry
	err := json.Unmarshal(buf.Bytes(), &entry)
	require.NoError(t, err)

	assert.Equal(t, "DEBUG", entry.Level)
	assert.Equal(t, "debug message", entry.Message)
	assert.NotEmpty(t, entry.Timestamp)
}

func TestLogger_Info(t *testing.T) {
	buf := &bytes.Buffer{}
	logger := New(buf, LevelInfo)

	logger.Info("info message")

	var entry Entry
	err := json.Unmarshal(buf.Bytes(), &entry)
	require.NoError(t, err)

	assert.Equal(t, "INFO", entry.Level)
	assert.Equal(t, "info message", entry.Message)
}

func TestLogger_Warn(t *testing.T) {
	buf := &bytes.Buffer{}
	logger := New(buf, LevelWarn)

	logger.Warn("warning message")

	var entry Entry
	err := json.Unmarshal(buf.Bytes(), &entry)
	require.NoError(t, err)

	assert.Equal(t, "WARN", entry.Level)
	assert.Equal(t, "warning message", entry.Message)
}

func TestLogger_Error(t *testing.T) {
	buf := &bytes.Buffer{}
	logger := New(buf, LevelError)

	logger.Error("error message")

	var entry Entry
	err := json.Unmarshal(buf.Bytes(), &entry)
	require.NoError(t, err)

	assert.Equal(t, "ERROR", entry.Level)
	assert.Equal(t, "error message", entry.Message)
}

func TestLogger_LevelFiltering(t *testing.T) {
	buf := &bytes.Buffer{}
	logger := New(buf, LevelWarn)

	logger.Debug("debug")
	logger.Info("info")
	logger.Warn("warn")
	logger.Error("error")

	// Only WARN and ERROR should be logged
	output := buf.String()
	assert.NotContains(t, output, "DEBUG")
	assert.NotContains(t, output, "INFO")
	assert.Contains(t, output, "WARN")
	assert.Contains(t, output, "ERROR")
}

func TestLogger_Debugf(t *testing.T) {
	buf := &bytes.Buffer{}
	logger := New(buf, LevelDebug)

	logger.Debugf("formatted %s %d", "message", 42)

	var entry Entry
	err := json.Unmarshal(buf.Bytes(), &entry)
	require.NoError(t, err)

	assert.Equal(t, "DEBUG", entry.Level)
	assert.Equal(t, "formatted message 42", entry.Message)
}

func TestLogger_Infof(t *testing.T) {
	buf := &bytes.Buffer{}
	logger := New(buf, LevelInfo)

	logger.Infof("user %s logged in from %s", "john", "192.168.1.1")

	var entry Entry
	err := json.Unmarshal(buf.Bytes(), &entry)
	require.NoError(t, err)

	assert.Equal(t, "INFO", entry.Level)
	assert.Equal(t, "user john logged in from 192.168.1.1", entry.Message)
}

func TestLogger_Warnf(t *testing.T) {
	buf := &bytes.Buffer{}
	logger := New(buf, LevelWarn)

	logger.Warnf("connection timeout after %d seconds", 30)

	var entry Entry
	err := json.Unmarshal(buf.Bytes(), &entry)
	require.NoError(t, err)

	assert.Equal(t, "WARN", entry.Level)
	assert.Equal(t, "connection timeout after 30 seconds", entry.Message)
}

func TestLogger_Errorf(t *testing.T) {
	buf := &bytes.Buffer{}
	logger := New(buf, LevelError)

	logger.Errorf("failed to process request %s: %s", "req-123", "timeout")

	var entry Entry
	err := json.Unmarshal(buf.Bytes(), &entry)
	require.NoError(t, err)

	assert.Equal(t, "ERROR", entry.Level)
	assert.Equal(t, "failed to process request req-123: timeout", entry.Message)
}

func TestLogger_WithRequestID(t *testing.T) {
	buf := &bytes.Buffer{}
	logger := New(buf, LevelInfo)

	requestLogger := logger.WithRequestID("req-12345")
	requestLogger.Info("processing request")

	var entry Entry
	err := json.Unmarshal(buf.Bytes(), &entry)
	require.NoError(t, err)

	assert.Equal(t, "req-12345", entry.RequestID)
	assert.Equal(t, "processing request", entry.Message)
}

func TestLogger_WithTenantID(t *testing.T) {
	buf := &bytes.Buffer{}
	logger := New(buf, LevelInfo)

	tenantLogger := logger.WithTenantID("tenant-abc")
	tenantLogger.Info("tenant operation")

	var entry Entry
	err := json.Unmarshal(buf.Bytes(), &entry)
	require.NoError(t, err)

	assert.Equal(t, "tenant-abc", entry.TenantID)
	assert.Equal(t, "tenant operation", entry.Message)
}

func TestLogger_WithUserID(t *testing.T) {
	buf := &bytes.Buffer{}
	logger := New(buf, LevelInfo)

	userLogger := logger.WithUserID("user-xyz")
	userLogger.Info("user action")

	var entry Entry
	err := json.Unmarshal(buf.Bytes(), &entry)
	require.NoError(t, err)

	assert.Equal(t, "user-xyz", entry.UserID)
	assert.Equal(t, "user action", entry.Message)
}

func TestLogger_WithField(t *testing.T) {
	buf := &bytes.Buffer{}
	logger := New(buf, LevelInfo)

	fieldLogger := logger.WithField("order_id", "ord-456")
	fieldLogger.Info("order processed")

	var entry Entry
	err := json.Unmarshal(buf.Bytes(), &entry)
	require.NoError(t, err)

	assert.Equal(t, "order processed", entry.Message)
	require.NotNil(t, entry.Fields)
	assert.Equal(t, "ord-456", entry.Fields["order_id"])
}

func TestLogger_WithFields(t *testing.T) {
	buf := &bytes.Buffer{}
	logger := New(buf, LevelInfo)

	fieldLogger := logger.WithFields(map[string]interface{}{
		"order_id":    "ord-789",
		"customer_id": "cust-123",
		"amount":      99.99,
	})
	fieldLogger.Info("order created")

	var entry Entry
	err := json.Unmarshal(buf.Bytes(), &entry)
	require.NoError(t, err)

	assert.Equal(t, "order created", entry.Message)
	require.NotNil(t, entry.Fields)
	assert.Equal(t, "ord-789", entry.Fields["order_id"])
	assert.Equal(t, "cust-123", entry.Fields["customer_id"])
	assert.Equal(t, 99.99, entry.Fields["amount"])
}

func TestLogger_ChainedContexts(t *testing.T) {
	buf := &bytes.Buffer{}
	logger := New(buf, LevelInfo)

	chainedLogger := logger.
		WithRequestID("req-001").
		WithTenantID("tenant-001").
		WithUserID("user-001").
		WithField("action", "create").
		WithField("resource", "order")

	chainedLogger.Info("complex operation")

	var entry Entry
	err := json.Unmarshal(buf.Bytes(), &entry)
	require.NoError(t, err)

	assert.Equal(t, "req-001", entry.RequestID)
	assert.Equal(t, "tenant-001", entry.TenantID)
	assert.Equal(t, "user-001", entry.UserID)
	assert.Equal(t, "complex operation", entry.Message)
	assert.Equal(t, "create", entry.Fields["action"])
	assert.Equal(t, "order", entry.Fields["resource"])
}

func TestLogger_MultipleLogLines(t *testing.T) {
	buf := &bytes.Buffer{}
	logger := New(buf, LevelInfo)

	logger.Info("first message")
	logger.Info("second message")
	logger.Info("third message")

	lines := bytes.Split(bytes.TrimSpace(buf.Bytes()), []byte("\n"))
	assert.Equal(t, 3, len(lines))

	var entries []Entry
	for _, line := range lines {
		var entry Entry
		err := json.Unmarshal(line, &entry)
		require.NoError(t, err)
		entries = append(entries, entry)
	}

	assert.Equal(t, "first message", entries[0].Message)
	assert.Equal(t, "second message", entries[1].Message)
	assert.Equal(t, "third message", entries[2].Message)
}

func TestLogger_EmptyFields(t *testing.T) {
	buf := &bytes.Buffer{}
	logger := New(buf, LevelInfo)

	logger.Info("message without fields")

	var entry Entry
	err := json.Unmarshal(buf.Bytes(), &entry)
	require.NoError(t, err)

	assert.Nil(t, entry.Fields)
}

func TestLogger_CopyFields(t *testing.T) {
	fields := map[string]interface{}{
		"key1": "value1",
		"key2": "value2",
	}
	copied := copyFields(fields)

	// Verify it's a copy, not a reference
	assert.Equal(t, fields, copied)

	// Modify original
	fields["key1"] = "modified"
	assert.NotEqual(t, fields["key1"], copied["key1"])
}

func TestLogger_JSONEncoding(t *testing.T) {
	buf := &bytes.Buffer{}
	logger := New(buf, LevelInfo)

	logger.WithFields(map[string]interface{}{
		"string": "value",
		"number": 42,
		"bool":   true,
		"null":   nil,
	}).Info("message")

	var entry Entry
	err := json.Unmarshal(buf.Bytes(), &entry)
	require.NoError(t, err)

	assert.Equal(t, "value", entry.Fields["string"])
	assert.Equal(t, float64(42), entry.Fields["number"])
	assert.Equal(t, true, entry.Fields["bool"])
	assert.Nil(t, entry.Fields["null"])
}

func TestLogger_ConcurrentLogging(t *testing.T) {
	buf := &bytes.Buffer{}
	logger := New(buf, LevelInfo)

	// Simulate concurrent logging (just verify no race conditions)
	done := make(chan bool, 2)

	go func() {
		logger.Info("message 1")
		done <- true
	}()

	go func() {
		logger.Info("message 2")
		done <- true
	}()

	<-done
	<-done

	lines := bytes.Split(bytes.TrimSpace(buf.Bytes()), []byte("\n"))
	assert.Equal(t, 2, len(lines))
}
