// Package totp generates Time-based One-Time Passwords
package totp

import (
	"crypto/hmac"
	"crypto/sha1"
	"encoding/base32"
	"encoding/binary"
	"fmt"
	"strings"
	"time"
)

// Result contains the generated code and metadata
type Result struct {
	Code             string `json:"code"`
	RemainingSeconds int64  `json:"remainingSeconds"`
	ValidUntil       string `json:"validUntil"`
}

// Generate generates a 6-digit TOTP code from a secret key
func Generate(secret string) (string, int64, error) {
	secret = strings.ToUpper(strings.ReplaceAll(secret, " ", ""))
	secret = strings.ReplaceAll(secret, "-", "")

	if m := len(secret) % 8; m != 0 {
		secret += strings.Repeat("=", 8-m)
	}

	key, err := base32.StdEncoding.DecodeString(secret)
	if err != nil {
		return "", 0, fmt.Errorf("invalid secret key: %w", err)
	}

	now := time.Now().Unix()
	counter := now / 30
	remaining := 30 - (now % 30)

	code := hotp(key, counter)

	return fmt.Sprintf("%06d", code), remaining, nil
}

// GenerateWithInfo returns code with additional info
func GenerateWithInfo(secret string) (*Result, error) {
	code, remaining, err := Generate(secret)
	if err != nil {
		return nil, err
	}

	return &Result{
		Code:             code,
		RemainingSeconds: remaining,
		ValidUntil:       time.Now().Add(time.Duration(remaining) * time.Second).Format("15:04:05"),
	}, nil
}

func hotp(key []byte, counter int64) int {
	buf := make([]byte, 8)
	binary.BigEndian.PutUint64(buf, uint64(counter))

	mac := hmac.New(sha1.New, key)
	mac.Write(buf)
	hash := mac.Sum(nil)

	offset := hash[len(hash)-1] & 0x0f
	code := binary.BigEndian.Uint32(hash[offset:offset+4]) & 0x7fffffff

	return int(code % 1000000)
}
