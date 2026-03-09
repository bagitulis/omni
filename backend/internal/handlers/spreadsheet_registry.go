package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/omni/backend/internal/dto/response"
	"github.com/omni/backend/internal/services/spreadsheet"
	"gorm.io/gorm"
)

// SpreadsheetRegistryHandler handles spreadsheet registry endpoints
type SpreadsheetRegistryHandler struct {
	fallbackDB *gorm.DB
}

// NewSpreadsheetRegistryHandler creates a new handler
func NewSpreadsheetRegistryHandler(db *gorm.DB) *SpreadsheetRegistryHandler {
	return &SpreadsheetRegistryHandler{fallbackDB: db}
}

// getDB returns the appropriate database for the current request
func (h *SpreadsheetRegistryHandler) getDB(c *gin.Context) (*gorm.DB, error) {
	return GetTenantDBFromContext(c, h.fallbackDB)
}

// List handles GET /api/spreadsheets
func (h *SpreadsheetRegistryHandler) List(c *gin.Context) {
	tenantID := c.GetString("tenantID")
	if tenantID == "" {
		c.JSON(http.StatusUnauthorized, response.Error("Missing tenantId"))
		return
	}

	db, err := h.getDB(c)
	if err != nil {
		c.JSON(http.StatusInternalServerError, response.Error(err.Error()))
		return
	}

	svc := spreadsheet.NewRegistryService(db, tenantID)
	filter := spreadsheet.ListFilter{
		Purpose:    c.Query("purpose"),
		ActiveOnly: c.Query("active") == "true",
	}

	spreadsheets, err := svc.List(filter)
	if err != nil {
		c.JSON(http.StatusInternalServerError, response.Error(err.Error()))
		return
	}

	c.JSON(http.StatusOK, response.Success(gin.H{"spreadsheets": spreadsheets}))
}

// Get handles GET /api/spreadsheets/:id
func (h *SpreadsheetRegistryHandler) Get(c *gin.Context) {
	tenantID := c.GetString("tenantID")
	if tenantID == "" {
		c.JSON(http.StatusUnauthorized, response.Error("Missing tenantId"))
		return
	}

	db, err := h.getDB(c)
	if err != nil {
		c.JSON(http.StatusInternalServerError, response.Error(err.Error()))
		return
	}

	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, response.Error("invalid ID"))
		return
	}

	svc := spreadsheet.NewRegistryService(db, tenantID)
	sp, err := svc.Get(uint(id))
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, response.Error("spreadsheet not found"))
			return
		}
		c.JSON(http.StatusInternalServerError, response.Error(err.Error()))
		return
	}

	c.JSON(http.StatusOK, response.Success(sp))
}

// Register handles POST /api/spreadsheets
func (h *SpreadsheetRegistryHandler) Register(c *gin.Context) {
	tenantID := c.GetString("tenantID")
	if tenantID == "" {
		c.JSON(http.StatusUnauthorized, response.Error("Missing tenantId"))
		return
	}

	db, err := h.getDB(c)
	if err != nil {
		c.JSON(http.StatusInternalServerError, response.Error(err.Error()))
		return
	}

	var req spreadsheet.RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, response.Error(err.Error()))
		return
	}

	svc := spreadsheet.NewRegistryService(db, tenantID)
	sp, err := svc.Register(req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, response.Error(err.Error()))
		return
	}

	c.JSON(http.StatusCreated, response.Success(sp))
}

// Update handles PUT /api/spreadsheets/:id
func (h *SpreadsheetRegistryHandler) Update(c *gin.Context) {
	tenantID := c.GetString("tenantID")
	if tenantID == "" {
		c.JSON(http.StatusUnauthorized, response.Error("Missing tenantId"))
		return
	}

	db, err := h.getDB(c)
	if err != nil {
		c.JSON(http.StatusInternalServerError, response.Error(err.Error()))
		return
	}

	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, response.Error("invalid ID"))
		return
	}

	var req spreadsheet.UpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, response.Error(err.Error()))
		return
	}

	svc := spreadsheet.NewRegistryService(db, tenantID)
	sp, err := svc.Update(uint(id), req)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, response.Error("spreadsheet not found"))
			return
		}
		c.JSON(http.StatusInternalServerError, response.Error(err.Error()))
		return
	}

	c.JSON(http.StatusOK, response.Success(sp))
}

// Delete handles DELETE /api/spreadsheets/:id
func (h *SpreadsheetRegistryHandler) Delete(c *gin.Context) {
	tenantID := c.GetString("tenantID")
	if tenantID == "" {
		c.JSON(http.StatusUnauthorized, response.Error("Missing tenantId"))
		return
	}

	db, err := h.getDB(c)
	if err != nil {
		c.JSON(http.StatusInternalServerError, response.Error(err.Error()))
		return
	}

	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, response.Error("invalid ID"))
		return
	}

	svc := spreadsheet.NewRegistryService(db, tenantID)
	if err := svc.Delete(uint(id)); err != nil {
		c.JSON(http.StatusInternalServerError, response.Error(err.Error()))
		return
	}

	c.JSON(http.StatusOK, response.Success(gin.H{"message": "spreadsheet deleted"}))
}

// MarkSynced handles POST /api/spreadsheets/:id/synced
func (h *SpreadsheetRegistryHandler) MarkSynced(c *gin.Context) {
	tenantID := c.GetString("tenantID")
	if tenantID == "" {
		c.JSON(http.StatusUnauthorized, response.Error("Missing tenantId"))
		return
	}

	db, err := h.getDB(c)
	if err != nil {
		c.JSON(http.StatusInternalServerError, response.Error(err.Error()))
		return
	}

	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, response.Error("invalid ID"))
		return
	}

	svc := spreadsheet.NewRegistryService(db, tenantID)
	if err := svc.UpdateLastSync(uint(id)); err != nil {
		c.JSON(http.StatusInternalServerError, response.Error(err.Error()))
		return
	}

	c.JSON(http.StatusOK, response.Success(gin.H{"message": "sync time updated"}))
}
