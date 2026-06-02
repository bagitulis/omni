package handlers

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestParseTimestampMs_EmptyString verifies empty string returns error
func TestParseTimestampMs_EmptyString(t *testing.T) {
	ms, err := parseTimestampMs("")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "empty string")
	assert.Equal(t, int64(0), ms)
}

// TestParseTimestampMs_NonNumeric verifies non-numeric string returns error
func TestParseTimestampMs_NonNumeric(t *testing.T) {
	ms, err := parseTimestampMs("abc")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid timestamp")
	assert.Equal(t, int64(0), ms)
}

// TestParseTimestampMs_InvalidMixed verifies partially numeric string returns error
func TestParseTimestampMs_InvalidMixed(t *testing.T) {
	ms, err := parseTimestampMs("123abc")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid timestamp")
	assert.Equal(t, int64(0), ms)
}

// TestParseTimestampMs_ValidTimestamp verifies valid millisecond timestamp is parsed
func TestParseTimestampMs_ValidTimestamp(t *testing.T) {
	ms, err := parseTimestampMs("1234567890123")
	assert.NoError(t, err)
	assert.Equal(t, int64(1234567890123), ms)
}

// TestParseTimestampMs_Zero verifies "0" is parsed successfully (valid number)
func TestParseTimestampMs_Zero(t *testing.T) {
	ms, err := parseTimestampMs("0")
	assert.NoError(t, err)
	assert.Equal(t, int64(0), ms)
}

// TestParseTimestampMs_WhitespaceOnly verifies whitespace-only string returns error
func TestParseTimestampMs_WhitespaceOnly(t *testing.T) {
	ms, err := parseTimestampMs("   ")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid timestamp")
	assert.Equal(t, int64(0), ms)
}

// TestBuildPlatformTokenStatus_InvalidConfig verifies invalid tokenExpiry returns invalid_config status
func TestBuildPlatformTokenStatus_InvalidConfig(t *testing.T) {
	configMap := map[string]string{
		"tokenExpiry": "not-a-number",
		"shopId":      "test-shop",
		"shopName":    "Test Shop",
	}

	nowMs := int64(1700000000000)
	result := buildPlatformTokenStatus(configMap, nowMs)

	assert.Equal(t, "invalid_config", result["status"])
	assert.Equal(t, false, result["valid"])
	assert.Equal(t, true, result["isExpired"])
	assert.Nil(t, result["expiresAt"])
	assert.Nil(t, result["refreshTokenExpiresAt"])
}

// TestBuildPlatformTokenStatus_EmptyTokenExpiry verifies empty tokenExpiry returns invalid_config status
func TestBuildPlatformTokenStatus_EmptyTokenExpiry(t *testing.T) {
	configMap := map[string]string{
		"tokenExpiry": "",
		"shopId":      "test-shop",
		"shopName":    "Test Shop",
	}

	nowMs := int64(1700000000000)
	result := buildPlatformTokenStatus(configMap, nowMs)

	assert.Equal(t, "invalid_config", result["status"])
	assert.Equal(t, false, result["valid"])
	assert.Equal(t, true, result["isExpired"])
}

// TestBuildPlatformTokenStatus_ValidToken verifies valid token expiry computes correct status
func TestBuildPlatformTokenStatus_ValidToken(t *testing.T) {
	futureMs := int64(1730000000000) // some future timestamp
	pastMs := int64(1600000000000)   // some past timestamp
	nowMs := int64(1700000000000)

	t.Run("valid future token", func(t *testing.T) {
		configMap := map[string]string{
			"tokenExpiry": "1730000000000",
			"shopId":      "test-shop",
			"shopName":    "Test Shop",
		}
		result := buildPlatformTokenStatus(configMap, nowMs)
		assert.Equal(t, "valid", result["status"])
		assert.Equal(t, true, result["valid"])
		assert.Equal(t, false, result["isExpired"])
		assert.Equal(t, "test-shop", result["shopId"])
		assert.Equal(t, "Test Shop", result["shopName"])
	})

	t.Run("expired past token", func(t *testing.T) {
		configMap := map[string]string{
			"tokenExpiry": "1600000000000",
			"shopId":      "test-shop",
			"shopName":    "Test Shop",
		}
		// nowMs is 1700000000000, so 1600000000000 is in the past
		result := buildPlatformTokenStatus(configMap, nowMs)
		assert.Equal(t, "expired", result["status"])
		assert.Equal(t, false, result["valid"])
		assert.Equal(t, true, result["isExpired"])
	})

	_ = futureMs
	_ = pastMs
}
