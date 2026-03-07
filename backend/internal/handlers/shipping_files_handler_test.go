package handlers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

// TestShippingFilesHandler_GetShippingFiles_EmptyWhenDirNotExist tests GetShippingFiles returns empty list when directory doesn't exist
func TestShippingFilesHandler_GetShippingFiles_EmptyWhenDirNotExist(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	// Create temp directory for test
	tempDir := t.TempDir()
	handler := NewShippingFilesHandler(tempDir)

	// Set up route with middleware that sets tenantID
	r.GET("/api/shipping/files", func(c *gin.Context) {
		c.Set("tenantID", "test-tenant")
		handler.GetShippingFiles(c)
	})

	req, _ := http.NewRequest("GET", "/api/shipping/files", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.True(t, resp["success"].(bool))

	data := resp["data"].(map[string]interface{})
	files := data["files"].([]interface{})
	assert.Empty(t, files)
	assert.Equal(t, "No shipping files found", data["message"])
}

// TestShippingFilesHandler_GetShippingFiles_MissingTenant tests GetShippingFiles returns 400 when tenantID is empty
func TestShippingFilesHandler_GetShippingFiles_MissingTenant(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	handler := NewShippingFilesHandler("")
	r.GET("/api/shipping/files", handler.GetShippingFiles)

	req, _ := http.NewRequest("GET", "/api/shipping/files", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var resp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.False(t, resp["success"].(bool))
	assert.Equal(t, "Missing tenantId", resp["error"])
}

// TestShippingFilesHandler_GetShippingFiles_WithFiles tests GetShippingFiles returns files when they exist
func TestShippingFilesHandler_GetShippingFiles_WithFiles(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	// Create temp directory and files
	tempDir := t.TempDir()
	tenantID := "test-tenant"
	shippingDir := filepath.Join(tempDir, "data", tenantID, "shipping")
	err := os.MkdirAll(shippingDir, 0755)
	assert.NoError(t, err)

	// Create test files
	testFiles := []string{"label1.pdf", "label2.pdf", "shipment.pdf"}
	for _, filename := range testFiles {
		err = os.WriteFile(filepath.Join(shippingDir, filename), []byte("test"), 0644)
		assert.NoError(t, err)
	}

	handler := NewShippingFilesHandler(tempDir)

	r.GET("/api/shipping/files", func(c *gin.Context) {
		c.Set("tenantID", tenantID)
		handler.GetShippingFiles(c)
	})

	req, _ := http.NewRequest("GET", "/api/shipping/files", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]interface{}
	err = json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.True(t, resp["success"].(bool))

	data := resp["data"].(map[string]interface{})
	files := data["files"].([]interface{})
	assert.Len(t, files, 3)
}

// TestShippingFilesHandler_ProcessShippingFile_Success tests ProcessShippingFile returns 200 with valid filename
func TestShippingFilesHandler_ProcessShippingFile_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	handler := NewShippingFilesHandler("")

	r.POST("/api/shipping/process-file", func(c *gin.Context) {
		c.Set("tenantID", "test-tenant")
		handler.ProcessShippingFile(c)
	})

	reqBody := map[string]string{"filename": "label1.pdf"}
	bodyBytes, _ := json.Marshal(reqBody)

	req, _ := http.NewRequest("POST", "/api/shipping/process-file", bytes.NewBuffer(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.True(t, resp["success"].(bool))
	assert.Equal(t, "File processed successfully", resp["message"])
}

// TestShippingFilesHandler_ProcessShippingFile_MissingTenant tests ProcessShippingFile returns 400 when tenantID is empty
func TestShippingFilesHandler_ProcessShippingFile_MissingTenant(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	handler := NewShippingFilesHandler("")
	r.POST("/api/shipping/process-file", handler.ProcessShippingFile)

	reqBody := map[string]string{"filename": "label1.pdf"}
	bodyBytes, _ := json.Marshal(reqBody)

	req, _ := http.NewRequest("POST", "/api/shipping/process-file", bytes.NewBuffer(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var resp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.False(t, resp["success"].(bool))
	assert.Equal(t, "Missing tenantId", resp["error"])
}

// TestShippingFilesHandler_ProcessShippingFile_EmptyBody tests ProcessShippingFile returns 400 with empty body
func TestShippingFilesHandler_ProcessShippingFile_EmptyBody(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	handler := NewShippingFilesHandler("")

	r.POST("/api/shipping/process-file", func(c *gin.Context) {
		c.Set("tenantID", "test-tenant")
		handler.ProcessShippingFile(c)
	})

	req, _ := http.NewRequest("POST", "/api/shipping/process-file", bytes.NewBuffer([]byte("{}")))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var resp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.False(t, resp["success"].(bool))
}

// TestShippingFilesHandler_ProcessShippingFile_InvalidJSON tests ProcessShippingFile returns 400 with invalid JSON
func TestShippingFilesHandler_ProcessShippingFile_InvalidJSON(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	handler := NewShippingFilesHandler("")

	r.POST("/api/shipping/process-file", func(c *gin.Context) {
		c.Set("tenantID", "test-tenant")
		handler.ProcessShippingFile(c)
	})

	req, _ := http.NewRequest("POST", "/api/shipping/process-file", bytes.NewBuffer([]byte("invalid json")))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var resp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.False(t, resp["success"].(bool))
}
