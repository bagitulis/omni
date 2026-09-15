package platform

import (
	"context"
	"strconv"

	"github.com/omni/backend/internal/config"
	shopeePkg "github.com/omni/backend/pkg/shopee"
)

// ShopeeConfigManager handles Shopee-specific configuration
type ShopeeConfigManager struct {
	*BaseConfigManager
	PartnerID  string
	PartnerKey string
	ShopID     string
}

// NewShopeeConfigManager creates a Shopee config manager
func NewShopeeConfigManager(tenantID string) *ShopeeConfigManager {
	return &ShopeeConfigManager{
		BaseConfigManager: NewBaseConfigManager(PlatformShopee, tenantID),
	}
}

// LoadConfig loads Shopee config from database.
//
// Bug C (Phase 9+): prefer canonical credential_app_configs +
// credential_connections; only fall back to legacy platform_configs when
// canonical rows are empty. Also applies ActivePartnerEnv via
// ApplyActivePair so the Live/Test switch column added in Phase 8 is
// honored at signing time.
func (m *ShopeeConfigManager) LoadConfig(ctx context.Context) error {
	// 1. Try canonical tables first.
	canonical, cErr := LoadShopeeCanonical(ctx, m.tenantID)
	if cErr != nil {
		m.logger.WithTenantID(m.tenantID).Warn("Canonical Shopee load failed, will try legacy: " + cErr.Error())
	}
	if canonical.LivePair.IsValid() {
		canonical.applyTo(m)
	} else {
		// 2a. Fallback: load global credentials from system.db.
		globalConfig := config.GetGlobalConfigService()
		if globalConfig != nil {
			creds, err := globalConfig.GetShopeeCredentials()
			if err == nil {
				m.PartnerID = creds.PartnerID
				m.PartnerKey = creds.PartnerKey
			}
		}
	}

	// 2b. Fallback: load tenant-specific legacy config (ShopID, tokens).
	// Even when canonical.LivePair is valid, some legacy-only fields (e.g.
	// pushPartnerKey) may still be needed; the legacy loader is additive
	// and won't overwrite anything the canonical loader just populated.
	if err := m.LoadConfigFromDB(ctx); err != nil {
		m.logger.WithTenantID(m.tenantID).Warn("Failed to load legacy config from DB: " + err.Error())
	}

	// Parse ShopID from configs if canonical didn't supply one.
	if m.ShopID == "" {
		if shopIDStr, ok := m.GetConfig("shopId"); ok && shopIDStr != "" {
			m.ShopID = shopIDStr
		}
	}

	// Log status.
	if m.PartnerID != "" && m.PartnerKey != "" && m.ShopID != "" {
		m.logger.WithTenantID(m.tenantID).Info("Shopee config loaded successfully")
	} else {
		if m.PartnerID == "" || m.PartnerKey == "" {
			m.logger.WithTenantID(m.tenantID).Warn("Shopee credentials missing (canonical + legacy)")
		}
		if m.ShopID == "" {
			m.logger.WithTenantID(m.tenantID).Warn("Shopee shopId missing")
		}
	}

	return nil
}

// IsConfigured returns whether Shopee is properly configured
func (m *ShopeeConfigManager) IsConfigured() bool {
	return m.PartnerID != "" && m.PartnerKey != "" && m.ShopID != ""
}

// ApplyActivePair picks between the live and test partner pairs per the
// tenant's `active_partner_env` toggle and writes the winning pair into
// PartnerID / PartnerKey. Safety default is live (see
// pkg/shopee.SelectActivePair for the exact rule).
//
// Callers with access to `credential_app_configs` build both pairs from
// PartnerID/PartnerKey + TestPartnerID/TestPartnerKey and pass ActivePartnerEnv
// as the env; the config manager stays free of DB imports.
func (m *ShopeeConfigManager) ApplyActivePair(live, test shopeePkg.PartnerPair, env string) {
	chosen := shopeePkg.SelectActivePair(live, test, env)
	m.PartnerID = strconv.FormatInt(chosen.ID, 10)
	m.PartnerKey = chosen.Key
}
