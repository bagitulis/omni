package platform

import (
	"testing"

	"github.com/omni/backend/internal/models"
)

// Phase 11.4 — RED tests for TikTok canonical projection.
// Same pattern as Shopee canonical (Batch 2) but for TikTok's
// AppKey/AppSecret/ShopID/ShopCipher/AccessToken/RefreshToken shape.

func TestCanonicalToTiktokFields_HappyProjection(t *testing.T) {
	app := &models.CredentialAppConfig{
		Platform:  "tiktok",
		AppKey:    "6h2cj0a07atkt",
		AppSecret: "shhh-app-secret",
	}
	conn := &models.CredentialConnection{
		StoreIdentifier: "7495139858797791519",
		ShopCipher:      "TT-cipher-abc123",
		AccessToken:     "at-fresh",
		RefreshToken:    "rt-fresh",
		TokenExpiry:     1_800_000_000_000,
		RefreshExpiry:   1_900_000_000_000,
	}
	got := CanonicalToTiktokFields(app, conn)
	if got.AppKey != "6h2cj0a07atkt" || got.AppSecret != "shhh-app-secret" {
		t.Fatalf("app credentials not projected: %+v", got)
	}
	if got.ShopID != "7495139858797791519" {
		t.Fatalf("shop_id not projected: got %q", got.ShopID)
	}
	if got.ShopCipher != "TT-cipher-abc123" {
		t.Fatalf("shop_cipher not projected: got %q", got.ShopCipher)
	}
	if got.AccessToken != "at-fresh" || got.RefreshToken != "rt-fresh" {
		t.Fatalf("tokens not projected: %+v", got)
	}
	if got.TokenExpiry != 1_800_000_000_000 || got.RefreshExpiry != 1_900_000_000_000 {
		t.Fatalf("expiries not projected: %+v", got)
	}
}

func TestCanonicalToTiktokFields_NilInputsReturnZero(t *testing.T) {
	got := CanonicalToTiktokFields(nil, nil)
	if got.AppKey != "" || got.AppSecret != "" || got.ShopID != "" || got.ShopCipher != "" || got.AccessToken != "" {
		t.Fatalf("expected all-zero fields, got %+v", got)
	}
}

func TestCanonicalToTiktokFields_AppOnly(t *testing.T) {
	app := &models.CredentialAppConfig{Platform: "tiktok", AppKey: "K", AppSecret: "S"}
	got := CanonicalToTiktokFields(app, nil)
	if got.AppKey != "K" || got.AppSecret != "S" {
		t.Fatalf("app-only projection wrong: %+v", got)
	}
	// Connection-derived fields must remain zero — signals caller to fall back.
	if got.ShopID != "" || got.ShopCipher != "" || got.AccessToken != "" {
		t.Fatalf("connection fields must stay empty when conn is nil: %+v", got)
	}
}

func TestCanonicalToTiktokFields_ConnOnly(t *testing.T) {
	conn := &models.CredentialConnection{
		StoreIdentifier: "shop-1",
		ShopCipher:      "cipher-1",
		AccessToken:     "at",
	}
	got := CanonicalToTiktokFields(nil, conn)
	if got.ShopID != "shop-1" || got.ShopCipher != "cipher-1" || got.AccessToken != "at" {
		t.Fatalf("conn-only projection wrong: %+v", got)
	}
	if got.AppKey != "" || got.AppSecret != "" {
		t.Fatalf("app fields must stay empty when app is nil: %+v", got)
	}
}
