package routes

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/omni/backend/internal/handlers"
	"github.com/stretchr/testify/assert"
)

// TestRegisterExtensionRoutes_ResumeRouteExists pins that the resume endpoint is
// actually wired. A service method nobody can reach is not a feature.
func TestRegisterExtensionRoutes_ResumeRouteExists(t *testing.T) {
	gin.SetMode(gin.TestMode)

	router := gin.New()
	api := router.Group("/api")
	RegisterExtensionRoutes(api, nil, nil, handlers.NewScrapeHandler(nil, nil))

	req := httptest.NewRequest(http.MethodPost, "/api/extensions/scrape/job-1/resume", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.NotEqual(t, http.StatusNotFound, w.Code,
		"POST /api/extensions/scrape/:job_id/resume must be registered")
}

// TestRegisterExtensionRoutes_ResumeRequiresAuth keeps the resume endpoint on the
// authenticated side of the group: it re-queues work inside a tenant schema.
func TestRegisterExtensionRoutes_ResumeRequiresAuth(t *testing.T) {
	gin.SetMode(gin.TestMode)

	router := gin.New()
	api := router.Group("/api")
	RegisterExtensionRoutes(api, nil, nil, handlers.NewScrapeHandler(nil, nil))

	req := httptest.NewRequest(http.MethodPost, "/api/extensions/scrape/job-1/resume", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code,
		"an unauthenticated resume must be rejected by the auth middleware")
}
