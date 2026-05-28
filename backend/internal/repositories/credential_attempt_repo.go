package repositories

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/omni/backend/internal/models"
	"gorm.io/gorm"
)

// CreateAttempt creates a pending OAuth connection attempt.
func (r *CredentialRepository) CreateAttempt(ctx context.Context, attempt *models.OAuthConnectionAttempt) error {
	if err := validateAttemptScope(attempt.TenantID, attempt.Platform, attempt.AttemptID); err != nil {
		return err
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

// GetAttempt retrieves an OAuth attempt by tenant, platform, and attempt_id.
func (r *CredentialRepository) GetAttempt(ctx context.Context, tenantID, platform, attemptID string) (*models.OAuthConnectionAttempt, error) {
	if err := validateAttemptScope(tenantID, platform, attemptID); err != nil {
		return nil, err
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
	if err := validateAttemptScope(tenantID, platform, attemptID); err != nil {
		return err
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
