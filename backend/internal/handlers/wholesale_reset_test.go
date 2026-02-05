package handlers

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

// TestWholesaleExtendedHandler_BatchWholesaleReset_MissingTenant tests BatchWholesaleReset without tenant
func TestWholesaleExtendedHandler_BatchWholesaleReset_MissingTenant(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	handler := NewWholesaleExtendedHandler("", nil)
	r.POST("/api/wholesale/shopee/batch-reset", handler.BatchWholesaleReset)

	req, _ := http.NewRequest("POST", "/api/wholesale/shopee/batch-reset", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

// TestWholesaleExtendedHandler_BatchWholesaleReset_InvalidJSON tests BatchWholesaleReset with invalid JSON
func TestWholesaleExtendedHandler_BatchWholesaleReset_InvalidJSON(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	r.Use(func(c *gin.Context) {
		c.Set("tenantID", "test-tenant")
		c.Next()
	})

	handler := NewWholesaleExtendedHandler("", nil)
	r.POST("/api/wholesale/shopee/batch-reset", handler.BatchWholesaleReset)

	body := `{invalid json`
	req, _ := http.NewRequest("POST", "/api/wholesale/shopee/batch-reset", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

// TestWholesaleExtendedHandler_BatchWholesaleReset_EmptyItems tests BatchWholesaleReset with empty items
func TestWholesaleExtendedHandler_BatchWholesaleReset_EmptyItems(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	r.Use(func(c *gin.Context) {
		c.Set("tenantID", "test-tenant")
		c.Next()
	})

	handler := NewWholesaleExtendedHandler("", nil)
	r.POST("/api/wholesale/shopee/batch-reset", handler.BatchWholesaleReset)

	body := `{"items": []}`
	req, _ := http.NewRequest("POST", "/api/wholesale/shopee/batch-reset", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

// TestWholesaleExtendedHandler_BatchWholesaleReset_DBError tests BatchWholesaleReset with DB error
func TestWholesaleExtendedHandler_BatchWholesaleReset_DBError(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	r.Use(func(c *gin.Context) {
		c.Set("tenantID", "test-tenant")
		c.Next()
	})

	handler := NewWholesaleExtendedHandler("", nil)
	r.POST("/api/wholesale/shopee/batch-reset", handler.BatchWholesaleReset)

	body := `{"items": [{"sku": "SKU001", "price": 100000}]}`
	req, _ := http.NewRequest("POST", "/api/wholesale/shopee/batch-reset", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	// Without DB, should return 500
	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

// TestGenerateTiers tests the GenerateTiers helper function
func TestGenerateTiers(t *testing.T) {
	basePrice := 100000.0
	discountRates := []int{5, 10, 15}

	tiers := GenerateTiers(basePrice, discountRates)

	assert.Len(t, tiers, 3)

	// First tier: 5% discount
	assert.Equal(t, 5, tiers[0].MinQty)
	assert.Equal(t, 10, tiers[0].MaxQty)
	assert.Equal(t, 95000.0, tiers[0].Price) // 100000 * (100-5) / 100

	// Second tier: 10% discount
	assert.Equal(t, 10, tiers[1].MinQty)
	assert.Equal(t, 15, tiers[1].MaxQty)
	assert.Equal(t, 90000.0, tiers[1].Price) // 100000 * (100-10) / 100

	// Third tier (last): 15% discount, max qty = 999
	assert.Equal(t, 15, tiers[2].MinQty)
	assert.Equal(t, 999, tiers[2].MaxQty)
	assert.Equal(t, 85000.0, tiers[2].Price) // 100000 * (100-15) / 100
}

// TestGenerateTiers_Empty tests GenerateTiers with empty discount rates
func TestGenerateTiers_Empty(t *testing.T) {
	basePrice := 100000.0
	discountRates := []int{}

	tiers := GenerateTiers(basePrice, discountRates)

	assert.Len(t, tiers, 0)
}

// TestGenerateTiers_SingleTier tests GenerateTiers with single tier
func TestGenerateTiers_SingleTier(t *testing.T) {
	basePrice := 100000.0
	discountRates := []int{10}

	tiers := GenerateTiers(basePrice, discountRates)

	assert.Len(t, tiers, 1)
	assert.Equal(t, 5, tiers[0].MinQty)
	assert.Equal(t, 999, tiers[0].MaxQty) // Single tier gets max qty of 999
	assert.Equal(t, 90000.0, tiers[0].Price)
}
