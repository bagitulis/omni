package services

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewCaptchaService(t *testing.T) {
	config := &CaptchaConfig{
		SecretKey: "test-secret",
		SiteKey:   "test-site-key",
		Enabled:   true,
	}
	svc := NewCaptchaService(config)

	assert.NotNil(t, svc)
	assert.Equal(t, "test-secret", svc.secretKey)
	assert.Equal(t, "test-site-key", svc.siteKey)
	assert.True(t, svc.enabled)
	assert.Equal(t, "https://www.google.com/recaptcha/api/siteverify", svc.verifyURL)
}

func TestNewCaptchaService_CustomVerifyURL(t *testing.T) {
	config := &CaptchaConfig{
		SecretKey: "test-secret",
		SiteKey:   "test-site-key",
		VerifyURL: "https://custom.verify.url",
		Enabled:   true,
	}
	svc := NewCaptchaService(config)

	assert.Equal(t, "https://custom.verify.url", svc.verifyURL)
}

func TestCaptchaService_IsEnabled(t *testing.T) {
	tests := []struct {
		name    string
		enabled bool
	}{
		{"enabled", true},
		{"disabled", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := &CaptchaConfig{Enabled: tt.enabled}
			svc := NewCaptchaService(config)
			assert.Equal(t, tt.enabled, svc.IsEnabled())
		})
	}
}

func TestCaptchaService_GetSiteKey(t *testing.T) {
	config := &CaptchaConfig{
		SiteKey: "6LdXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXX",
		Enabled: true,
	}
	svc := NewCaptchaService(config)

	assert.Equal(t, "6LdXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXX", svc.GetSiteKey())
}

func TestCaptchaService_Verify_Disabled(t *testing.T) {
	config := &CaptchaConfig{
		Enabled: false,
	}
	svc := NewCaptchaService(config)

	result, err := svc.Verify(context.Background(), &VerifyRequest{Token: ""})

	require.NoError(t, err)
	assert.True(t, result.Success)
	assert.Equal(t, 1.0, result.Score)
}

func TestCaptchaService_Verify_EmptyToken(t *testing.T) {
	config := &CaptchaConfig{
		SecretKey: "test-secret",
		Enabled:   true,
	}
	svc := NewCaptchaService(config)

	_, err := svc.Verify(context.Background(), &VerifyRequest{Token: ""})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "captcha token is required")
}

func TestCaptchaService_Verify_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "POST", r.Method)
		assert.Equal(t, "application/x-www-form-urlencoded", r.Header.Get("Content-Type"))

		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"success": true, "score": 0.9, "action": "login", "hostname": "example.com"}`))
	}))
	defer server.Close()

	config := &CaptchaConfig{
		SecretKey: "test-secret",
		VerifyURL: server.URL,
		Enabled:   true,
	}
	svc := NewCaptchaService(config)

	result, err := svc.Verify(context.Background(), &VerifyRequest{
		Token:    "test-token",
		RemoteIP: "127.0.0.1",
		Action:   "login",
	})

	require.NoError(t, err)
	assert.True(t, result.Success)
	assert.Equal(t, 0.9, result.Score)
	assert.Equal(t, "login", result.Action)
}

func TestCaptchaService_Verify_ActionMismatch(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"success": true, "score": 0.9, "action": "register"}`))
	}))
	defer server.Close()

	config := &CaptchaConfig{
		SecretKey: "test-secret",
		VerifyURL: server.URL,
		Enabled:   true,
	}
	svc := NewCaptchaService(config)

	_, err := svc.Verify(context.Background(), &VerifyRequest{
		Token:  "test-token",
		Action: "login",
	})

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "captcha action mismatch")
}

func TestCaptchaService_VerifyWithScore_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"success": true, "score": 0.9, "action": "login"}`))
	}))
	defer server.Close()

	config := &CaptchaConfig{
		SecretKey: "test-secret",
		VerifyURL: server.URL,
		Enabled:   true,
	}
	svc := NewCaptchaService(config)

	err := svc.VerifyWithScore(context.Background(), "test-token", "127.0.0.1", "login", 0.5)
	assert.NoError(t, err)
}

func TestCaptchaService_VerifyWithScore_ScoreTooLow(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"success": true, "score": 0.3}`))
	}))
	defer server.Close()

	config := &CaptchaConfig{
		SecretKey: "test-secret",
		VerifyURL: server.URL,
		Enabled:   true,
	}
	svc := NewCaptchaService(config)

	err := svc.VerifyWithScore(context.Background(), "test-token", "127.0.0.1", "", 0.5)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "captcha score too low")
}

func TestCaptchaService_VerifyWithScore_VerificationFailed(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"success": false, "error-codes": ["invalid-input-secret"]}`))
	}))
	defer server.Close()

	config := &CaptchaConfig{
		SecretKey: "test-secret",
		VerifyURL: server.URL,
		Enabled:   true,
	}
	svc := NewCaptchaService(config)

	err := svc.VerifyWithScore(context.Background(), "test-token", "127.0.0.1", "", 0.5)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "captcha verification failed")
}

func TestCaptchaService_VerifyWithScore_VerificationFailedNoErrorCodes(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"success": false}`))
	}))
	defer server.Close()

	config := &CaptchaConfig{
		SecretKey: "test-secret",
		VerifyURL: server.URL,
		Enabled:   true,
	}
	svc := NewCaptchaService(config)

	err := svc.VerifyWithScore(context.Background(), "test-token", "127.0.0.1", "", 0.5)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "captcha verification failed")
}

func TestCaptchaConfig(t *testing.T) {
	config := CaptchaConfig{
		SecretKey: "secret",
		SiteKey:   "site",
		VerifyURL: "https://verify.url",
		Enabled:   true,
	}

	assert.Equal(t, "secret", config.SecretKey)
	assert.Equal(t, "site", config.SiteKey)
	assert.Equal(t, "https://verify.url", config.VerifyURL)
	assert.True(t, config.Enabled)
}

func TestCaptchaResponse(t *testing.T) {
	resp := CaptchaResponse{
		Success:    true,
		Score:      0.9,
		Action:     "login",
		Hostname:   "example.com",
		ErrorCodes: nil,
	}

	assert.True(t, resp.Success)
	assert.Equal(t, 0.9, resp.Score)
	assert.Equal(t, "login", resp.Action)
	assert.Equal(t, "example.com", resp.Hostname)
}

func TestVerifyRequest(t *testing.T) {
	req := VerifyRequest{
		Token:    "test-token",
		RemoteIP: "127.0.0.1",
		Action:   "login",
	}

	assert.Equal(t, "test-token", req.Token)
	assert.Equal(t, "127.0.0.1", req.RemoteIP)
	assert.Equal(t, "login", req.Action)
}
