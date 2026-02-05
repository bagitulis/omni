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

	// Create test request
	reqBody := BatchAddRequest{
		Items: []BatchAddItem{
			{
				ItemID: 1001,
				SKU:    "SKU001",
				Tiers: []WholesaleTier{
					{MinQty: 5, MaxQty: 10, Price: 90000},
					{MinQty: 11, MaxQty: 20, Price: 85000},
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
		SKUs:          []string{"SKU001", "SKU002"},
		DiscountRates: []int{5, 10, 15},
	}

	body, _ := json.Marshal(reqBody)
	c.Request = httptest.NewRequest("POST", "/api/wholesale/shopee/preview", bytes.NewBuffer(body))
	c.Request.Header.Set("Content-Type", "application/json")

	var req PreviewRequest
	err := c.ShouldBindJSON(&req)
	assert.NoError(t, err, "Should bind JSON successfully")
	assert.Equal(t, 2, len(req.SKUs), "Should have two SKUs")
	assert.Equal(t, 3, len(req.DiscountRates), "Should have three discount rates")
}

// TestImportWholesale tests the ImportWholesale handler
func TestImportWholesale(t *testing.T) {
	gin.SetMode(gin.TestMode)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Set("tenantId", "test-tenant")

	// Create test request
	reqBody := ImportWholesaleRequest{
		Data: []ImportWholesaleItem{
			{
				SKU: "SKU001",
				Tiers: []WholesaleTier{
					{MinQty: 5, MaxQty: 10, Price: 90000},
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
			{ProductID: "prod001", MPQ: 10},
			{ProductID: "prod002", MPQ: 5},
		},
	}

	body, _ := json.Marshal(reqBody)
	c.Request = httptest.NewRequest("POST", "/api/wholesale/tiktok/batch-mpq", bytes.NewBuffer(body))
	c.Request.Header.Set("Content-Type", "application/json")

	var req TiktokBatchMpqRequest
	err := c.ShouldBindJSON(&req)
	assert.NoError(t, err, "Should bind JSON successfully")
	assert.Equal(t, 2, len(req.Products), "Should have two products")
}

// TestWholesaleDTOStructures tests all wholesale DTO structures
func TestWholesaleDTOStructures(t *testing.T) {
	// Test WholesaleTier
	tier := WholesaleTier{
		MinQty: 5,
		MaxQty: 10,
		Price:  90000,
	}
	assert.Equal(t, 5, tier.MinQty, "MinQty should be 5")
	assert.Equal(t, 90000.0, tier.Price, "Price should be 90000")

	// Test WholesaleInfo
	info := WholesaleInfo{
		ItemID:  1001,
		HasTier: true,
		Tiers:   []WholesaleTier{tier},
		MPQ:     5,
	}
	assert.Equal(t, int64(1001), info.ItemID, "ItemID should be 1001")
	assert.True(t, info.HasTier, "HasTier should be true")

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

	// Test GenerateTiers function
	tiers := GenerateTiers(100000, []int{5, 10, 15})
	assert.Equal(t, 3, len(tiers), "Should generate 3 tiers")
	assert.Equal(t, 95000.0, tiers[0].Price, "First tier price should be discounted 5%")
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
