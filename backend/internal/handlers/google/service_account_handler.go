// Package google handles Google API related endpoints
package google

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/omni/backend/internal/services/google"
)

// ServiceAccountHandler handles Google service account management
type ServiceAccountHandler struct {
	authService *google.AuthService
}

// NewServiceAccountHandler creates a new service account handler
func NewServiceAccountHandler(authService *google.AuthService) *ServiceAccountHandler {
	return &ServiceAccountHandler{authService: authService}
}

// ListAccounts handles GET /api/google/service-accounts
// Returns list of available service accounts
func (h *ServiceAccountHandler) ListAccounts(c *gin.Context) {
	accounts := h.authService.ListServiceAccounts()

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    accounts,
	})
}

// GetStats handles GET /api/google/service-accounts/stats
// Returns detailed statistics for all service accounts
func (h *ServiceAccountHandler) GetStats(c *gin.Context) {
	stats := h.authService.GetServiceAccountStats()

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    stats,
	})
}

// SwitchAccount handles POST /api/google/service-accounts/switch
// Switches to a different service account
func (h *ServiceAccountHandler) SwitchAccount(c *gin.Context) {
	var req struct {
		AccountID string `json:"account_id" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	if err := h.authService.SwitchServiceAccount(req.AccountID); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Switched to service account: " + req.AccountID,
	})
}

// DetectSheet handles POST /api/google/service-accounts/detect-sheet
// Detects sheet type from URL
func (h *ServiceAccountHandler) DetectSheet(c *gin.Context) {
	tenantID := c.GetString("tenantID")
	if tenantID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{
			"success": false,
			"error":   "Missing tenantId",
		})
		return
	}

	var req struct {
		URL string `json:"url" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	spreadsheetID := extractSpreadsheetID(req.URL)

	sheetsService := google.NewSheetsService(h.authService, tenantID)
	info, err := sheetsService.GetSpreadsheetInfo(c.Request.Context(), spreadsheetID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	// Detect sheet type based on content
	detectedType := detectSheetType(info)

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"spreadsheet_id": info.ID,
			"title":          info.Title,
			"detected_type":  detectedType,
			"sheets":         info.Sheets,
		},
	})
}

// detectSheetType attempts to detect the type of spreadsheet based on title/content
func detectSheetType(info *google.SpreadsheetInfo) string {
	title := info.Title

	// Check title for common keywords
	keywords := map[string][]string{
		"inventory": {"inventory", "stok", "stock", "persediaan"},
		"wallet":    {"wallet", "dompet", "transaksi", "transaction"},
		"shipping":  {"shipping", "ongkir", "pengiriman", "ekspedisi"},
		"order":     {"order", "pesanan", "penjualan", "sales"},
	}

	for sheetType, kws := range keywords {
		for _, kw := range kws {
			if containsIgnoreCase(title, kw) {
				return sheetType
			}
		}
	}

	return "unknown"
}

// containsIgnoreCase checks if s contains substr (case-insensitive)
func containsIgnoreCase(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr ||
		len(s) > 0 && len(substr) > 0 &&
			(s[0]|0x20 == substr[0]|0x20) && containsIgnoreCase(s[1:], substr[1:]) ||
		len(s) > 0 && containsIgnoreCase(s[1:], substr))
}
