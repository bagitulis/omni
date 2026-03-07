package handlers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/glebarez/sqlite"
	"github.com/omni/backend/internal/config"
	"github.com/omni/backend/internal/models"
	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"
)

func TestInventoryHandler_UpdateConfig(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("missing tenant ID returns 400", func(t *testing.T) {
		handler := NewInventoryHandler(nil)

		body := map[string]interface{}{"key_column": "TOTAL"}
		bodyBytes, _ := json.Marshal(body)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodPut, "/api/inventory/config", bytes.NewReader(bodyBytes))
		c.Request.Header.Set("Content-Type", "application/json")

		handler.UpdateConfig(c)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		var resp map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &resp)
		assert.Equal(t, false, resp["success"])
		assert.Contains(t, resp["error"], "Missing tenantId")
	})

	t.Run("invalid selected columns type returns 400", func(t *testing.T) {
		handler := NewInventoryHandler(nil)

		body := map[string]interface{}{"selected_columns": 123}
		bodyBytes, _ := json.Marshal(body)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodPut, "/api/inventory/config", bytes.NewReader(bodyBytes))
		c.Request.Header.Set("Content-Type", "application/json")
		c.Set("tenant_id", "test-tenant")

		handler.UpdateConfig(c)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		var resp map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &resp)
		assert.Equal(t, false, resp["success"])
		assert.Contains(t, resp["error"], "selected_columns and all_columns")
	})

	t.Run("valid request with no DB returns error", func(t *testing.T) {
		handler := NewInventoryHandler(nil)

		body := map[string]interface{}{"key_column": "TOTAL", "selected_columns": "TOTAL,AUTO"}
		bodyBytes, _ := json.Marshal(body)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodPut, "/api/inventory/config", bytes.NewReader(bodyBytes))
		c.Request.Header.Set("Content-Type", "application/json")
		c.Set("tenant_id", "test-tenant")

		handler.UpdateConfig(c)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
	})
}

func TestInventoryHandler_UpdateConfig_DoesNotInsertDuplicateEmptyPrimaryKey(t *testing.T) {
	gin.SetMode(gin.TestMode)
	config.SetDatabaseDriver(config.DBDriver("sqlite"), nil)
	defer config.SetDatabaseDriver(config.DriverPostgres, nil)

	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})
	if err != nil {
		t.Fatalf("failed to open sqlite: %v", err)
	}

	if err := db.AutoMigrate(&models.InventorySettings{}); err != nil {
		t.Fatalf("failed to migrate inventory settings: %v", err)
	}

	existing := models.InventorySettings{
		ID:              "",
		TenantID:        "test-tenant",
		SelectedColumns: `"OLD"`,
		HeaderRow:       1,
		DataStartRow:    2,
		SyncIntervalSec: 300,
	}
	if err := db.Create(&existing).Error; err != nil {
		t.Fatalf("failed to seed existing settings: %v", err)
	}

	handler := NewInventoryHandler(db)
	body := map[string]interface{}{
		"key_column":       "TOTAL",
		"selected_columns": []string{"TOTAL", "AUTO"},
	}
	bodyBytes, _ := json.Marshal(body)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPut, "/api/inventory/config", bytes.NewReader(bodyBytes))
	c.Request.Header.Set("Content-Type", "application/json")
	c.Set("tenant_id", "test-tenant")
	c.Set("tenantID", "test-tenant")

	handler.UpdateConfig(c)

	assert.Equal(t, http.StatusOK, w.Code)

	var count int64
	err = db.Model(&models.InventorySettings{}).Where("tenant_id = ?", "test-tenant").Count(&count).Error
	if err != nil {
		t.Fatalf("failed to count settings rows: %v", err)
	}
	assert.Equal(t, int64(1), count)

	var updated models.InventorySettings
	err = db.Where("tenant_id = ?", "test-tenant").First(&updated).Error
	if err != nil {
		t.Fatalf("failed to load updated settings: %v", err)
	}
	assert.Equal(t, "TOTAL", updated.KeyColumn)
	assert.Equal(t, `["TOTAL","AUTO"]`, updated.SelectedColumns)
}
