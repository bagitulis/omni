package analytics

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// tiktok_ads_helpers.go contains only unexported methods on *TiktokAdsHandler.
// They are exercised through GetDashboard in tiktok_ads_dashboard_test.go.
// This file verifies the constructor creates a valid handler to be used as receiver.

func TestTiktokAdsHandler_HelpersViaCtor(t *testing.T) {
	t.Run("NewTiktokAdsHandler returns non-nil handler", func(t *testing.T) {
		handler := NewTiktokAdsHandler("/data")
		assert.NotNil(t, handler)
		assert.Equal(t, "/data", handler.basePath)
	})
}
