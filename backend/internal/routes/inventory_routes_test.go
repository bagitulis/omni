package routes

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/omni/backend/internal/handlers"
	"github.com/stretchr/testify/assert"
)

func TestRegisterInventorySimpleRoutes_ConfigUpdateRouteExists(t *testing.T) {
	gin.SetMode(gin.TestMode)

	router := gin.New()
	api := router.Group("/api")
	handler := handlers.NewInventoryHandler(nil)
	RegisterInventorySimpleRoutes(api, handler)

	req := httptest.NewRequest(http.MethodPut, "/api/inventory/config", bytes.NewBufferString(`{"key_column":"TOTAL"}`))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.NotEqual(t, http.StatusNotFound, w.Code)
}
