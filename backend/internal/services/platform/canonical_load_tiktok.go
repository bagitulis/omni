// Phase 11.4 — TikTok canonical projection (Batch 2 gap fix).
//
// Same pattern as Shopee canonical (Batch 2 6e565401): the sync worker
// used to read from the legacy platform_configs key-value store, which
// fell out of sync for tenants whose only OAuth flow was via the
// canonical credential_connections tables. TikTok Yumna is exactly this
// case — refresh flow writes credential_connections.shop_cipher but
// TiktokConfigManager.LoadConfig only reads platform_configs, so the
// client reports "shopCipher is missing" and all sync calls fail with
// "tiktok client not initialized" even though the token is fresh.

package platform

import (
	"context"
	"strconv"

	"github.com/omni/backend/internal/config"
	"github.com/omni/backend/internal/models"
	"github.com/omni/backend/internal/repositories"
)

// TiktokCanonicalFields is the read-only projection of canonical credential
// tables into the fields TiktokConfigManager consumes.
type TiktokCanonicalFields struct {
	AppKey        string
	AppSecret     string
	ShopID        string // from CredentialConnection.StoreIdentifier
	ShopCipher    string // from CredentialConnection.ShopCipher — the missing bit
	AccessToken   string
	RefreshToken  string
	TokenExpiry   int64
	RefreshExpiry int64
}

// CanonicalToTiktokFields projects an app_config + optional connection
// into the flat TiktokCanonicalFields shape. Both inputs may be nil.
func CanonicalToTiktokFields(app *models.CredentialAppConfig, conn *models.CredentialConnection) TiktokCanonicalFields {
	out := TiktokCanonicalFields{}
	if app != nil {
		out.AppKey = app.AppKey
		out.AppSecret = app.AppSecret
	}
	if conn != nil {
		out.ShopID = conn.StoreIdentifier
		out.ShopCipher = conn.ShopCipher
		out.AccessToken = conn.AccessToken
		out.RefreshToken = conn.RefreshToken
		out.TokenExpiry = conn.TokenExpiry
		out.RefreshExpiry = conn.RefreshExpiry
	}
	return out
}

// LoadTiktokCanonical fetches (app_config, first-active-connection) from
// the canonical tables for the tenant. Returns zero-value fields (not an
// error) when either row is missing so the caller can fall back to legacy.
func LoadTiktokCanonical(ctx context.Context, tenantID string) (TiktokCanonicalFields, error) {
	db, err := config.GetTenantDBByID(tenantID)
	if err != nil {
		return TiktokCanonicalFields{}, err
	}
	repo := repositories.NewCredentialRepository(db)
	app, err := repo.GetAppConfig(ctx, tenantID, models.PlatformTiktok)
	if err != nil {
		return TiktokCanonicalFields{}, err
	}
	conns, err := repo.ListConnections(ctx, tenantID, models.PlatformTiktok)
	if err != nil {
		return CanonicalToTiktokFields(app, nil), err
	}
	var conn *models.CredentialConnection
	for i := range conns {
		if conns[i].DisabledAt == nil {
			conn = &conns[i]
			break
		}
	}
	return CanonicalToTiktokFields(app, conn), nil
}

// applyTo writes the canonical fields into a TiktokConfigManager, mirroring
// the applyTo helper on ShopeeCanonicalFields.
func (f TiktokCanonicalFields) applyTo(m *TiktokConfigManager) {
	if f.AppKey != "" {
		m.AppKey = f.AppKey
	}
	if f.AppSecret != "" {
		m.AppSecret = f.AppSecret
	}
	if f.ShopID != "" {
		m.ShopID = f.ShopID
	}
	if f.ShopCipher != "" {
		m.ShopCipher = f.ShopCipher
	}
	if f.AccessToken != "" {
		_ = m.SetTokens(f.AccessToken, f.RefreshToken)
		if f.TokenExpiry > 0 {
			_ = m.SetConfig("tokenExpiry", strconv.FormatInt(f.TokenExpiry, 10))
		}
		if f.RefreshExpiry > 0 {
			_ = m.SetConfig("refreshTokenExpiry", strconv.FormatInt(f.RefreshExpiry, 10))
		}
	}
}
