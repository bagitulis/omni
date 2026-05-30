package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/omni/backend/internal/models"
	"github.com/omni/backend/internal/utils"
	credentialsvc "github.com/omni/backend/internal/services"
	"github.com/omni/backend/internal/testutils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCredentialApiHandler_MissingTenant(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	handler := NewPlatformAuthHandler(nil, "http://localhost:3000", "./data", nil)
	r.GET("/api/credentials/platforms", handler.GetCredentialPlatforms)

	req, _ := http.NewRequest("GET", "/api/credentials/platforms", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
	var resp map[string]any
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	assert.Equal(t, false, resp["success"])
	assert.Equal(t, "Missing tenant_id", resp["error"])
}

func TestCredentialApiHandler_ListAndMutate(t *testing.T) {
	gin.SetMode(gin.TestMode)
	key, err := utils.GenerateKey()
	require.NoError(t, err)
	os.Setenv("ENCRYPTION_KEY", key)
	r := gin.New()
	db := testutils.SetupTestPostgresWithModels(t, &models.CredentialConnection{}, &models.CredentialAppConfig{}, &models.OAuthConnectionAttempt{}, &models.CredentialAuditEvent{})
	service := credentialsvc.NewCredentialApiService(db)
	handler := NewPlatformAuthHandler(nil, "http://localhost:3000", "./data", service)
	r.Use(func(c *gin.Context) {
		c.Set("tenant_id", "tenant-test")
		c.Set("role", "admin")
		c.Set("userID", "user-test")
		c.Next()
	})
	r.GET("/api/credentials/platforms", handler.GetCredentialPlatforms)
	r.PUT("/api/credentials/platforms/:platform/app", handler.PutCredentialApp)
	r.POST("/api/credentials/platforms/:platform/app/rotate", handler.RotateCredentialApp)
	r.POST("/api/credentials/platforms/:platform/connections/manual-token", handler.PostCredentialManualToken)

	appReq := `{"region":"Indonesia","app_key":"app-key-1","app_secret":"app-secret-1","reason":"initial_setup"}`
	req, _ := http.NewRequest("PUT", "/api/credentials/platforms/lazada/app", strings.NewReader(appReq))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
	var resp map[string]any
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	assert.Equal(t, true, resp["success"])
	assert.NotContains(t, w.Body.String(), "access-token-123")
	assert.NotContains(t, w.Body.String(), "refresh-token-123")
	assert.NotContains(t, w.Body.String(), "app-secret-1")

	manualReq := `{"store_identifier":"store-123","region":"Indonesia","access_token":"access-token-123","refresh_token":"refresh-token-123","reason":"emergency_recovery"}`
	req, _ = http.NewRequest("POST", "/api/credentials/platforms/tiktok/connections/manual-token", strings.NewReader(manualReq))
	req.Header.Set("Content-Type", "application/json")
	w = httptest.NewRecorder()
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	assert.Equal(t, true, resp["success"])
	assert.NotContains(t, w.Body.String(), "access-token-123")
	assert.NotContains(t, w.Body.String(), "refresh-token-123")
	assert.NotContains(t, w.Body.String(), "app-secret-1")

	req, _ = http.NewRequest("GET", "/api/credentials/platforms", nil)
	w = httptest.NewRecorder()
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	assert.Equal(t, true, resp["success"])
	assert.NotContains(t, w.Body.String(), "access-token-123")
	assert.NotContains(t, w.Body.String(), "refresh-token-123")
}

func TestCredentialApiHandler_ManualTokenRoleGate(t *testing.T) {
	gin.SetMode(gin.TestMode)
	key, err := utils.GenerateKey()
	require.NoError(t, err)
	os.Setenv("ENCRYPTION_KEY", key)
	r := gin.New()
	db := testutils.SetupTestPostgresWithModels(t, &models.CredentialConnection{}, &models.CredentialAppConfig{}, &models.OAuthConnectionAttempt{}, &models.CredentialAuditEvent{})
	service := credentialsvc.NewCredentialApiService(db)
	handler := NewPlatformAuthHandler(nil, "http://localhost:3000", "./data", service)
	r.Use(func(c *gin.Context) {
		c.Set("tenant_id", "tenant-test")
		c.Set("role", "user")
		c.Set("userID", "user-test")
		c.Next()
	})
	r.POST("/api/credentials/platforms/:platform/connections/manual-token", handler.PostCredentialManualToken)

	req, _ := http.NewRequest("POST", "/api/credentials/platforms/tiktok/connections/manual-token", strings.NewReader(`{"store_identifier":"store-123","access_token":"access-token-123","reason":"emergency_recovery"}`))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	var resp map[string]any
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	assert.Equal(t, false, resp["success"])
	assert.Equal(t, "forbidden", resp["error"])
}

func TestCredentialApiHandler_ManualTokenRequiresReason(t *testing.T) {
	gin.SetMode(gin.TestMode)
	key, err := utils.GenerateKey()
	require.NoError(t, err)
	os.Setenv("ENCRYPTION_KEY", key)
	r := gin.New()
	db := testutils.SetupTestPostgresWithModels(t, &models.CredentialConnection{}, &models.CredentialAppConfig{}, &models.OAuthConnectionAttempt{}, &models.CredentialAuditEvent{})
	service := credentialsvc.NewCredentialApiService(db)
	handler := NewPlatformAuthHandler(nil, "http://localhost:3000", "./data", service)
	r.Use(func(c *gin.Context) {
		c.Set("tenant_id", "tenant-test")
		c.Set("role", "admin")
		c.Set("userID", "user-test")
		c.Next()
	})
	r.POST("/api/credentials/platforms/:platform/connections/manual-token", handler.PostCredentialManualToken)

	req, _ := http.NewRequest("POST", "/api/credentials/platforms/tiktok/connections/manual-token", strings.NewReader(`{"store_identifier":"store-123","access_token":"access-token-123"}`))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	var resp map[string]any
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	assert.Equal(t, false, resp["success"])
	assert.Equal(t, "reason is required", resp["error"])
}

func TestCredentialErrorStatus(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		expected int
	}{
		{"forbidden", fmt.Errorf("forbidden action"), 403},
		{"not authorized", fmt.Errorf("not authorized for this"), 403},
		{"not found", fmt.Errorf("connection not found"), 404},
		{"missing", fmt.Errorf("missing tenant_id"), 400},
		{"invalid", fmt.Errorf("invalid platform"), 400},
		{"unsupported", fmt.Errorf("unsupported platform"), 400},
		{"generic error", fmt.Errorf("something went wrong"), 500},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, credentialErrorStatus(tt.err))
		})
	}
}
