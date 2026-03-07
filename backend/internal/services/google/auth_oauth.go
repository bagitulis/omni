package google

import (
	"fmt"
	"time"
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
			return nil
		}
	}
	return fmt.Errorf("service account not found: %s", accountID)
}
