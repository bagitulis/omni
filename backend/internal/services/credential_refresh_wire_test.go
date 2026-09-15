package services

import (
	"context"
	"testing"

	"github.com/omni/backend/internal/models"
)

// Bug A — POST /credentials/.../refresh must actually invoke the remote OAuth
// refresh endpoint. Previously it only flipped the row status and wrote a
// success audit event, misleading callers.
//
// The fix injects a TokenRefresher into CredentialApiService. This unit test
// covers the wire: given a fake refresher, the service reports refresh
// success / failure faithfully.

// fakeRefresher implements CredentialTokenRefresher and records the calls it
// received so the test can assert the service actually delegated to it.
type fakeRefresher struct {
	calls        []string
	returnErr    error
	returnAccess string
	returnRefresh string
}

func (f *fakeRefresher) RefreshShopeeToken(ctx context.Context, tenantID string) (string, string, error) {
	f.calls = append(f.calls, "shopee:"+tenantID)
	return f.returnAccess, f.returnRefresh, f.returnErr
}
func (f *fakeRefresher) RefreshLazadaToken(ctx context.Context, tenantID string) (string, string, error) {
	f.calls = append(f.calls, "lazada:"+tenantID)
	return f.returnAccess, f.returnRefresh, f.returnErr
}
func (f *fakeRefresher) RefreshTiktokToken(ctx context.Context, tenantID string) (string, string, error) {
	f.calls = append(f.calls, "tiktok:"+tenantID)
	return f.returnAccess, f.returnRefresh, f.returnErr
}

// TestRefreshConnection_DelegatesToRefresher_Success — happy path.
// Uses the small business helper (not the full ChangeConnectionStatus) so
// this test doesn't require a live DB.
func TestRefreshConnection_DelegatesToRefresher_Success(t *testing.T) {
	fake := &fakeRefresher{returnAccess: "new-at", returnRefresh: "new-rt"}
	access, refresh, err := invokeRemoteRefresh(context.Background(), fake, models.PlatformShopee, "tenant-a")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if access != "new-at" || refresh != "new-rt" {
		t.Errorf("got (%q, %q), want (new-at, new-rt)", access, refresh)
	}
	if len(fake.calls) != 1 || fake.calls[0] != "shopee:tenant-a" {
		t.Errorf("calls = %v, want [shopee:tenant-a]", fake.calls)
	}
}

// TestRefreshConnection_DelegatesToRefresher_ErrorPropagates — negative path.
// A real remote failure (e.g. refresh_token expired) MUST bubble up so the
// service can return proper HTTP 500 + record an audit event with
// status=failure. This is exactly the bug from the smoke report: the service
// silently reported success.
func TestRefreshConnection_DelegatesToRefresher_ErrorPropagates(t *testing.T) {
	fake := &fakeRefresher{returnErr: errRemoteRefreshFailed("refresh_token invalid")}
	_, _, err := invokeRemoteRefresh(context.Background(), fake, models.PlatformLazada, "tenant-b")
	if err == nil {
		t.Fatal("expected error from failing refresher, got nil")
	}
	if len(fake.calls) != 1 || fake.calls[0] != "lazada:tenant-b" {
		t.Errorf("calls = %v, want [lazada:tenant-b]", fake.calls)
	}
}

// TestRefreshConnection_UnknownPlatform — negative path: platform value
// outside the supported set is a caller bug; return a descriptive error
// rather than silently no-op.
func TestRefreshConnection_UnknownPlatform(t *testing.T) {
	fake := &fakeRefresher{}
	_, _, err := invokeRemoteRefresh(context.Background(), fake, "shopify", "tenant-x")
	if err == nil {
		t.Fatal("expected error for unknown platform")
	}
	if len(fake.calls) != 0 {
		t.Errorf("no delegation should happen for unknown platform; calls=%v", fake.calls)
	}
}

// TestRefreshConnection_NilRefresher — negative path: service constructed
// without a token refresher (backward compat mode). Must return a specific
// error so callers know why refresh is a no-op.
func TestRefreshConnection_NilRefresher(t *testing.T) {
	_, _, err := invokeRemoteRefresh(context.Background(), nil, models.PlatformShopee, "t")
	if err == nil {
		t.Fatal("expected error when refresher is nil")
	}
}

// TestRefreshConnection_RoutesTiktok — smoke path for the third platform.
func TestRefreshConnection_RoutesTiktok(t *testing.T) {
	fake := &fakeRefresher{returnAccess: "at"}
	_, _, err := invokeRemoteRefresh(context.Background(), fake, models.PlatformTiktok, "t3")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(fake.calls) != 1 || fake.calls[0] != "tiktok:t3" {
		t.Errorf("calls = %v, want [tiktok:t3]", fake.calls)
	}
}
