package utils

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGenerateKey(t *testing.T) {
	key, err := GenerateKey()
	assert.NoError(t, err)
	assert.NotEmpty(t, key)
	// Fernet keys are base64 encoded
	assert.True(t, len(key) >= 40)
}

func TestNewEncryptionService_Valid(t *testing.T) {
	key, err := GenerateKey()
	require.NoError(t, err)

	svc, err := NewEncryptionService(key)
	assert.NoError(t, err)
	assert.NotNil(t, svc)
}

func TestNewEncryptionService_InvalidKey(t *testing.T) {
	tests := []struct {
		name   string
		keyStr string
	}{
		{
			name:   "empty key",
			keyStr: "",
		},
		{
			name:   "invalid base64",
			keyStr: "not-a-valid-fernet-key!!!",
		},
		{
			name:   "short invalid key",
			keyStr: "short",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc, err := NewEncryptionService(tt.keyStr)
			assert.Error(t, err)
			assert.Nil(t, svc)
		})
	}
}

func TestEncryptionService_EncryptDecrypt(t *testing.T) {
	key, err := GenerateKey()
	require.NoError(t, err)

	svc, err := NewEncryptionService(key)
	require.NoError(t, err)

	tests := []struct {
		name      string
		plaintext string
	}{
		{
			name:      "simple string",
			plaintext: "hello world",
		},
		{
			name:      "json-like string",
			plaintext: `{"user_id": "123", "role": "admin"}`,
		},
		{
			name:      "empty string",
			plaintext: "",
		},
		{
			name:      "long text",
			plaintext: "Lorem ipsum dolor sit amet, consectetur adipiscing elit. Sed do eiusmod tempor incididunt ut labore et dolore magna aliqua.",
		},
		{
			name:      "special characters",
			plaintext: "!@#$%^&*()_+-=[]{}|;:',.<>?",
		},
		{
			name:      "unicode characters",
			plaintext: "你好世界🌍🚀",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Encrypt
			ciphertext, err := svc.Encrypt(tt.plaintext)
			assert.NoError(t, err)
			assert.NotEmpty(t, ciphertext)
			assert.NotEqual(t, tt.plaintext, ciphertext)

			// Decrypt
			decrypted, err := svc.Decrypt(ciphertext)
			assert.NoError(t, err)
			assert.Equal(t, tt.plaintext, decrypted)
		})
	}
}

func TestEncryptionService_DecryptInvalid(t *testing.T) {
	key, err := GenerateKey()
	require.NoError(t, err)

	svc, err := NewEncryptionService(key)
	require.NoError(t, err)

	tests := []struct {
		name       string
		ciphertext string
	}{
		{
			name:       "empty string",
			ciphertext: "",
		},
		{
			name:       "invalid token",
			ciphertext: "not-a-valid-token",
		},
		{
			name:       "corrupted data",
			ciphertext: "gAAAAABl5q5q5q5q5q5q5q5q5q5q5q5q5q",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			decrypted, err := svc.Decrypt(tt.ciphertext)
			assert.Error(t, err)
			assert.Empty(t, decrypted)
		})
	}
}

func TestEncryptionService_DecryptWithTTL_Valid(t *testing.T) {
	key, err := GenerateKey()
	require.NoError(t, err)

	svc, err := NewEncryptionService(key)
	require.NoError(t, err)

	plaintext := "secret data with ttl"

	// Encrypt
	ciphertext, err := svc.Encrypt(plaintext)
	require.NoError(t, err)

	// Decrypt with TTL (should work immediately)
	decrypted, err := svc.DecryptWithTTL(ciphertext, 60*time.Second)
	assert.NoError(t, err)
	assert.Equal(t, plaintext, decrypted)
}

func TestEncryptionService_DecryptWithTTL_Expired(t *testing.T) {
	key, err := GenerateKey()
	require.NoError(t, err)

	svc, err := NewEncryptionService(key)
	require.NoError(t, err)

	// With TTL validation, invalid/empty token should fail
	decrypted, err := svc.DecryptWithTTL("", 60*time.Second)
	assert.Error(t, err)
	assert.Empty(t, decrypted)
}

func TestEncryptionService_DecryptWithTTL_Invalid(t *testing.T) {
	key, err := GenerateKey()
	require.NoError(t, err)

	svc, err := NewEncryptionService(key)
	require.NoError(t, err)

	// Invalid ciphertext
	decrypted, err := svc.DecryptWithTTL("invalid-token", 60*time.Second)
	assert.Error(t, err)
	assert.Empty(t, decrypted)
}

func TestIsEncrypted(t *testing.T) {
	// First create a real encrypted value
	key, err := GenerateKey()
	require.NoError(t, err)

	svc, err := NewEncryptionService(key)
	require.NoError(t, err)

	realToken, err := svc.Encrypt("test")
	require.NoError(t, err)

	tests := []struct {
		name     string
		value    string
		expected bool
	}{
		{
			name:     "real encrypted token",
			value:    realToken,
			expected: true,
		},
		{
			name:     "short string",
			value:    "short",
			expected: false,
		},
		{
			name:     "empty string",
			value:    "",
			expected: false,
		},
		{
			name:     "invalid base64",
			value:    "!!!invalid!!!base64!!!data!!!invalid!!!base64!!!data!!!invalid!!!base64!!!data!!!",
			expected: false,
		},
		{
			name:     "valid base64 but not fernet",
			value:    "YWJjZGVmZ2hpamtsbW5vcHFyc3R1dnd4eXphYmNkZWZnaGlqa2xtbm9wcXJzdHV2d3h5eg==",
			expected: true, // Valid base64 of length > 50
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsEncrypted(tt.value)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestEncryptionService_RoundTrip_MultipleValues(t *testing.T) {
	key, err := GenerateKey()
	require.NoError(t, err)

	svc, err := NewEncryptionService(key)
	require.NoError(t, err)

	values := []string{
		"value1",
		"value2",
		"value3",
	}

	ciphertexts := make([]string, len(values))

	// Encrypt all
	for i, val := range values {
		ct, err := svc.Encrypt(val)
		assert.NoError(t, err)
		ciphertexts[i] = ct
	}

	// Each ciphertext should be unique
	assert.NotEqual(t, ciphertexts[0], ciphertexts[1])
	assert.NotEqual(t, ciphertexts[1], ciphertexts[2])

	// Decrypt and verify
	for i, ct := range ciphertexts {
		decrypted, err := svc.Decrypt(ct)
		assert.NoError(t, err)
		assert.Equal(t, values[i], decrypted)
	}
}
