package handlers

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

// TestSettingsHandler_GetInventorySettings_MissingTenant tests GetInventorySettings without tenant
func TestSettingsHandler_GetInventorySettings_MissingTenant(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	handler := NewSettingsHandler(nil)
	r.GET("/api/settings/inventory", handler.GetInventorySettings)

	req, _ := http.NewRequest("GET", "/api/settings/inventory", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)

	var resp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.False(t, resp["success"].(bool))
}

// TestSettingsHandler_UpdateInventorySettings_MissingTenant tests UpdateInventorySettings without tenant
func TestSettingsHandler_UpdateInventorySettings_MissingTenant(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	handler := NewSettingsHandler(nil)
	r.PUT("/api/settings/inventory", handler.UpdateInventorySettings)

	req, _ := http.NewRequest("PUT", "/api/settings/inventory", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

// TestSettingsHandler_GetGoogleSheetsSettings_MissingTenant tests GetGoogleSheetsSettings without tenant
func TestSettingsHandler_GetGoogleSheetsSettings_MissingTenant(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	handler := NewSettingsHandler(nil)
	r.GET("/api/settings/google-sheets", handler.GetGoogleSheetsSettings)

	req, _ := http.NewRequest("GET", "/api/settings/google-sheets", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

// TestSettingsHandler_UpdateGoogleSheetsSettings_MissingTenant tests UpdateGoogleSheetsSettings without tenant
func TestSettingsHandler_UpdateGoogleSheetsSettings_MissingTenant(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	handler := NewSettingsHandler(nil)
	r.PUT("/api/settings/google-sheets", handler.UpdateGoogleSheetsSettings)

	req, _ := http.NewRequest("PUT", "/api/settings/google-sheets", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

// TestInventorySettingsRequest_Structure tests InventorySettingsRequest struct
func TestInventorySettingsRequest_Structure(t *testing.T) {
	req := InventorySettingsRequest{
		SpreadsheetID: "sheet-123",
		SheetName:     "Inventory",
		AutoSync:      true,
		SyncInterval:  300,
	}

	assert.Equal(t, "sheet-123", req.SpreadsheetID)
	assert.Equal(t, "Inventory", req.SheetName)
	assert.True(t, req.AutoSync)
	assert.Equal(t, 300, req.SyncInterval)
}

// TestGoogleSheetsSettingsRequest_Structure tests GoogleSheetsSettingsRequest struct
func TestGoogleSheetsSettingsRequest_Structure(t *testing.T) {
	req := GoogleSheetsSettingsRequest{
		ServiceAccountEmail:  "test@project.iam.gserviceaccount.com",
		DefaultSpreadsheetID: "sheet-abc",
		IsConnected:          true,
	}

	assert.Equal(t, "test@project.iam.gserviceaccount.com", req.ServiceAccountEmail)
	assert.Equal(t, "sheet-abc", req.DefaultSpreadsheetID)
	assert.True(t, req.IsConnected)
}

// TestGenerateID_Unique tests that generateID produces unique IDs
func TestGenerateID_Unique(t *testing.T) {
	id1 := generateID()
	id2 := generateID()

	assert.NotEmpty(t, id1)
	assert.NotEmpty(t, id2)
	assert.NotEqual(t, id1, id2)
}

// TestSettingsHandler_GetGeneralSettings_MissingTenant tests GetGeneralSettings without tenant
func TestSettingsHandler_GetGeneralSettings_MissingTenant(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	handler := NewSettingsHandler(nil)
	r.GET("/api/settings/general", handler.GetGeneralSettings)

	req, _ := http.NewRequest("GET", "/api/settings/general", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var resp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.False(t, resp["success"].(bool))
}

// TestSettingsHandler_UpdateGeneralSettings_MissingTenant tests UpdateGeneralSettings without tenant
func TestSettingsHandler_UpdateGeneralSettings_MissingTenant(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	handler := NewSettingsHandler(nil)
	r.POST("/api/settings/general", handler.UpdateGeneralSettings)

	req, _ := http.NewRequest("POST", "/api/settings/general", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

// TestGeneralSettingsRequest_Structure tests GeneralSettingsRequest struct
func TestGeneralSettingsRequest_Structure(t *testing.T) {
	req := GeneralSettingsRequest{
		Language:             "en",
		Timezone:             "Asia/Jakarta",
		NotificationsEmail:   true,
		NotificationsBrowser: true,
		AutoSync:             true,
		SyncInterval:         "30",
	}

	assert.Equal(t, "en", req.Language)
	assert.Equal(t, "Asia/Jakarta", req.Timezone)
	assert.True(t, req.NotificationsEmail)
	assert.True(t, req.NotificationsBrowser)
	assert.True(t, req.AutoSync)
	assert.Equal(t, "30", req.SyncInterval)
}
