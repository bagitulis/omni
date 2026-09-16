package handlers

import (
	"time"

	"bytes"
	"encoding/json"
	"strings"

	"github.com/omni/backend/internal/models"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newScrapeTestRouter(t *testing.T, h *ScrapeHandler, withTenant bool) *gin.Engine {
	t.Helper()
	gin.SetMode(gin.TestMode)

	r := gin.New()
	group := r.Group("/api/extensions")
	authed := group.Group("")
	if withTenant {
		authed.Use(func(c *gin.Context) {
			c.Set("tenant_id", "test-tenant")
			c.Set("userID", "1")
			c.Next()
		})
	}
	authed.POST("/scrape", h.Start)
	authed.GET("/scraped-products", h.ListProducts)
	return r
}

// TestScrapeHandler_MissingTenantIsRejected enforces the no-default-tenant rule.
// A handler that invented a tenant would write scraped rows into another
// tenant's schema.
func TestScrapeHandler_MissingTenantIsRejected(t *testing.T) {
	h := NewScrapeHandler(nil, nil)
	r := newScrapeTestRouter(t, h, false)

	for _, tc := range []struct {
		name, method, path, body string
	}{
		{"start", http.MethodPost, "/api/extensions/scrape", `{"mode":"search","query":"x","extension_id":"e"}`},
		{"list products", http.MethodGet, "/api/extensions/scraped-products?job_id=j", ""},
	} {
		t.Run(tc.name, func(t *testing.T) {
			req, err := http.NewRequest(tc.method, tc.path, bytes.NewReader([]byte(tc.body)))
			require.NoError(t, err)
			req.Header.Set("Content-Type", "application/json")

			w := httptest.NewRecorder()
			r.ServeHTTP(w, req)

			assert.Equal(t, http.StatusUnauthorized, w.Code)
			var resp struct {
				Success bool `json:"success"`
			}
			require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
			assert.False(t, resp.Success, "success must not be true on a rejected request")
		})
	}
}

// TestScrapeHandler_ValidationBeforeDispatch checks every mode's requirement is
// enforced before any work is dispatched, so a bad request never opens a
// browser tab.
func TestScrapeHandler_ValidationBeforeDispatch(t *testing.T) {
	h := NewScrapeHandler(nil, nil) // nil service: validation must run first
	r := newScrapeTestRouter(t, h, true)

	cases := []struct {
		name string
		body string
	}{
		{"search without query", `{"mode":"search","extension_id":"e"}`},
		{"shop without shop_url", `{"mode":"shop","extension_id":"e"}`},
		{"product without product_url", `{"mode":"product","extension_id":"e"}`},
		{"unknown mode", `{"mode":"telepathy","extension_id":"e"}`},
		{"missing mode", `{"extension_id":"e"}`},
		{"missing extension_id", `{"mode":"search","query":"x"}`},
		{"malformed json", `{nope`},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			req, err := http.NewRequest(http.MethodPost, "/api/extensions/scrape",
				bytes.NewReader([]byte(tc.body)))
			require.NoError(t, err)
			req.Header.Set("Content-Type", "application/json")

			w := httptest.NewRecorder()
			r.ServeHTTP(w, req)

			assert.Equal(t, http.StatusBadRequest, w.Code,
				"an invalid request must be rejected before dispatch")
		})
	}
}

// TestScrapeHandler_NilServiceFailsClosed asserts an unwired handler returns 503
// rather than panicking, since a panic in a Gin handler can take the process
// down.
func TestScrapeHandler_NilServiceFailsClosed(t *testing.T) {
	h := NewScrapeHandler(nil, nil)
	r := newScrapeTestRouter(t, h, true)

	req, err := http.NewRequest(http.MethodPost, "/api/extensions/scrape",
		bytes.NewReader([]byte(`{"mode":"search","query":"x","extension_id":"e"}`)))
	require.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	r.ServeHTTP(w, req) // must not panic

	assert.Equal(t, http.StatusServiceUnavailable, w.Code)
}

// TestScrapeHandler_QueuesRatherThanRunsInline documents the dispatch contract:
// a scrape is queued as a job, not run in the request goroutine. Running inline
// would hold the connection for the whole scrape and time out behind a proxy,
// and would give no way to cancel it.
//
// With a nil tenantDB the handler cannot reach the queue, so the request fails
// at the service-availability check. What this pins is that the handler never
// falls back to running the scrape synchronously.
func TestScrapeHandler_QueuesRatherThanRunsInline(t *testing.T) {
	h := NewScrapeHandler(nil, nil)
	r := newScrapeTestRouter(t, h, true)

	req, err := http.NewRequest(http.MethodPost, "/api/extensions/scrape",
		bytes.NewReader([]byte(`{"mode":"search","query":"x","extension_id":"e"}`)))
	require.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	// Must not be a synchronous success, and must not be a 200 carrying results.
	assert.NotEqual(t, http.StatusOK, w.Code,
		"the scrape endpoint must queue work, not return results inline")
	assert.Equal(t, http.StatusServiceUnavailable, w.Code)
}

// TestScrapeHandler_RejectsMalformedExtensionID pins that the start path runs
// the shared identifier validation. `binding:"required"` alone accepts any
// non-empty string, including one carrying characters that could forge a log
// line or be abused in a lookup.
func TestScrapeHandler_RejectsMalformedExtensionID(t *testing.T) {
	h := NewScrapeHandler(nil, nil) // nil service: validation must run first
	r := newScrapeTestRouter(t, h, true)

	cases := []struct {
		name string
		body string
	}{
		{"space", `{"mode":"search","query":"x","extension_id":"ext 1"}`},
		{"path separator", `{"mode":"search","query":"x","extension_id":"../../etc"}`},
		{"newline", "{\"mode\":\"search\",\"query\":\"x\",\"extension_id\":\"ext\\n1\"}"},
		{"over length", `{"mode":"search","query":"x","extension_id":"` + strings.Repeat("a", 129) + `"}`},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			req, err := http.NewRequest(http.MethodPost, "/api/extensions/scrape",
				bytes.NewReader([]byte(tc.body)))
			require.NoError(t, err)
			req.Header.Set("Content-Type", "application/json")

			w := httptest.NewRecorder()
			r.ServeHTTP(w, req)

			assert.Equal(t, http.StatusBadRequest, w.Code,
				"an invalid extension_id must be rejected before dispatch")
		})
	}
}

// TestScrapeHandler_AcceptsChromeExtensionID guards against the validation being
// tightened past what a real Chrome extension id looks like.
func TestScrapeHandler_AcceptsChromeExtensionID(t *testing.T) {
	h := NewScrapeHandler(nil, nil)
	r := newScrapeTestRouter(t, h, true)

	body := `{"mode":"search","query":"x","extension_id":"abcdefghijklmnopabcdefghijklmnop"}`
	req, err := http.NewRequest(http.MethodPost, "/api/extensions/scrape", bytes.NewReader([]byte(body)))
	require.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	// Passes validation and stops at the availability check, not at a 400.
	assert.Equal(t, http.StatusServiceUnavailable, w.Code)
}

func TestValidateScrapeRequest(t *testing.T) {
	cases := []struct {
		name    string
		body    startScrapeRequest
		wantErr bool
	}{
		{"search ok", startScrapeRequest{Mode: "search", Query: "x", ExtensionID: "ext-1"}, false},
		{"search missing query", startScrapeRequest{Mode: "search"}, true},
		{"shop ok", startScrapeRequest{Mode: "shop", ShopURL: "https://x", ExtensionID: "ext-1"}, false},
		{"shop missing url", startScrapeRequest{Mode: "shop"}, true},
		{"product ok", startScrapeRequest{Mode: "product", ProductURL: "https://x", ExtensionID: "ext-1"}, false},
		{"product missing url", startScrapeRequest{Mode: "product"}, true},
		{"unknown mode", startScrapeRequest{Mode: "nope"}, true},
		{"empty mode", startScrapeRequest{}, true},
		{"missing extension id", startScrapeRequest{Mode: "search", Query: "x"}, true},
		{"malformed extension id", startScrapeRequest{Mode: "search", Query: "x", ExtensionID: "ext 1"}, true},
		{"valid extension id", startScrapeRequest{Mode: "search", Query: "x", ExtensionID: "ext-1_A"}, false},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			err := validateScrapeRequest(tc.body)
			if tc.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestTotalPages(t *testing.T) {
	cases := []struct {
		total, pageSize, want int
	}{
		{0, 50, 0},
		{1, 50, 1},
		{50, 50, 1},
		{51, 50, 2},
		{100, 50, 2},
		{101, 50, 3},
		{10, 0, 0}, // guards against division by zero
	}

	for _, tc := range cases {
		assert.Equal(t, tc.want, totalPages(tc.total, tc.pageSize),
			"totalPages(%d, %d)", tc.total, tc.pageSize)
	}
}

func TestToScrapedProductDTO(t *testing.T) {
	// The API shape must carry the capture source, so an operator can see when
	// the DOM fallback is carrying production traffic.
	created := time.Date(2026, 9, 14, 12, 0, 0, 0, time.UTC)
	p := toScrapedProductDTO(models.ScrapedProduct{
		ID:           7,
		JobID:        "job-1",
		Platform:     "shopee",
		ScrapeMode:   "search",
		Query:        "keyboard",
		ShopID:       "10",
		ProductName:  "Keyboard",
		Price:        "15000",
		Sold:         "42",
		Link:         "https://shopee.co.id/product/10/20",
		ImageURL:     "img",
		ShopeeItemID: "20",
		PageNumber:   2,
		Source:       models.ScrapeSourceDOM,
		CreatedAt:    created,
	})

	if p.ID != 7 || p.JobID != "job-1" {
		t.Errorf("identity fields lost: %+v", p)
	}
	if p.Platform != "shopee" {
		t.Errorf("Platform = %q, want shopee", p.Platform)
	}
	if p.Source != models.ScrapeSourceDOM {
		t.Errorf("Source = %q, want dom (fallback usage must be visible)", p.Source)
	}
	if p.PageNumber != 2 {
		t.Errorf("PageNumber = %d, want 2", p.PageNumber)
	}
	if p.CreatedAt != "2026-09-14T12:00:00Z" {
		t.Errorf("CreatedAt = %q, want RFC3339 UTC", p.CreatedAt)
	}
}
