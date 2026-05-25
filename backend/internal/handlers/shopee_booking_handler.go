package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/omni/backend/internal/middleware"
	"github.com/omni/backend/internal/services/sync"
)

// ShopeeBookingHandler handles Shopee booking order endpoints
type ShopeeBookingHandler struct{}

// NewShopeeBookingHandler creates a new Shopee booking handler
func NewShopeeBookingHandler() *ShopeeBookingHandler {
	return &ShopeeBookingHandler{}
}

// SyncBookingOrders syncs booking orders from Shopee API
func (h *ShopeeBookingHandler) SyncBookingOrders(c *gin.Context) {
	tenantID := middleware.GetTenantID(c)
	if tenantID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{
			"success": false,
			"error":   "Missing tenant_id",
		})
		return
	}

	days := 7
	if daysQuery := c.Query("days"); daysQuery != "" {
		if parsed, err := strconv.Atoi(daysQuery); err == nil {
			days = parsed
		}
	}

	bookingService := sync.GetBookingSyncService(tenantID, nil)
	if bookingService == nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Failed to create booking sync service",
		})
		return
	}

	result := bookingService.SyncBookings(c.Request.Context(), days)
	if !result.Success && len(result.Errors) > 0 {
		c.JSON(http.StatusOK, gin.H{
			"success":  false,
			"code":     "PARTIAL_SYNC_FAILURE",
			"error":    "One or more booking sync operations failed",
			"data":     result,
			"category": "booking",
			"days":     days,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success":  true,
		"data":     result,
		"category": "booking",
		"days":     days,
	})
}

// GetBookingOrders lists booking orders with pagination and filters
func (h *ShopeeBookingHandler) GetBookingOrders(c *gin.Context) {
	tenantID := middleware.GetTenantID(c)
	if tenantID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{
			"success": false,
			"error":   "Missing tenant_id",
		})
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "50"))
	search := c.Query("search")
	bookingStatus := c.Query("booking_status")
	matchStatus := c.Query("match_status")
	platform := c.Query("platform")

	if platform != "" && platform != "shopee" {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Booking orders are only available for Shopee",
		})
		return
	}

	bookingService := sync.GetBookingSyncService(tenantID, nil)
	if bookingService == nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Failed to create booking sync service",
		})
		return
	}

	bookings, total, err := bookingService.ListBookings(c.Request.Context(), sync.BookingListParams{
		Page:          page,
		PageSize:      pageSize,
		Search:        search,
		BookingStatus: bookingStatus,
		MatchStatus:   matchStatus,
		Platform:      platform,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success":   true,
		"data":      bookings,
		"count":     total,
		"page":      page,
		"page_size": pageSize,
	})
}

// GetBookingOrderDetail gets a single booking order detail by booking_sn
func (h *ShopeeBookingHandler) GetBookingOrderDetail(c *gin.Context) {
	tenantID := middleware.GetTenantID(c)
	if tenantID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{
			"success": false,
			"error":   "Missing tenant_id",
		})
		return
	}

	bookingSN := c.Param("booking_sn")
	if bookingSN == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Missing booking_sn",
		})
		return
	}

	bookingService := sync.GetBookingSyncService(tenantID, nil)
	if bookingService == nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Failed to create booking sync service",
		})
		return
	}

	booking, items, err := bookingService.GetBookingDetail(c.Request.Context(), bookingSN)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}
	if booking == nil {
		c.JSON(http.StatusNotFound, gin.H{
			"success": false,
			"error":   "Booking not found",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"booking": booking,
			"items":   items,
		},
	})
}
