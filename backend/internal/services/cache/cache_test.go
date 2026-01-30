package cache

import (
	"testing"
	"time"
)

func TestCache_SetAndGet(t *testing.T) {
	c := New(5*time.Minute, 10*time.Minute)
	defer c.Stop()

	tenantID := "test_tenant"
	key := "test_key"
	value := "test_value"

	// Set value
	err := c.Set(tenantID, key, value, time.Minute)
	if err != nil {
		t.Fatalf("Set failed: %v", err)
	}

	// Get value
	got, found := c.Get(tenantID, key)
	if !found {
		t.Fatal("Expected to find value in cache")
	}
	if got != value {
		t.Errorf("Expected %v, got %v", value, got)
	}
}

func TestCache_GetNotFound(t *testing.T) {
	c := New(5*time.Minute, 10*time.Minute)
	defer c.Stop()

	_, found := c.Get("tenant", "nonexistent")
	if found {
		t.Error("Expected not to find nonexistent key")
	}
}

func TestCache_Delete(t *testing.T) {
	c := New(5*time.Minute, 10*time.Minute)
	defer c.Stop()

	tenantID := "test_tenant"
	key := "test_key"

	c.Set(tenantID, key, "value", time.Minute)

	// Delete
	err := c.Delete(tenantID, key)
	if err != nil {
		t.Fatalf("Delete failed: %v", err)
	}

	// Verify deleted
	_, found := c.Get(tenantID, key)
	if found {
		t.Error("Expected key to be deleted")
	}
}

func TestCache_MultiTenantIsolation(t *testing.T) {
	c := New(5*time.Minute, 10*time.Minute)
	defer c.Stop()

	tenant1 := "tenant_a"
	tenant2 := "tenant_b"
	key := "shared_key"

	// Set same key for different tenants with different values
	c.Set(tenant1, key, "value_a", time.Minute)
	c.Set(tenant2, key, "value_b", time.Minute)

	// Each tenant should get their own value
	got1, found1 := c.Get(tenant1, key)
	got2, found2 := c.Get(tenant2, key)

	if !found1 || !found2 {
		t.Fatal("Expected to find values for both tenants")
	}

	if got1 != "value_a" {
		t.Errorf("Tenant 1: expected value_a, got %v", got1)
	}
	if got2 != "value_b" {
		t.Errorf("Tenant 2: expected value_b, got %v", got2)
	}
}

func TestCache_DeletePattern(t *testing.T) {
	c := New(5*time.Minute, 10*time.Minute)
	defer c.Stop()

	tenantID := "test_tenant"

	// Set multiple keys with pattern
	c.Set(tenantID, "analytics:summary", "v1", time.Minute)
	c.Set(tenantID, "analytics:health", "v2", time.Minute)
	c.Set(tenantID, "products:list", "v3", time.Minute)

	// Delete analytics pattern
	err := c.DeletePattern(tenantID, "analytics:*")
	if err != nil {
		t.Fatalf("DeletePattern failed: %v", err)
	}

	// Analytics keys should be deleted
	_, found1 := c.Get(tenantID, "analytics:summary")
	_, found2 := c.Get(tenantID, "analytics:health")
	if found1 || found2 {
		t.Error("Expected analytics keys to be deleted")
	}

	// Products key should still exist
	_, found3 := c.Get(tenantID, "products:list")
	if !found3 {
		t.Error("Expected products key to still exist")
	}
}

func TestCache_Expiration(t *testing.T) {
	c := New(5*time.Minute, time.Second)
	defer c.Stop()

	tenantID := "test_tenant"
	key := "expiring_key"

	// Set with short TTL
	c.Set(tenantID, key, "value", 50*time.Millisecond)

	// Should exist immediately
	_, found := c.Get(tenantID, key)
	if !found {
		t.Fatal("Expected to find value immediately after set")
	}

	// Wait for expiration
	time.Sleep(100 * time.Millisecond)

	// Should be expired
	_, found = c.Get(tenantID, key)
	if found {
		t.Error("Expected key to be expired")
	}
}

func TestCache_Count(t *testing.T) {
	c := New(5*time.Minute, 10*time.Minute)
	defer c.Stop()

	if c.Count() != 0 {
		t.Error("Expected empty cache")
	}

	c.Set("t1", "k1", "v1", time.Minute)
	c.Set("t1", "k2", "v2", time.Minute)
	c.Set("t2", "k1", "v3", time.Minute)

	if c.Count() != 3 {
		t.Errorf("Expected 3 items, got %d", c.Count())
	}
}

func TestCache_ClearTenant(t *testing.T) {
	c := New(5*time.Minute, 10*time.Minute)
	defer c.Stop()

	tenant1 := "tenant_a"
	tenant2 := "tenant_b"

	c.Set(tenant1, "k1", "v1", time.Minute)
	c.Set(tenant1, "k2", "v2", time.Minute)
	c.Set(tenant2, "k1", "v3", time.Minute)

	// Clear tenant1
	c.ClearTenant(tenant1)

	// Tenant1 keys should be gone
	_, found1 := c.Get(tenant1, "k1")
	_, found2 := c.Get(tenant1, "k2")
	if found1 || found2 {
		t.Error("Expected tenant1 keys to be cleared")
	}

	// Tenant2 key should still exist
	_, found3 := c.Get(tenant2, "k1")
	if !found3 {
		t.Error("Expected tenant2 key to still exist")
	}
}

func TestCache_Clear(t *testing.T) {
	c := New(5*time.Minute, 10*time.Minute)
	defer c.Stop()

	c.Set("t1", "k1", "v1", time.Minute)
	c.Set("t2", "k2", "v2", time.Minute)

	c.Clear()

	if c.Count() != 0 {
		t.Error("Expected cache to be empty after Clear")
	}
}

func TestBuildKey(t *testing.T) {
	key := buildKey("my_tenant", "my_key")
	expected := "tenant:my_tenant:my_key"
	if key != expected {
		t.Errorf("Expected %s, got %s", expected, key)
	}
}
