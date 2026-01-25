package repositories

import (
	"context"
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"github.com/omni/backend/internal/models"
	"gorm.io/gorm"
)

// WebhookRepository handles webhook data access
type WebhookRepository struct {
	db *gorm.DB
}

// NewWebhookRepository creates a new webhook repository
func NewWebhookRepository(db *gorm.DB) *WebhookRepository {
	return &WebhookRepository{db: db}
}

// CreateLog creates a new webhook log
func (r *WebhookRepository) CreateLog(ctx context.Context, tenantID, platform, eventType string, payload, headers interface{}) (*models.WebhookLog, error) {
	payloadJSON, _ := json.Marshal(payload)
	headersJSON, _ := json.Marshal(headers)

	log := &models.WebhookLog{
		ID:        uuid.New().String(),
		TenantID:  tenantID,
		Platform:  platform,
		EventType: eventType,
		Payload:   string(payloadJSON),
		Headers:   string(headersJSON),
		Status:    models.WebhookStatusReceived,
		CreatedAt: time.Now(),
	}
	err := r.db.WithContext(ctx).Create(log).Error
	if err != nil {
		return nil, err
	}
	return log, nil
}

// UpdateLogStatus updates webhook log status
func (r *WebhookRepository) UpdateLogStatus(ctx context.Context, id, status, errorMsg string) error {
	now := time.Now()
	updates := map[string]interface{}{
		"status":       status,
		"processed_at": now,
	}
	if errorMsg != "" {
		updates["error_msg"] = errorMsg
	}
	return r.db.WithContext(ctx).Model(&models.WebhookLog{}).
		Where("id = ?", id).Updates(updates).Error
}

// CreateOrderEvent creates a new order event
func (r *WebhookRepository) CreateOrderEvent(ctx context.Context, event *models.WebhookOrderEvent) error {
	event.CreatedAt = time.Now()
	return r.db.WithContext(ctx).Create(event).Error
}

// CreateProductEvent creates a new product event
func (r *WebhookRepository) CreateProductEvent(ctx context.Context, event *models.WebhookProductEvent) error {
	event.CreatedAt = time.Now()
	return r.db.WithContext(ctx).Create(event).Error
}

// CreateReturnEvent creates a new return event
func (r *WebhookRepository) CreateReturnEvent(ctx context.Context, event *models.WebhookReturnEvent) error {
	event.CreatedAt = time.Now()
	return r.db.WithContext(ctx).Create(event).Error
}

// FindLogsByTenant finds webhook logs by tenant
func (r *WebhookRepository) FindLogsByTenant(ctx context.Context, tenantID string, limit, offset int) ([]models.WebhookLog, int64, error) {
	var logs []models.WebhookLog
	var total int64

	r.db.WithContext(ctx).Model(&models.WebhookLog{}).
		Where("tenant_id = ?", tenantID).Count(&total)

	err := r.db.WithContext(ctx).
		Where("tenant_id = ?", tenantID).
		Order("created_at DESC").
		Limit(limit).Offset(offset).
		Find(&logs).Error

	return logs, total, err
}

// FindOrderEventsByOrder finds order events by order SN
func (r *WebhookRepository) FindOrderEventsByOrder(ctx context.Context, tenantID, orderSN string) ([]models.WebhookOrderEvent, error) {
	var events []models.WebhookOrderEvent
	err := r.db.WithContext(ctx).
		Where("tenant_id = ? AND order_sn = ?", tenantID, orderSN).
		Order("created_at DESC").
		Find(&events).Error
	return events, err
}

// FindOrderEventsByPlatform finds order events by platform
func (r *WebhookRepository) FindOrderEventsByPlatform(ctx context.Context, tenantID, platform string, limit, offset int) ([]models.WebhookOrderEvent, int64, error) {
	var events []models.WebhookOrderEvent
	var total int64

	query := r.db.WithContext(ctx).Model(&models.WebhookOrderEvent{}).
		Where("tenant_id = ? AND platform = ?", tenantID, platform)

	query.Count(&total)

	err := query.Order("created_at DESC").
		Limit(limit).Offset(offset).
		Find(&events).Error

	return events, total, err
}

// FindRecentOrderEvents finds recent order events
func (r *WebhookRepository) FindRecentOrderEvents(ctx context.Context, tenantID string, since time.Time) ([]models.WebhookOrderEvent, error) {
	var events []models.WebhookOrderEvent
	err := r.db.WithContext(ctx).
		Where("tenant_id = ? AND created_at > ?", tenantID, since).
		Order("created_at DESC").
		Find(&events).Error
	return events, err
}

// DeleteOldLogs deletes logs older than specified duration
func (r *WebhookRepository) DeleteOldLogs(ctx context.Context, olderThan time.Duration) error {
	cutoff := time.Now().Add(-olderThan)
	return r.db.WithContext(ctx).
		Where("created_at < ?", cutoff).
		Delete(&models.WebhookLog{}).Error
}
