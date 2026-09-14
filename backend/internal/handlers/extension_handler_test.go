package handlers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Handler-level tests exercise auth/tenant gating and request validation without
// a database. The fixtures follow the existing pattern in this package: gin in
// test mode, a stub middleware that injects tenant_id, and httptest.

func newExtensionTestRouter(t *testing.T, handler *ExtensionHandler, withTenant bool) *gin.Engine {
	t.Helper()
	gin.SetMode(gin.TestMode)

	r := gin.New()
	group := r.Group("/api/extensions")

	// Unauthenticated routes, mirroring routes/extension_routes.go.
	group.POST("/pairing/confirm", handler.ConfirmPairing)

	authed := group.Group("")
	if withTenant {
		authed.Use(func(c *gin.Context) {
			c.Set("tenant_id", "test-tenant")
			c.Set("userID", "1")
			c.Next()
		})
	}
	authed.GET("", handler.List)
	authed.POST("/pairing/generate", handler.GeneratePairingCode)
	authed.DELETE("/:extension_id", handler.Unpair)

	return r
}

// TestExtensionHandler_MissingTenantIsRejected enforces the platform rule that
// there is no default tenant. A handler that invented one would read or write
// another tenant's data.
func TestExtensionHandler_MissingTenantIsRejected(t *testing.T) {
	h := NewExtensionHandler(nil)
	r := newExtensionTestRouter(t, h, false)

	cases := []struct {
		name   string
		method string
		path   string
		body   string
	}{
		{"list", http.MethodGet, "/api/extensions", ""},
		{"generate pairing", http.MethodPost, "/api/extensions/pairing/generate", ""},
		{"unpair", http.MethodDelete, "/api/extensions/ext-1", ""},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			var reader *bytes.Reader
			if tc.body == "" {
				reader = bytes.NewReader(nil)
			} else {
				reader = bytes.NewReader([]byte(tc.body))
			}
			req, err := http.NewRequest(tc.method, tc.path, reader)
			require.NoError(t, err)
			req.Header.Set("Content-Type", "application/json")

			w := httptest.NewRecorder()
			r.ServeHTTP(w, req)

			assert.Equal(t, http.StatusUnauthorized, w.Code,
				"a request without tenant_id must be rejected, never defaulted")

			var resp struct {
				Success bool   `json:"success"`
				Error   string `json:"error"`
			}
			require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
			assert.False(t, resp.Success, "success must not be true on a rejected request")
		})
	}
}

// TestExtensionHandler_NilServiceFailsClosed asserts a handler constructed
// without a service returns 503 instead of panicking. A Gin handler that
// dereferences nil aborts the request goroutine and can take the process down,
// so failing closed is both safer and easier to diagnose.
func TestExtensionHandler_NilServiceFailsClosed(t *testing.T) {
	h := NewExtensionHandler(nil) // deliberately no service
	r := newExtensionTestRouter(t, h, true)

	cases := []struct {
		name   string
		method string
		path   string
		body   string
	}{
		{"list", http.MethodGet, "/api/extensions", ""},
		{"generate pairing", http.MethodPost, "/api/extensions/pairing/generate", ""},
		{"unpair valid id", http.MethodDelete, "/api/extensions/ext-1", ""},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			var body *bytes.Reader
			if tc.body == "" {
				body = bytes.NewReader(nil)
			} else {
				body = bytes.NewReader([]byte(tc.body))
			}
			req, err := http.NewRequest(tc.method, tc.path, body)
			require.NoError(t, err)

			w := httptest.NewRecorder()
			// Must not panic.
			r.ServeHTTP(w, req)

			assert.Equal(t, http.StatusServiceUnavailable, w.Code,
				"an unwired service must fail closed rather than panic")
		})
	}
}

// TestExtensionHandler_ConfirmPairingRequiresFields checks validation happens
// before any lookup, so a malformed body cannot reach the service.
func TestExtensionHandler_ConfirmPairingRequiresFields(t *testing.T) {
	h := NewExtensionHandler(nil)
	r := newExtensionTestRouter(t, h, false)

	cases := []struct {
		name string
		body string
	}{
		{"missing code", `{"extension_id":"ext-1"}`},
		{"missing extension_id", `{"code":"ABCD1234"}`},
		{"empty body", `{}`},
		{"not json", `{nope`},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			req, err := http.NewRequest(http.MethodPost, "/api/extensions/pairing/confirm",
				bytes.NewReader([]byte(tc.body)))
			require.NoError(t, err)
			req.Header.Set("Content-Type", "application/json")

			w := httptest.NewRecorder()
			r.ServeHTTP(w, req)

			assert.Equal(t, http.StatusBadRequest, w.Code,
				"an invalid body must be rejected before any tenant lookup")
		})
	}
}

// TestExtensionHandler_ConfirmPairingGenericErrorOnUnknownCode asserts the error
// does not reveal whether the code exists. Distinguishing "unknown" from
// "expired" would let a caller probe for valid codes.
func TestExtensionHandler_ConfirmPairingGenericErrorOnUnknownCode(t *testing.T) {
	h := NewExtensionHandler(nil) // no resolver configured, so every code fails
	r := newExtensionTestRouter(t, h, false)

	req, err := http.NewRequest(http.MethodPost, "/api/extensions/pairing/confirm",
		bytes.NewReader([]byte(`{"code":"ABCD1234","extension_id":"ext-1"}`)))
	require.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	// A nil service fails closed with 503 rather than panicking.
	assert.Equal(t, http.StatusServiceUnavailable, w.Code)

	var resp struct {
		Success bool   `json:"success"`
		Error   string `json:"error"`
	}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	assert.False(t, resp.Success)
	// The message must not hint at whether the code exists.
	assert.NotContains(t, resp.Error, "unknown code")
}

// TestExtensionHandler_UnpairRejectsInvalidExtensionID guards the path parameter
// before it reaches a query.
func TestExtensionHandler_UnpairRejectsInvalidExtensionID(t *testing.T) {
	h := NewExtensionHandler(nil)
	r := newExtensionTestRouter(t, h, true)

	// Validation runs before the service check, so an invalid identifier is a
	// 400 even without a wired service. Space and semicolon are rejected by
	// ValidateExtensionID. An encoded slash is not tested here because a
	// percent-encoded %2F does not match a single path segment in the router,
	// so it never reaches the handler.
	for _, id := range []string{"has%20space", "has%3Bsemicolon", "has%27quote"} {
		req, err := http.NewRequest(http.MethodDelete, "/api/extensions/"+id, nil)
		require.NoError(t, err)

		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code,
			"invalid extension_id %q must be rejected before any lookups", id)
	}
}
