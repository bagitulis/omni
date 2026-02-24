package quota_test

import (
	"testing"

	"github.com/omni/backend/internal/services/quota"
)

// ---------------------------------------------------------------------------
// QuotaType constants
// ---------------------------------------------------------------------------

func TestQuotaTypeConstants(t *testing.T) {
	if quota.QuotaAPICall != "api_call" {
		t.Errorf("expected QuotaAPICall='api_call', got %q", quota.QuotaAPICall)
	}
	if quota.QuotaOrderSync != "order_sync" {
		t.Errorf("expected QuotaOrderSync='order_sync', got %q", quota.QuotaOrderSync)
	}
	if quota.QuotaProductSync != "product_sync" {
		t.Errorf("expected QuotaProductSync='product_sync', got %q", quota.QuotaProductSync)
	}
	if quota.QuotaShipment != "shipment" {
		t.Errorf("expected QuotaShipment='shipment', got %q", quota.QuotaShipment)
	}
}

// ---------------------------------------------------------------------------
// QuotaConfig TableName
// ---------------------------------------------------------------------------

func TestQuotaConfig_TableName(t *testing.T) {
	cfg := quota.QuotaConfig{}
	got := cfg.TableName()
	if got != "quota_configs" {
		t.Errorf("expected TableName='quota_configs', got %q", got)
	}
}

// ---------------------------------------------------------------------------
// QuotaUsage TableName
// ---------------------------------------------------------------------------

func TestQuotaUsage_TableName(t *testing.T) {
	u := quota.QuotaUsage{}
	got := u.TableName()
	if got != "quota_usage" {
		t.Errorf("expected TableName='quota_usage', got %q", got)
	}
}

// ---------------------------------------------------------------------------
// NewQuotaManagementService constructor
// ---------------------------------------------------------------------------

func TestNewQuotaManagementService_NotNil(t *testing.T) {
	svc := quota.NewQuotaManagementService(nil)
	if svc == nil {
		t.Fatal("NewQuotaManagementService returned nil")
	}
}

func TestNewQuotaManagementService_NilDB_NoHang(t *testing.T) {
	// Constructor with nil db should not panic
	svc := quota.NewQuotaManagementService(nil)
	if svc == nil {
		t.Fatal("expected non-nil service even with nil db")
	}
}

// ---------------------------------------------------------------------------
// QuotaStatus struct field assignment
// ---------------------------------------------------------------------------

func TestQuotaStatus_FieldAssignment(t *testing.T) {
	s := quota.QuotaStatus{
		QuotaType:       quota.QuotaAPICall,
		DailyUsage:      50,
		DailyLimit:      100,
		DailyRemaining:  50,
		HourlyUsage:     5,
		HourlyLimit:     20,
		HourlyRemaining: 15,
		IsExhausted:     false,
	}

	if s.QuotaType != quota.QuotaAPICall {
		t.Errorf("expected QuotaAPICall, got %q", s.QuotaType)
	}
	if s.DailyUsage != 50 {
		t.Errorf("expected DailyUsage=50, got %d", s.DailyUsage)
	}
	if s.DailyLimit != 100 {
		t.Errorf("expected DailyLimit=100, got %d", s.DailyLimit)
	}
	if s.DailyRemaining != 50 {
		t.Errorf("expected DailyRemaining=50, got %d", s.DailyRemaining)
	}
	if s.HourlyUsage != 5 {
		t.Errorf("expected HourlyUsage=5, got %d", s.HourlyUsage)
	}
	if s.HourlyLimit != 20 {
		t.Errorf("expected HourlyLimit=20, got %d", s.HourlyLimit)
	}
	if s.HourlyRemaining != 15 {
		t.Errorf("expected HourlyRemaining=15, got %d", s.HourlyRemaining)
	}
	if s.IsExhausted {
		t.Error("expected IsExhausted=false")
	}
}

func TestQuotaStatus_IsExhausted_True(t *testing.T) {
	s := quota.QuotaStatus{
		QuotaType:   quota.QuotaOrderSync,
		DailyUsage:  100,
		DailyLimit:  100,
		IsExhausted: true,
	}
	if !s.IsExhausted {
		t.Error("expected IsExhausted=true when usage equals limit")
	}
}

// ---------------------------------------------------------------------------
// QuotaConfig struct field assignment
// ---------------------------------------------------------------------------

func TestQuotaConfig_FieldAssignment(t *testing.T) {
	cfg := quota.QuotaConfig{
		TenantID:    "tenant-1",
		Platform:    "shopee",
		QuotaType:   quota.QuotaProductSync,
		DailyLimit:  1000,
		HourlyLimit: 200,
		MinuteLimit: 50,
		IsEnabled:   true,
	}

	if cfg.TenantID != "tenant-1" {
		t.Errorf("expected TenantID='tenant-1', got %q", cfg.TenantID)
	}
	if cfg.Platform != "shopee" {
		t.Errorf("expected Platform='shopee', got %q", cfg.Platform)
	}
	if cfg.QuotaType != quota.QuotaProductSync {
		t.Errorf("expected QuotaProductSync, got %q", cfg.QuotaType)
	}
	if cfg.DailyLimit != 1000 {
		t.Errorf("expected DailyLimit=1000, got %d", cfg.DailyLimit)
	}
	if cfg.HourlyLimit != 200 {
		t.Errorf("expected HourlyLimit=200, got %d", cfg.HourlyLimit)
	}
	if cfg.MinuteLimit != 50 {
		t.Errorf("expected MinuteLimit=50, got %d", cfg.MinuteLimit)
	}
	if !cfg.IsEnabled {
		t.Error("expected IsEnabled=true")
	}
}

// ---------------------------------------------------------------------------
// QuotaUsage struct field assignment
// ---------------------------------------------------------------------------

func TestQuotaUsage_FieldAssignment(t *testing.T) {
	u := quota.QuotaUsage{
		TenantID:  "tenant-2",
		Platform:  "lazada",
		QuotaType: quota.QuotaShipment,
		Period:    "2026-02-24",
		Usage:     42,
		Limit:     100,
	}

	if u.TenantID != "tenant-2" {
		t.Errorf("expected TenantID='tenant-2', got %q", u.TenantID)
	}
	if u.Platform != "lazada" {
		t.Errorf("expected Platform='lazada', got %q", u.Platform)
	}
	if u.QuotaType != quota.QuotaShipment {
		t.Errorf("expected QuotaShipment, got %q", u.QuotaType)
	}
	if u.Period != "2026-02-24" {
		t.Errorf("expected Period='2026-02-24', got %q", u.Period)
	}
	if u.Usage != 42 {
		t.Errorf("expected Usage=42, got %d", u.Usage)
	}
	if u.Limit != 100 {
		t.Errorf("expected Limit=100, got %d", u.Limit)
	}
}
