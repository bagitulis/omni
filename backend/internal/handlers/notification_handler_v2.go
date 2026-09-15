package handlers

import (
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/omni/backend/internal/repositories"
)

// V2 endpoints for notifications. All handlers here delegate to the service
// layer; no DB access or business logic lives in this file.
//
// Layout: one method per endpoint, each keeps to the standard pattern
// (tenant/user extract → parse → svc → response).

// getUser returns the UUID user ID stashed by middleware.Auth (see
// internal/middleware/auth.go which uses c.Set("userID", claims.UserID)).
// Kept tolerant of alternate keys used by other middlewares.
func (h *NotificationHandler) getUser(c *gin.Context) string {
	if s := c.GetString("userID"); s != "" {
		return s
	}
	if s := c.GetString("user_id"); s != "" {
		return s
	}
	if v, ok := c.Get("userID"); ok {
		if s, ok := v.(string); ok {
			return s
		}
	}
	return ""
}

// GetCounts returns total, unread-per-user, and severity breakdown.
// GET /api/notifications/counts
func (h *NotificationHandler) GetCounts(c *gin.Context) {
	svc, err := h.newService(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": err.Error()})
		return
	}
	uid := h.getUser(c)
	counts, err := svc.CountsForUser(c.Request.Context(), uid)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"total":       counts.Total,
			"unread":      counts.Unread,
			"by_severity": counts.BySeverity,
		},
	})
}

// ListV2 lists notifications with filters, search, per-user unread flag.
// GET /api/notifications  (V2 handler; superseds ListNotifications for new clients)
func (h *NotificationHandler) ListV2(c *gin.Context) {
	svc, err := h.newService(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": err.Error()})
		return
	}
	f := parseListFilter(c, h.getUser(c))
	items, err := svc.ListActive(c.Request.Context(), f)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    gin.H{"items": items, "count": len(items)},
	})
}

func parseListFilter(c *gin.Context, userID string) repositories.ListFilter {
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "50"))
	sinceID, _ := strconv.ParseInt(c.DefaultQuery("since_id", "0"), 10, 64)
	unreadOnly := c.DefaultQuery("unread_only", "false") == "true"
	minSev, _ := strconv.Atoi(c.DefaultQuery("min_severity", "0"))
	f := repositories.ListFilter{
		Limit:       limit,
		SinceID:     sinceID,
		UnreadOnly:  unreadOnly,
		UserID:      userID,
		Category:    c.Query("category"),
		MinSeverity: int16(minSev),
		Search:      strings.TrimSpace(c.Query("q")),
	}
	if from := c.Query("from"); from != "" {
		if t, err := time.Parse(time.RFC3339, from); err == nil {
			f.From = &t
		}
	}
	if to := c.Query("to"); to != "" {
		if t, err := time.Parse(time.RFC3339, to); err == nil {
			f.To = &t
		}
	}
	return f
}

// MarkReadV2 marks a notification read for the current user only.
// PATCH /api/notifications/:id/read
func (h *NotificationHandler) MarkReadV2(c *gin.Context) {
	svc, err := h.newService(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": err.Error()})
		return
	}
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "invalid notification id"})
		return
	}
	uid := h.getUser(c)
	if uid == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"success": false, "error": "missing user_id"})
		return
	}
	if err := svc.MarkReadForUser(c.Request.Context(), id, uid); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "data": gin.H{"affected": int64(1)}})
}

// MarkAllReadV2 marks every unread notification as read for the user.
// PATCH /api/notifications/read-all
func (h *NotificationHandler) MarkAllReadV2(c *gin.Context) {
	svc, err := h.newService(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": err.Error()})
		return
	}
	uid := h.getUser(c)
	if uid == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"success": false, "error": "missing user_id"})
		return
	}
	affected, err := svc.MarkAllReadForUser(c.Request.Context(), uid)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "data": gin.H{"affected": affected}})
}

// BulkMarkRead marks the supplied IDs read for the current user.
// POST /api/notifications/bulk/read  body: {"ids":[1,2,3]}
func (h *NotificationHandler) BulkMarkRead(c *gin.Context) {
	svc, err := h.newService(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": err.Error()})
		return
	}
	var body struct {
		IDs []int64 `json:"ids"`
	}
	if err := c.ShouldBindJSON(&body); err != nil || len(body.IDs) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "ids required"})
		return
	}
	uid := h.getUser(c)
	if uid == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"success": false, "error": "missing user_id"})
		return
	}
	affected, err := svc.BulkMarkReadForUser(c.Request.Context(), uid, body.IDs)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "data": gin.H{"affected": affected}})
}

// BulkDelete deletes the supplied IDs.
// POST /api/notifications/bulk/delete  body: {"ids":[1,2,3]}
func (h *NotificationHandler) BulkDelete(c *gin.Context) {
	svc, err := h.newService(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": err.Error()})
		return
	}
	var body struct {
		IDs []int64 `json:"ids"`
	}
	if err := c.ShouldBindJSON(&body); err != nil || len(body.IDs) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "ids required"})
		return
	}
	affected, err := svc.BulkDelete(c.Request.Context(), body.IDs)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "data": gin.H{"affected": affected}})
}

// Snooze hides a notification until the supplied timestamp.
// POST /api/notifications/:id/snooze  body: {"until":"2026-09-17T10:00:00Z"}
func (h *NotificationHandler) Snooze(c *gin.Context) {
	svc, err := h.newService(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": err.Error()})
		return
	}
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "invalid notification id"})
		return
	}
	var body struct {
		Until time.Time `json:"until" binding:"required"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "until (RFC3339) required"})
		return
	}
	if err := svc.Snooze(c.Request.Context(), id, body.Until); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "data": gin.H{"snoozed_until": body.Until}})
}
