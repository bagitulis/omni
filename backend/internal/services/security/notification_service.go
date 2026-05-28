package security

import (
	"crypto/rand"
	"fmt"
	"math/big"
	"sync"
	"time"

	"github.com/rs/zerolog/log"
)

// NotificationService handles security notifications and alerts
type NotificationService struct {
	mu          sync.RWMutex
	alerts      []SecurityAlert
	maxAlerts   int
	rateLimiter map[string]time.Time // alertType -> lastSent
	minInterval time.Duration

	// canonicalPush, if set, routes high/critical alerts to the main notification feed
	canonicalPush func(tenantID, notifType, category, title, message string) error
}

// NewNotificationService creates a new notification service
func NewNotificationService() *NotificationService {
	return &NotificationService{
		alerts:      make([]SecurityAlert, 0),
		maxAlerts:   1000,
		rateLimiter: make(map[string]time.Time),
		minInterval: 5 * time.Minute,
	}
}

// SetCanonicalPush sets a callback that routes high/critical security alerts
// into the main persisted notification feed via the canonical service.
// This is a one-way bridge: security stats remain independent.
func (s *NotificationService) SetCanonicalPush(fn func(tenantID, notifType, category, title, message string) error) {
	s.canonicalPush = fn
}

// SecurityAlert represents a security alert
type SecurityAlert struct {
	ID        string                 `json:"id"`
	Type      string                 `json:"type"`
	Severity  string                 `json:"severity"` // low, medium, high, critical
	Message   string                 `json:"message"`
	Details   map[string]interface{} `json:"details,omitempty"`
	TenantID  string                 `json:"tenant_id,omitempty"`
	IPAddress string                 `json:"ip_address,omitempty"`
	CreatedAt time.Time              `json:"created_at"`
}

// RecordAlert records a security alert
func (s *NotificationService) RecordAlert(alert SecurityAlert) bool {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Rate limit check
	if lastSent, ok := s.rateLimiter[alert.Type]; ok {
		if time.Since(lastSent) < s.minInterval {
			return false // Rate limited
		}
	}

	// Generate ID
	alert.ID = generateAlertID()
	alert.CreatedAt = time.Now()

	// Add to alerts
	s.alerts = append(s.alerts, alert)
	if len(s.alerts) > s.maxAlerts {
		s.alerts = s.alerts[1:] // Remove oldest
	}

	// Update rate limiter
	s.rateLimiter[alert.Type] = time.Now()

	// Bridge high/critical security alerts to main notification feed
	// LOW/MEDIUM severity alerts do NOT create persisted notifications
	if s.canonicalPush != nil && (alert.Severity == "high" || alert.Severity == "critical") {
		notifType := "warning"
		if alert.Severity == "critical" {
			notifType = "error"
		}
		title := fmt.Sprintf("Security %s: %s", alert.Severity, alert.Type)
		if err := s.canonicalPush(alert.TenantID, notifType, "security", title, alert.Message); err != nil {
			log.Error().Err(err).
				Str("severity", alert.Severity).
				Str("alert_type", alert.Type).
				Msg("Failed to push security alert to main notification feed")
		}
	}

	return true
}

// GetAlerts retrieves alerts with optional filters
func (s *NotificationService) GetAlerts(filter AlertFilter) []SecurityAlert {
	s.mu.RLock()
	defer s.mu.RUnlock()

	result := make([]SecurityAlert, 0)
	for i := len(s.alerts) - 1; i >= 0 && len(result) < filter.Limit; i-- {
		alert := s.alerts[i]

		if filter.Type != "" && alert.Type != filter.Type {
			continue
		}
		if filter.Severity != "" && alert.Severity != filter.Severity {
			continue
		}
		if filter.TenantID != "" && alert.TenantID != filter.TenantID {
			continue
		}
		if !filter.Since.IsZero() && alert.CreatedAt.Before(filter.Since) {
			continue
		}

		result = append(result, alert)
	}

	return result
}

// AlertFilter represents alert query filters
type AlertFilter struct {
	Type     string
	Severity string
	TenantID string
	Since    time.Time
	Limit    int
}

// TrackSuspiciousActivity tracks suspicious activity
func (s *NotificationService) TrackSuspiciousActivity(tenantID, ipAddress, activityType, description string) {
	alert := SecurityAlert{
		Type:      "suspicious_activity",
		Severity:  "medium",
		Message:   description,
		TenantID:  tenantID,
		IPAddress: ipAddress,
		Details: map[string]interface{}{
			"activity_type": activityType,
		},
	}
	s.RecordAlert(alert)
}

// AlertCritical records a critical security alert
func (s *NotificationService) AlertCritical(tenantID, message string, details map[string]interface{}) {
	alert := SecurityAlert{
		Type:     "critical",
		Severity: "critical",
		Message:  message,
		TenantID: tenantID,
		Details:  details,
	}
	s.RecordAlert(alert)
}

// GetStats returns alert statistics
func (s *NotificationService) GetStats() map[string]interface{} {
	s.mu.RLock()
	defer s.mu.RUnlock()

	severityCounts := make(map[string]int)
	typeCounts := make(map[string]int)

	for _, alert := range s.alerts {
		severityCounts[alert.Severity]++
		typeCounts[alert.Type]++
	}

	return map[string]interface{}{
		"total":       len(s.alerts),
		"by_severity": severityCounts,
		"by_type":     typeCounts,
	}
}

func generateAlertID() string {
	return time.Now().Format("20060102150405") + "-" + randomString(8)
}

func randomString(n int) string {
	const letters = "abcdefghijklmnopqrstuvwxyz0123456789"
	b := make([]byte, n)
	for i := range b {
		idx, err := rand.Int(rand.Reader, big.NewInt(int64(len(letters))))
		if err != nil {
			b[i] = letters[0] // fallback, should never happen
			continue
		}
		b[i] = letters[idx.Int64()]
	}
	return string(b)
}
