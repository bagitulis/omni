package services

import (
	"context"
	"time"

	"github.com/omni/backend/internal/models"
	"github.com/omni/backend/internal/repositories"
)

// AuditService handles audit logging operations
type AuditService struct {
	repo *repositories.AuditRepository
}

// NewAuditService creates a new audit service
func NewAuditService(repo *repositories.AuditRepository) *AuditService {
	return &AuditService{repo: repo}
}

// Log logs an audit entry
func (s *AuditService) Log(ctx context.Context, entry *models.AuditLogEntry) error {
	return s.repo.Create(ctx, entry)
}

// LogAction logs a simple action
func (s *AuditService) LogAction(ctx context.Context, tenantID, action, userID string) error {
	return s.repo.Create(ctx, &models.AuditLogEntry{
		TenantID: tenantID,
		Action:   action,
		UserID:   userID,
		Status:   "success",
	})
}

// LogActionWithDetails logs an action with details
func (s *AuditService) LogActionWithDetails(ctx context.Context, tenantID, action, userID string, details map[string]interface{}) error {
	return s.repo.Create(ctx, &models.AuditLogEntry{
		TenantID: tenantID,
		Action:   action,
		UserID:   userID,
		Status:   "success",
		Details:  details,
	})
}

// LogError logs a failed action
func (s *AuditService) LogError(ctx context.Context, tenantID, action, userID, errorMsg string) error {
	return s.repo.Create(ctx, &models.AuditLogEntry{
		TenantID:     tenantID,
		Action:       action,
		UserID:       userID,
		Status:       "error",
		ErrorMessage: errorMsg,
	})
}

// GetByTenant gets audit logs by tenant with pagination
func (s *AuditService) GetByTenant(ctx context.Context, tenantID string, page, pageSize int) (*AuditListResponse, error) {
	offset := (page - 1) * pageSize
	logs, total, err := s.repo.FindByTenant(ctx, tenantID, pageSize, offset)
	if err != nil {
		return nil, err
	}

	return &AuditListResponse{
		Data:       logs,
		Total:      total,
		Page:       page,
		PageSize:   pageSize,
		TotalPages: (total + int64(pageSize) - 1) / int64(pageSize),
	}, nil
}

// AuditListResponse represents paginated audit log response
type AuditListResponse struct {
	Data       []models.AuditLog `json:"data"`
	Total      int64             `json:"total"`
	Page       int               `json:"page"`
	PageSize   int               `json:"pageSize"`
	TotalPages int64             `json:"totalPages"`
}

// GetByUser gets audit logs by user
func (s *AuditService) GetByUser(ctx context.Context, userID string, limit int) ([]models.AuditLog, error) {
	return s.repo.FindByUser(ctx, userID, limit)
}

// GetByAction gets audit logs by action type
func (s *AuditService) GetByAction(ctx context.Context, tenantID, action string, limit int) ([]models.AuditLog, error) {
	return s.repo.FindByAction(ctx, tenantID, action, limit)
}

// GetByDateRange gets audit logs within date range
func (s *AuditService) GetByDateRange(ctx context.Context, tenantID string, startDate, endDate time.Time, page, pageSize int) (*AuditListResponse, error) {
	offset := (page - 1) * pageSize
	logs, total, err := s.repo.FindByDateRange(ctx, tenantID, startDate, endDate, pageSize, offset)
	if err != nil {
		return nil, err
	}

	return &AuditListResponse{
		Data:       logs,
		Total:      total,
		Page:       page,
		PageSize:   pageSize,
		TotalPages: (total + int64(pageSize) - 1) / int64(pageSize),
	}, nil
}

// CleanupOldLogs deletes audit logs older than specified duration
func (s *AuditService) CleanupOldLogs(ctx context.Context, tenantID string, olderThan time.Duration) error {
	return s.repo.DeleteOldLogs(ctx, tenantID, olderThan)
}
