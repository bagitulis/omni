package tiktok

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func init() {
	gin.SetMode(gin.TestMode)
}

// --- UploadImageRequest JSON marshaling tests ---

func TestUploadImageRequest_JSONMarshal(t *testing.T) {
	req := UploadImageRequest{
		ImageURL:  "https://example.com/image.jpg",
		ImageData: "base64data==",
		UseCase:   "MAIN_IMAGE",
	}
	b, err := json.Marshal(req)
	require.NoError(t, err)

	var m map[string]interface{}
	require.NoError(t, json.Unmarshal(b, &m))

	assert.Equal(t, "https://example.com/image.jpg", m["image_url"])
	assert.Equal(t, "base64data==", m["image_data"])
	assert.Equal(t, "MAIN_IMAGE", m["use_case"])
}

func TestUploadImageRequest_JSONUnmarshal(t *testing.T) {
	raw := `{"image_url":"https://example.com/img.png","use_case":"SWATCH_IMAGE"}`
	var req UploadImageRequest
	require.NoError(t, json.Unmarshal([]byte(raw), &req))
	assert.Equal(t, "https://example.com/img.png", req.ImageURL)
	assert.Equal(t, "SWATCH_IMAGE", req.UseCase)
	assert.Empty(t, req.ImageData)
}

func TestUploadImageRequest_OmitEmptyFields(t *testing.T) {
	req := UploadImageRequest{UseCase: "MAIN_IMAGE"}
	b, err := json.Marshal(req)
	require.NoError(t, err)

	// image_url and image_data are omitempty — should not appear
	assert.NotContains(t, string(b), "image_url")
	assert.NotContains(t, string(b), "image_data")
	assert.Contains(t, string(b), "use_case")
}

// --- ImageUploadResult JSON marshaling tests ---

func TestImageUploadResult_JSONMarshal(t *testing.T) {
	result := ImageUploadResult{
		URI:     "tiktok://image/abc123",
		URL:     "https://cdn.example.com/image.png",
		Width:   800,
		Height:  600,
		Success: true,
	}
	b, err := json.Marshal(result)
	require.NoError(t, err)

	var m map[string]interface{}
	require.NoError(t, json.Unmarshal(b, &m))

	assert.Equal(t, "tiktok://image/abc123", m["uri"])
	assert.Equal(t, "https://cdn.example.com/image.png", m["url"])
	assert.EqualValues(t, 800, m["width"])
	assert.EqualValues(t, 600, m["height"])
	assert.Equal(t, true, m["success"])
}

func TestImageUploadResult_JSONUnmarshal(t *testing.T) {
	raw := `{"uri":"tiktok://image/xyz","url":"https://cdn.example.com/x.png","width":1024,"height":768,"success":true}`
	var r ImageUploadResult
	require.NoError(t, json.Unmarshal([]byte(raw), &r))
	assert.Equal(t, "tiktok://image/xyz", r.URI)
	assert.Equal(t, "https://cdn.example.com/x.png", r.URL)
	assert.Equal(t, 1024, r.Width)
	assert.Equal(t, 768, r.Height)
	assert.True(t, r.Success)
}

// --- ImageUploadTask JSON marshaling tests ---

func TestImageUploadTask_JSONMarshal(t *testing.T) {
	task := ImageUploadTask{
		TaskID:    "task-001",
		Status:    "PENDING",
		ImageURI:  "tiktok://image/abc",
		Error:     "",
		CreatedAt: 1700000000,
	}
	b, err := json.Marshal(task)
	require.NoError(t, err)

	var m map[string]interface{}
	require.NoError(t, json.Unmarshal(b, &m))

	assert.Equal(t, "task-001", m["task_id"])
	assert.Equal(t, "PENDING", m["status"])
	assert.Equal(t, "tiktok://image/abc", m["image_uri"])
	assert.EqualValues(t, 1700000000, m["created_at"])
}

func TestImageUploadTask_OmitEmptyFields(t *testing.T) {
	task := ImageUploadTask{
		TaskID:    "task-002",
		Status:    "DONE",
		CreatedAt: 1700000001,
	}
	b, err := json.Marshal(task)
	require.NoError(t, err)

	// image_uri and error are omitempty — should not appear when empty
	assert.NotContains(t, string(b), "image_uri")
	assert.NotContains(t, string(b), `"error"`)
}

// --- UploadImage handler tests ---

func TestProductImageHandler_UploadImage_MissingTenantID(t *testing.T) {
	r := gin.New()
	h := NewProductImageHandler("/test/path")
	r.POST("/upload-image", h.UploadImage)

	body := `{"image_url":"https://example.com/img.jpg","use_case":"MAIN_IMAGE"}`
	req, _ := http.NewRequest(http.MethodPost, "/upload-image", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
	var resp map[string]interface{}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	assert.Equal(t, false, resp["success"])
}

func TestProductImageHandler_UploadImage_InvalidJSON(t *testing.T) {
	r := gin.New()
	r.Use(func(c *gin.Context) {
		c.Set("tenant_id", "test-tenant")
		c.Next()
	})
	h := NewProductImageHandler("/test/path")
	r.POST("/upload-image", h.UploadImage)

	req, _ := http.NewRequest(http.MethodPost, "/upload-image", bytes.NewBufferString("not-json{{{"))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	var resp map[string]interface{}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	assert.Equal(t, false, resp["success"])
}

func TestProductImageHandler_UploadImage_Success(t *testing.T) {
	r := gin.New()
	r.Use(func(c *gin.Context) {
		c.Set("tenant_id", "test-tenant")
		c.Next()
	})
	h := NewProductImageHandler("/test/path")
	r.POST("/upload-image", h.UploadImage)

	body := `{"image_url":"https://example.com/img.jpg","use_case":"MAIN_IMAGE"}`
	req, _ := http.NewRequest(http.MethodPost, "/upload-image", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	var resp map[string]interface{}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	assert.Equal(t, true, resp["success"])

	data, ok := resp["data"].(map[string]interface{})
	require.True(t, ok, "expected data to be a map")
	assert.NotEmpty(t, data["uri"])
	assert.NotEmpty(t, data["url"])
	assert.EqualValues(t, 800, data["width"])
	assert.EqualValues(t, 800, data["height"])
	assert.Equal(t, true, data["success"])
}

func TestProductImageHandler_UploadImage_EmptyBody(t *testing.T) {
	// Empty body is valid JSON-wise (no required fields in UploadImageRequest)
	r := gin.New()
	r.Use(func(c *gin.Context) {
		c.Set("tenant_id", "test-tenant")
		c.Next()
	})
	h := NewProductImageHandler("/test/path")
	r.POST("/upload-image", h.UploadImage)

	req, _ := http.NewRequest(http.MethodPost, "/upload-image", bytes.NewBufferString("{}"))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	// Should succeed — no required fields in the struct
	assert.Equal(t, http.StatusOK, w.Code)
}

// --- GetUploadTasks handler tests ---

func TestProductImageHandler_GetUploadTasks_MissingTenantID(t *testing.T) {
	r := gin.New()
	h := NewProductImageHandler("/test/path")
	r.GET("/image-upload-tasks", h.GetUploadTasks)

	req, _ := http.NewRequest(http.MethodGet, "/image-upload-tasks", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
	var resp map[string]interface{}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	assert.Equal(t, false, resp["success"])
}

func TestProductImageHandler_GetUploadTasks_Success(t *testing.T) {
	r := gin.New()
	r.Use(func(c *gin.Context) {
		c.Set("tenant_id", "test-tenant")
		c.Next()
	})
	h := NewProductImageHandler("/test/path")
	r.GET("/image-upload-tasks", h.GetUploadTasks)

	req, _ := http.NewRequest(http.MethodGet, "/image-upload-tasks", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	var resp map[string]interface{}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	assert.Equal(t, true, resp["success"])

	data, ok := resp["data"].(map[string]interface{})
	require.True(t, ok, "expected data to be a map")
	tasks, ok := data["tasks"].([]interface{})
	require.True(t, ok, "expected tasks to be an array")
	assert.Empty(t, tasks)
}

// --- Constructor test ---

func TestNewProductImageHandler(t *testing.T) {
	h := NewProductImageHandler("/some/base/path")
	assert.NotNil(t, h)
	assert.Equal(t, "/some/base/path", h.basePath)
}
