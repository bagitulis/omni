package google

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"sync"
	"time"

	"golang.org/x/oauth2"
)

// AuthStatus represents the current authentication status
type AuthStatus struct {
	IsAuthenticated bool       `json:"is_authenticated"`
	ServiceAccount  string     `json:"service_account,omitempty"`
	Email           string     `json:"email,omitempty"`
	Scopes          []string   `json:"scopes,omitempty"`
	ExpiresAt       *time.Time `json:"expires_at,omitempty"`
}

// ServiceAccount represents a Google service account
type ServiceAccount struct {
	ID          string `json:"id"`
	Email       string `json:"email"`
	Description string `json:"description,omitempty"`
	IsActive    bool   `json:"is_active"`
}

// ServiceAccountStats represents statistics for service accounts
type ServiceAccountStats struct {
	Accounts       []ServiceAccount `json:"accounts"`
	ActiveAccount  string           `json:"active_account"`
	TotalQuotaUsed int              `json:"total_quota_used"`
	QuotaLimit     int              `json:"quota_limit"`
}

// OAuthStateStore stores OAuth states for validation
type OAuthStateStore struct {
	mu     sync.RWMutex
	states map[string]oauthState
}

type oauthState struct {
	tenantID  string
	createdAt time.Time
}

var stateStore = &OAuthStateStore{
	states: make(map[string]oauthState),
}

// GetStatus returns the current authentication status
func (s *AuthService) GetStatus() *AuthStatus {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if len(s.credentials) == 0 {
		return &AuthStatus{IsAuthenticated: false}
	}

	return &AuthStatus{
		IsAuthenticated: true,
		ServiceAccount:  "default",
		Scopes:          []string{"https://www.googleapis.com/auth/spreadsheets"},
	}
}

// GenerateOAuthURL generates OAuth authorization URL
func (s *AuthService) GenerateOAuthURL(redirectURI, tenantID string) (string, string, error) {
	// Generate random state
	stateBytes := make([]byte, 16)
	if _, err := rand.Read(stateBytes); err != nil {
		return "", "", fmt.Errorf("generate state: %w", err)
	}
	state := hex.EncodeToString(stateBytes)

	// Store state with tenant ID
	stateStore.mu.Lock()
	stateStore.states[state] = oauthState{
		tenantID:  tenantID,
		createdAt: time.Now(),
	}
	stateStore.mu.Unlock()

	// Build OAuth URL (placeholder - would use actual OAuth config)
	url := fmt.Sprintf(
		"https://accounts.google.com/o/oauth2/v2/auth?client_id=YOUR_CLIENT_ID&redirect_uri=%s&response_type=code&scope=https://www.googleapis.com/auth/spreadsheets&state=%s",
		redirectURI,
		state,
	)

	return url, state, nil
}

// ValidateOAuthState validates and returns the tenant ID for a given state
func (s *AuthService) ValidateOAuthState(state string) (string, error) {
	stateStore.mu.Lock()
	defer stateStore.mu.Unlock()

	stored, ok := stateStore.states[state]
	if !ok {
		return "", fmt.Errorf("invalid or expired state")
	}

	// Check expiration (10 minutes)
	if time.Since(stored.createdAt) > 10*time.Minute {
		delete(stateStore.states, state)
		return "", fmt.Errorf("state expired")
	}

	// Clean up state
	delete(stateStore.states, state)
	return stored.tenantID, nil
}

// ExchangeCode exchanges authorization code for OAuth token
func (s *AuthService) ExchangeCode(ctx context.Context, code, tenantID string) (*oauth2.Token, error) {
	// Placeholder implementation - would use actual OAuth config
	// In production, this would exchange the code for a real token
	token := &oauth2.Token{
		AccessToken:  "placeholder_access_token",
		TokenType:    "Bearer",
		RefreshToken: "placeholder_refresh_token",
		Expiry:       time.Now().Add(time.Hour),
	}
	return token, nil
}

// RevokeAccess revokes Google access for a tenant
func (s *AuthService) RevokeAccess(tenantID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Remove cached clients
	delete(s.clients, tenantID)
	delete(s.driveClients, tenantID)

	return nil
}

// ListServiceAccounts returns list of available service accounts
func (s *AuthService) ListServiceAccounts() []ServiceAccount {
	// Return default service account
	return []ServiceAccount{
		{
			ID:          "default",
			Email:       "service@project.iam.gserviceaccount.com",
			Description: "Default service account",
			IsActive:    true,
		},
	}
}

// GetServiceAccountStats returns statistics for all service accounts
func (s *AuthService) GetServiceAccountStats() *ServiceAccountStats {
	return &ServiceAccountStats{
		Accounts:       s.ListServiceAccounts(),
		ActiveAccount:  "default",
		TotalQuotaUsed: 0,
		QuotaLimit:     60,
	}
}

// SwitchServiceAccount switches to a different service account
func (s *AuthService) SwitchServiceAccount(accountID string) error {
	accounts := s.ListServiceAccounts()
	for _, acc := range accounts {
		if acc.ID == accountID {
			// In production, this would update the active service account
			return nil
		}
	}
	return fmt.Errorf("service account not found: %s", accountID)
}
