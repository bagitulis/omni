package platform_test

import (
	"context"
	"testing"

	"github.com/omni/backend/internal/services/platform"
)

// ---------------------------------------------------------------------------
// PlatformType constants
// ---------------------------------------------------------------------------

func TestPlatformTypeConstants(t *testing.T) {
	if platform.PlatformShopee != "shopee" {
		t.Errorf("expected PlatformShopee='shopee', got %q", platform.PlatformShopee)
	}
	if platform.PlatformLazada != "lazada" {
		t.Errorf("expected PlatformLazada='lazada', got %q", platform.PlatformLazada)
	}
	if platform.PlatformTiktok != "tiktok" {
		t.Errorf("expected PlatformTiktok='tiktok', got %q", platform.PlatformTiktok)
	}
}

// ---------------------------------------------------------------------------
// NewBaseConfigManager
// ---------------------------------------------------------------------------

func TestNewBaseConfigManager_GetPlatform(t *testing.T) {
	m := platform.NewBaseConfigManager(platform.PlatformShopee, "tenant-1")
	if m == nil {
		t.Fatal("NewBaseConfigManager returned nil")
	}
	if m.GetPlatform() != platform.PlatformShopee {
		t.Errorf("expected platform shopee, got %s", m.GetPlatform())
	}
}

func TestNewBaseConfigManager_GetTenantID(t *testing.T) {
	m := platform.NewBaseConfigManager(platform.PlatformLazada, "tenant-abc")
	if m.GetTenantID() != "tenant-abc" {
		t.Errorf("expected tenant-abc, got %s", m.GetTenantID())
	}
}

func TestNewBaseConfigManager_InitialTokensEmpty(t *testing.T) {
	m := platform.NewBaseConfigManager(platform.PlatformTiktok, "t1")
	if m.GetAccessToken() != "" {
		t.Errorf("expected empty access token, got %q", m.GetAccessToken())
	}
	if m.GetRefreshToken() != "" {
		t.Errorf("expected empty refresh token, got %q", m.GetRefreshToken())
	}
}

func TestBaseConfigManager_SetTokens(t *testing.T) {
	m := platform.NewBaseConfigManager(platform.PlatformShopee, "t1")

	err := m.SetTokens("access-abc", "refresh-xyz")
	if err != nil {
		t.Fatalf("SetTokens returned error: %v", err)
	}
	if m.GetAccessToken() != "access-abc" {
		t.Errorf("expected access-abc, got %q", m.GetAccessToken())
	}
	if m.GetRefreshToken() != "refresh-xyz" {
		t.Errorf("expected refresh-xyz, got %q", m.GetRefreshToken())
	}
}

func TestBaseConfigManager_SetConfig_GetConfig(t *testing.T) {
	m := platform.NewBaseConfigManager(platform.PlatformShopee, "t1")

	err := m.SetConfig("myKey", "myValue")
	if err != nil {
		t.Fatalf("SetConfig returned error: %v", err)
	}

	val, ok := m.GetConfig("myKey")
	if !ok {
		t.Fatal("GetConfig returned false for existing key")
	}
	if val != "myValue" {
		t.Errorf("expected myValue, got %q", val)
	}
}

func TestBaseConfigManager_GetConfig_MissingKey(t *testing.T) {
	m := platform.NewBaseConfigManager(platform.PlatformShopee, "t1")

	val, ok := m.GetConfig("nonexistent")
	if ok {
		t.Error("GetConfig should return false for missing key")
	}
	if val != "" {
		t.Errorf("expected empty string for missing key, got %q", val)
	}
}

func TestBaseConfigManager_GetTokenExpiry_Default(t *testing.T) {
	m := platform.NewBaseConfigManager(platform.PlatformShopee, "t1")
	// No tokenExpiry set — should return 0
	expiry := m.GetTokenExpiry()
	if expiry != 0 {
		t.Errorf("expected 0 default token expiry, got %d", expiry)
	}
}

func TestBaseConfigManager_GetTokenExpiry_SetViaConfig(t *testing.T) {
	m := platform.NewBaseConfigManager(platform.PlatformShopee, "t1")
	// Set a future expiry as string
	m.SetConfig("tokenExpiry", "9999999999999")

	expiry := m.GetTokenExpiry()
	if expiry != 9999999999999 {
		t.Errorf("expected 9999999999999, got %d", expiry)
	}
}

func TestBaseConfigManager_IsTokenExpired_NoExpiry(t *testing.T) {
	m := platform.NewBaseConfigManager(platform.PlatformShopee, "t1")
	// No expiry set — should be considered expired
	if !m.IsTokenExpired() {
		t.Error("expected IsTokenExpired=true when no expiry is set")
	}
}

func TestBaseConfigManager_IsTokenExpired_FutureExpiry(t *testing.T) {
	m := platform.NewBaseConfigManager(platform.PlatformShopee, "t1")
	// Set expiry far in the future (year 2317): 10000000000000 ms
	m.SetConfig("tokenExpiry", "10000000000000")

	if m.IsTokenExpired() {
		t.Error("expected IsTokenExpired=false for far-future expiry")
	}
}

func TestBaseConfigManager_IsTokenExpired_PastExpiry(t *testing.T) {
	m := platform.NewBaseConfigManager(platform.PlatformShopee, "t1")
	// Set expiry in the past: 1 ms
	m.SetConfig("tokenExpiry", "1")

	if !m.IsTokenExpired() {
		t.Error("expected IsTokenExpired=true for past expiry")
	}
}

func TestBaseConfigManager_SetTokenRefresher(t *testing.T) {
	m := platform.NewBaseConfigManager(platform.PlatformShopee, "t1")
	// Set a valid TokenRefreshFunc — verifies the method is callable without panicking
	m.SetTokenRefresher(platform.TokenRefreshFunc(func(ctx context.Context, tenantID string) (string, string, error) {
		return "new-access", "new-refresh", nil
	}))
	// No assertion needed beyond no panic — the refresher is stored and called on EnsureValidToken
}

// ---------------------------------------------------------------------------
// NewShopeeConfigManager
// ---------------------------------------------------------------------------

func TestNewShopeeConfigManager(t *testing.T) {
	m := platform.NewShopeeConfigManager("tenant-shopee")
	if m == nil {
		t.Fatal("NewShopeeConfigManager returned nil")
	}
	if m.GetPlatform() != platform.PlatformShopee {
		t.Errorf("expected shopee platform, got %s", m.GetPlatform())
	}
	if m.GetTenantID() != "tenant-shopee" {
		t.Errorf("expected tenant-shopee, got %s", m.GetTenantID())
	}
}

func TestShopeeConfigManager_IsConfigured_WhenEmpty(t *testing.T) {
	m := platform.NewShopeeConfigManager("t1")
	if m.IsConfigured() {
		t.Error("expected IsConfigured=false when PartnerID/PartnerKey/ShopID are empty")
	}
}

func TestShopeeConfigManager_IsConfigured_WhenFilled(t *testing.T) {
	m := platform.NewShopeeConfigManager("t1")
	m.PartnerID = "pid"
	m.PartnerKey = "pkey"
	m.ShopID = "sid"
	if !m.IsConfigured() {
		t.Error("expected IsConfigured=true when all fields set")
	}
}

// ---------------------------------------------------------------------------
// NewLazadaConfigManager
// ---------------------------------------------------------------------------

func TestNewLazadaConfigManager(t *testing.T) {
	m := platform.NewLazadaConfigManager("tenant-lazada")
	if m == nil {
		t.Fatal("NewLazadaConfigManager returned nil")
	}
	if m.GetPlatform() != platform.PlatformLazada {
		t.Errorf("expected lazada platform, got %s", m.GetPlatform())
	}
	if m.GetTenantID() != "tenant-lazada" {
		t.Errorf("expected tenant-lazada, got %s", m.GetTenantID())
	}
	// Default region should be ID
	if m.Region != "ID" {
		t.Errorf("expected default region ID, got %s", m.Region)
	}
}

func TestLazadaConfigManager_IsConfigured_WhenEmpty(t *testing.T) {
	m := platform.NewLazadaConfigManager("t1")
	if m.IsConfigured() {
		t.Error("expected IsConfigured=false when AppKey/AppSecret/AccessToken are empty")
	}
}

// ---------------------------------------------------------------------------
// NewTiktokConfigManager
// ---------------------------------------------------------------------------

func TestNewTiktokConfigManager(t *testing.T) {
	m := platform.NewTiktokConfigManager("tenant-tiktok")
	if m == nil {
		t.Fatal("NewTiktokConfigManager returned nil")
	}
	if m.GetPlatform() != platform.PlatformTiktok {
		t.Errorf("expected tiktok platform, got %s", m.GetPlatform())
	}
	if m.GetTenantID() != "tenant-tiktok" {
		t.Errorf("expected tenant-tiktok, got %s", m.GetTenantID())
	}
}

func TestTiktokConfigManager_IsConfigured_WhenEmpty(t *testing.T) {
	m := platform.NewTiktokConfigManager("t1")
	if m.IsConfigured() {
		t.Error("expected IsConfigured=false when credentials are empty")
	}
}

// ---------------------------------------------------------------------------
// NewPlatformCoordinationService
// ---------------------------------------------------------------------------

func TestNewPlatformCoordinationService(t *testing.T) {
	svc := platform.NewPlatformCoordinationService("tenant-coord")
	if svc == nil {
		t.Fatal("NewPlatformCoordinationService returned nil")
	}
	if svc.GetTenantID() != "tenant-coord" {
		t.Errorf("expected tenant-coord, got %s", svc.GetTenantID())
	}
}

func TestPlatformCoordinationService_NotInitializedByDefault(t *testing.T) {
	svc := platform.NewPlatformCoordinationService("t1")
	if svc.IsInitialized() {
		t.Error("expected IsInitialized=false before InitializePlatforms is called")
	}
}

func TestPlatformCoordinationService_GetClient_UnregisteredPlatform(t *testing.T) {
	svc := platform.NewPlatformCoordinationService("t1")
	// No InitializePlatforms called — GetClient should return error
	_, err := svc.GetClient(platform.PlatformShopee)
	if err == nil {
		t.Error("expected error when getting client for uninitialized platform")
	}
}

func TestPlatformCoordinationService_GetConfigManager_UnregisteredPlatform(t *testing.T) {
	svc := platform.NewPlatformCoordinationService("t1")
	_, err := svc.GetConfigManager(platform.PlatformLazada)
	if err == nil {
		t.Error("expected error when getting config manager for uninitialized platform")
	}
}

func TestPlatformCoordinationService_GetAllClients_EmptyBeforeInit(t *testing.T) {
	svc := platform.NewPlatformCoordinationService("t1")
	clients := svc.GetAllClients()
	if len(clients) != 0 {
		t.Errorf("expected 0 clients before init, got %d", len(clients))
	}
}

func TestPlatformCoordinationService_GetShopeeClient_NilBeforeInit(t *testing.T) {
	svc := platform.NewPlatformCoordinationService("t1")
	client := svc.GetShopeeClient()
	if client != nil {
		t.Error("expected nil Shopee client before initialization")
	}
}

func TestPlatformCoordinationService_GetLazadaClient_NilBeforeInit(t *testing.T) {
	svc := platform.NewPlatformCoordinationService("t1")
	client := svc.GetLazadaClient()
	if client != nil {
		t.Error("expected nil Lazada client before initialization")
	}
}

func TestPlatformCoordinationService_GetTiktokClient_NilBeforeInit(t *testing.T) {
	svc := platform.NewPlatformCoordinationService("t1")
	client := svc.GetTiktokClient()
	if client != nil {
		t.Error("expected nil TikTok client before initialization")
	}
}

// ---------------------------------------------------------------------------
// Registry: RegisterTokenRefreshService / GetTokenRefreshService
// ---------------------------------------------------------------------------

func TestRegisterTokenRefreshService(t *testing.T) {
	// Reset state after test
	original := platform.GetTokenRefreshService()
	defer platform.RegisterTokenRefreshService(original)

	platform.RegisterTokenRefreshService(nil)
	if platform.GetTokenRefreshService() != nil {
		t.Error("expected nil after registering nil")
	}
}

func TestGetPlatformCoordinationService_EmptyTenantID_ReturnsNil(t *testing.T) {
	svc := platform.GetPlatformCoordinationService("")
	if svc != nil {
		t.Error("expected nil for empty tenantID")
	}
}

func TestGetPlatformCoordinationService_ValidTenantID_ReturnsService(t *testing.T) {
	svc := platform.GetPlatformCoordinationService("tenant-x")
	if svc == nil {
		t.Error("expected non-nil service for valid tenantID")
	}
	if svc.GetTenantID() != "tenant-x" {
		t.Errorf("expected tenant-x, got %s", svc.GetTenantID())
	}
}

func TestClearPlatformServices_NoOp(t *testing.T) {
	// Should not panic
	platform.ClearPlatformServices()
}

func TestInvalidateTenantPlatformService_NoOp(t *testing.T) {
	// Should not panic
	platform.InvalidateTenantPlatformService("tenant-x")
}
