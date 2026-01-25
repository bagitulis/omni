package sync

import (
	"context"
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"time"

	"gorm.io/gorm"
)

// DeltaSyncRecord represents a delta sync checkpoint
type DeltaSyncRecord struct {
	ID             uint      `gorm:"primaryKey" json:"id"`
	TenantID       string    `gorm:"index;not null" json:"tenantId"`
	Platform       string    `gorm:"index;not null" json:"platform"`
	EntityType     string    `gorm:"index;not null" json:"entityType"` // orders, products, inventory
	LastSyncTime   time.Time `json:"lastSyncTime"`
	LastCursor     string    `json:"lastCursor,omitempty"`
	LastHash       string    `json:"lastHash,omitempty"`
	RecordsSynced  int       `json:"recordsSynced"`
	Status         string    `json:"status"` // success, failed, partial
	ErrorMessage   string    `json:"errorMessage,omitempty"`
	CreatedAt      time.Time `json:"createdAt"`
	UpdatedAt      time.Time `json:"updatedAt"`
}

// TableName returns the table name for GORM
func (DeltaSyncRecord) TableName() string {
	return "delta_sync_records"
}

// ChangeRecord represents a tracked change for delta sync
type ChangeRecord struct {
	ID         uint      `gorm:"primaryKey" json:"id"`
	TenantID   string    `gorm:"index;not null" json:"tenantId"`
	Platform   string    `gorm:"index;not null" json:"platform"`
	EntityType string    `gorm:"index;not null" json:"entityType"`
	EntityID   string    `gorm:"index;not null" json:"entityId"`
	ChangeType string    `json:"changeType"` // create, update, delete
	OldData    string    `json:"oldData,omitempty"` // JSON
	NewData    string    `json:"newData,omitempty"` // JSON
	Processed  bool      `gorm:"default:false" json:"processed"`
	ProcessedAt time.Time `json:"processedAt,omitempty"`
	CreatedAt  time.Time `json:"createdAt"`
}

// TableName returns the table name for GORM
func (ChangeRecord) TableName() string {
	return "change_records"
}

// DeltaSyncService handles delta synchronization
type DeltaSyncService struct {
	db *gorm.DB
}

// NewDeltaSyncService creates a new delta sync service
func NewDeltaSyncService(db *gorm.DB) *DeltaSyncService {
	return &DeltaSyncService{db: db}
}

// GetLastSyncCheckpoint retrieves the last sync checkpoint
func (s *DeltaSyncService) GetLastSyncCheckpoint(
	ctx context.Context,
	tenantID string,
	platform string,
	entityType string,
) (*DeltaSyncRecord, error) {
	var record DeltaSyncRecord

	err := s.db.WithContext(ctx).
		Where("tenant_id = ? AND platform = ? AND entity_type = ?",
			tenantID, platform, entityType).
		Order("last_sync_time DESC").
		First(&record).Error

	if err != nil {
		return nil, err
	}

	return &record, nil
}

// SaveSyncCheckpoint saves a sync checkpoint
func (s *DeltaSyncService) SaveSyncCheckpoint(
	ctx context.Context,
	record *DeltaSyncRecord,
) error {
	return s.db.WithContext(ctx).Create(record).Error
}

// RecordChange records a data change for tracking
func (s *DeltaSyncService) RecordChange(
	ctx context.Context,
	tenantID string,
	platform string,
	entityType string,
	entityID string,
	changeType string,
	oldData interface{},
	newData interface{},
) error {
	change := &ChangeRecord{
		TenantID:   tenantID,
		Platform:   platform,
		EntityType: entityType,
		EntityID:   entityID,
		ChangeType: changeType,
	}

	if oldData != nil {
		data, _ := json.Marshal(oldData)
		change.OldData = string(data)
	}
	if newData != nil {
		data, _ := json.Marshal(newData)
		change.NewData = string(data)
	}

	return s.db.WithContext(ctx).Create(change).Error
}

// GetUnprocessedChanges retrieves unprocessed changes
func (s *DeltaSyncService) GetUnprocessedChanges(
	ctx context.Context,
	tenantID string,
	platform string,
	entityType string,
	limit int,
) ([]ChangeRecord, error) {
	var changes []ChangeRecord

	query := s.db.WithContext(ctx).
		Where("tenant_id = ? AND platform = ? AND entity_type = ? AND processed = ?",
			tenantID, platform, entityType, false).
		Order("created_at ASC")

	if limit > 0 {
		query = query.Limit(limit)
	}

	err := query.Find(&changes).Error
	return changes, err
}

// MarkChangesProcessed marks changes as processed
func (s *DeltaSyncService) MarkChangesProcessed(
	ctx context.Context,
	ids []uint,
) error {
	return s.db.WithContext(ctx).
		Model(&ChangeRecord{}).
		Where("id IN ?", ids).
		Updates(map[string]interface{}{
			"processed":    true,
			"processed_at": time.Now(),
		}).Error
}

// ComputeDataHash computes hash for data comparison
func (s *DeltaSyncService) ComputeDataHash(data interface{}) string {
	jsonData, _ := json.Marshal(data)
	hash := md5.Sum(jsonData)
	return hex.EncodeToString(hash[:])
}

// NeedsDeltaSync checks if delta sync is needed
func (s *DeltaSyncService) NeedsDeltaSync(
	ctx context.Context,
	tenantID string,
	platform string,
	entityType string,
	minInterval time.Duration,
) (bool, error) {
	checkpoint, err := s.GetLastSyncCheckpoint(ctx, tenantID, platform, entityType)
	if err != nil {
		// No checkpoint found, needs sync
		return true, nil
	}

	// Check if enough time has passed
	return time.Since(checkpoint.LastSyncTime) >= minInterval, nil
}

// CleanupOldChanges removes old processed changes
func (s *DeltaSyncService) CleanupOldChanges(
	ctx context.Context,
	tenantID string,
	olderThan time.Duration,
) (int64, error) {
	cutoff := time.Now().Add(-olderThan)

	result := s.db.WithContext(ctx).
		Where("tenant_id = ? AND processed = ? AND processed_at < ?",
			tenantID, true, cutoff).
		Delete(&ChangeRecord{})

	return result.RowsAffected, result.Error
}
