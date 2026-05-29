package utils

import (
	"strings"
	"errors"
	"time"

	"github.com/fernet/fernet-go"
)

// EncryptionService handles Fernet encryption/decryption
type EncryptionService struct {
	key *fernet.Key
}

// NewEncryptionService creates a new encryption service
func NewEncryptionService(keyString string) (*EncryptionService, error) {
	if keyString == "" {
		return nil, errors.New("encryption key is required")
	}

	key, err := fernet.DecodeKey(keyString)
	if err != nil {
		return nil, err
	}

	return &EncryptionService{key: key}, nil
}

// Encrypt encrypts plaintext using Fernet
func (s *EncryptionService) Encrypt(plaintext string) (string, error) {
	token, err := fernet.EncryptAndSign([]byte(plaintext), s.key)
	if err != nil {
		return "", err
	}
	return string(token), nil
}

// Decrypt decrypts Fernet token
func (s *EncryptionService) Decrypt(ciphertext string) (string, error) {
	token := []byte(ciphertext)
	msg := fernet.VerifyAndDecrypt(token, 0, []*fernet.Key{s.key})
	if msg == nil {
		return "", errors.New("decryption failed")
	}
	return string(msg), nil
}

// DecryptWithTTL decrypts with time-to-live validation
func (s *EncryptionService) DecryptWithTTL(ciphertext string, ttl time.Duration) (string, error) {
	token := []byte(ciphertext)
	msg := fernet.VerifyAndDecrypt(token, ttl, []*fernet.Key{s.key})
	if msg == nil {
		return "", errors.New("decryption failed or token expired")
	}
	return string(msg), nil
}

// GenerateKey generates a new Fernet key
func GenerateKey() (string, error) {
	key := fernet.Key{}
	if err := key.Generate(); err != nil {
		return "", err
	}
	return key.Encode(), nil
}

// IsEncrypted checks if a string looks like a Fernet token
func IsEncrypted(value string) bool {
	// Fernet tokens are base64url encoded, version byte 0x80 → always start with "gAAAAA"
	return strings.HasPrefix(value, "gAAAAA")
}
