package repositories

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/omni/backend/internal/models"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// GetActiveConnectionForRefresh returns the active credential row for a tenant/platform/store.
// The returned row is locked for update inside the current transaction scope when possible.
func (r *CredentialRepository) GetActiveConnectionForRefresh(ctx context.Context, tenantID, platform, storeIdentifier string) (*models.CredentialConnection, error) {
	if err := validateConnectionScope(tenantID, platform, storeIdentifier); err != nil {
		return nil, err
	}
	var conn models.CredentialConnection
	err := r.db.WithContext(ctx).
		Clauses(clause.Locking{Strength: "UPDATE"}).
		Where("tenant_id = ? AND platform = ? AND store_identifier = ? AND disabled_at IS NULL", tenantID, platform, storeIdentifier).
		First(&conn).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, fmt.Errorf("get active connection for refresh: %w", err)
	}
	if err := r.decryptSecrets(&conn); err != nil {
		return nil, err
	}
	return &conn, nil
}

// ErrStaleVersion is returned when an optimistic lock check fails
// because the row version has changed since it was last read.
var ErrStaleVersion = errors.New("stale or disabled connection")

// UpdateConnectionTokensWithVersion updates tokens only when the row version matches.
// It preserves disabled rows and prevents stale refresh writes from overwriting newer tokens.
func (r *CredentialRepository) UpdateConnectionTokensWithVersion(ctx context.Context, conn *models.CredentialConnection, expectedVersion int, status, code string, lastRefreshAt *time.Time) error {
	if err := validateConnectionScope(conn.TenantID, conn.Platform, conn.StoreIdentifier); err != nil {
		return err
	}
	if expectedVersion <= 0 {
		return fmt.Errorf("expected version is required")
	}
	if err := r.encryptSecrets(conn); err != nil {
		return err
	}
	updates := map[string]any{
		"access_token":    conn.AccessToken,
		"refresh_token":    conn.RefreshToken,
		"shop_cipher":     conn.ShopCipher,
		"token_expiry":    conn.TokenExpiry,
		"refresh_expiry":  conn.RefreshExpiry,
		"status":          status,
		"version":         gorm.Expr("version + 1"),
		"updated_at":      time.Now(),
		"updated_by":      conn.UpdatedBy,
		"last_refresh_at": lastRefreshAt,
	}
	if code != "" {
		updates["disabled_reason"] = code
	}
	result := r.db.WithContext(ctx).
		Model(&models.CredentialConnection{}).
		Where("tenant_id = ? AND platform = ? AND store_identifier = ? AND disabled_at IS NULL AND version = ?", conn.TenantID, conn.Platform, conn.StoreIdentifier, expectedVersion).
		Updates(updates)
	if result.Error != nil {
		return fmt.Errorf("update connection tokens: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("%w: %s/%s/%s", ErrStaleVersion, conn.TenantID, conn.Platform, conn.StoreIdentifier)
	}
	return nil
}

// UpdateConnectionStatusWithVersion updates only the status field when the version matches.
func (r *CredentialRepository) UpdateConnectionStatusWithVersion(ctx context.Context, conn *models.CredentialConnection, expectedVersion int, status, code string) error {
	if err := validateConnectionScope(conn.TenantID, conn.Platform, conn.StoreIdentifier); err != nil {
		return err
	}
	if expectedVersion <= 0 {
		return fmt.Errorf("expected version is required")
	}
	updates := map[string]any{
		"status":     status,
		"version":    gorm.Expr("version + 1"),
		"updated_at": time.Now(),
		"updated_by": conn.UpdatedBy,
	}
	if code != "" {
		updates["disabled_reason"] = code
	}
	result := r.db.WithContext(ctx).
		Model(&models.CredentialConnection{}).
		Where("tenant_id = ? AND platform = ? AND store_identifier = ? AND disabled_at IS NULL AND version = ?", conn.TenantID, conn.Platform, conn.StoreIdentifier, expectedVersion).
		Updates(updates)
	if result.Error != nil {
		return fmt.Errorf("update connection status: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("%w: %s/%s/%s", ErrStaleVersion, conn.TenantID, conn.Platform, conn.StoreIdentifier)
	}
	return nil
}

// DisableConnectionWithVersion disables a credential connection only when the version matches.
func (r *CredentialRepository) DisableConnectionWithVersion(ctx context.Context, conn *models.CredentialConnection, expectedVersion int, reason string) error {
	if err := validateConnectionScope(conn.TenantID, conn.Platform, conn.StoreIdentifier); err != nil {
		return err
	}
	if expectedVersion <= 0 {
		return fmt.Errorf("expected version is required")
	}
	now := time.Now()
	result := r.db.WithContext(ctx).
		Model(&models.CredentialConnection{}).
		Where("tenant_id = ? AND platform = ? AND store_identifier = ? AND disabled_at IS NULL AND version = ?", conn.TenantID, conn.Platform, conn.StoreIdentifier, expectedVersion).
		Updates(map[string]any{
			"status":          "disconnected",
			"disabled_at":     &now,
			"disabled_reason": reason,
			"version":         gorm.Expr("version + 1"),
			"updated_at":      now,
			"updated_by":      conn.UpdatedBy,
		})
	if result.Error != nil {
		return fmt.Errorf("disable connection: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("%w: %s/%s/%s", ErrStaleVersion, conn.TenantID, conn.Platform, conn.StoreIdentifier)
	}
	return nil
}
