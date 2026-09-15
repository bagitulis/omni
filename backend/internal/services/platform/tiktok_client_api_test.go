package platform

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"testing"
)

// TikTok API 202502 migration — Phase 7 Item A.
//
// Original code used POST /product/202309/products/search. Per TikTok
// changelog, /product/202502/products/search is the successor; both are live
// (dual-run), but we adopt the newer version to stay on the maintained path.

// fakeTransport captures the last request URL for later assertions.
type fakeTransport struct {
	lastPath string
	lastMethod string
	body       []byte
}

func (t *fakeTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	t.lastMethod = req.Method
	t.lastPath = req.URL.Path
	if req.Body != nil {
		defer req.Body.Close()
		b, _ := io.ReadAll(req.Body)
		t.body = b
	}
	// Minimal well-formed response so GetProductList returns cleanly.
	resp := map[string]interface{}{
		"code":    0,
		"message": "success",
		"data": map[string]interface{}{
			"products":    []interface{}{},
			"total_count": 0,
		},
	}
	respBody, _ := json.Marshal(resp)
	return &http.Response{
		StatusCode: 200,
		Body:       io.NopCloser(strings.NewReader(string(respBody))),
		Header:     make(http.Header),
	}, nil
}

// newFakeTikTokClient assembles a TiktokAPIClient with fake HTTP so tests
// can inspect the outgoing path without a real network. Directly reaches into
// package internals (same package) so no exported API is bent for tests.
func newFakeTikTokClient(t *testing.T) (*TiktokAPIClient, *fakeTransport) {
	t.Helper()
	fake := &fakeTransport{}
	client := &TiktokAPIClient{
		config: &TiktokConfigManager{
			BaseConfigManager: NewBaseConfigManager(PlatformTiktok, "tenant-a"),
			AppKey:            "test-app-key",
			AppSecret:         "test-app-secret",
			ShopCipher:        "test-shop-cipher",
		},
		initialized: true,
		httpClient:  &http.Client{Transport: fake},
	}
	// IsConfigured() also checks GetAccessToken(); seed via SetTokens (the
	// exported helper that writes to BaseConfigManager's in-memory field).
	_ = client.config.SetTokens("test-access-token", "test-refresh-token")
	return client, fake
}

// TestGetProductList_UsesV202502 — RED-first: the current code calls
// /product/202309/products/search. This test asserts the new expected path.
func TestGetProductList_UsesV202502(t *testing.T) {
	client, fake := newFakeTikTokClient(t)

	_, err := client.GetProductList(context.Background(), 0, 10)
	if err != nil {
		t.Fatalf("GetProductList: %v", err)
	}

	wantPath := "/product/202502/products/search"
	if fake.lastPath != wantPath {
		t.Errorf("path = %q, want %q (Phase 7 Item A migration 202309 → 202502)",
			fake.lastPath, wantPath)
	}
	if fake.lastMethod != http.MethodPost {
		t.Errorf("method = %q, want POST", fake.lastMethod)
	}
}
