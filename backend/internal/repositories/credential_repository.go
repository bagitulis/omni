package repositories

import (
	"context"
	"fmt"
	"os"
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
		enc, _ = utils.NewEncryptionService(encKey)
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
	query := r.db.WithContext(ctx).Where("tenant_id = ?", tenantID)
	if platform != "" {
		query = query.Where("platform = ?", platform)
	}
	if err := query.Find(&conns).Error; err != nil {
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
	if tenantID == "" {
		return fmt.Errorf("tenant_id is required")
	}
	if platform == "" {
		return fmt.Errorf("platform is required")
	}
	if storeIdentifier == "" {
		return fmt.Errorf("store_identifier is required")
	}
	return nil
}

//------------------------------------------------------------------------------
// App Config Methods
//------------------------------------------------------------------------------

// GetAppConfig retrieves the app config for a tenant/platform.
func (r *CredentialRepository) GetAppConfig(ctx context.Context, tenantID, platform string) (*models.CredentialAppConfig, error) {
	if tenantID == "" {
		return nil, fmt.Errorf("tenant_id is required")
	}
	var cfg models.CredentialAppConfig
	err := r.db.WithContext(ctx).
		Where("tenant_id = ? AND platform = ?", tenantID, platform).
		First(&cfg).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, fmt.Errorf("get app config: %w", err)
	}
	if err := r.decryptAppConfigSecrets(&cfg); err != nil {
		return nil, err
	}
	return &cfg, nil
}

// UpsertAppConfig encrypts secrets and creates or updates an app config.
func (r *CredentialRepository) UpsertAppConfig(ctx context.Context, cfg *models.CredentialAppConfig) error {
	if cfg.TenantID == "" {
		return fmt.Errorf("tenant_id is required")
	}
	if cfg.ID == "" {
		cfg.ID = uuid.New().String()
	}
	if cfg.Region == "" {
		cfg.Region = "id"
	}
	if err := r.encryptAppConfigSecrets(cfg); err != nil {
		return err
	}
	var existing models.CredentialAppConfig
	result := r.db.WithContext(ctx).
		Where("tenant_id = ? AND platform = ? AND COALESCE(store_identifier, '') = COALESCE(?, '')", cfg.TenantID, cfg.Platform, cfg.StoreIdentifier).
		First(&existing)
	if result.Error != nil && result.Error != gorm.ErrRecordNotFound {
		return fmt.Errorf("upsert app config lookup: %w", result.Error)
	}
	if result.Error == gorm.ErrRecordNotFound {
		if err := r.db.WithContext(ctx).Create(cfg).Error; err != nil {
			return fmt.Errorf("create app config: %w", err)
		}
		return nil
	}
	cfg.ID = existing.ID
	if err := r.db.WithContext(ctx).
		Model(&existing).
		Where("id = ?", existing.ID).
		Updates(cfg).Error; err != nil {
		return fmt.Errorf("upsert app config: %w", err)
	}
	return nil
}

//------------------------------------------------------------------------------
// OAuth Attempt Methods
//------------------------------------------------------------------------------

// CreateAttempt creates a pending OAuth connection attempt.
func (r *CredentialRepository) CreateAttempt(ctx context.Context, attempt *models.OAuthConnectionAttempt) error {
	if attempt.TenantID == "" {
		return fmt.Errorf("tenant_id is required")
	}
	if attempt.ID == "" {
		attempt.ID = uuid.New().String()
	}
	if attempt.Status == "" {
		attempt.Status = "pending"
	}
	if err := r.db.WithContext(ctx).Create(attempt).Error; err != nil {
		return fmt.Errorf("create attempt: %w", err)
	}
	return nil
}

// GetAttempt retrieves a pending OAuth attempt by tenant, platform, and attempt_id.
func (r *CredentialRepository) GetAttempt(ctx context.Context, tenantID, platform, attemptID string) (*models.OAuthConnectionAttempt, error) {
	if tenantID == "" {
		return nil, fmt.Errorf("tenant_id is required")
	}
	var attempt models.OAuthConnectionAttempt
	err := r.db.WithContext(ctx).
		Where("tenant_id = ? AND platform = ? AND attempt_id = ?", tenantID, platform, attemptID).
		First(&attempt).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, fmt.Errorf("get attempt: %w", err)
	}
	return &attempt, nil
}

// CompleteAttempt marks an OAuth attempt as completed or failed.
func (r *CredentialRepository) CompleteAttempt(ctx context.Context, tenantID, platform, attemptID, status string) error {
	if tenantID == "" {
		return fmt.Errorf("tenant_id is required")
	}
	now := time.Now()
	result := r.db.WithContext(ctx).
		Model(&models.OAuthConnectionAttempt{}).
		Where("tenant_id = ? AND platform = ? AND attempt_id = ?", tenantID, platform, attemptID).
		Updates(map[string]any{
			"status":       status,
			"completed_at": &now,
		})
	if result.Error != nil {
		return fmt.Errorf("complete attempt: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("attempt not found: %s", attemptID)
	}
	return nil
}

//------------------------------------------------------------------------------
// Audit Methods
//------------------------------------------------------------------------------

// ListAuditEvents returns sanitized audit events for a tenant/platform.
func (r *CredentialRepository) ListAuditEvents(ctx context.Context, tenantID, platform string, limit, offset int) ([]models.CredentialAuditEvent, error) {
	if tenantID == "" {
		return nil, fmt.Errorf("tenant_id is required")
	}
	if limit <= 0 {
		limit = 50
	}
	var events []models.CredentialAuditEvent
	err := r.db.WithContext(ctx).
		Where("tenant_id = ? AND platform = ?", tenantID, platform).
		Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&events).Error
	if err != nil {
		return nil, fmt.Errorf("list audit events: %w", err)
	}
	return events, nil
}

// CreateAuditEvent creates a sanitized audit event.
func (r *CredentialRepository) CreateAuditEvent(ctx context.Context, event *models.CredentialAuditEvent) error {
	if event.TenantID == "" {
		return fmt.Errorf("tenant_id is required")
	}
	if event.ID == "" {
		event.ID = uuid.New().String()
	}
	if err := r.db.WithContext(ctx).Create(event).Error; err != nil {
		return fmt.Errorf("create audit event: %w", err)
	}
	return nil
}
