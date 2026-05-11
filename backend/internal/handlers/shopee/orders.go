package shopee

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/omni/backend/internal/config"
	"github.com/omni/backend/internal/dto/response"
	"github.com/omni/backend/internal/middleware"
	"github.com/omni/backend/internal/services/shopee"
	"github.com/omni/backend/internal/utils"
)

// OrderHandler handles Shopee order HTTP requests
type OrderHandler struct {
	basePath string
}

// NewOrderHandler creates a new order handler
func NewOrderHandler(basePath string) *OrderHandler {
	return &OrderHandler{basePath: basePath}
}

// GetOrders handles GET /api/shopee/orders
func (h *OrderHandler) GetOrders(c *gin.Context) {
	tenantID := middleware.GetTenantID(c)
	if tenantID == "" {
		c.JSON(http.StatusUnauthorized, response.Error("Missing tenant_id"))
		return
	}

	// Parse pagination
	page, pageSize, err := utils.ValidatePagination(c.DefaultQuery("page", ""), c.DefaultQuery("pageSize", ""))
	if err != nil {
		c.JSON(http.StatusBadRequest, response.Error(err.Error()))
		return
	}

	// Get database connection
	db, err := config.GetTenantDB(tenantID, h.basePath)
	if err != nil {
		c.JSON(http.StatusInternalServerError, response.Error("Database connection failed"))
		return
	}

	// Use Service instead of Repo directly
	service := shopee.NewOrderService(db)
	orders, total, err := service.GetOrders(c.Request.Context(), page, pageSize)
	if err != nil {
		c.JSON(http.StatusInternalServerError, response.Error("Failed to fetch orders: "+err.Error()))
		return
	}

	// Calculate pagination
	totalPages := int(total) / pageSize
	if int(total)%pageSize > 0 {
		totalPages++
	}

	c.JSON(http.StatusOK, response.SuccessWithMeta(orders, &response.Meta{
		Total:      int(total),
		Page:       page,
		PageSize:   pageSize,
		TotalPages: totalPages,
	}))
}

// GetOrderByID handles GET /api/shopee/orders/:orderSn
func (h *OrderHandler) GetOrderByID(c *gin.Context) {
	tenantID := middleware.GetTenantID(c)
	if tenantID == "" {
		c.JSON(http.StatusUnauthorized, response.Error("Missing tenant_id"))
		return
	}

	orderSN := c.Param("orderSn")
	if orderSN == "" {
		c.JSON(http.StatusBadRequest, response.Error("Order SN required"))
		return
	}

	db, err := config.GetTenantDB(tenantID, h.basePath)
	if err != nil {
		c.JSON(http.StatusInternalServerError, response.Error("Database connection failed"))
		return
	}

	service := shopee.NewOrderService(db)
	order, err := service.GetOrderBySN(c.Request.Context(), orderSN)
	if err != nil {
		c.JSON(http.StatusNotFound, response.Error("Order not found"))
		return
	}

	c.JSON(http.StatusOK, response.Success(order))
}
