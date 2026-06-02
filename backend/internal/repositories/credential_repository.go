package repositories

import (
"context"
"fmt"
"os"
"strings"
"time"
"github.com/google/uuid"
"github.com/omni/backend/internal/models"
"github.com/omni/backend/internal/utils"
"github.com/rs/zerolog/log"
"gorm.io/gorm"
)

// CredentialRepository handles tenant-scoped credential CRUD with encryption.
// All store credential reads/writes require tenant_id + platform + store_identifier scope.
type CredentialRepository struct {
	db         *gorm.DB
	encryption *utils.EncryptionService
}

// NewCredentialRepository creates a new CredentialRepository with Fernet encryption.
func NewCredentialRepository(db *gorm.DB) *CredentialRepository {
	encKey := os.Getenv("ENCRYPTION_KEY")
	var enc *utils.EncryptionService
	if encKey != "" {
		var err error
		enc, err = utils.NewEncryptionService(encKey)
		if err != nil {
			log.Warn().Err(err).Msg("Invalid ENCRYPTION_KEY, credential encryption disabled")
		}
	} else {
		log.Warn().Msg("ENCRYPTION_KEY not set, credential encryption disabled")
	}
	return &CredentialRepository{db: db, encryption: enc}
}

//------------------------------------------------------------------------------
// Store Connection Methods
//------------------------------------------------------------------------------

// GetConnection retrieves an active (non-disabled) credential connection by scope.
// Returns nil, nil if not found.
func (r *CredentialRepository) GetConnection(ctx context.Context, tenantID, platform, storeIdentifier string) (*models.CredentialConnection, error) {
	if err := validateConnectionScope(tenantID, platform, storeIdentifier); err != nil {
		return nil, err
	}
	var conn models.CredentialConnection
	err := r.db.WithContext(ctx).
		Where("tenant_id = ? AND platform = ? AND store_identifier = ? AND disabled_at IS NULL", tenantID, platform, storeIdentifier).
		First(&conn).Error
	if err != nil {
		if isMissingRelationError(err) {
			return nil, nil
		}
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, fmt.Errorf("get connection: %w", err)
	}
	if err := r.decryptSecrets(&conn); err != nil {
		return nil, err
	}
	return &conn, nil
}

// ListConnections lists credential connections for a tenant, optionally filtered by platform.
func (r *CredentialRepository) ListConnections(ctx context.Context, tenantID, platform string) ([]models.CredentialConnection, error) {
	if tenantID == "" {
		return nil, fmt.Errorf("tenant_id is required")
	}
	var conns []models.CredentialConnection
	query := r.db.WithContext(ctx).Where("tenant_id = ? AND disabled_at IS NULL", tenantID)
	if platform != "" {
		query = query.Where("platform = ?", platform)
	}
	if err := query.Find(&conns).Error; err != nil {
		if isMissingRelationError(err) {
			return []models.CredentialConnection{}, nil
		}
		return nil, fmt.Errorf("list connections: %w", err)
	}
	for i := range conns {
		if err := r.decryptSecrets(&conns[i]); err != nil {
			return nil, err
		}
	}
	return conns, nil
}

// CreateConnection encrypts secrets and creates a new credential connection.
func (r *CredentialRepository) CreateConnection(ctx context.Context, conn *models.CredentialConnection) error {
	if err := validateConnectionScope(conn.TenantID, conn.Platform, conn.StoreIdentifier); err != nil {
		return err
	}
	if conn.ID == "" {
		conn.ID = uuid.New().String()
	}
	if conn.Status == "" {
		conn.Status = "disconnected"
	}
	if conn.Version == 0 {
		conn.Version = 1
	}
	if err := r.encryptSecrets(conn); err != nil {
		return err
	}
	if err := r.db.WithContext(ctx).Create(conn).Error; err != nil {
		return fmt.Errorf("create connection: %w", err)
	}
	log.Debug().
		Str("tenant_id", conn.TenantID).
		Str("platform", conn.Platform).
		Str("store_identifier", conn.StoreIdentifier).
		Msg("Created credential connection")
	return nil
}

// ReconnectWithToken re-enables a disabled connection and updates its tokens.
// Unlike UpdateConnection, this works on disabled connections too.
func (r *CredentialRepository) ReconnectWithToken(ctx context.Context, conn *models.CredentialConnection) error {
	if err := validateConnectionScope(conn.TenantID, conn.Platform, conn.StoreIdentifier); err != nil {
		return err
	}
	if err := r.encryptSecrets(conn); err != nil {
		return err
	}
	result := r.db.WithContext(ctx).Model(&models.CredentialConnection{}).
		Where("tenant_id = ? AND platform = ? AND store_identifier = ?", conn.TenantID, conn.Platform, conn.StoreIdentifier).
		Updates(map[string]interface{}{
			"access_token":    conn.AccessToken,
			"refresh_token":   conn.RefreshToken,
			"shop_cipher":     conn.ShopCipher,
			"token_expiry":    conn.TokenExpiry,
			"status":          "connected",
			"region":          conn.Region,
			"updated_by":      conn.UpdatedBy,
			"disabled_at":     nil,
			"disabled_reason": "",
		})
	if result.Error != nil {
		return fmt.Errorf("reconnect with token: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("connection not found for %s/%s/%s", conn.TenantID, conn.Platform, conn.StoreIdentifier)
	}
	return nil
}

// UpdateConnection encrypts changed secrets and updates a credential connection.
func (r *CredentialRepository) UpdateConnection(ctx context.Context, conn *models.CredentialConnection) error {
	if err := validateConnectionScope(conn.TenantID, conn.Platform, conn.StoreIdentifier); err != nil {
		return err
	}
	if err := r.encryptSecrets(conn); err != nil {
		return err
	}
	// Prevent updating ID or timestamps via Update
	result := r.db.WithContext(ctx).Model(&models.CredentialConnection{}).
		Where("tenant_id = ? AND platform = ? AND store_identifier = ? AND disabled_at IS NULL", conn.TenantID, conn.Platform, conn.StoreIdentifier).
		Updates(conn)
	if result.Error != nil {
		return fmt.Errorf("update connection: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("connection not found for id: %s", conn.ID)
	}
	return nil
}

// UpdateManualConnectionWithVersion updates manual token fields with optimistic locking.
// Prevents manual token application from silently overwriting auto-refreshed tokens.
// Unlike UpdateConnection(), this checks version to detect stale writes.
// Unlike UpdateConnectionTokensWithVersion(), this also updates region and disabled_at.
func (r *CredentialRepository) UpdateManualConnectionWithVersion(ctx context.Context, conn *models.CredentialConnection, expectedVersion int) error {
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
		"refresh_token":   conn.RefreshToken,
		"shop_cipher":     conn.ShopCipher,
		"token_expiry":    conn.TokenExpiry,
		"status":          "connected",
		"region":          conn.Region,
		"disabled_at":     nil,
		"disabled_reason": "",
		"version":         gorm.Expr("version + 1"),
		"updated_at":      time.Now(),
		"updated_by":      conn.UpdatedBy,
	}
	result := r.db.WithContext(ctx).
		Model(&models.CredentialConnection{}).
		Where("tenant_id = ? AND platform = ? AND store_identifier = ? AND disabled_at IS NULL AND version = ?", conn.TenantID, conn.Platform, conn.StoreIdentifier, expectedVersion).
		Updates(updates)
	if result.Error != nil {
		return fmt.Errorf("update manual connection: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("%w: %s/%s/%s", ErrStaleVersion, conn.TenantID, conn.Platform, conn.StoreIdentifier)
	}
	return nil
}

// UpdateConnectionStatus updates only the status field for a connection.
func (r *CredentialRepository) UpdateConnectionStatus(ctx context.Context, tenantID, platform, storeIdentifier, status string) error {
	if err := validateConnectionScope(tenantID, platform, storeIdentifier); err != nil {
		return err
	}
	result := r.db.WithContext(ctx).
		Model(&models.CredentialConnection{}).
		Where("tenant_id = ? AND platform = ? AND store_identifier = ? AND disabled_at IS NULL", tenantID, platform, storeIdentifier).
		Update("status", status)
	if result.Error != nil {
		return fmt.Errorf("update connection status: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("connection not found for %s/%s/%s", tenantID, platform, storeIdentifier)
	}
	return nil
}

func validateConnectionScope(tenantID, platform, storeIdentifier string) error {
	if err := validateTenantPlatformScope(tenantID, platform); err != nil {
		return err
	}
	if storeIdentifier == "" {
		return fmt.Errorf("store_identifier is required")
	}
	return nil
}

func validateTenantPlatformScope(tenantID, platform string) error {
	if tenantID == "" {
		return fmt.Errorf("tenant_id is required")
	}
	if platform == "" {
		return fmt.Errorf("platform is required")
	}
	return nil
}

func validateAttemptScope(tenantID, platform, attemptID string) error {
	if err := validateTenantPlatformScope(tenantID, platform); err != nil {
		return err
	}
	if attemptID == "" {
		return fmt.Errorf("attempt_id is required")
	}
	return nil
}

// isMissingRelationError checks if a database error is caused by a missing table.
func isMissingRelationError(err error) bool {
	if err == nil {
		return false
	}
	message := strings.ToLower(err.Error())
	return strings.Contains(message, "sqlstate 42p01") ||
		strings.Contains(message, "relation ") && strings.Contains(message, " does not exist") ||
		strings.Contains(message, "no such table")
}
