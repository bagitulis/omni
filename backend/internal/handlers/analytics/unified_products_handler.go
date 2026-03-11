package analytics

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/omni/backend/internal/config"
	"github.com/omni/backend/internal/dto/response"
	"github.com/omni/backend/internal/middleware"
)

// GetClassifiedProducts handles GET /api/analytics/products/classified
// Returns products grouped by recommended action
// Action is calculated based on ROAS thresholds:
// - ROAS >= 5: SCALE_UP (high performers worth scaling)
// - ROAS >= 2: MAINTAIN (profitable, keep as is)
// - ROAS >= 1: REDUCE (break-even, reduce budget)
// - ROAS < 1: STOP (losing money, stop immediately)
func (h *UnifiedHandler) GetClassifiedProducts(c *gin.Context) {
	tenantID := middleware.GetTenantID(c)
	if tenantID == "" {
		c.JSON(http.StatusUnauthorized, response.Error("Missing tenant_id"))
		return
	}

	db, err := config.GetTenantDB(tenantID, h.basePath)
	if err != nil {
		c.JSON(http.StatusInternalServerError, response.Error("Database connection failed"))
		return
	}

	ctx := c.Request.Context()

	// Get products from MV - only those with spend (total_cost > 0)
	var products []struct {
		ProductID    string  `gorm:"column:product_id"`
		ProductName  string  `gorm:"column:product_name"`
		TotalCost    float64 `gorm:"column:total_cost"`
		TotalRevenue float64 `gorm:"column:total_revenue"`
		Roas         float64 `gorm:"column:roas"`
		TotalOrders  int64   `gorm:"column:total_orders"`
		Ctr          float64 `gorm:"column:ctr"`
	}

	db.WithContext(ctx).Table("mv_ml_product_analysis").
		Where("tenant_id = ? AND total_cost > 0", tenantID).
		Order("total_revenue DESC").
		Find(&products)

	// Group by action based on ROAS thresholds
	scaleUp := make([]gin.H, 0)
	maintain := make([]gin.H, 0)
	reduce := make([]gin.H, 0)
	stop := make([]gin.H, 0)

	for _, p := range products {
		item, action := h.classifyProduct(p.ProductID, p.ProductName, p.TotalCost,
			p.TotalRevenue, p.Roas, p.TotalOrders, p.Ctr)

		switch action {
		case "SCALE_UP":
			scaleUp = append(scaleUp, item)
		case "MAINTAIN":
			maintain = append(maintain, item)
		case "REDUCE":
			reduce = append(reduce, item)
		case "STOP":
			stop = append(stop, item)
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"scale_up": scaleUp,
			"maintain": maintain,
			"reduce":   reduce,
			"stop":     stop,
		},
		"counts": gin.H{
			"scale_up": len(scaleUp),
			"maintain": len(maintain),
			"reduce":   len(reduce),
			"stop":     len(stop),
		},
	})
}

// classifyProduct determines the action for a product based on ROAS
func (h *UnifiedHandler) classifyProduct(
	productID, productName string,
	totalCost, totalRevenue, roas float64,
	totalOrders int64, ctr float64,
) (gin.H, string) {
	var action string
	var actionLabel string
	var recommendation string

	switch {
	case roas >= 5:
		action = "SCALE_UP"
		actionLabel = "Scale Up"
		recommendation = "Produk sangat menguntungkan. Pertimbangkan untuk meningkatkan budget."
	case roas >= 2:
		action = "MAINTAIN"
		actionLabel = "Maintain"
		recommendation = "Produk profitable. Pertahankan budget saat ini."
	case roas >= 1:
		action = "REDUCE"
		actionLabel = "Reduce"
		recommendation = "Produk break-even. Kurangi budget atau optimalkan."
	default:
		action = "STOP"
		actionLabel = "Stop"
		recommendation = "Produk merugi. Hentikan iklan atau evaluasi ulang."
	}

	item := gin.H{
		"product_id":     productID,
		"product_name":   productName,
		"total_cost":     totalCost,
		"total_revenue":  totalRevenue,
		"roas":           roas,
		"total_orders":   totalOrders,
		"ctr":            ctr,
		"action":         action,
		"action_label":   actionLabel,
		"recommendation": recommendation,
	}

	return item, action
}

// GetTopProducts handles GET /api/analytics/products/top
func (h *UnifiedHandler) GetTopProducts(c *gin.Context) {
	tenantID := middleware.GetTenantID(c)
	if tenantID == "" {
		c.JSON(http.StatusUnauthorized, response.Error("Missing tenant_id"))
		return
	}

	db, err := config.GetTenantDB(tenantID, h.basePath)
	if err != nil {
		c.JSON(http.StatusInternalServerError, response.Error("Database connection failed"))
		return
	}

	ctx := c.Request.Context()
	limit := 10

	// Get top products by revenue from MV
	var topProducts []struct {
		ProductID   string  `gorm:"column:product_id"`
		ProductName string  `gorm:"column:product_name"`
		Revenue     float64 `gorm:"column:total_revenue"`
		Cost        float64 `gorm:"column:total_cost"`
		Roas        float64 `gorm:"column:roas"`
	}

	db.WithContext(ctx).Table("mv_ml_product_analysis").
		Where("tenant_id = ?", tenantID).
		Order("total_revenue DESC").
		Limit(limit).
		Find(&topProducts)

	// Format response
	products := make([]gin.H, 0, len(topProducts))
	for _, p := range topProducts {
		products = append(products, gin.H{
			"product_id":   p.ProductID,
			"product_name": p.ProductName,
			"revenue":      p.Revenue,
			"cost":         p.Cost,
			"roas":         p.Roas,
			"source":       "tiktok",
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    products,
	})
}
