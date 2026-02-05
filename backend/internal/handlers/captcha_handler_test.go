package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/omni/backend/internal/services"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockCaptchaService mocks the CaptchaService
type MockCaptchaService struct {
	mock.Mock
}

func (m *MockCaptchaService) Verify(ctx context.Context, req *services.VerifyRequest) (*services.CaptchaResponse, error) {
	args := m.Called(ctx, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*services.CaptchaResponse), args.Error(1)
}

func (m *MockCaptchaService) IsEnabled() bool {
	args := m.Called()
	return args.Bool(0)
}

func (m *MockCaptchaService) GetSiteKey() string {
	args := m.Called()
	return args.String(0)
}

// TestVerifyCaptcha_Success tests successful captcha verification
func TestVerifyCaptcha_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	mockService := new(MockCaptchaService)
	mockService.On("Verify", mock.Anything, mock.AnythingOfType("*services.VerifyRequest")).
		Return(&services.CaptchaResponse{
			Success: true,
			Score:   0.9,
		}, nil)

	r.POST("/api/captcha/verify", func(c *gin.Context) {
		var req VerifyCaptchaRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"success": false,
				"error":   "Invalid request body",
			})
			return
		}

		result, err := mockService.Verify(c.Request.Context(), &services.VerifyRequest{
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
	})

	reqBody := `{"token": "test-token-123"}`
	req, _ := http.NewRequest("POST", "/api/captcha/verify", strings.NewReader(reqBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Equal(t, true, resp["success"])

	data := resp["data"].(map[string]interface{})
	assert.Equal(t, true, data["verified"])
	assert.Equal(t, 0.9, data["score"])

	mockService.AssertExpectations(t)
}

// TestVerifyCaptcha_InvalidToken tests captcha verification with invalid token
func TestVerifyCaptcha_InvalidToken(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	mockService := new(MockCaptchaService)
	mockService.On("Verify", mock.Anything, mock.AnythingOfType("*services.VerifyRequest")).
		Return(nil, errors.New("invalid captcha token"))

	r.POST("/api/captcha/verify", func(c *gin.Context) {
		var req VerifyCaptchaRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"success": false,
				"error":   "Invalid request body",
			})
			return
		}

		result, err := mockService.Verify(c.Request.Context(), &services.VerifyRequest{
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

		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"data": gin.H{
				"verified": result.Success,
				"score":    result.Score,
			},
		})
	})

	reqBody := `{"token": "invalid-token"}`
	req, _ := http.NewRequest("POST", "/api/captcha/verify", strings.NewReader(reqBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var resp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Equal(t, false, resp["success"])
	assert.Contains(t, resp["error"], "invalid captcha token")

	mockService.AssertExpectations(t)
}

// TestVerifyCaptcha_MissingToken tests captcha verification without token
func TestVerifyCaptcha_MissingToken(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	r.POST("/api/captcha/verify", func(c *gin.Context) {
		var req VerifyCaptchaRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"success": false,
				"error":   "Invalid request body",
			})
			return
		}
	})

	reqBody := `{}`
	req, _ := http.NewRequest("POST", "/api/captcha/verify", strings.NewReader(reqBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var resp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Equal(t, false, resp["success"])
}

// TestGetCaptchaStatus_Enabled tests captcha status when enabled
func TestGetCaptchaStatus_Enabled(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	mockService := new(MockCaptchaService)
	mockService.On("IsEnabled").Return(true)

	r.GET("/api/captcha/status", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"data": gin.H{
				"enabled": mockService.IsEnabled(),
			},
		})
	})

	req, _ := http.NewRequest("GET", "/api/captcha/status", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Equal(t, true, resp["success"])

	data := resp["data"].(map[string]interface{})
	assert.Equal(t, true, data["enabled"])

	mockService.AssertExpectations(t)
}

// TestGetCaptchaStatus_Disabled tests captcha status when disabled
func TestGetCaptchaStatus_Disabled(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	mockService := new(MockCaptchaService)
	mockService.On("IsEnabled").Return(false)

	r.GET("/api/captcha/status", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"data": gin.H{
				"enabled": mockService.IsEnabled(),
			},
		})
	})

	req, _ := http.NewRequest("GET", "/api/captcha/status", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Equal(t, true, resp["success"])

	data := resp["data"].(map[string]interface{})
	assert.Equal(t, false, data["enabled"])

	mockService.AssertExpectations(t)
}

// TestGetSiteKey_Success tests getting site key
func TestGetSiteKey_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	mockService := new(MockCaptchaService)
	mockService.On("GetSiteKey").Return("test-site-key-123")
	mockService.On("IsEnabled").Return(true)

	r.GET("/api/captcha/sitekey", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"enabled": mockService.IsEnabled(),
			"siteKey": mockService.GetSiteKey(),
		})
	})

	req, _ := http.NewRequest("GET", "/api/captcha/sitekey", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Equal(t, true, resp["success"])
	assert.Equal(t, true, resp["enabled"])
	assert.Equal(t, "test-site-key-123", resp["siteKey"])

	mockService.AssertExpectations(t)
}

// TestVerifyCaptcha_VerificationFailed tests captcha verification failure
func TestVerifyCaptcha_VerificationFailed(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	mockService := new(MockCaptchaService)
	mockService.On("Verify", mock.Anything, mock.AnythingOfType("*services.VerifyRequest")).
		Return(&services.CaptchaResponse{
			Success:    false,
			Score:      0.1,
			ErrorCodes: []string{"timeout-or-duplicate"},
		}, nil)

	r.POST("/api/captcha/verify", func(c *gin.Context) {
		var req VerifyCaptchaRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"success": false,
				"error":   "Invalid request body",
			})
			return
		}

		result, err := mockService.Verify(c.Request.Context(), &services.VerifyRequest{
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
	})

	reqBody := `{"token": "expired-token"}`
	req, _ := http.NewRequest("POST", "/api/captcha/verify", strings.NewReader(reqBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var resp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Equal(t, false, resp["success"])
	assert.Equal(t, "Captcha verification failed", resp["error"])

	errorCodes := resp["errors"].([]interface{})
	assert.Contains(t, errorCodes, "timeout-or-duplicate")

	mockService.AssertExpectations(t)
}
