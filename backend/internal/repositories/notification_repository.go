package repositories

import (
	"context"
	"time"

	"github.com/omni/backend/internal/models"
	"gorm.io/gorm"
)

// NotificationRepository handles notification data access.
type NotificationRepository struct {
	db *gorm.DB
}

// NewNotificationRepository creates a new notification repository.
func NewNotificationRepository(db *gorm.DB) *NotificationRepository {
	return &NotificationRepository{db: db}
}

// ListFilter drives ListActive with the V2 filter set.
type ListFilter struct {
	Limit       int
	SinceID     int64
	UnreadOnly  bool
	UserID      string // required when UnreadOnly is set (UUID from auth claims)
	Category    string // exact match, "" = any
	MinSeverity int16  // >= filter, 0 = no filter
	Search      string // ILIKE substring against title/message, "" = no filter
	From        *time.Time
	To          *time.Time
}

// Counts is the shape returned by CountsForUser.
type Counts struct {
	Total      int64
	Unread     int64
	BySeverity map[int16]int64
}

// Create saves a notification to the database.
func (r *NotificationRepository) Create(ctx context.Context, notif *models.Notification) error {
	return r.db.WithContext(ctx).Create(notif).Error
}

// List returns recent notifications with pagination and optional unread filter.
func (r *NotificationRepository) List(ctx context.Context, limit int, sinceID int64, unreadOnly bool) ([]models.Notification, error) {
	if limit <= 0 || limit > 100 {
		limit = 50
	}

	query := r.db.WithContext(ctx).Order("created_at DESC").Limit(limit)
	if sinceID > 0 {
		query = query.Where("id < ?", sinceID)
	}
	if unreadOnly {
		query = query.Where("read = false")
	}

	var items []models.Notification
	err := query.Find(&items).Error
	return items, err
}

// GetByID returns a single notification by ID.
func (r *NotificationRepository) GetByID(ctx context.Context, id int64) (*models.Notification, error) {
	var notif models.Notification
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&notif).Error
	if err != nil {
		return nil, err
	}
	return &notif, nil
}

// UnreadCount returns the number of unread notifications.
func (r *NotificationRepository) UnreadCount(ctx context.Context) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&models.Notification{}).Where("read = false").Count(&count).Error
	return count, err
}

// MarkAsRead marks a single notification as read.
func (r *NotificationRepository) MarkAsRead(ctx context.Context, id int64) error {
	return r.db.WithContext(ctx).Model(&models.Notification{}).
		Where("id = ?", id).
		Update("read", true).Error
}

// MarkAllAsRead marks all notifications as read.
func (r *NotificationRepository) MarkAllAsRead(ctx context.Context) error {
	return r.db.WithContext(ctx).Model(&models.Notification{}).
		Where("read = false").
		Update("read", true).Error
}

// Delete removes a single notification.
func (r *NotificationRepository) Delete(ctx context.Context, id int64) error {
	return r.db.WithContext(ctx).Delete(&models.Notification{}, id).Error
}

// DeleteAll removes all notifications.
func (r *NotificationRepository) DeleteAll(ctx context.Context) error {
	return r.db.WithContext(ctx).Session(&gorm.Session{AllowGlobalUpdate: true}).Delete(&models.Notification{}).Error
}

// CleanupOlderThan removes notifications older than the given number of days.
// Returns the number of rows affected.
func (r *NotificationRepository) CleanupOlderThan(ctx context.Context, days int) (int64, error) {
	if days <= 0 {
		return 0, nil
	}
	cutoff := time.Now().AddDate(0, 0, -days)
	result := r.db.WithContext(ctx).Where("created_at < ?", cutoff).Delete(&models.Notification{})
	return result.RowsAffected, result.Error
}

// GetSettings returns notification settings. Returns default if not found.
func (r *NotificationRepository) GetSettings(ctx context.Context) (*models.NotificationSettings, error) {
	var settings models.NotificationSettings
	err := r.db.WithContext(ctx).First(&settings).Error
	if err == gorm.ErrRecordNotFound {
		return &models.NotificationSettings{RetentionDays: 30}, nil
	}
	return &settings, err
}

// SaveSettings upserts notification settings.
func (r *NotificationRepository) SaveSettings(ctx context.Context, retentionDays int) error {
	var existing models.NotificationSettings
	err := r.db.WithContext(ctx).First(&existing).Error

	if err == gorm.ErrRecordNotFound {
		return r.db.WithContext(ctx).Create(&models.NotificationSettings{
			RetentionDays: retentionDays,
			CreatedAt:     time.Now(),
			UpdatedAt:     time.Now(),
		}).Error
	}
	if err != nil {
		return err
	}

	existing.RetentionDays = retentionDays
	existing.UpdatedAt = time.Now()
	return r.db.WithContext(ctx).Save(&existing).Error
}
