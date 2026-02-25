package google

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

// ---- SpreadsheetLink / SaveLinksRequest JSON ----

func TestSpreadsheetLink_JSONFields(t *testing.T) {
	link := SpreadsheetLink{
		Type:          "inventory",
		SpreadsheetID: "1BxiMVs0XRA5nFMdKvBdBZjgmUUqptlbs74OgVE2upms",
		URL:           "https://docs.google.com/spreadsheets/d/1BxiMVs0XRA5nFMdKvBdBZjgmUUqptlbs74OgVE2upms",
		Title:         "Inventory Sheet",
	}

	data, err := json.Marshal(link)
	assert.NoError(t, err)

	var out map[string]interface{}
	assert.NoError(t, json.Unmarshal(data, &out))

	assert.Equal(t, "inventory", out["type"])
	assert.Equal(t, "1BxiMVs0XRA5nFMdKvBdBZjgmUUqptlbs74OgVE2upms", out["spreadsheet_id"])
	assert.NotEmpty(t, out["url"])
	assert.Equal(t, "Inventory Sheet", out["title"])
}

func TestSaveLinksRequest_JSONFields(t *testing.T) {
	inventoryURL := "https://docs.google.com/spreadsheets/d/abc"
	req := SaveLinksRequest{
		InventoryURL: &inventoryURL,
	}

	data, err := json.Marshal(req)
	assert.NoError(t, err)

	var out map[string]interface{}
	assert.NoError(t, json.Unmarshal(data, &out))

	assert.Equal(t, inventoryURL, out["inventory_url"])
}

// ---- getStringValue ----

func TestGetStringValue_Nil(t *testing.T) {
	result := getStringValue(nil)
	assert.Equal(t, "", result)
}

func TestGetStringValue_NonNil(t *testing.T) {
	val := "hello"
	result := getStringValue(&val)
	assert.Equal(t, "hello", result)
}

func TestGetStringValue_Empty(t *testing.T) {
	empty := ""
	result := getStringValue(&empty)
	assert.Equal(t, "", result)
}

// ---- getPreferredStringValue ----

func TestGetPreferredStringValue_BothNil(t *testing.T) {
	result := getPreferredStringValue(nil, nil)
	assert.Equal(t, "", result)
}

func TestGetPreferredStringValue_PrimarySet(t *testing.T) {
	primary := "primary-value"
	fallback := "fallback-value"
	result := getPreferredStringValue(&primary, &fallback)
	assert.Equal(t, "primary-value", result)
}

func TestGetPreferredStringValue_PrimaryNilFallbackSet(t *testing.T) {
	fallback := "fallback-value"
	result := getPreferredStringValue(nil, &fallback)
	assert.Equal(t, "fallback-value", result)
}

func TestGetPreferredStringValue_PrimaryNilFallbackNil(t *testing.T) {
	result := getPreferredStringValue(nil, nil)
	assert.Equal(t, "", result)
}

// ---- buildLinksFromRequest ----

func TestBuildLinksFromRequest_WithURLFields(t *testing.T) {
	inv := "https://docs.google.com/spreadsheets/d/inv"
	wallet := "https://docs.google.com/spreadsheets/d/wallet"
	req := SaveLinksRequest{
		InventoryURL: &inv,
		WalletURL:    &wallet,
	}

	links := buildLinksFromRequest(req)
	assert.NotNil(t, links)
	assert.Equal(t, inv, links.Inventory)
	assert.Equal(t, wallet, links.Wallet)
	assert.Equal(t, "", links.Shipping)
	assert.Equal(t, "", links.Order)
}

func TestBuildLinksFromRequest_FallbackFields(t *testing.T) {
	inv := "https://docs.google.com/spreadsheets/d/inv-legacy"
	req := SaveLinksRequest{
		Inventory: &inv,
	}

	links := buildLinksFromRequest(req)
	assert.NotNil(t, links)
	assert.Equal(t, inv, links.Inventory)
}

func TestBuildLinksFromRequest_PrimaryOverridesFallback(t *testing.T) {
	primary := "https://docs.google.com/spreadsheets/d/primary"
	fallback := "https://docs.google.com/spreadsheets/d/fallback"
	req := SaveLinksRequest{
		InventoryURL: &primary,
		Inventory:    &fallback,
	}

	links := buildLinksFromRequest(req)
	assert.Equal(t, primary, links.Inventory)
}

func TestBuildLinksFromRequest_Empty(t *testing.T) {
	links := buildLinksFromRequest(SaveLinksRequest{})
	assert.NotNil(t, links)
	assert.Equal(t, "", links.Inventory)
	assert.Equal(t, "", links.Wallet)
	assert.Equal(t, "", links.Shipping)
	assert.Equal(t, "", links.Order)
}

// ---- SaveLinks handler ----

func TestSettingsHandler_SaveLinks_MissingTenantID(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	handler := NewSettingsHandler(nil, nil)
	r.POST("/api/google/settings/save-links", handler.SaveLinks)

	body := `{"inventory_url": "https://docs.google.com/spreadsheets/d/abc"}`
	req, _ := http.NewRequest("POST", "/api/google/settings/save-links", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
	var resp map[string]interface{}
	assert.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	assert.Equal(t, false, resp["success"])
}

func TestSettingsHandler_SaveLinks_InvalidJSON(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	handler := NewSettingsHandler(nil, nil)
	r.Use(func(c *gin.Context) {
		c.Set("tenantID", "test-tenant")
		c.Next()
	})
	r.POST("/api/google/settings/save-links", handler.SaveLinks)

	req, _ := http.NewRequest("POST", "/api/google/settings/save-links", bytes.NewBufferString(`{invalid}`))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	var resp map[string]interface{}
	assert.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	assert.Equal(t, false, resp["success"])
}

// ---- GetSavedLinks handler ----

func TestSettingsHandler_GetSavedLinks_MissingTenantID(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	handler := NewSettingsHandler(nil, nil)
	r.GET("/api/google/settings/saved-links", handler.GetSavedLinks)

	req, _ := http.NewRequest("GET", "/api/google/settings/saved-links", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
	var resp map[string]interface{}
	assert.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	assert.Equal(t, false, resp["success"])
}

// ---- ValidateLink handler ----

func TestSettingsHandler_ValidateLink_MissingTenantID(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	handler := NewSettingsHandler(nil, nil)
	r.POST("/api/google/settings/validate-link", handler.ValidateLink)

	body := `{"spreadsheet_url": "https://docs.google.com/spreadsheets/d/abc", "type": "inventory"}`
	req, _ := http.NewRequest("POST", "/api/google/settings/validate-link", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
	var resp map[string]interface{}
	assert.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	assert.Equal(t, false, resp["success"])
}

func TestSettingsHandler_ValidateLink_InvalidJSON(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	handler := NewSettingsHandler(nil, nil)
	r.Use(func(c *gin.Context) {
		c.Set("tenantID", "test-tenant")
		c.Next()
	})
	r.POST("/api/google/settings/validate-link", handler.ValidateLink)

	req, _ := http.NewRequest("POST", "/api/google/settings/validate-link", bytes.NewBufferString(`{invalid}`))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	var resp map[string]interface{}
	assert.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	assert.Equal(t, false, resp["success"])
}

func TestSettingsHandler_ValidateLink_MissingURL(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	handler := NewSettingsHandler(nil, nil)
	r.Use(func(c *gin.Context) {
		c.Set("tenantID", "test-tenant")
		c.Next()
	})
	r.POST("/api/google/settings/validate-link", handler.ValidateLink)

	// Neither spreadsheet_url nor url provided
	req, _ := http.NewRequest("POST", "/api/google/settings/validate-link", bytes.NewBufferString(`{"type":"inventory"}`))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	var resp map[string]interface{}
	assert.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	assert.Equal(t, false, resp["success"])
}

func TestSettingsHandler_ValidateLink_InvalidType(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	handler := NewSettingsHandler(nil, nil)
	r.Use(func(c *gin.Context) {
		c.Set("tenantID", "test-tenant")
		c.Next()
	})
	r.POST("/api/google/settings/validate-link", handler.ValidateLink)

	body := `{"spreadsheet_url": "https://docs.google.com/spreadsheets/d/abc", "type": "unknown_type"}`
	req, _ := http.NewRequest("POST", "/api/google/settings/validate-link", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	var resp map[string]interface{}
	assert.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	assert.Equal(t, false, resp["success"])
}

func TestSettingsHandler_ValidateLink_NilAuthPanicsWithValidURL(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// extractSpreadsheetID returns the raw input for non-Google-Sheets URLs (no empty-string path).
	// A valid Google Sheets URL passes ID extraction and reaches GetSpreadsheetInfo, which
	// panics when authService is nil. Document this as expected behavior.
	handler := NewSettingsHandler(nil, nil)

	r := gin.New()
	r.Use(func(c *gin.Context) {
		c.Set("tenantID", "test-tenant")
		c.Next()
	})
	r.POST("/api/google/settings/validate-link", handler.ValidateLink)

	body := `{"spreadsheet_url": "https://docs.google.com/spreadsheets/d/FAKE_ID_123/edit", "type": "inventory"}`
	req, _ := http.NewRequest("POST", "/api/google/settings/validate-link", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	// Nil authService causes a panic inside GetSpreadsheetInfo.
	// Production code always provides a real authService; nil is only a test concern.
	assert.Panics(t, func() {
		r.ServeHTTP(w, req)
	})
}
