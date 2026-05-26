package services

import (
	"context"
	"fmt"
	"time"

	"github.com/omni/backend/internal/models"
	"github.com/omni/backend/internal/repositories"
	"gorm.io/gorm"
)

type CredentialRefreshOutcome struct {
	AccessToken          string
	RefreshToken         string
	ShopCipher           string
	TokenExpiry          int64
	RefreshExpiry        int64
	Status               string
	Code                 string
	PreserveCurrentToken bool
}

type CredentialLifecycleService struct {
	db *gorm.DB
}

func NewCredentialLifecycleService(db *gorm.DB) *CredentialLifecycleService {
	return &CredentialLifecycleService{db: db}
}

func (s *CredentialLifecycleService) RefreshConnection(ctx context.Context, tenantID, platform, actor, actorRole string, refreshFn func(*models.CredentialConnection) (*CredentialRefreshOutcome, error)) (*models.CredentialConnection, error) {
	return s.applyLifecycleChange(ctx, tenantID, platform, actor, actorRole, "refresh", refreshFn)
}

func (s *CredentialLifecycleService) DisconnectConnection(ctx context.Context, tenantID, platform, actor, actorRole string, disconnectFn func(*models.CredentialConnection) (*CredentialRefreshOutcome, error)) (*models.CredentialConnection, error) {
	return s.applyLifecycleChange(ctx, tenantID, platform, actor, actorRole, "disconnect", disconnectFn)
}

func (s *CredentialLifecycleService) applyLifecycleChange(ctx context.Context, tenantID, platform, actor, actorRole, eventType string, changeFn func(*models.CredentialConnection) (*CredentialRefreshOutcome, error)) (*models.CredentialConnection, error) {
	if s.db == nil {
		return nil, fmt.Errorf("credential database is required")
	}
	var updated *models.CredentialConnection
	err := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		repo := repositories.NewCredentialRepository(tx)
		conn, err := repo.GetActiveConnectionForRefresh(ctx, tenantID, platform)
		if err != nil {
			return err
		}
		if conn == nil {
			return fmt.Errorf("connection not found for %s/%s", tenantID, platform)
		}
		initialVersion := conn.Version
		outcome, err := changeFn(conn)
		if err != nil {
			return s.writeAudit(ctx, repo, conn, actor, actorRole, eventType+"_failed", "failed", "refresh_failed", map[string]any{"store_identifier_mask": conn.StoreIdentifierMask()})
		}
		if outcome == nil {
			return fmt.Errorf("credential lifecycle outcome is required")
		}
		if eventType == "disconnect" {
			if err := repo.DisableConnectionWithVersion(ctx, conn, initialVersion, outcome.Code); err != nil {
				return err
			}
			conn.Status = "disconnected"
			conn.DisabledReason = outcome.Code
			conn.DisabledBy = actor
			updated = conn
			return s.writeAudit(ctx, repo, conn, actor, actorRole, "disconnect", "success", outcome.Code, map[string]any{"store_identifier_mask": conn.StoreIdentifierMask()})
		}
		conn.AccessToken = outcome.AccessToken
		conn.RefreshToken = outcome.RefreshToken
		conn.ShopCipher = outcome.ShopCipher
		conn.TokenExpiry = outcome.TokenExpiry
		conn.RefreshExpiry = outcome.RefreshExpiry
		conn.Status = outcome.Status
		conn.UpdatedBy = actor
		now := time.Now()
		if err := repo.UpdateConnectionTokensWithVersion(ctx, conn, initialVersion, outcome.Status, outcome.Code, &now); err != nil {
			if outcome.PreserveCurrentToken {
				return s.writeAudit(ctx, repo, conn, actor, actorRole, "refresh_failed", "failed", "stale_refresh_blocked", map[string]any{"store_identifier_mask": conn.StoreIdentifierMask()})
			}
			return err
		}
		conn.Version = initialVersion + 1
		conn.LastRefreshAt = &now
		updated = conn
		return s.writeAudit(ctx, repo, conn, actor, actorRole, eventType, "success", outcome.Code, map[string]any{"store_identifier_mask": conn.StoreIdentifierMask()})
	})
	return updated, err
}

func (s *CredentialLifecycleService) writeAudit(ctx context.Context, repo *repositories.CredentialRepository, conn *models.CredentialConnection, actor, actorRole, eventType, status, code string, metadata map[string]any) error {
	return repo.CreateAuditEvent(ctx, &models.CredentialAuditEvent{
		TenantID:        conn.TenantID,
		Platform:        conn.Platform,
		StoreIdentifier: conn.StoreIdentifier,
		EventType:       eventType,
		Status:          status,
		Code:            code,
		Actor:           actor,
		ActorRole:       actorRole,
		Metadata:        models.JSONMap(metadata),
	})
}
