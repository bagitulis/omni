package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/omni/backend/internal/services"
)

// CaptchaHandler handles captcha verification endpoints
type CaptchaHandler struct {
	captchaService *services.CaptchaService
}

// NewCaptchaHandler creates a new captcha handler
func NewCaptchaHandler(captchaService *services.CaptchaService) *CaptchaHandler {
	return &CaptchaHandler{captchaService: captchaService}
}

// VerifyCaptchaRequest represents captcha verification request
type VerifyCaptchaRequest struct {
	Token  string `json:"token" binding:"required"`
	Action string `json:"action,omitempty"`
}

// VerifyCaptcha verifies a captcha token
func (h *CaptchaHandler) VerifyCaptcha(c *gin.Context) {
	var req VerifyCaptchaRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Invalid request body",
		})
		return
	}

	result, err := h.captchaService.Verify(c.Request.Context(), &services.VerifyRequest{
		Token:    req.Token,
		RemoteIP: c.ClientIP(),
		Action:   req.Action,
	})
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	if !result.Success {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Captcha verification failed",
			"errors":  result.ErrorCodes,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"verified": true,
			"score":    result.Score,
		},
	})
}

// GetCaptchaStatus returns captcha service status
func (h *CaptchaHandler) GetCaptchaStatus(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"enabled": h.captchaService.IsEnabled(),
		},
	})
}

// GetSiteKey returns the reCAPTCHA site key for frontend
func (h *CaptchaHandler) GetSiteKey(c *gin.Context) {
	siteKey := h.captchaService.GetSiteKey()
	enabled := h.captchaService.IsEnabled()

	c.JSON(http.StatusOK, gin.H{
		"success":  true,
		"enabled":  enabled,
		"site_key": siteKey,
	})
}
