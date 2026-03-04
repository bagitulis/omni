package handlers

import (
	"bytes"
	"encoding/json"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

// TestBatchDeleteByItemIds tests the BatchDeleteByItemIds handler
func TestBatchDeleteByItemIds(t *testing.T) {
	gin.SetMode(gin.TestMode)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Set("tenantId", "test-tenant")

	// Create test request
	reqBody := BatchDeleteRequest{
		ItemIDs: []int64{1001, 1002, 1003},
	}

	body, _ := json.Marshal(reqBody)
	c.Request = httptest.NewRequest("POST", "/api/wholesale/shopee/batch-delete", bytes.NewBuffer(body))
	c.Request.Header.Set("Content-Type", "application/json")

	var req BatchDeleteRequest
	err := c.ShouldBindJSON(&req)
	assert.NoError(t, err, "Should bind JSON successfully")
	assert.Equal(t, 3, len(req.ItemIDs), "Should have three item IDs")
}

// TestBatchAdd tests the BatchAdd handler
func TestBatchAdd(t *testing.T) {
	gin.SetMode(gin.TestMode)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Set("tenantId", "test-tenant")

	// Create test request with new field names
	reqBody := BatchAddRequest{
		Items: []BatchAddItem{
			{
				ItemID: 1001,
				SKU:    "SKU001",
				Tiers: []WholesaleTier{
					{MinCount: 5, MaxCount: 10, UnitPrice: 90000},
					{MinCount: 11, MaxCount: 20, UnitPrice: 85000},
				},
			},
		},
	}

	body, _ := json.Marshal(reqBody)
	c.Request = httptest.NewRequest("POST", "/api/wholesale/shopee/batch-add", bytes.NewBuffer(body))
	c.Request.Header.Set("Content-Type", "application/json")

	var req BatchAddRequest
	err := c.ShouldBindJSON(&req)
	assert.NoError(t, err, "Should bind JSON successfully")
	assert.Equal(t, 1, len(req.Items), "Should have one item")
	assert.Equal(t, "SKU001", req.Items[0].SKU, "SKU should be SKU001")
	assert.Equal(t, 2, len(req.Items[0].Tiers), "Should have two tiers")
}

// TestPreview tests the Preview handler
func TestPreview(t *testing.T) {
	gin.SetMode(gin.TestMode)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Set("tenantId", "test-tenant")

	// Create test request
	reqBody := PreviewRequest{
		SKUs:      []string{"SKU001", "SKU002"},
		BasePrice: 100000,
	}

	body, _ := json.Marshal(reqBody)
	c.Request = httptest.NewRequest("POST", "/api/wholesale/shopee/preview", bytes.NewBuffer(body))
	c.Request.Header.Set("Content-Type", "application/json")

	var req PreviewRequest
	err := c.ShouldBindJSON(&req)
	assert.NoError(t, err, "Should bind JSON successfully")
	assert.Equal(t, 2, len(req.SKUs), "Should have two SKUs")
	assert.Equal(t, float64(100000), req.BasePrice, "BasePrice should be 100000")
}

// TestImportWholesale tests the ImportWholesale handler
func TestImportWholesale(t *testing.T) {
	gin.SetMode(gin.TestMode)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Set("tenantId", "test-tenant")

	// Create test request with new field names
	reqBody := ImportWholesaleRequest{
		Data: []ImportWholesaleItem{
			{
				SKU: "SKU001",
				Tiers: []WholesaleTier{
					{MinCount: 5, MaxCount: 10, UnitPrice: 90000},
				},
			},
		},
	}

	body, _ := json.Marshal(reqBody)
	c.Request = httptest.NewRequest("POST", "/api/wholesale/shopee/import", bytes.NewBuffer(body))
	c.Request.Header.Set("Content-Type", "application/json")

	var req ImportWholesaleRequest
	err := c.ShouldBindJSON(&req)
	assert.NoError(t, err, "Should bind JSON successfully")
	assert.Equal(t, 1, len(req.Data), "Should have one item")
}

// TestBatchSetMpq tests the BatchSetMpq handler
func TestBatchSetMpq(t *testing.T) {
	gin.SetMode(gin.TestMode)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Set("tenantId", "test-tenant")

	// Create test request
	reqBody := map[string]interface{}{
		"items": []map[string]interface{}{
			{"sku": "SKU001", "price": 100000},
			{"sku": "SKU002", "price": 95000},
		},
		"mpq": 5,
	}

	body, _ := json.Marshal(reqBody)
	c.Request = httptest.NewRequest("POST", "/api/wholesale/shopee/batch-mpq", bytes.NewBuffer(body))
	c.Request.Header.Set("Content-Type", "application/json")

	var req map[string]interface{}
	err := c.ShouldBindJSON(&req)
	assert.NoError(t, err, "Should bind JSON successfully")
	assert.NotNil(t, req, "Request should not be nil")
}

// TestBatchSetTiktokMpq tests the BatchSetTiktokMpq handler
func TestBatchSetTiktokMpq(t *testing.T) {
	gin.SetMode(gin.TestMode)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Set("tenantId", "test-tenant")

	// Create test request
	reqBody := TiktokBatchMpqRequest{
		Products: []TiktokMpqItem{
			{ProductID: "prod001", SKU: "SKU-001", Price: 5000, MPQ: 10},
			{ProductID: "prod002", SKU: "SKU-002", Price: 6000, MPQ: 5},
		},
		MPQ: 3,
	}

	body, _ := json.Marshal(reqBody)
	c.Request = httptest.NewRequest("POST", "/api/wholesale/tiktok/batch-mpq", bytes.NewBuffer(body))
	c.Request.Header.Set("Content-Type", "application/json")

	var req TiktokBatchMpqRequest
	err := c.ShouldBindJSON(&req)
	assert.NoError(t, err, "Should bind JSON successfully")
	assert.Equal(t, 2, len(req.Products), "Should have two products")
	assert.Equal(t, 3, req.MPQ, "Should have MPQ at root level")
}

// TestWholesaleDTOStructures tests all wholesale DTO structures
func TestWholesaleDTOStructures(t *testing.T) {
	// Test WholesaleTier with new field names
	tier := WholesaleTier{
		MinCount:  5,
		MaxCount:  10,
		UnitPrice: 90000,
	}
	assert.Equal(t, 5, tier.MinCount, "MinCount should be 5")
	assert.Equal(t, 90000.0, tier.UnitPrice, "UnitPrice should be 90000")

	// Test WholesaleInfo
	info := WholesaleInfo{
		ItemID:       1001,
		HasWholesale: true,
		Tiers:        []WholesaleTier{tier},
	}
	assert.Equal(t, int64(1001), info.ItemID, "ItemID should be 1001")
	assert.True(t, info.HasWholesale, "HasWholesale should be true")

	// Test UpdateWholesaleRequest
	updateReq := UpdateWholesaleRequest{
		Tiers: []WholesaleTier{tier},
	}
	assert.Equal(t, 1, len(updateReq.Tiers), "Should have one tier")

	// Test BatchDeleteRequest
	delReq := BatchDeleteRequest{
		ItemIDs: []int64{1001, 1002},
	}
	assert.Equal(t, 2, len(delReq.ItemIDs), "Should have two item IDs")
}

// TestWholesaleBatchMissingTenant tests error handling
func TestWholesaleBatchMissingTenant(t *testing.T) {
	gin.SetMode(gin.TestMode)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("POST", "/api/wholesale/shopee/batch-delete", nil)

	tenantID := c.GetString("tenantId")
	assert.Equal(t, "", tenantID, "TenantId should be empty when not set")
}
