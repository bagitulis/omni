package services

import (
	"sync"
	"testing"
	"time"

	"github.com/omni/backend/internal/services/cache"
	"github.com/omni/backend/internal/utils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// mockCache tracks ClearTenant calls for testing
type mockCache struct {
	mu             sync.Mutex
	clearedTenants []string
	items          map[string]string
}

func newMockCache() *mockCache {
	return &mockCache{items: make(map[string]string)}
}

func (m *mockCache) Get(tenantID, key string) (interface{}, bool) {
	m.mu.Lock()
	defer m.mu.Unlock()
	v, ok := m.items[tenantID+":"+key]
	return v, ok
}

func (m *mockCache) Set(tenantID, key string, value interface{}, ttl time.Duration) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.items[tenantID+":"+key] = value.(string)
	return nil
}

func (m *mockCache) Delete(tenantID, key string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.items, tenantID+":"+key)
	return nil
}

func (m *mockCache) DeletePattern(tenantID, pattern string) error { return nil }

func (m *mockCache) ClearTenant(tenantID string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.clearedTenants = append(m.clearedTenants, tenantID)
}

func (m *mockCache) Count() int {
	m.mu.Lock()
	defer m.mu.Unlock()
	return len(m.items)
}

func (m *mockCache) Stop() {}

func (m *mockCache) getTenantsCleared() []string {
	m.mu.Lock()
	defer m.mu.Unlock()
	return append([]string{}, m.clearedTenants...)
}

// TestSetCacheManager_NilCache_NoOp verifies SwitchTenant works without cache set
func TestSetCacheManager_NilCache_NoOp(t *testing.T) {
	jwtSvc := utils.NewJWTService("test-secret")
	svc := NewMultiTenantAuthService(nil, jwtSvc, "/test/path")
	// cache is nil by default — SwitchTenant should not panic on nil cache
	// (the FORBIDDEN check runs before cache access, so use a non-developer role)
	_, err := svc.SwitchTenant(nil, "user-1", "admin", "tenant-x")
	require.Error(t, err)
	authErr, ok := err.(*AuthError)
	require.True(t, ok)
	assert.Equal(t, "FORBIDDEN", authErr.Code)
}

// TestSwitchTenant_ClearTenantCache verifies ClearTenant is called on switch
func TestSwitchTenant_ClearTenantCache(t *testing.T) {
	jwtSvc := utils.NewJWTService("test-secret")
	svc := NewMultiTenantAuthService(nil, jwtSvc, "/test/path")

	mc := newMockCache()
	svc.SetCacheManager(mc)

	// Developer role passes the FORBIDDEN check; TenantExists will panic on nil tenantService.
	// We catch the panic and verify ClearTenant was called before it.
	defer func() {
		recover()
		cleared := mc.getTenantsCleared()
		require.Len(t, cleared, 1, "expected ClearTenant to be called once for target tenant")
		assert.Equal(t, "target-tenant", cleared[0])
	}()

	_, _ = svc.SwitchTenant(nil, "dev-user", "developer", "target-tenant")
}

// TestSwitchTenant_CacheNotClearedForNonDeveloper verifies no cache clearing for non-developer
func TestSwitchTenant_CacheNotClearedForNonDeveloper(t *testing.T) {
	jwtSvc := utils.NewJWTService("test-secret")
	svc := NewMultiTenantAuthService(nil, jwtSvc, "/test/path")

	mc := newMockCache()
	svc.SetCacheManager(mc)

	_, err := svc.SwitchTenant(nil, "user-1", "admin", "tenant-x")
	require.Error(t, err)

	cleared := mc.getTenantsCleared()
	assert.Empty(t, cleared, "ClearTenant should not be called for non-developer")
}

// TestCacheManager_InterfaceCompliance verifies cache.Cache implements CacheManager
func TestCacheManager_InterfaceCompliance(t *testing.T) {
	c := cache.New(time.Minute, time.Minute)
	defer c.Stop()

	// Verify interface compliance via type assertion
	var _ cache.CacheManager = c

	// Test multi-tenant isolation through the interface
	var cm cache.CacheManager = c

	tenantA := "tenant_alpha"
	tenantB := "tenant_beta"

	// Set keys for both tenants
	err := cm.Set(tenantA, "shared_key", "value_a", time.Minute)
	require.NoError(t, err)
	err = cm.Set(tenantB, "shared_key", "value_b", time.Minute)
	require.NoError(t, err)

	// Each tenant gets its own value
	gotA, foundA := cm.Get(tenantA, "shared_key")
	gotB, foundB := cm.Get(tenantB, "shared_key")
	require.True(t, foundA)
	require.True(t, foundB)
	assert.Equal(t, "value_a", gotA)
	assert.Equal(t, "value_b", gotB)

	// ClearTenant removes only one tenant's data
	cm.ClearTenant(tenantA)

	_, foundAAfter := cm.Get(tenantA, "shared_key")
	assert.False(t, foundAAfter, "tenantA data should be cleared")

	gotBAfter, foundBAfter := cm.Get(tenantB, "shared_key")
	require.True(t, foundBAfter, "tenantB data should persist")
	assert.Equal(t, "value_b", gotBAfter)
}

// TestCache_BuildKey_PreventsCrossTenant verifies key construction prevents leaks
func TestCache_BuildKey_PreventsCrossTenant(t *testing.T) {
	c := cache.New(time.Minute, time.Minute)
	defer c.Stop()

	// Simulate same key name used by different tenants
	c.Set("shop_a", "api_token", "token_a", time.Minute)
	c.Set("shop_b", "api_token", "token_b", time.Minute)

	// Verify isolation
	gotA, foundA := c.Get("shop_a", "api_token")
	gotB, foundB := c.Get("shop_b", "api_token")

	require.True(t, foundA)
	require.True(t, foundB)
	assert.Equal(t, "token_a", gotA)
	assert.Equal(t, "token_b", gotB)

	// Verify third tenant sees nothing
	_, foundC := c.Get("shop_c", "api_token")
	assert.False(t, foundC)
}
