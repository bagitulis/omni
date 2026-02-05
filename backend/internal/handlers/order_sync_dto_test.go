package handlers

import (
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

// TestNotInitializedSyncResponse tests the helper function
func TestNotInitializedSyncResponse(t *testing.T) {
	resp := NotInitializedSyncResponse("unpaid", 7)

	// Verify structure
	assert.Equal(t, true, resp["success"])
	assert.Equal(t, "unpaid", resp["category"])
	assert.Equal(t, 7, resp["days"])
	assert.Contains(t, resp["message"].(string), "OAuth")

	// Verify platform errors
	data := resp["data"].(gin.H)
	shopeeErr := data["shopee"].(gin.H)
	assert.Equal(t, false, shopeeErr["success"])
	assert.Equal(t, 0, shopeeErr["count"])

	lazadaErr := data["lazada"].(gin.H)
	assert.Equal(t, false, lazadaErr["success"])

	tiktokErr := data["tiktok"].(gin.H)
	assert.Equal(t, false, tiktokErr["success"])
}

// TestSyncByCategoryBody_Structure tests SyncByCategoryBody struct
func TestSyncByCategoryBody_Structure(t *testing.T) {
	body := SyncByCategoryBody{
		Days: 14,
	}
	assert.Equal(t, 14, body.Days)
}
