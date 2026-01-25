package handlers

import "github.com/gin-gonic/gin"

// SyncByCategoryBody represents the request body for sync by category
type SyncByCategoryBody struct {
	Days int `json:"days"`
}

// NotInitializedSyncResponse returns response for uninitialized sync service
func NotInitializedSyncResponse(category string, days int) gin.H {
	platformError := gin.H{
		"success": false,
		"count":   0,
		"error":   "Platform not configured - complete OAuth setup first",
	}
	return gin.H{
		"success":  true,
		"data":     gin.H{"shopee": platformError, "lazada": platformError, "tiktok": platformError},
		"message":  "Sync requires platform OAuth configuration",
		"category": category,
		"days":     days,
	}
}
