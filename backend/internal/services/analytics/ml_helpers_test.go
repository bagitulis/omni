package analytics

import (
	"testing"

	"github.com/omni/backend/internal/models"
	"github.com/stretchr/testify/assert"
)

func TestApplyFilters(t *testing.T) {
	service := &MLAnalyticsService{}
	products := []models.MLProductAnalysis{
		{ProductID: "p1", Category: "STAR", Action: "SCALE_UP"},
		{ProductID: "p2", Category: "WATCH", Action: "MONITOR"},
		{ProductID: "p3", Category: "STAR", Action: "MONITOR"},
	}

	all := service.applyFilters(products, "", "")
	assert.Len(t, all, 3)

	starOnly := service.applyFilters(products, "STAR", "")
	assert.Len(t, starOnly, 2)
	assert.Equal(t, "p1", starOnly[0].ProductID)
	assert.Equal(t, "p3", starOnly[1].ProductID)

	monitorOnly := service.applyFilters(products, "", "MONITOR")
	assert.Len(t, monitorOnly, 2)
	assert.Equal(t, "p2", monitorOnly[0].ProductID)
	assert.Equal(t, "p3", monitorOnly[1].ProductID)

	combined := service.applyFilters(products, "STAR", "MONITOR")
	assert.Len(t, combined, 1)
	assert.Equal(t, "p3", combined[0].ProductID)

	none := service.applyFilters(products, "PROBLEM", "STOP")
	assert.Empty(t, none)
}

func TestSortProducts(t *testing.T) {
	service := &MLAnalyticsService{}
	products := []models.MLProductAnalysis{
		{ProductID: "p1", UnifiedScore: 50, ROAS: 2.0, TotalRevenue: 100, TotalProfit: 40, TotalCost: 60},
		{ProductID: "p2", UnifiedScore: 70, ROAS: 1.2, TotalRevenue: 200, TotalProfit: 90, TotalCost: 110},
		{ProductID: "p3", UnifiedScore: 30, ROAS: 3.5, TotalRevenue: 150, TotalProfit: 80, TotalCost: 70},
	}

	service.sortProducts(products, "unified_score", "desc")
	assert.Equal(t, []string{"p2", "p1", "p3"}, []string{products[0].ProductID, products[1].ProductID, products[2].ProductID})

	service.sortProducts(products, "roas", "asc")
	assert.Equal(t, []string{"p2", "p1", "p3"}, []string{products[0].ProductID, products[1].ProductID, products[2].ProductID})

	service.sortProducts(products, "revenue", "desc")
	assert.Equal(t, []string{"p2", "p3", "p1"}, []string{products[0].ProductID, products[1].ProductID, products[2].ProductID})

	service.sortProducts(products, "profit", "asc")
	assert.Equal(t, []string{"p1", "p3", "p2"}, []string{products[0].ProductID, products[1].ProductID, products[2].ProductID})

	service.sortProducts(products, "cost", "asc")
	assert.Equal(t, []string{"p1", "p3", "p2"}, []string{products[0].ProductID, products[1].ProductID, products[2].ProductID})

	service.sortProducts(products, "unknown", "asc")
	assert.Equal(t, []string{"p3", "p1", "p2"}, []string{products[0].ProductID, products[1].ProductID, products[2].ProductID})
}
