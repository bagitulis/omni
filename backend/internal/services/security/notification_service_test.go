package security_test

import (
	"testing"
	"time"

	"github.com/omni/backend/internal/services/security"
)

// TestNewNotificationService verifies the constructor initializes state correctly.
func TestNewNotificationService(t *testing.T) {
	svc := security.NewNotificationService()
	if svc == nil {
		t.Fatal("NewNotificationService() returned nil")
	}

	// Fresh service should have zero alerts
	filter := security.AlertFilter{Limit: 100}
	alerts := svc.GetAlerts(filter)
	if len(alerts) != 0 {
		t.Errorf("expected 0 alerts on new service, got %d", len(alerts))
	}

	stats := svc.GetStats()
	total, ok := stats["total"].(int)
	if !ok {
		t.Fatal("stats[total] is not an int")
	}
	if total != 0 {
		t.Errorf("expected stats total=0, got %d", total)
	}
}

// TestRecordAlert_HappyPath verifies a basic alert is stored and returned.
func TestRecordAlert_HappyPath(t *testing.T) {
	svc := security.NewNotificationService()

	alert := security.SecurityAlert{
		Type:      "login_failure",
		Severity:  "medium",
		Message:   "Failed login attempt",
		TenantID:  "tenant-1",
		IPAddress: "192.168.1.1",
	}

	recorded := svc.RecordAlert(alert)
	if !recorded {
		t.Fatal("RecordAlert() returned false for first alert")
	}

	filter := security.AlertFilter{Limit: 10}
	alerts := svc.GetAlerts(filter)
	if len(alerts) != 1 {
		t.Fatalf("expected 1 alert, got %d", len(alerts))
	}

	got := alerts[0]
	if got.Type != "login_failure" {
		t.Errorf("expected type login_failure, got %s", got.Type)
	}
	if got.Severity != "medium" {
		t.Errorf("expected severity medium, got %s", got.Severity)
	}
	if got.Message != "Failed login attempt" {
		t.Errorf("unexpected message: %s", got.Message)
	}
	if got.TenantID != "tenant-1" {
		t.Errorf("expected tenant tenant-1, got %s", got.TenantID)
	}
	if got.IPAddress != "192.168.1.1" {
		t.Errorf("expected IP 192.168.1.1, got %s", got.IPAddress)
	}
	if got.ID == "" {
		t.Error("expected non-empty ID after recording")
	}
	if got.CreatedAt.IsZero() {
		t.Error("expected non-zero CreatedAt")
	}
}

// TestRecordAlert_RateLimiting verifies duplicate alert types are rate-limited.
func TestRecordAlert_RateLimiting(t *testing.T) {
	svc := security.NewNotificationService()

	alert := security.SecurityAlert{
		Type:     "brute_force",
		Severity: "high",
		Message:  "Brute force detected",
		TenantID: "tenant-2",
	}

	// First alert: should succeed
	first := svc.RecordAlert(alert)
	if !first {
		t.Fatal("first RecordAlert() should succeed")
	}

	// Second alert of same type immediately: should be rate-limited
	second := svc.RecordAlert(alert)
	if second {
		t.Error("second RecordAlert() of same type should be rate-limited (return false)")
	}

	// Only 1 alert should be stored
	filter := security.AlertFilter{Limit: 100}
	alerts := svc.GetAlerts(filter)
	if len(alerts) != 1 {
		t.Errorf("expected 1 alert due to rate limiting, got %d", len(alerts))
	}
}

// TestRecordAlert_DifferentTypes verifies different alert types are NOT rate-limited by each other.
func TestRecordAlert_DifferentTypes(t *testing.T) {
	svc := security.NewNotificationService()

	types := []string{"login_failure", "brute_force", "suspicious_ip", "token_abuse"}
	for _, alertType := range types {
		a := security.SecurityAlert{
			Type:     alertType,
			Severity: "low",
			Message:  "test " + alertType,
		}
		recorded := svc.RecordAlert(a)
		if !recorded {
			t.Errorf("expected alert type %s to be recorded (different type), got false", alertType)
		}
	}

	filter := security.AlertFilter{Limit: 100}
	alerts := svc.GetAlerts(filter)
	if len(alerts) != len(types) {
		t.Errorf("expected %d alerts, got %d", len(types), len(alerts))
	}
}

// TestGetAlerts_FilterByType verifies type-based filtering.
func TestGetAlerts_FilterByType(t *testing.T) {
	svc := security.NewNotificationService()

	for i := 0; i < 3; i++ {
		svc.RecordAlert(security.SecurityAlert{Type: "login_failure", Severity: "low", Message: "fail"})
		// Bypass rate limiter by sleeping is not practical; use different types in a real test
	}

	// Add alerts with distinct types to avoid rate limiting
	svc.RecordAlert(security.SecurityAlert{Type: "type_a", Severity: "low", Message: "a"})
	svc.RecordAlert(security.SecurityAlert{Type: "type_b", Severity: "high", Message: "b"})
	svc.RecordAlert(security.SecurityAlert{Type: "type_c", Severity: "critical", Message: "c"})

	filter := security.AlertFilter{Type: "type_b", Limit: 10}
	alerts := svc.GetAlerts(filter)
	if len(alerts) != 1 {
		t.Fatalf("expected 1 alert with type type_b, got %d", len(alerts))
	}
	if alerts[0].Type != "type_b" {
		t.Errorf("expected type type_b, got %s", alerts[0].Type)
	}
}

// TestGetAlerts_FilterBySeverity verifies severity-based filtering.
func TestGetAlerts_FilterBySeverity(t *testing.T) {
	svc := security.NewNotificationService()

	svc.RecordAlert(security.SecurityAlert{Type: "evt_low", Severity: "low", Message: "low event"})
	svc.RecordAlert(security.SecurityAlert{Type: "evt_critical", Severity: "critical", Message: "critical event"})
	svc.RecordAlert(security.SecurityAlert{Type: "evt_high", Severity: "high", Message: "high event"})

	filter := security.AlertFilter{Severity: "critical", Limit: 10}
	alerts := svc.GetAlerts(filter)
	if len(alerts) != 1 {
		t.Fatalf("expected 1 critical alert, got %d", len(alerts))
	}
	if alerts[0].Severity != "critical" {
		t.Errorf("expected severity critical, got %s", alerts[0].Severity)
	}
}

// TestGetAlerts_FilterByTenant verifies tenant-based filtering.
func TestGetAlerts_FilterByTenant(t *testing.T) {
	svc := security.NewNotificationService()

	svc.RecordAlert(security.SecurityAlert{Type: "evt_t1", Severity: "low", Message: "t1 msg", TenantID: "tenant-a"})
	svc.RecordAlert(security.SecurityAlert{Type: "evt_t2", Severity: "low", Message: "t2 msg", TenantID: "tenant-b"})
	svc.RecordAlert(security.SecurityAlert{Type: "evt_t3", Severity: "low", Message: "t3 msg", TenantID: "tenant-a"})

	filter := security.AlertFilter{TenantID: "tenant-a", Limit: 10}
	alerts := svc.GetAlerts(filter)
	if len(alerts) != 2 {
		t.Fatalf("expected 2 alerts for tenant-a, got %d", len(alerts))
	}
	for _, a := range alerts {
		if a.TenantID != "tenant-a" {
			t.Errorf("unexpected tenant in result: %s", a.TenantID)
		}
	}
}

// TestGetAlerts_FilterBySince verifies time-based filtering.
func TestGetAlerts_FilterBySince(t *testing.T) {
	svc := security.NewNotificationService()
	// Record an old alert, sleep to ensure its CreatedAt is strictly before cutoff
	svc.RecordAlert(security.SecurityAlert{Type: "old_evt", Severity: "low", Message: "old"})
	time.Sleep(20 * time.Millisecond) // ensure old_evt.CreatedAt is strictly before cutoff
	cutoff := time.Now()
	time.Sleep(20 * time.Millisecond) // ensure new_evt.CreatedAt is strictly after cutoff
	svc.RecordAlert(security.SecurityAlert{Type: "new_evt", Severity: "low", Message: "new"})
	filter := security.AlertFilter{Since: cutoff, Limit: 10}
	alerts := svc.GetAlerts(filter)
	if len(alerts) != 1 {
		t.Fatalf("expected 1 alert since cutoff, got %d", len(alerts))
	}
	if alerts[0].Type != "new_evt" {
		t.Errorf("expected new_evt, got %s", alerts[0].Type)
	}
}

// TestGetAlerts_LimitRespected verifies the Limit field caps results.
func TestGetAlerts_LimitRespected(t *testing.T) {
	svc := security.NewNotificationService()

	// Use distinct types to avoid rate limiting
	for i := 0; i < 10; i++ {
		alertType := "type_limit_" + string(rune('a'+i))
		svc.RecordAlert(security.SecurityAlert{Type: alertType, Severity: "low", Message: "msg"})
	}

	filter := security.AlertFilter{Limit: 5}
	alerts := svc.GetAlerts(filter)
	if len(alerts) != 5 {
		t.Errorf("expected 5 alerts with Limit=5, got %d", len(alerts))
	}
}

// TestGetAlerts_ZeroLimit verifies zero Limit returns no results.
func TestGetAlerts_ZeroLimit(t *testing.T) {
	svc := security.NewNotificationService()
	svc.RecordAlert(security.SecurityAlert{Type: "some_type", Severity: "low", Message: "msg"})

	filter := security.AlertFilter{Limit: 0}
	alerts := svc.GetAlerts(filter)
	if len(alerts) != 0 {
		t.Errorf("expected 0 alerts with Limit=0, got %d", len(alerts))
	}
}

// TestGetAlerts_MostRecentFirst verifies alerts are returned newest-first.
func TestGetAlerts_MostRecentFirst(t *testing.T) {
	svc := security.NewNotificationService()

	svc.RecordAlert(security.SecurityAlert{Type: "first_recorded", Severity: "low", Message: "first"})
	time.Sleep(5 * time.Millisecond)
	svc.RecordAlert(security.SecurityAlert{Type: "second_recorded", Severity: "low", Message: "second"})

	filter := security.AlertFilter{Limit: 10}
	alerts := svc.GetAlerts(filter)
	if len(alerts) < 2 {
		t.Fatalf("expected at least 2 alerts, got %d", len(alerts))
	}
	// Newest should be first
	if alerts[0].Type != "second_recorded" {
		t.Errorf("expected newest alert first (second_recorded), got %s", alerts[0].Type)
	}
}

// TestMaxAlerts_BoundaryEnforced verifies that alerts beyond maxAlerts (1000) evict the oldest.
func TestMaxAlerts_BoundaryEnforced(t *testing.T) {
	svc := security.NewNotificationService()

	// Insert 1001 alerts with distinct types using a loop
	for i := 0; i < 1001; i++ {
		alertType := "flood_type_" + intToStr(i)
		svc.RecordAlert(security.SecurityAlert{Type: alertType, Severity: "low", Message: "flood"})
	}

	// Total stored should not exceed 1000
	filter := security.AlertFilter{Limit: 2000}
	alerts := svc.GetAlerts(filter)
	if len(alerts) > 1000 {
		t.Errorf("expected at most 1000 alerts, got %d", len(alerts))
	}

	// The very first alert (type flood_type_0) should have been evicted
	filter2 := security.AlertFilter{Type: "flood_type_0", Limit: 10}
	evicted := svc.GetAlerts(filter2)
	if len(evicted) != 0 {
		t.Error("expected flood_type_0 to be evicted (oldest), but it's still present")
	}
}

// TestTrackSuspiciousActivity verifies it stores a suspicious_activity alert.
func TestTrackSuspiciousActivity(t *testing.T) {
	svc := security.NewNotificationService()

	svc.TrackSuspiciousActivity("tenant-x", "10.0.0.1", "rapid_requests", "Too many requests")

	filter := security.AlertFilter{Type: "suspicious_activity", Limit: 10}
	alerts := svc.GetAlerts(filter)
	if len(alerts) != 1 {
		t.Fatalf("expected 1 suspicious_activity alert, got %d", len(alerts))
	}

	a := alerts[0]
	if a.Severity != "medium" {
		t.Errorf("expected medium severity, got %s", a.Severity)
	}
	if a.TenantID != "tenant-x" {
		t.Errorf("expected tenant-x, got %s", a.TenantID)
	}
	if a.IPAddress != "10.0.0.1" {
		t.Errorf("expected IP 10.0.0.1, got %s", a.IPAddress)
	}
	if a.Message != "Too many requests" {
		t.Errorf("unexpected message: %s", a.Message)
	}
	if a.Details == nil {
		t.Fatal("expected details map to be non-nil")
	}
	if a.Details["activity_type"] != "rapid_requests" {
		t.Errorf("expected activity_type=rapid_requests in details, got %v", a.Details["activity_type"])
	}
}

// TestAlertCritical verifies it stores a critical severity alert.
func TestAlertCritical(t *testing.T) {
	svc := security.NewNotificationService()

	details := map[string]interface{}{
		"source": "api_gateway",
	}
	svc.AlertCritical("tenant-y", "Unauthorized access attempt", details)

	filter := security.AlertFilter{Type: "critical", Severity: "critical", Limit: 10}
	alerts := svc.GetAlerts(filter)
	if len(alerts) != 1 {
		t.Fatalf("expected 1 critical alert, got %d", len(alerts))
	}
	if alerts[0].Message != "Unauthorized access attempt" {
		t.Errorf("unexpected message: %s", alerts[0].Message)
	}
	if alerts[0].TenantID != "tenant-y" {
		t.Errorf("expected tenant-y, got %s", alerts[0].TenantID)
	}
}

// TestGetStats verifies stats aggregation by severity and type.
func TestGetStats(t *testing.T) {
	svc := security.NewNotificationService()

	svc.RecordAlert(security.SecurityAlert{Type: "type_x", Severity: "high", Message: "a"})
	svc.RecordAlert(security.SecurityAlert{Type: "type_y", Severity: "low", Message: "b"})
	svc.RecordAlert(security.SecurityAlert{Type: "type_z", Severity: "high", Message: "c"})

	stats := svc.GetStats()

	total, ok := stats["total"].(int)
	if !ok {
		t.Fatal("stats[total] is not int")
	}
	if total != 3 {
		t.Errorf("expected total=3, got %d", total)
	}

	bySev, ok := stats["by_severity"].(map[string]int)
	if !ok {
		t.Fatal("stats[by_severity] is not map[string]int")
	}
	if bySev["high"] != 2 {
		t.Errorf("expected high=2, got %d", bySev["high"])
	}
	if bySev["low"] != 1 {
		t.Errorf("expected low=1, got %d", bySev["low"])
	}

	byType, ok := stats["by_type"].(map[string]int)
	if !ok {
		t.Fatal("stats[by_type] is not map[string]int")
	}
	if byType["type_x"] != 1 || byType["type_y"] != 1 || byType["type_z"] != 1 {
		t.Errorf("unexpected by_type counts: %v", byType)
	}
}

// TestSecurityAlertStruct verifies the SecurityAlert struct fields.
func TestSecurityAlertStruct(t *testing.T) {
	now := time.Now()
	a := security.SecurityAlert{
		ID:        "alert-001",
		Type:      "test_type",
		Severity:  "critical",
		Message:   "test message",
		TenantID:  "tenant-1",
		IPAddress: "127.0.0.1",
		CreatedAt: now,
		Details:   map[string]interface{}{"key": "value"},
	}

	if a.ID != "alert-001" {
		t.Errorf("ID mismatch: %s", a.ID)
	}
	if a.Type != "test_type" {
		t.Errorf("Type mismatch: %s", a.Type)
	}
	if a.Severity != "critical" {
		t.Errorf("Severity mismatch: %s", a.Severity)
	}
	if a.TenantID != "tenant-1" {
		t.Errorf("TenantID mismatch: %s", a.TenantID)
	}
	if a.IPAddress != "127.0.0.1" {
		t.Errorf("IPAddress mismatch: %s", a.IPAddress)
	}
	if !a.CreatedAt.Equal(now) {
		t.Errorf("CreatedAt mismatch")
	}
	if a.Details["key"] != "value" {
		t.Errorf("Details mismatch")
	}
}

// TestAlertFilterStruct verifies AlertFilter struct fields are assignable.
func TestAlertFilterStruct(t *testing.T) {
	since := time.Now()
	f := security.AlertFilter{
		Type:     "login",
		Severity: "high",
		TenantID: "t1",
		Since:    since,
		Limit:    50,
	}

	if f.Type != "login" {
		t.Errorf("Type mismatch: %s", f.Type)
	}
	if f.Severity != "high" {
		t.Errorf("Severity mismatch: %s", f.Severity)
	}
	if f.TenantID != "t1" {
		t.Errorf("TenantID mismatch: %s", f.TenantID)
	}
	if !f.Since.Equal(since) {
		t.Error("Since mismatch")
	}
	if f.Limit != 50 {
		t.Errorf("Limit mismatch: %d", f.Limit)
	}
}

// intToStr converts int to string without importing strconv (avoids circular issues).
func intToStr(n int) string {
	if n == 0 {
		return "0"
	}
	neg := false
	if n < 0 {
		neg = true
		n = -n
	}
	buf := make([]byte, 0, 20)
	for n > 0 {
		buf = append([]byte{byte('0' + n%10)}, buf...)
		n /= 10
	}
	if neg {
		buf = append([]byte{'-'}, buf...)
	}
	return string(buf)
}
