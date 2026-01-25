package repositories

import (
	"context"
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"github.com/omni/backend/internal/models"
	"gorm.io/gorm"
)

// AuditRepository handles audit log data access
type AuditRepository struct {
	db *gorm.DB
}

// NewAuditRepository creates a new audit repository
func NewAuditRepository(db *gorm.DB) *AuditRepository {
	return &AuditRepository{db: db}
}

// Create creates a new audit log entry
func (r *AuditRepository) Create(ctx context.Context, entry *models.AuditLogEntry) error {
	detailsJSON := ""
	if entry.Details != nil {
		bytes, _ := json.Marshal(entry.Details)
		detailsJSON = string(bytes)
	}

	log := &models.AuditLog{
		ID:             uuid.New().String(),
		TenantID:       entry.TenantID,
		Action:         entry.Action,
		UserID:         entry.UserID,
		TargetUserID:   entry.TargetUserID,
		TargetTenantID: entry.TargetTenantID,
		Details:        detailsJSON,
		Status:         entry.Status,
		ErrorMessage:   entry.ErrorMessage,
		IPAddress:      entry.IPAddress,
		UserAgent:      entry.UserAgent,
		CreatedAt:      time.Now(),
	}
	return r.db.WithContext(ctx).Create(log).Error
}

// FindByTenant finds audit logs by tenant
func (r *AuditRepository) FindByTenant(ctx context.Context, tenantID string, limit, offset int) ([]models.AuditLog, int64, error) {
	var logs []models.AuditLog
	var total int64

	r.db.WithContext(ctx).Model(&models.AuditLog{}).
		Where("tenant_id = ?", tenantID).Count(&total)

	err := r.db.WithContext(ctx).
		Where("tenant_id = ?", tenantID).
		Order("created_at DESC").
		Limit(limit).Offset(offset).
		Find(&logs).Error

	return logs, total, err
}

// FindByUser finds audit logs by user
func (r *AuditRepository) FindByUser(ctx context.Context, userID string, limit int) ([]models.AuditLog, error) {
	var logs []models.AuditLog
	err := r.db.WithContext(ctx).
		Where("user_id = ?", userID).
		Order("created_at DESC").
		Limit(limit).
		Find(&logs).Error
	return logs, err
}

// FindByAction finds audit logs by action
func (r *AuditRepository) FindByAction(ctx context.Context, tenantID, action string, limit int) ([]models.AuditLog, error) {
	var logs []models.AuditLog
	err := r.db.WithContext(ctx).
		Where("tenant_id = ? AND action = ?", tenantID, action).
		Order("created_at DESC").
		Limit(limit).
		Find(&logs).Error
	return logs, err
}

// FindByDateRange finds audit logs within a date range
func (r *AuditRepository) FindByDateRange(ctx context.Context, tenantID string, startDate, endDate time.Time, limit, offset int) ([]models.AuditLog, int64, error) {
	var logs []models.AuditLog
	var total int64

	query := r.db.WithContext(ctx).Model(&models.AuditLog{}).
		Where("tenant_id = ? AND created_at BETWEEN ? AND ?", tenantID, startDate, endDate)

	query.Count(&total)

	err := query.Order("created_at DESC").
		Limit(limit).Offset(offset).
		Find(&logs).Error

	return logs, total, err
}

// DeleteOldLogs deletes logs older than specified duration
func (r *AuditRepository) DeleteOldLogs(ctx context.Context, tenantID string, olderThan time.Duration) error {
	cutoff := time.Now().Add(-olderThan)
	return r.db.WithContext(ctx).
		Where("tenant_id = ? AND created_at < ?", tenantID, cutoff).
		Delete(&models.AuditLog{}).Error
}
