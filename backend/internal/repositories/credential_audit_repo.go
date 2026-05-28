package repositories

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/omni/backend/internal/models"
)

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
		if isMissingRelationError(err) {
			return []models.CredentialAuditEvent{}, nil
		}
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
