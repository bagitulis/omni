package handlers

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

// TestRespondUnauthorized tests respondUnauthorized helper
func TestRespondUnauthorized(t *testing.T) {
	gin.SetMode(gin.TestMode)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	respondUnauthorized(c, "test error")

	assert.Equal(t, http.StatusUnauthorized, w.Code)
	assert.Contains(t, w.Body.String(), "test error")
	assert.Contains(t, w.Body.String(), `"success":false`)
}

// TestRespondBadRequest tests respondBadRequest helper
func TestRespondBadRequest(t *testing.T) {
	gin.SetMode(gin.TestMode)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	respondBadRequest(c, "invalid input")

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "invalid input")
	assert.Contains(t, w.Body.String(), `"success":false`)
}

// TestRespondNotFound tests respondNotFound helper
func TestRespondNotFound(t *testing.T) {
	gin.SetMode(gin.TestMode)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	respondNotFound(c, "resource not found")

	assert.Equal(t, http.StatusNotFound, w.Code)
	assert.Contains(t, w.Body.String(), "resource not found")
	assert.Contains(t, w.Body.String(), `"success":false`)
}

// TestRespondInternalError tests respondInternalError helper
func TestRespondInternalError(t *testing.T) {
	gin.SetMode(gin.TestMode)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	respondInternalError(c, assert.AnError)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	assert.Contains(t, w.Body.String(), `"success":false`)
}

// TestRespondSuccess tests respondSuccess helper
func TestRespondSuccess(t *testing.T) {
	gin.SetMode(gin.TestMode)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	data := []string{"item1", "item2"}
	respondSuccess(c, data, 2)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), `"success":true`)
	assert.Contains(t, w.Body.String(), `"count":2`)
	assert.Contains(t, w.Body.String(), `"data"`)
}

// TestRespondWithConfig tests respondWithConfig helper
func TestRespondWithConfig(t *testing.T) {
	gin.SetMode(gin.TestMode)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	config := map[string]string{"key": "value"}
	respondWithConfig(c, config)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), `"success":true`)
	assert.Contains(t, w.Body.String(), `"data"`)
}

// TestRespondWithConfigs tests respondWithConfigs helper
func TestRespondWithConfigs(t *testing.T) {
	gin.SetMode(gin.TestMode)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	configs := []map[string]string{{"key": "value1"}, {"key": "value2"}}
	respondWithConfigs(c, configs)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), `"success":true`)
	assert.Contains(t, w.Body.String(), `"data"`)
}

// TestRespondCreated tests respondCreated helper
func TestRespondCreated(t *testing.T) {
	gin.SetMode(gin.TestMode)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	config := map[string]string{"id": "123"}
	respondCreated(c, config)

	assert.Equal(t, http.StatusCreated, w.Code)
	assert.Contains(t, w.Body.String(), `"success":true`)
	assert.Contains(t, w.Body.String(), `"data"`)
}

// TestRespondDeleted tests respondDeleted helper
func TestRespondDeleted(t *testing.T) {
	gin.SetMode(gin.TestMode)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	respondDeleted(c, "Item deleted successfully")

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), `"success":true`)
	assert.Contains(t, w.Body.String(), "Item deleted successfully")
}
