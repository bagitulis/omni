package handlers

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

// TestGetCSRFToken_Success tests successful CSRF token generation
func TestGetCSRFToken_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	csrfHandler := NewCSRFHandler()
	r.GET("/api/csrf-token", csrfHandler.GetCSRFToken)

	req, _ := http.NewRequest("GET", "/api/csrf-token", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Equal(t, true, resp["success"])

	data, ok := resp["data"].(map[string]interface{})
	assert.True(t, ok)
	assert.NotEmpty(t, data["token"])
	assert.Equal(t, float64(86400), data["expires_in"])
}

// TestGetCSRFToken_SetsCookie tests that CSRF token is set in cookie
func TestGetCSRFToken_SetsCookie(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	csrfHandler := NewCSRFHandler()
	r.GET("/api/csrf-token", csrfHandler.GetCSRFToken)

	req, _ := http.NewRequest("GET", "/api/csrf-token", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	// Check that cookie is set
	cookies := w.Result().Cookies()
	var csrfCookie *http.Cookie
	for _, c := range cookies {
		if c.Name == "csrf_token" {
			csrfCookie = c
			break
		}
	}

	assert.NotNil(t, csrfCookie, "CSRF cookie should be set")
	assert.NotEmpty(t, csrfCookie.Value)
	assert.Equal(t, "/", csrfCookie.Path)
	assert.False(t, csrfCookie.HttpOnly)
}

// TestGetCSRFToken_ResponseMatchesCookie tests token in response matches cookie
func TestGetCSRFToken_ResponseMatchesCookie(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	csrfHandler := NewCSRFHandler()
	r.GET("/api/csrf-token", csrfHandler.GetCSRFToken)

	req, _ := http.NewRequest("GET", "/api/csrf-token", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	// Get token from response
	var resp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)

	data := resp["data"].(map[string]interface{})
	responseToken := data["token"].(string)

	// Get token from cookie (may be URL encoded)
	cookies := w.Result().Cookies()
	var cookieToken string
	for _, c := range cookies {
		if c.Name == "csrf_token" {
			cookieToken = c.Value
			break
		}
	}

	// URL decode cookie token for comparison (cookies may be URL encoded)
	decodedCookie, decodeErr := url.QueryUnescape(cookieToken)
	if decodeErr != nil {
		decodedCookie = cookieToken
	}

	// Tokens should match (double-submit pattern)
	assert.Equal(t, responseToken, decodedCookie)
}

// TestGetCSRFToken_ResponseFormat tests the response format is correct
func TestGetCSRFToken_ResponseFormat(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	csrfHandler := NewCSRFHandler()
	r.GET("/api/csrf-token", csrfHandler.GetCSRFToken)

	req, _ := http.NewRequest("GET", "/api/csrf-token", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)

	// Check required fields
	assert.Contains(t, resp, "success")
	assert.Contains(t, resp, "data")

	data := resp["data"].(map[string]interface{})
	assert.Contains(t, data, "token")
	assert.Contains(t, data, "expires_in")
}

// TestCSRFHandler_NewCSRFHandler tests handler creation
func TestCSRFHandler_NewCSRFHandler(t *testing.T) {
	handler := NewCSRFHandler()
	assert.NotNil(t, handler)
}
