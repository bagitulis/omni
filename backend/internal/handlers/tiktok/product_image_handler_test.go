package tiktok

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestImageHandler_Constructor(t *testing.T) {
	handler := NewImageHandler("/data/path")
	assert.NotNil(t, handler)
	assert.Equal(t, "/data/path", handler.basePath)
}

func TestImageHandler_UploadImage_MissingTenantID(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	handler := NewImageHandler("/test/path")
	r.POST("/upload-image", handler.UploadImage)

	body := `{"url": "https://example.com/image.jpg"}`
	req, _ := http.NewRequest("POST", "/upload-image", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
	var resp map[string]interface{}
	assert.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	assert.Equal(t, false, resp["success"])
}

func TestImageHandler_UploadImage_InvalidJSON(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	handler := NewImageHandler("/test/path")
	r.Use(func(c *gin.Context) {
		c.Set("tenantID", "test-tenant")
		c.Next()
	})
	r.POST("/upload-image", handler.UploadImage)

	body := `{invalid}`
	req, _ := http.NewRequest("POST", "/upload-image", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestImageHandler_UploadImage_FailsWithoutDB(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	handler := NewImageHandler("/test/path")
	r.Use(func(c *gin.Context) {
		c.Set("tenantID", "test-tenant")
		c.Next()
	})
	r.POST("/upload-image", handler.UploadImage)

	body := `{"url": "https://example.com/image.jpg"}`
	req, _ := http.NewRequest("POST", "/upload-image", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestImageHandler_BatchUploadImages_MissingTenantID(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	handler := NewImageHandler("/test/path")
	r.POST("/batch-upload", handler.BatchUploadImages)

	body := `{"urls": ["https://example.com/image.jpg"]}`
	req, _ := http.NewRequest("POST", "/batch-upload", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestImageHandler_GetUploadTasks_Returns501(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	handler := NewImageHandler("/test/path")
	r.GET("/upload-tasks", handler.GetUploadTasks)

	req, _ := http.NewRequest("GET", "/upload-tasks", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotImplemented, w.Code)
}

func TestUploadImageRequest_Validation(t *testing.T) {
	req := UploadImageRequest{
		URL:     "https://example.com/test.jpg",
		UseCase: "MAIN_IMAGE",
	}
	assert.Equal(t, "https://example.com/test.jpg", req.URL)
	assert.Equal(t, "MAIN_IMAGE", req.UseCase)
}
