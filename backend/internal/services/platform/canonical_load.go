package platform

import (
	"context"
	"strconv"

	"github.com/omni/backend/internal/config"
	"github.com/omni/backend/internal/models"
	"github.com/omni/backend/internal/repositories"
	shopeePkg "github.com/omni/backend/pkg/shopee"
)

// ShopeeCanonicalFields is the read-only projection of the canonical
// credential tables (credential_app_configs + credential_connections) into
// the fields ShopeeConfigManager consumes.
//
// Bug C: sync worker previously loaded from the legacy platform_configs
// key-value store. After the Phase 5 migration to canonical tables that
// store fell out of sync — some tenants now have canonical rows only.
// This projection lets the config manager prefer canonical, then fall back
// to legacy.
type ShopeeCanonicalFields struct {
	LivePair         shopeePkg.PartnerPair
	TestPair         shopeePkg.PartnerPair
	ActivePartnerEnv string // "live" | "test", defaults to "live" when unset
	ShopID           string // from CredentialConnection.StoreIdentifier
	AccessToken      string
	RefreshToken     string
	TokenExpiry      int64
	RefreshExpiry    int64
}

// CanonicalToShopeeFields projects an app_config + optional connection into
// the flat ShopeeCanonicalFields shape. Both inputs may be nil (fresh
// tenant); the resulting zero values instruct the caller to fall back to
// legacy loading.
func CanonicalToShopeeFields(app *models.CredentialAppConfig, conn *models.CredentialConnection) ShopeeCanonicalFields {
	out := ShopeeCanonicalFields{ActivePartnerEnv: "live"}
	if app != nil {
		out.LivePair = shopeePkg.PartnerPair{ID: app.PartnerID, Key: app.PartnerKey}
		out.TestPair = shopeePkg.PartnerPair{ID: app.TestPartnerID, Key: app.TestPartnerKey}
		if app.ActivePartnerEnv != "" {
			out.ActivePartnerEnv = app.ActivePartnerEnv
		}
	}
	if conn != nil {
		out.ShopID = conn.StoreIdentifier
		out.AccessToken = conn.AccessToken
		out.RefreshToken = conn.RefreshToken
		out.TokenExpiry = conn.TokenExpiry
		out.RefreshExpiry = conn.RefreshExpiry
	}
	return out
}

// LoadShopeeCanonical fetches (app_config, first-active-connection) from the
// canonical tables for the tenant and translates them via
// CanonicalToShopeeFields. Returns zero-value fields (not an error) when
// either row is missing so the caller can decide to fall back to legacy.
//
// Kept off ShopeeConfigManager (free function) so it's easier to test in
// isolation and doesn't require reaching for the manager's private fields.
func LoadShopeeCanonical(ctx context.Context, tenantID string) (ShopeeCanonicalFields, error) {
	db, err := config.GetTenantDBByID(tenantID)
	if err != nil {
		return ShopeeCanonicalFields{ActivePartnerEnv: "live"}, err
	}
	repo := repositories.NewCredentialRepository(db)
	app, err := repo.GetAppConfig(ctx, tenantID, models.PlatformShopee)
	if err != nil {
		return ShopeeCanonicalFields{ActivePartnerEnv: "live"}, err
	}
	conns, err := repo.ListConnections(ctx, tenantID, models.PlatformShopee)
	if err != nil {
		return CanonicalToShopeeFields(app, nil), err
	}
	var conn *models.CredentialConnection
	for i := range conns {
		if conns[i].DisabledAt == nil {
			conn = &conns[i]
			break
		}
	}
	return CanonicalToShopeeFields(app, conn), nil
}

// helper: canonical fields → strings the config manager fields want.
func (f ShopeeCanonicalFields) applyTo(m *ShopeeConfigManager) {
	m.ApplyActivePair(f.LivePair, f.TestPair, f.ActivePartnerEnv)
	if f.ShopID != "" {
		m.ShopID = f.ShopID
	}
	if f.AccessToken != "" {
		_ = m.SetTokens(f.AccessToken, f.RefreshToken)
		// Also mirror into the key-value store used by GetTokenExpiry etc.
		if f.TokenExpiry > 0 {
			_ = m.SetConfig("tokenExpiry", strconv.FormatInt(f.TokenExpiry, 10))
		}
		if f.RefreshExpiry > 0 {
			_ = m.SetConfig("refreshTokenExpiry", strconv.FormatInt(f.RefreshExpiry, 10))
		}
	}
}
