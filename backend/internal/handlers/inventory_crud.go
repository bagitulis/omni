package handlers

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/omni/backend/internal/models"
	"github.com/omni/backend/internal/services/inventory"
	"github.com/rs/zerolog/log"
	"gorm.io/gorm"
)

// ============================================================================
// CRUD Operations by Key Value
// Added: 2026-01-27 - Fix 404 error on PUT /api/inventory/:keyValue
// Matching Node.js backend: backend-node/src/routes/inventoryDataRoutes.ts
// ============================================================================

// GetRecordByKey handles GET /api/inventory/:keyValue
func (h *InventoryHandler) GetRecordByKey(c *gin.Context) {
	tenantID := c.GetString("tenantID")
	if tenantID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"success": false, "error": "tenant ID required"})
		return
	}

	keyValue := c.Param("keyValue")
	if keyValue == "" {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "keyValue is required"})
		return
	}

	db, err := h.getDB(c)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": err.Error()})
		return
	}

	service := inventory.NewInventoryService(db, tenantID)
	record, err := service.GetByKeyValue(c.Request.Context(), keyValue)

	if err != nil {
		log.Error().Err(err).Str("tenant_id", tenantID).Str("key_value", keyValue).Msg("Failed to get inventory record")
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": err.Error()})
		return
	}
	if record == nil {
		c.JSON(http.StatusNotFound, gin.H{"success": false, "error": "Inventory record not found"})
		return
	}

	// Parse JSONB data
	var data map[string]interface{}
	if record.Data != "" {
		json.Unmarshal([]byte(record.Data), &data)
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"id":              record.ID,
			"key_value":       record.KeyValue,
			"key_column_name": record.KeyColumnName,
			"data":            data,
			"created_at":      record.CreatedAt.Format(time.RFC3339),
			"updated_at":      record.UpdatedAt.Format(time.RFC3339),
		},
	})
}

// UpdateRecordByKey handles PUT /api/inventory/:keyValue
// 🔴 CRITICAL - This fixes the 404 error!
func (h *InventoryHandler) UpdateRecordByKey(c *gin.Context) {
	tenantID := c.GetString("tenantID")
	if tenantID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"success": false, "error": "tenant ID required"})
		return
	}

	keyValue := c.Param("keyValue")
	if keyValue == "" {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "keyValue is required"})
		return
	}

	var requestBody map[string]interface{}
	if err := c.ShouldBindJSON(&requestBody); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "Invalid request body"})
		return
	}

	db, err := h.getDB(c)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": err.Error()})
		return
	}

	service := inventory.NewInventoryService(db, tenantID)
	record, err := service.UpdateByKeyValue(c.Request.Context(), keyValue, requestBody)

	if err == gorm.ErrRecordNotFound {
		c.JSON(http.StatusNotFound, gin.H{"success": false, "error": "Inventory record not found"})
		return
	}
	if err != nil {
		log.Error().Err(err).Str("tenant_id", tenantID).Str("key_value", keyValue).Msg("Failed to update inventory record")
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": err.Error()})
		return
	}

	var data map[string]interface{}
	if record.Data != "" {
		json.Unmarshal([]byte(record.Data), &data)
	}

	log.Info().
		Str("tenant_id", tenantID).
		Str("key_value", keyValue).
		Msg("Updated inventory record")

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"id":              record.ID,
			"key_value":       record.KeyValue,
			"key_column_name": record.KeyColumnName,
			"data":            data,
			"updated_at":      record.UpdatedAt.Format(time.RFC3339),
		},
	})
}

// CreateRecord handles POST /api/inventory
func (h *InventoryHandler) CreateRecord(c *gin.Context) {
	tenantID := c.GetString("tenantID")
	if tenantID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"success": false, "error": "tenant ID required"})
		return
	}

	var requestBody map[string]interface{}
	if err := c.ShouldBindJSON(&requestBody); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "Invalid request body"})
		return
	}

	keyValue, ok := requestBody["keyValue"].(string)
	if !ok || keyValue == "" {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "keyValue is required"})
		return
	}

	db, err := h.getDB(c)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": err.Error()})
		return
	}

	// Get settings to determine key column name
	var settings models.InventorySettings
	db.WithContext(c.Request.Context()).Where("tenant_id = ?", tenantID).First(&settings)
	keyColumnName := settings.KeyColumn
	if keyColumnName == "" {
		keyColumnName = "SKU"
	}

	service := inventory.NewInventoryService(db, tenantID)
	record, err := service.CreateRecord(c.Request.Context(), keyColumnName, keyValue, requestBody)

	if err != nil {
		log.Error().Err(err).Str("tenant_id", tenantID).Str("key_value", keyValue).Msg("Failed to create inventory record")
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": err.Error()})
		return
	}

	var data map[string]interface{}
	if record.Data != "" {
		json.Unmarshal([]byte(record.Data), &data)
	}

	log.Info().
		Str("tenant_id", tenantID).
		Str("key_value", keyValue).
		Msg("Created inventory record")

	c.JSON(http.StatusCreated, gin.H{
		"success": true,
		"data": gin.H{
			"id":              record.ID,
			"key_value":       record.KeyValue,
			"key_column_name": record.KeyColumnName,
			"data":            data,
			"created_at":      record.CreatedAt.Format(time.RFC3339),
		},
	})
}

// DeleteRecordByKey handles DELETE /api/inventory/:keyValue
func (h *InventoryHandler) DeleteRecordByKey(c *gin.Context) {
	tenantID := c.GetString("tenantID")
	if tenantID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"success": false, "error": "tenant ID required"})
		return
	}

	keyValue := c.Param("keyValue")
	if keyValue == "" {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "keyValue is required"})
		return
	}

	db, err := h.getDB(c)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": err.Error()})
		return
	}

	service := inventory.NewInventoryService(db, tenantID)
	err = service.DeleteByKeyValue(c.Request.Context(), keyValue)

	if err == gorm.ErrRecordNotFound {
		c.JSON(http.StatusNotFound, gin.H{"success": false, "error": "Inventory record not found"})
		return
	}
	if err != nil {
		log.Error().Err(err).Str("tenant_id", tenantID).Str("key_value", keyValue).Msg("Failed to delete inventory record")
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": err.Error()})
		return
	}

	log.Info().
		Str("tenant_id", tenantID).
		Str("key_value", keyValue).
		Msg("Deleted inventory record")

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Inventory record deleted",
	})
}
