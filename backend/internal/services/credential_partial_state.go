package services

import (
	"context"
	"fmt"

	"github.com/omni/backend/internal/repositories"
	"github.com/rs/zerolog/log"
	"gorm.io/gorm"
)

// PartialState represents the completeness of a platform's canonical credentials.
type PartialState int

const (
	// StateBothEmpty means no app config and no connections — platform not connected.
	StateBothEmpty PartialState = iota
	// StateAppOnly means app config exists but no store connections — incomplete, warns.
	StateAppOnly
	// StateConnectionOnly means a store connection exists but app config is missing — incomplete, repairs.
	StateConnectionOnly
	// StateBothPresent means both app config and at least one connection exist — OK.
	StateBothPresent
)

// String returns a human-readable label for the partial state.
func (s PartialState) String() string {
	switch s {
	case StateBothEmpty:
		return "both_empty"
	case StateAppOnly:
		return "app_only"
	case StateConnectionOnly:
		return "connection_only"
	case StateBothPresent:
		return "both_present"
	default:
		return "unknown"
	}
}

// classifyPartialState determines which of the 4 canonical states applies.
func classifyPartialState(state canonicalCredentialLoadState) PartialState {
	if state.AppConfigured && state.StoreConfigured {
		return StateBothPresent
	}
	if state.AppConfigured && !state.StoreConfigured {
		return StateAppOnly
	}
	if !state.AppConfigured && state.StoreConfigured {
		return StateConnectionOnly
	}
	return StateBothEmpty
}

// validateAndRepairPartialState checks the credential load state and repairs
// connection-only by creating a missing app config from platform_configs.
// Returns the classified state and a non-nil error only for unrecoverable situations.
func (s *CredentialService) validateAndRepairPartialState(
	ctx context.Context, db *gorm.DB, tenantID, platform string,
	loadState canonicalCredentialLoadState,
) (PartialState, error) {

	state := classifyPartialState(loadState)

	switch state {
	case StateBothEmpty:
		log.Info().Str("tenant_id", tenantID).Str("platform", platform).
			Msg("[CredentialService] Platform not connected — no app config or connections")
		return state, nil

	case StateAppOnly:
		log.Warn().Str("tenant_id", tenantID).Str("platform", platform).
			Msg("[CredentialService] App config exists but no store connection — incomplete setup")
		return state, nil

	case StateBothPresent:
		return state, nil

	case StateConnectionOnly:
		log.Warn().Str("tenant_id", tenantID).Str("platform", platform).
			Msg("[CredentialService] Store connection exists but app config missing — attempting repair")
		if err := s.repairMissingAppConfig(ctx, db, tenantID, platform); err != nil {
			return state, fmt.Errorf("repair app config for %s/%s: %w", tenantID, platform, err)
		}
		log.Info().Str("tenant_id", tenantID).Str("platform", platform).
			Msg("[CredentialService] App config repaired from platform_configs")
		return state, nil
	}

	return state, nil
}

// repairMissingAppConfig creates an app config by reading platform_configs.
// Uses the same backfill helpers that BackfillAppConfigs uses.
func (s *CredentialService) repairMissingAppConfig(ctx context.Context, db *gorm.DB, tenantID, platform string) error {
	requiredKeys, ok := appCredentialKeys[platform]
	if !ok {
		return fmt.Errorf("unsupported platform %s for app config repair", platform)
	}

	credRepo := repositories.NewCredentialRepository(db)

	// Check again — idempotency guard
	existing, err := credRepo.GetAppConfig(ctx, tenantID, platform)
	if err != nil {
		return fmt.Errorf("check existing app config: %w", err)
	}
	if existing != nil {
		return nil // already exists, nothing to repair
	}

	// Read platform_configs to build app config
	credRows, _, err := readAndClassify(ctx, db)
	if err != nil {
		return fmt.Errorf("read platform_configs for repair: %w", err)
	}

	platformRows := groupByPlatform(credRows)
	pRows, ok := platformRows[platform]
	if !ok || len(pRows) == 0 {
		return fmt.Errorf("no platform_configs rows found for %s — cannot repair app config", platform)
	}

	configMap := buildConfigMap(pRows)

	for _, key := range requiredKeys {
		if configMap[key] == "" {
			return fmt.Errorf("missing required key %q in platform_configs for %s — cannot repair app config", key, platform)
		}
	}

	cfg := buildAppConfig(tenantID, platform, configMap)
	if err := credRepo.UpsertAppConfig(ctx, cfg); err != nil {
		return fmt.Errorf("upsert repaired app config: %w", err)
	}

	return nil
}
