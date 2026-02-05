package handlers

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

// TestInventoryHandler_GetSelectedColumns_MissingTenant tests GetSelectedColumns without tenant
func TestInventoryHandler_GetSelectedColumns_MissingTenant(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	handler := NewInventoryHandler(nil)
	r.GET("/api/inventory/columns/selected", handler.GetSelectedColumns)

	req, _ := http.NewRequest("GET", "/api/inventory/columns/selected", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)

	var resp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.False(t, resp["success"].(bool))
}

// TestInventoryHandler_UpdateSelectedColumns_MissingTenant tests UpdateSelectedColumns without tenant
func TestInventoryHandler_UpdateSelectedColumns_MissingTenant(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	handler := NewInventoryHandler(nil)
	r.POST("/api/inventory/columns/selected", handler.UpdateSelectedColumns)

	req, _ := http.NewRequest("POST", "/api/inventory/columns/selected", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

// TestInventoryHandler_GetAvailableColumns_MissingTenant tests GetAvailableColumns without tenant
func TestInventoryHandler_GetAvailableColumns_MissingTenant(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	handler := NewInventoryHandler(nil)
	r.GET("/api/inventory/columns/available", handler.GetAvailableColumns)

	req, _ := http.NewRequest("GET", "/api/inventory/columns/available", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

// TestColumnIndexToLetter tests column index to letter conversion
func TestColumnIndexToLetter_SingleLetter(t *testing.T) {
	assert.Equal(t, "A", columnIndexToLetter(0))
	assert.Equal(t, "B", columnIndexToLetter(1))
	assert.Equal(t, "Z", columnIndexToLetter(25))
}

// TestColumnIndexToLetter_DoubleLetter tests double letter columns
func TestColumnIndexToLetter_DoubleLetter(t *testing.T) {
	assert.Equal(t, "AA", columnIndexToLetter(26))
	assert.Equal(t, "AB", columnIndexToLetter(27))
	assert.Equal(t, "AZ", columnIndexToLetter(51))
	assert.Equal(t, "BA", columnIndexToLetter(52))
}

// TestSortColumnsAlphabetically tests column sorting with priorities
func TestSortColumnsAlphabetically_Basic(t *testing.T) {
	columns := []string{"ZZZ", "AAA", "BBB"}
	sortColumnsAlphabetically(columns)
	assert.Equal(t, "AAA", columns[0])
	assert.Equal(t, "BBB", columns[1])
	assert.Equal(t, "ZZZ", columns[2])
}

// TestSortColumnsAlphabetically_WithPriority tests that priority columns come first
func TestSortColumnsAlphabetically_WithPriority(t *testing.T) {
	columns := []string{"ZZZ", "TOTAL", "Masuk", "AAA"}
	sortColumnsAlphabetically(columns)
	// Priority columns should be first (Masuk has priority 1, TOTAL has priority 2)
	assert.Equal(t, "Masuk", columns[0])
	assert.Equal(t, "TOTAL", columns[1])
}

// TestAvailableColumn_Structure tests AvailableColumn struct
func TestAvailableColumn_Structure(t *testing.T) {
	col := AvailableColumn{
		Name:           "SKU",
		SpreadsheetCol: "A",
		Type:           "text",
		Position:       1,
	}

	assert.Equal(t, "SKU", col.Name)
	assert.Equal(t, "A", col.SpreadsheetCol)
	assert.Equal(t, "text", col.Type)
	assert.Equal(t, 1, col.Position)
}
