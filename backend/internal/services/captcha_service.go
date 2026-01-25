package services

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// CaptchaService handles CAPTCHA verification
type CaptchaService struct {
	secretKey  string
	siteKey    string
	verifyURL  string
	httpClient *http.Client
	enabled    bool
}

// CaptchaConfig holds captcha configuration
type CaptchaConfig struct {
	SecretKey string
	SiteKey   string
	VerifyURL string
	Enabled   bool
}

// NewCaptchaService creates a new captcha service
func NewCaptchaService(config *CaptchaConfig) *CaptchaService {
	verifyURL := config.VerifyURL
	if verifyURL == "" {
		verifyURL = "https://www.google.com/recaptcha/api/siteverify"
	}

	return &CaptchaService{
		secretKey: config.SecretKey,
		siteKey:   config.SiteKey,
		verifyURL: verifyURL,
		enabled:   config.Enabled,
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

// CaptchaResponse represents Google reCAPTCHA response
type CaptchaResponse struct {
	Success     bool      `json:"success"`
	Score       float64   `json:"score"`
	Action      string    `json:"action"`
	ChallengeTS time.Time `json:"challenge_ts"`
	Hostname    string    `json:"hostname"`
	ErrorCodes  []string  `json:"error-codes"`
}

// VerifyRequest represents verification request
type VerifyRequest struct {
	Token    string
	RemoteIP string
	Action   string
}

// Verify verifies a CAPTCHA token
func (s *CaptchaService) Verify(ctx context.Context, req *VerifyRequest) (*CaptchaResponse, error) {
	if !s.enabled {
		// Return success if captcha is disabled
		return &CaptchaResponse{
			Success: true,
			Score:   1.0,
		}, nil
	}

	if req.Token == "" {
		return nil, errors.New("captcha token is required")
	}

	// Build form data
	formData := url.Values{}
	formData.Set("secret", s.secretKey)
	formData.Set("response", req.Token)
	if req.RemoteIP != "" {
		formData.Set("remoteip", req.RemoteIP)
	}

	// Make request
	httpReq, err := http.NewRequestWithContext(ctx, "POST", s.verifyURL, strings.NewReader(formData.Encode()))
	if err != nil {
		return nil, err
	}
	httpReq.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := s.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to verify captcha: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var result CaptchaResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, err
	}

	// Validate action if provided
	if req.Action != "" && result.Action != req.Action {
		return nil, errors.New("captcha action mismatch")
	}

	return &result, nil
}

// VerifyWithScore verifies captcha and checks score threshold
func (s *CaptchaService) VerifyWithScore(ctx context.Context, token, remoteIP, action string, minScore float64) error {
	result, err := s.Verify(ctx, &VerifyRequest{
		Token:    token,
		RemoteIP: remoteIP,
		Action:   action,
	})
	if err != nil {
		return err
	}

	if !result.Success {
		if len(result.ErrorCodes) > 0 {
			return fmt.Errorf("captcha verification failed: %s", strings.Join(result.ErrorCodes, ", "))
		}
		return errors.New("captcha verification failed")
	}

	if result.Score < minScore {
		return fmt.Errorf("captcha score too low: %.2f < %.2f", result.Score, minScore)
	}

	return nil
}

// IsEnabled returns whether captcha is enabled
func (s *CaptchaService) IsEnabled() bool {
	return s.enabled
}

// GetSiteKey returns the site key for reCAPTCHA (public key)
func (s *CaptchaService) GetSiteKey() string {
	return s.siteKey
}
