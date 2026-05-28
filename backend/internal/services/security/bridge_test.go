package security_test

import (
	"testing"

	"github.com/omni/backend/internal/services/security"
)

// TestRecordAlert_HighSeverity_BridgePush verifies a high severity alert triggers canonical push.
func TestRecordAlert_HighSeverity_BridgePush(t *testing.T) {
	svc := security.NewNotificationService()

	var (
		capturedTenant    string
		capturedNotifType string
		capturedCategory  string
		capturedTitle     string
		capturedMessage   string
	)

	mockPush := func(tenantID, notifType, category, title, message string) error {
		capturedTenant = tenantID
		capturedNotifType = notifType
		capturedCategory = category
		capturedTitle = title
		capturedMessage = message
		return nil
	}

	svc.SetCanonicalPush(mockPush)

	alert := security.SecurityAlert{
		Type:      "unauthorized_access",
		Severity:  "high",
		Message:   "Unauthorized access detected from IP 10.0.0.5",
		TenantID:  "tenant-1",
		IPAddress: "10.0.0.5",
	}

	result := svc.RecordAlert(alert)
	if !result {
		t.Fatal("RecordAlert() returned false for new alert")
	}

	if capturedTenant != "tenant-1" {
		t.Errorf("expected tenant tenant-1, got %s", capturedTenant)
	}
	if capturedNotifType != "warning" {
		t.Errorf("expected notifType warning for high severity, got %s", capturedNotifType)
	}
	if capturedCategory != "security" {
		t.Errorf("expected category security, got %s", capturedCategory)
	}
	if capturedTitle != "Security high: unauthorized_access" {
		t.Errorf("expected 'Security high: unauthorized_access', got '%s'", capturedTitle)
	}
	if capturedMessage != "Unauthorized access detected from IP 10.0.0.5" {
		t.Errorf("unexpected message: %s", capturedMessage)
	}
}

// TestRecordAlert_CriticalSeverity_BridgePush verifies a critical alert triggers with error type.
func TestRecordAlert_CriticalSeverity_BridgePush(t *testing.T) {
	svc := security.NewNotificationService()

	var capturedNotifType string
	mockPush := func(tenantID, notifType, category, title, message string) error {
		capturedNotifType = notifType
		return nil
	}

	svc.SetCanonicalPush(mockPush)

	alert := security.SecurityAlert{
		Type:     "breach",
		Severity: "critical",
		Message:  "Data breach detected",
		TenantID: "tenant-2",
	}

	svc.RecordAlert(alert)

	if capturedNotifType != "error" {
		t.Errorf("expected notifType error for critical severity, got %s", capturedNotifType)
	}

	// Verify security stats remain independent
	stats := svc.GetStats()
	total, ok := stats["total"].(int)
	if !ok {
		t.Fatal("stats[total] is not int")
	}
	if total != 1 {
		t.Errorf("expected total=1 from security stats, got %d", total)
	}
}

// TestRecordAlert_LowSeverity_NoBridge verifies low/medium severity alerts do NOT trigger canonical push.
func TestRecordAlert_LowSeverity_NoBridge(t *testing.T) {
	svc := security.NewNotificationService()

	pushCalled := false
	mockPush := func(tenantID, notifType, category, title, message string) error {
		pushCalled = true
		return nil
	}

	svc.SetCanonicalPush(mockPush)

	// Low severity
	alert := security.SecurityAlert{
		Type:     "info_event",
		Severity: "low",
		Message:  "Something happened",
		TenantID: "tenant-3",
	}

	svc.RecordAlert(alert)

	if pushCalled {
		t.Error("expected canonical push NOT to be called for low severity alert")
	}

	// Medium severity (suspicious activity)
	alert2 := security.SecurityAlert{
		Type:     "suspicious_activity",
		Severity: "medium",
		Message:  "Suspicious login",
		TenantID: "tenant-3",
	}

	svc.RecordAlert(alert2)

	if pushCalled {
		t.Error("expected canonical push NOT to be called for medium severity alert")
	}

	// Verify security stats still count all alerts
	stats := svc.GetStats()
	total, ok := stats["total"].(int)
	if !ok {
		t.Fatal("stats[total] is not int")
	}
	if total != 2 {
		t.Errorf("expected total=2 from security stats (both alerts recorded), got %d", total)
	}
}

// TestRecordAlert_NoCanonicalPush_NoPanic verifies without SetCanonicalPush, no panic occurs.
func TestRecordAlert_NoCanonicalPush_NoPanic(t *testing.T) {
	svc := security.NewNotificationService()

	alert := security.SecurityAlert{
		Type:     "test",
		Severity: "critical",
		Message:  "test",
		TenantID: "tenant-x",
	}

	result := svc.RecordAlert(alert)
	if !result {
		t.Fatal("RecordAlert() should succeed without canonical push")
	}

	stats := svc.GetStats()
	total, ok := stats["total"].(int)
	if !ok {
		t.Fatal("stats[total] is not int")
	}
	if total != 1 {
		t.Errorf("expected total=1, got %d", total)
	}
}
