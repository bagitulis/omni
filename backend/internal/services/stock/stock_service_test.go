package stock

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"

	"github.com/glebarez/sqlite"
	"github.com/omni/backend/internal/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

// setupStockTestDB creates an in-memory SQLite DB with the required schema
func setupStockTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err, "failed to open test database")

	err = db.AutoMigrate(&models.InventoryRecord{}, &models.InventorySkuPlatformStatus{})
	require.NoError(t, err, "failed to migrate tables")

	return db
}

// makeInventoryRecord creates a test InventoryRecord with stock in JSON data
func makeInventoryRecord(tenantID, sku string, stock int) models.InventoryRecord {
	data, _ := json.Marshal(map[string]interface{}{
		"SKU":   sku,
		"Stock": stock,
	})
	return models.InventoryRecord{
		ID:            fmt.Sprintf("rec-%s-%s", tenantID, sku),
		TenantID:      tenantID,
		KeyValue:      sku,
		KeyColumnName: "SKU",
		Data:          string(data),
	}
}

// --- Constructor ---

func TestNewStockService_ReturnsNonNil(t *testing.T) {
	db := setupStockTestDB(t)
	svc := NewStockService(db, "tenant1")
	assert.NotNil(t, svc)
	assert.Equal(t, "tenant1", svc.tenantID)
	assert.Nil(t, svc.credService)
}

func TestNewStockServiceWithCreds_ReturnsNonNil(t *testing.T) {
	db := setupStockTestDB(t)
	// Use a non-existent path — CredentialService is lazy so no error expected at construction
	svc := NewStockServiceWithCreds(db, "tenant1", ":memory:")
	assert.NotNil(t, svc)
	assert.Equal(t, "tenant1", svc.tenantID)
	assert.NotNil(t, svc.credService)
}

// --- GetStock ---

func TestStockService_GetStock_NotFound_ReturnsNil(t *testing.T) {
	db := setupStockTestDB(t)
	ctx := context.Background()
	svc := NewStockService(db, "tenant1")

	result, err := svc.GetStock(ctx, "MISSING-SKU")
	assert.NoError(t, err)
	assert.Nil(t, result)
}

func TestStockService_GetStock_Found_ReturnsDTO(t *testing.T) {
	db := setupStockTestDB(t)
	ctx := context.Background()
	svc := NewStockService(db, "tenant1")

	rec := makeInventoryRecord("tenant1", "SKU-001", 50)
	require.NoError(t, db.Create(&rec).Error)

	result, err := svc.GetStock(ctx, "SKU-001")
	assert.NoError(t, err)
	require.NotNil(t, result)
	assert.Equal(t, "SKU-001", result.SKU)
	assert.Equal(t, 50, result.CurrentStock)
}

func TestStockService_GetStock_TenantIsolation(t *testing.T) {
	db := setupStockTestDB(t)
	ctx := context.Background()

	// Insert record for tenant2
	rec := makeInventoryRecord("tenant2", "SKU-001", 10)
	require.NoError(t, db.Create(&rec).Error)

	// tenant1 service should not find tenant2's record
	svc := NewStockService(db, "tenant1")
	result, err := svc.GetStock(ctx, "SKU-001")
	assert.NoError(t, err)
	assert.Nil(t, result)
}

// --- GetAllStock ---

func TestStockService_GetAllStock_Empty(t *testing.T) {
	db := setupStockTestDB(t)
	ctx := context.Background()
	svc := NewStockService(db, "tenant1")

	stocks, total, err := svc.GetAllStock(ctx, false, 0, 0)
	assert.NoError(t, err)
	assert.Equal(t, int64(0), total)
	assert.Empty(t, stocks)
}

func TestStockService_GetAllStock_ReturnsTenantRecords(t *testing.T) {
	db := setupStockTestDB(t)
	ctx := context.Background()
	svc := NewStockService(db, "tenant1")

	recs := []models.InventoryRecord{
		makeInventoryRecord("tenant1", "SKU-A", 10),
		makeInventoryRecord("tenant1", "SKU-B", 20),
		makeInventoryRecord("tenant2", "SKU-C", 30), // different tenant
	}
	require.NoError(t, db.Create(&recs).Error)

	stocks, total, err := svc.GetAllStock(ctx, false, 50, 0)
	assert.NoError(t, err)
	assert.Equal(t, int64(2), total)
	assert.Len(t, stocks, 2)
}

func TestStockService_GetAllStock_DefaultLimit(t *testing.T) {
	db := setupStockTestDB(t)
	ctx := context.Background()
	svc := NewStockService(db, "tenant1")

	// Create 3 records
	for i := 0; i < 3; i++ {
		rec := makeInventoryRecord("tenant1", fmt.Sprintf("SKU-%d", i), i*5)
		require.NoError(t, db.Create(&rec).Error)
	}

	// Passing limit=0 should use default limit of 50
	stocks, total, err := svc.GetAllStock(ctx, false, 0, 0)
	assert.NoError(t, err)
	assert.Equal(t, int64(3), total)
	assert.Len(t, stocks, 3)
}

// --- BulkUpdateStock ---

func TestStockService_BulkUpdateStock_EmptyUpdates(t *testing.T) {
	db := setupStockTestDB(t)
	ctx := context.Background()
	svc := NewStockService(db, "tenant1")

	req := models.BulkStockUpdateRequest{Updates: []models.StockUpdateRequest{}}
	result, err := svc.BulkUpdateStock(ctx, req)
	assert.NoError(t, err)
	require.NotNil(t, result)
	assert.Equal(t, 0, result.TotalRequested)
	assert.Equal(t, 0, result.TotalSuccess)
	assert.Equal(t, 0, result.TotalFailed)
	assert.Empty(t, result.Results)
}

func TestStockService_BulkUpdateStock_SKUNotFound_CountsAsFailed(t *testing.T) {
	db := setupStockTestDB(t)
	ctx := context.Background()
	svc := NewStockService(db, "tenant1")

	req := models.BulkStockUpdateRequest{
		Updates: []models.StockUpdateRequest{
			{SKU: "NONEXISTENT", Quantity: 10},
		},
	}
	result, err := svc.BulkUpdateStock(ctx, req)
	assert.NoError(t, err)
	require.NotNil(t, result)
	assert.Equal(t, 1, result.TotalRequested)
	assert.Equal(t, 0, result.TotalSuccess)
	assert.Equal(t, 1, result.TotalFailed)
	require.Len(t, result.Results, 1)
	assert.False(t, result.Results[0].Success)
	assert.Equal(t, "NONEXISTENT", result.Results[0].SKU)
}

func TestStockService_BulkUpdateStock_ExistingSKU_UpdatesStock(t *testing.T) {
	db := setupStockTestDB(t)
	ctx := context.Background()
	svc := NewStockService(db, "tenant1")

	rec := makeInventoryRecord("tenant1", "SKU-001", 5)
	require.NoError(t, db.Create(&rec).Error)

	req := models.BulkStockUpdateRequest{
		Updates: []models.StockUpdateRequest{
			{SKU: "SKU-001", Quantity: 99, Platforms: []string{}},
		},
	}
	result, err := svc.BulkUpdateStock(ctx, req)
	assert.NoError(t, err)
	require.NotNil(t, result)
	assert.Equal(t, 1, result.TotalRequested)
	// credService is nil so platform sync returns message but UpdateStock itself succeeds
	assert.Equal(t, 1, result.TotalSuccess)
	assert.Equal(t, 0, result.TotalFailed)
}

func TestStockService_BulkUpdateStock_TotalRequested_MatchesInputCount(t *testing.T) {
	db := setupStockTestDB(t)
	ctx := context.Background()
	svc := NewStockService(db, "tenant1")

	req := models.BulkStockUpdateRequest{
		Updates: []models.StockUpdateRequest{
			{SKU: "SKU-A", Quantity: 1},
			{SKU: "SKU-B", Quantity: 2},
			{SKU: "SKU-C", Quantity: 3},
		},
	}
	result, err := svc.BulkUpdateStock(ctx, req)
	assert.NoError(t, err)
	assert.Equal(t, 3, result.TotalRequested)
}

// --- UpdateStock ---

func TestStockService_UpdateStock_SKUNotFound_ReturnsFailure(t *testing.T) {
	db := setupStockTestDB(t)
	ctx := context.Background()
	svc := NewStockService(db, "tenant1")

	req := models.StockUpdateRequest{SKU: "MISSING", Quantity: 5}
	result, err := svc.UpdateStock(ctx, req)
	assert.NoError(t, err)
	require.NotNil(t, result)
	assert.False(t, result.Success)
	assert.Equal(t, "SKU not found", result.Message)
}

func TestStockService_UpdateStock_ExistingSKU_Succeeds(t *testing.T) {
	db := setupStockTestDB(t)
	ctx := context.Background()
	svc := NewStockService(db, "tenant1")

	rec := makeInventoryRecord("tenant1", "SKU-X", 10)
	require.NoError(t, db.Create(&rec).Error)

	req := models.StockUpdateRequest{SKU: "SKU-X", Quantity: 42, Platforms: []string{"shopee"}}
	result, err := svc.UpdateStock(ctx, req)
	assert.NoError(t, err)
	require.NotNil(t, result)
	assert.True(t, result.Success)
	// Platform sync fails (credService nil) but the update itself succeeded
	require.Len(t, result.PlatformResults, 1)
	assert.Equal(t, "shopee", result.PlatformResults[0].Platform)
	assert.False(t, result.PlatformResults[0].Success)
	assert.Equal(t, "Credential service not initialized", result.PlatformResults[0].Message)
}

func TestStockService_UpdateStock_NilCredService_PlatformSyncFails(t *testing.T) {
	db := setupStockTestDB(t)
	ctx := context.Background()
	svc := NewStockService(db, "tenant1") // no creds

	rec := makeInventoryRecord("tenant1", "SKU-Y", 5)
	require.NoError(t, db.Create(&rec).Error)

	req := models.StockUpdateRequest{SKU: "SKU-Y", Quantity: 20, Platforms: []string{"lazada", "tiktok"}}
	result, err := svc.UpdateStock(ctx, req)
	assert.NoError(t, err)
	require.NotNil(t, result)
	assert.True(t, result.Success)
	for _, pr := range result.PlatformResults {
		assert.False(t, pr.Success)
		assert.Equal(t, "Credential service not initialized", pr.Message)
	}
}

// --- GetLowStockAlerts ---

func TestStockService_GetLowStockAlerts_Empty(t *testing.T) {
	db := setupStockTestDB(t)
	ctx := context.Background()
	svc := NewStockService(db, "tenant1")

	alerts, err := svc.GetLowStockAlerts(ctx)
	assert.NoError(t, err)
	assert.Empty(t, alerts)
}

func TestStockService_GetLowStockAlerts_BelowMinStock_ReturnsAlert(t *testing.T) {
	db := setupStockTestDB(t)
	ctx := context.Background()
	svc := NewStockService(db, "tenant1")

	// Stock=3, MinStock=10 → should alert (shortage = 7)
	data, _ := json.Marshal(map[string]interface{}{
		"SKU":      "SKU-LOW",
		"Stock":    3,
		"MinStock": 10,
	})
	rec := models.InventoryRecord{
		ID:            "rec-low",
		TenantID:      "tenant1",
		KeyValue:      "SKU-LOW",
		KeyColumnName: "SKU",
		Data:          string(data),
	}
	require.NoError(t, db.Create(&rec).Error)

	alerts, err := svc.GetLowStockAlerts(ctx)
	assert.NoError(t, err)
	require.Len(t, alerts, 1)
	assert.Equal(t, "SKU-LOW", alerts[0].SKU)
	assert.Equal(t, 3, alerts[0].CurrentStock)
	assert.Equal(t, 10, alerts[0].MinStock)
	assert.Equal(t, 7, alerts[0].Shortage)
}

func TestStockService_GetLowStockAlerts_AboveMinStock_NoAlert(t *testing.T) {
	db := setupStockTestDB(t)
	ctx := context.Background()
	svc := NewStockService(db, "tenant1")

	// Stock=50, MinStock=10 → no alert
	data, _ := json.Marshal(map[string]interface{}{
		"SKU":      "SKU-OK",
		"Stock":    50,
		"MinStock": 10,
	})
	rec := models.InventoryRecord{
		ID:            "rec-ok",
		TenantID:      "tenant1",
		KeyValue:      "SKU-OK",
		KeyColumnName: "SKU",
		Data:          string(data),
	}
	require.NoError(t, db.Create(&rec).Error)

	alerts, err := svc.GetLowStockAlerts(ctx)
	assert.NoError(t, err)
	assert.Empty(t, alerts)
}
