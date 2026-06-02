package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
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
		{"required", fmt.Errorf("reason is required"), 400},
		{"generic error", fmt.Errorf("something went wrong"), 500},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, credentialErrorStatus(tt.err))
		})
	}
}
