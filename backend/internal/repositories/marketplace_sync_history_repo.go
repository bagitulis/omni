package repositories

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/omni/backend/internal/models"
	"gorm.io/gorm"
)

// MarketplaceSyncHistoryRepo handles marketplace sync history data access
type MarketplaceSyncHistoryRepo struct {
	db *gorm.DB
}

// NewMarketplaceSyncHistoryRepo creates a new marketplace sync history repository
func NewMarketplaceSyncHistoryRepo(db *gorm.DB) *MarketplaceSyncHistoryRepo {
	return &MarketplaceSyncHistoryRepo{db: db}
}

// Create inserts a new marketplace sync history entry
func (r *MarketplaceSyncHistoryRepo) Create(ctx context.Context, entry *models.MarketplaceSyncHistory) error {
	if entry.ID == "" {
		entry.ID = uuid.New().String()
	}
	if entry.CreatedAt.IsZero() {
		entry.CreatedAt = time.Now()
	}
	return r.db.WithContext(ctx).Create(entry).Error
}

// CreateBatch inserts multiple marketplace sync history entries in a transaction
func (r *MarketplaceSyncHistoryRepo) CreateBatch(ctx context.Context, entries []models.MarketplaceSyncHistory) error {
	if len(entries) == 0 {
		return nil
	}
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		now := time.Now()
		for i := range entries {
			if entries[i].ID == "" {
				entries[i].ID = uuid.New().String()
			}
			if entries[i].CreatedAt.IsZero() {
				entries[i].CreatedAt = now
			}
		}
		return tx.Create(&entries).Error
	})
}

// List retrieves marketplace sync history with filters and pagination
func (r *MarketplaceSyncHistoryRepo) List(ctx context.Context, filter models.MarketplaceSyncHistoryFilter) (*models.MarketplaceSyncHistoryListResult, error) {
	query := r.db.WithContext(ctx).Model(&models.MarketplaceSyncHistory{}).
		Where("tenant_id = ?", filter.TenantID)

	if filter.Platform != "" {
		query = query.Where("platform = ?", filter.Platform)
	}
	if filter.Operation != "" {
		query = query.Where("operation = ?", filter.Operation)
	}
	if filter.Status != "" {
		query = query.Where("status = ?", filter.Status)
	}
	if filter.SKUSearch != "" {
		query = query.Where("sku LIKE ?", fmt.Sprintf("%%%s%%", filter.SKUSearch))
	}
	if filter.DateFrom != "" {
		query = query.Where("created_at >= ?", filter.DateFrom)
	}
	if filter.DateTo != "" {
		query = query.Where("created_at <= ?", filter.DateTo+" 23:59:59")
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, fmt.Errorf("failed to count marketplace sync history: %w", err)
	}

	// Apply pagination
	page := filter.Page
	pageSize := filter.PageSize
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}
	offset := (page - 1) * pageSize

	var entries []models.MarketplaceSyncHistory
	if err := query.Order("created_at DESC").
		Limit(pageSize).Offset(offset).
		Find(&entries).Error; err != nil {
		return nil, fmt.Errorf("failed to list marketplace sync history: %w", err)
	}

	return &models.MarketplaceSyncHistoryListResult{
		Entries:  entries,
		Total:    total,
		Page:     page,
		PageSize: pageSize,
	}, nil
}
