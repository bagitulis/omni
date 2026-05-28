package handlers

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/omni/backend/internal/models"
	credentialsvc "github.com/omni/backend/internal/services"
	"github.com/omni/backend/internal/testutils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDebugCredentialResponse(t *testing.T) {
	gin.SetMode(gin.TestMode)
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
	r.PUT("/api/credentials/platforms/:platform/app", handler.PutCredentialApp)
	r.POST("/api/credentials/platforms/:platform/connections/manual-token", handler.PostCredentialManualToken)

	appReq := `{"region":"Indonesia","app_key":"app-key-1","app_secret":"app-secret-1","reason":"initial_setup"}`
	req, _ := http.NewRequest("PUT", "/api/credentials/platforms/lazada/app", strings.NewReader(appReq))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	t.Logf("PUT app status: %d body: %s", w.Code, w.Body.String())

	manualReq := `{"store_identifier":"store-123","region":"Indonesia","access_token":"access-token-123","refresh_token":"refresh-token-123","reason":"emergency_recovery"}`
	req, _ = http.NewRequest("POST", "/api/credentials/platforms/tiktok/connections/manual-token", strings.NewReader(manualReq))
	req.Header.Set("Content-Type", "application/json")
	w = httptest.NewRecorder()
	r.ServeHTTP(w, req)
	t.Logf("POST manual-token status: %d body: %s", w.Code, w.Body.String())
	
	_ = assert.Equal
	_ = require.New
	_ = json.Marshal
}
