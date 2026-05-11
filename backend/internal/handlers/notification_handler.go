package handlers

import (
	"io"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/omni/backend/internal/services"
)

// NotificationHandler handles notification API endpoints.
type NotificationHandler struct{}

// NewNotificationHandler creates a new handler.
func NewNotificationHandler() *NotificationHandler {
	return &NotificationHandler{}
}

func (h *NotificationHandler) getService(c *gin.Context) (*services.NotificationService, error) {
	db, err := GetTenantDB(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": err.Error()})
		return nil, err
	}
	tenantID := c.GetString("tenant_id")
	return services.NewNotificationService(db).WithTenant(tenantID), nil
}

// CreateNotification creates a manual notification (from UI).
func (h *NotificationHandler) CreateNotification(c *gin.Context) {
	svc, err := h.getService(c)
	if err != nil {
		return
	}

	var body struct {
		Type      string `json:"type" binding:"required"`
		Category  string `json:"category" binding:"required"`
		Title     string `json:"title" binding:"required"`
		Message   string `json:"message"`
		ActionURL string `json:"action_url"`
	}

	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "invalid request body"})
		return
	}

	notif, err := svc.Push(body.Type, body.Category, body.Title, body.Message, body.ActionURL)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"success": true, "data": notif})
}

// StreamNotifications handles SSE connections for real-time notifications.
// GET /api/notifications/stream
func (h *NotificationHandler) StreamNotifications(c *gin.Context) {
	tenantID := c.GetString("tenant_id")
	// Support ticket-based auth for SSE connections
	if tenantID == "" {
		ticket := c.Query("ticket")
		if ticket != "" {
			_, ticketTenantID, valid := ValidateSSETicket(ticket)
			if !valid {
				c.JSON(http.StatusUnauthorized, gin.H{"success": false, "error": "invalid or expired ticket"})
				return
			}
			tenantID = ticketTenantID
			// Set tenant_id in context so GetTenantDB can find it
			c.Set("tenant_id", tenantID)
		}
	}

	if tenantID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"success": false, "error": "unauthorized"})
		return
	}
	db, err := GetTenantDB(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": err.Error()})
		return
	}

	svc := services.NewNotificationService(db)
	client := svc.RegisterClient(tenantID)
	defer svc.UnregisterClient(tenantID, client)

	// Set SSE headers
	c.Header("Content-Type", "text/event-stream")
	c.Header("Cache-Control", "no-cache")
	c.Header("Connection", "keep-alive")
	c.Header("Transfer-Encoding", "chunked")

	// Pinger to keep connection alive
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	c.Stream(func(w io.Writer) bool {
		select {
		case <-c.Request.Context().Done():
			return false
		case msg := <-client:
			c.SSEvent("notification", msg)
			return true
		case <-ticker.C:
			c.SSEvent("ping", "ping")
			return true
		}
	})
}

// (Other standard methods remain, just need to ensure getService used)

// ListNotifications returns recent notifications.
func (h *NotificationHandler) ListNotifications(c *gin.Context) {
	svc, err := h.getService(c)
	if err != nil {
		return
	}

	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "50"))
	sinceID, _ := strconv.ParseInt(c.DefaultQuery("since_id", "0"), 10, 64)
	unreadOnly := c.DefaultQuery("unread_only", "false") == "true"

	items, err := svc.List(limit, sinceID, unreadOnly)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "data": gin.H{"items": items, "count": len(items)}})
}

// GetUnreadCount returns the count of unread notifications.
func (h *NotificationHandler) GetUnreadCount(c *gin.Context) {
	svc, err := h.getService(c)
	if err != nil {
		return
	}

	count, err := svc.UnreadCount()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "data": gin.H{"unread_count": count}})
}

// MarkAsRead marks a single notification as read.
func (h *NotificationHandler) MarkAsRead(c *gin.Context) {
	svc, err := h.getService(c)
	if err != nil {
		return
	}

	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "invalid notification id"})
		return
	}

	if err := svc.MarkAsRead(id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true})
}

// MarkAllAsRead marks all notifications as read.
func (h *NotificationHandler) MarkAllAsRead(c *gin.Context) {
	svc, err := h.getService(c)
	if err != nil {
		return
	}

	if err := svc.MarkAllAsRead(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true})
}

// DeleteNotification deletes a single notification.
func (h *NotificationHandler) DeleteNotification(c *gin.Context) {
	svc, err := h.getService(c)
	if err != nil {
		return
	}

	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "invalid notification id"})
		return
	}

	if err := svc.Delete(id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true})
}

// DeleteAllNotifications deletes all notifications.
func (h *NotificationHandler) DeleteAllNotifications(c *gin.Context) {
	svc, err := h.getService(c)
	if err != nil {
		return
	}

	if err := svc.DeleteAll(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true})
}

// GetSettings returns notification settings.
func (h *NotificationHandler) GetSettings(c *gin.Context) {
	svc, err := h.getService(c)
	if err != nil {
		return
	}

	settings, err := svc.GetSettings()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "data": settings})
}

// UpdateSettings updates notification settings.
func (h *NotificationHandler) UpdateSettings(c *gin.Context) {
	svc, err := h.getService(c)
	if err != nil {
		return
	}

	var body struct {
		RetentionDays int `json:"retention_days"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "invalid request body"})
		return
	}

	if err := svc.SaveSettings(body.RetentionDays); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true})
}

// GetNotificationDetail returns full notification detail including metadata.
// GET /api/notifications/:id/detail
func (h *NotificationHandler) GetNotificationDetail(c *gin.Context) {
	svc, err := h.getService(c)
	if err != nil {
		return
	}

	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "invalid notification id"})
		return
	}

	notif, err := svc.GetByID(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"success": false, "error": "notification not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "data": notif})
}
