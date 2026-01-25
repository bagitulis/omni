package operations

import (
	"context"
	"time"

	"gorm.io/gorm"
)

// OperationType represents the type of operation
type OperationType string

const (
	OpPrintLabel    OperationType = "print_label"
	OpShipOrder     OperationType = "ship_order"
	OpCancelOrder   OperationType = "cancel_order"
	OpUpdateStock   OperationType = "update_stock"
	OpUpdatePrice   OperationType = "update_price"
	OpBulkUpdate    OperationType = "bulk_update"
)

// OperationStatus represents operation status
type OperationStatus string

const (
	StatusPending    OperationStatus = "pending"
	StatusProcessing OperationStatus = "processing"
	StatusCompleted  OperationStatus = "completed"
	StatusFailed     OperationStatus = "failed"
	StatusPartial    OperationStatus = "partial"
)

// ShopeeOperation represents a Shopee operation record
type ShopeeOperation struct {
	ID            uint            `gorm:"primaryKey" json:"id"`
	TenantID      string          `gorm:"index;not null" json:"tenantId"`
	OperationType OperationType   `gorm:"index;not null" json:"operationType"`
	Status        OperationStatus `gorm:"default:pending" json:"status"`
	OrderSN       string          `gorm:"index" json:"orderSn,omitempty"`
	ItemID        string          `json:"itemId,omitempty"`
	RequestData   string          `json:"requestData,omitempty"` // JSON
	ResponseData  string          `json:"responseData,omitempty"` // JSON
	ErrorMessage  string          `json:"errorMessage,omitempty"`
	RetryCount    int             `json:"retryCount"`
	MaxRetries    int             `gorm:"default:3" json:"maxRetries"`
	ScheduledAt   time.Time       `json:"scheduledAt,omitempty"`
	StartedAt     time.Time       `json:"startedAt,omitempty"`
	CompletedAt   time.Time       `json:"completedAt,omitempty"`
	CreatedAt     time.Time       `json:"createdAt"`
	UpdatedAt     time.Time       `json:"updatedAt"`
}

// TableName returns the table name for GORM
func (ShopeeOperation) TableName() string {
	return "shopee_operations"
}

// OperationBatch represents a batch of operations
type OperationBatch struct {
	ID            uint            `gorm:"primaryKey" json:"id"`
	TenantID      string          `gorm:"index;not null" json:"tenantId"`
	OperationType OperationType   `json:"operationType"`
	Status        OperationStatus `gorm:"default:pending" json:"status"`
	TotalItems    int             `json:"totalItems"`
	ProcessedItems int            `json:"processedItems"`
	SuccessItems  int             `json:"successItems"`
	FailedItems   int             `json:"failedItems"`
	StartedAt     time.Time       `json:"startedAt,omitempty"`
	CompletedAt   time.Time       `json:"completedAt,omitempty"`
	CreatedAt     time.Time       `json:"createdAt"`
	UpdatedAt     time.Time       `json:"updatedAt"`
}

// TableName returns the table name for GORM
func (OperationBatch) TableName() string {
	return "operation_batches"
}

// ShopeeOperationsService handles Shopee operations
type ShopeeOperationsService struct {
	db *gorm.DB
}

// NewShopeeOperationsService creates a new Shopee operations service
func NewShopeeOperationsService(db *gorm.DB) *ShopeeOperationsService {
	return &ShopeeOperationsService{db: db}
}

// CreateOperation creates a new operation
func (s *ShopeeOperationsService) CreateOperation(
	ctx context.Context,
	op *ShopeeOperation,
) error {
	op.Status = StatusPending
	return s.db.WithContext(ctx).Create(op).Error
}

// GetOperation retrieves an operation by ID
func (s *ShopeeOperationsService) GetOperation(
	ctx context.Context,
	tenantID string,
	id uint,
) (*ShopeeOperation, error) {
	var op ShopeeOperation

	err := s.db.WithContext(ctx).
		Where("tenant_id = ? AND id = ?", tenantID, id).
		First(&op).Error

	if err != nil {
		return nil, err
	}

	return &op, nil
}

// GetPendingOperations retrieves pending operations
func (s *ShopeeOperationsService) GetPendingOperations(
	ctx context.Context,
	tenantID string,
	limit int,
) ([]ShopeeOperation, error) {
	var ops []ShopeeOperation

	query := s.db.WithContext(ctx).
		Where("tenant_id = ? AND status = ?", tenantID, StatusPending).
		Where("scheduled_at <= ? OR scheduled_at IS NULL", time.Now()).
		Order("created_at ASC")

	if limit > 0 {
		query = query.Limit(limit)
	}

	err := query.Find(&ops).Error
	return ops, err
}

// UpdateOperationStatus updates operation status
func (s *ShopeeOperationsService) UpdateOperationStatus(
	ctx context.Context,
	id uint,
	status OperationStatus,
	response string,
	errorMsg string,
) error {
	updates := map[string]interface{}{
		"status":        status,
		"response_data": response,
		"error_message": errorMsg,
		"updated_at":    time.Now(),
	}

	if status == StatusProcessing {
		updates["started_at"] = time.Now()
	} else if status == StatusCompleted || status == StatusFailed {
		updates["completed_at"] = time.Now()
	}

	return s.db.WithContext(ctx).
		Model(&ShopeeOperation{}).
		Where("id = ?", id).
		Updates(updates).Error
}

// CreateBatch creates a new operation batch
func (s *ShopeeOperationsService) CreateBatch(
	ctx context.Context,
	batch *OperationBatch,
) error {
	batch.Status = StatusPending
	return s.db.WithContext(ctx).Create(batch).Error
}

// UpdateBatchProgress updates batch progress
func (s *ShopeeOperationsService) UpdateBatchProgress(
	ctx context.Context,
	id uint,
	processed int,
	success int,
	failed int,
) error {
	status := StatusProcessing
	if processed >= success+failed {
		if failed == 0 {
			status = StatusCompleted
		} else if success == 0 {
			status = StatusFailed
		} else {
			status = StatusPartial
		}
	}

	updates := map[string]interface{}{
		"status":          status,
		"processed_items": processed,
		"success_items":   success,
		"failed_items":    failed,
		"updated_at":      time.Now(),
	}

	if status == StatusCompleted || status == StatusFailed || status == StatusPartial {
		updates["completed_at"] = time.Now()
	}

	return s.db.WithContext(ctx).
		Model(&OperationBatch{}).
		Where("id = ?", id).
		Updates(updates).Error
}

// GetRecentOperations retrieves recent operations
func (s *ShopeeOperationsService) GetRecentOperations(
	ctx context.Context,
	tenantID string,
	opType OperationType,
	limit int,
) ([]ShopeeOperation, error) {
	var ops []ShopeeOperation

	query := s.db.WithContext(ctx).
		Where("tenant_id = ?", tenantID).
		Order("created_at DESC")

	if opType != "" {
		query = query.Where("operation_type = ?", opType)
	}
	if limit > 0 {
		query = query.Limit(limit)
	}

	err := query.Find(&ops).Error
	return ops, err
}
