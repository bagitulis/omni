package oauth

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"
)

const StateTTL = 10 * time.Minute

type StateClaims struct {
	TenantID     string `json:"tenant_id"`
	Platform     string `json:"platform"`
	AttemptID    string `json:"attempt_id"`
	Intent       string `json:"intent"`
	StoreID      string `json:"store_id,omitempty"`
	UserID       string `json:"user_id,omitempty"`
	SessionID    string `json:"session_id,omitempty"`
	CSRFNonce    string `json:"csrf_nonce"`
	RedirectPath string `json:"redirect_path"`
	ExpiresAt    int64  `json:"expires_at"`
}

func BuildSignedState(claims StateClaims) (string, error) {
	if claims.TenantID == "" || claims.Platform == "" || claims.AttemptID == "" || claims.CSRFNonce == "" || claims.ExpiresAt == 0 {
		return "", fmt.Errorf("state claims are incomplete")
	}
	payload, err := json.Marshal(claims)
	if err != nil {
		return "", fmt.Errorf("marshal state claims: %w", err)
	}
	encodedPayload := base64.RawURLEncoding.EncodeToString(payload)
	signature, err := signState(encodedPayload)
	if err != nil {
		return "", err
	}
	return encodedPayload + "." + signature, nil
}

func ParseSignedState(state string) (StateClaims, error) {
	parts := strings.Split(state, ".")
	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		return StateClaims{}, fmt.Errorf("invalid_state")
	}
	expectedSig, err := signState(parts[0])
	if err != nil {
		return StateClaims{}, fmt.Errorf("invalid_state")
	}
	if !hmac.Equal([]byte(expectedSig), []byte(parts[1])) {
		return StateClaims{}, fmt.Errorf("invalid_state")
	}
	payload, err := base64.RawURLEncoding.DecodeString(parts[0])
	if err != nil {
		return StateClaims{}, fmt.Errorf("invalid_state")
	}
	var claims StateClaims
	if err := json.Unmarshal(payload, &claims); err != nil {
		return StateClaims{}, fmt.Errorf("invalid_state")
	}
	if time.Now().Unix() > claims.ExpiresAt {
		return StateClaims{}, fmt.Errorf("expired_state")
	}
	return claims, nil
}

func NewNonce() (string, error) {
	buf := make([]byte, 24)
	if _, err := rand.Read(buf); err != nil {
		return "", fmt.Errorf("generate nonce: %w", err)
	}
	return base64.RawURLEncoding.EncodeToString(buf), nil
}

func signState(payload string) (string, error) {
	secret, err := stateSecret()
	if err != nil {
		return "", err
	}
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(payload))
	return base64.RawURLEncoding.EncodeToString(mac.Sum(nil)), nil
}

func stateSecret() (string, error) {
	secret := os.Getenv("OAUTH_STATE_SECRET")
	if secret == "" {
		secret = os.Getenv("JWT_SECRET")
	}
	if secret == "" {
		return "", fmt.Errorf("OAUTH_STATE_SECRET or JWT_SECRET must be configured")
	}
	return secret, nil
}
