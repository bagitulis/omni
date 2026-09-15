package repositories

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/omni/backend/internal/models"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// V2 API: per-user reads, dedup UPSERT, bulk ops, counts, snooze, richer filters.
// Legacy methods on notification_repository.go remain as-is for backward
// compatibility during the migration window. New producers/handlers should call
// only the V2 methods below.

// UpsertDedup inserts a notification or, if dedup_key matches an existing row,
// increments dedup_count and refreshes message/updated_at instead.
//
// When notif.DedupKey is nil or blank, this behaves exactly like Create.
//
// Active-window scoping: the DB unique index intentionally does NOT include a
// time predicate (Postgres forbids now() in partial-index predicates as it is
// non-IMMUTABLE). The "1 hour active window" is enforced by ExpireStaleDedupKeys
// which is called by the service layer periodically to NULL out old dedup_keys
// so a new occurrence starts a fresh row.
func (r *NotificationRepository) UpsertDedup(ctx context.Context, notif *models.Notification) (*models.Notification, error) {
	if notif == nil {
		return nil, errors.New("nil notification")
	}
	now := time.Now()
	if notif.CreatedAt.IsZero() {
		notif.CreatedAt = now
	}
	notif.UpdatedAt = now
	if notif.DedupCount <= 0 {
		notif.DedupCount = 1
	}
	if notif.Severity == 0 {
		notif.Severity = models.SeverityFromType(notif.Type)
	}
	// Postgres jsonb rejects empty strings; normalize to '{}' so producers can
	// pass "" without special-casing.
	if strings.TrimSpace(notif.Metadata) == "" {
		notif.Metadata = "{}"
	}

	// No dedup key → straight insert.
	if notif.DedupKey == nil || strings.TrimSpace(*notif.DedupKey) == "" {
		if err := r.db.WithContext(ctx).Create(notif).Error; err != nil {
			return nil, err
		}
		return notif, nil
	}

	err := r.db.WithContext(ctx).Clauses(clause.OnConflict{
		Columns: []clause.Column{{Name: "dedup_key"}},
		TargetWhere: clause.Where{Exprs: []clause.Expression{
			clause.Expr{SQL: "dedup_key IS NOT NULL"},
		}},
		DoUpdates: clause.Assignments(map[string]any{
			"dedup_count": gorm.Expr("notifications.dedup_count + 1"),
			"message":     notif.Message,
			"metadata":    notif.Metadata,
			"title":       notif.Title,
			"updated_at":  now,
		}),
	}).Create(notif).Error
	if err != nil {
		return nil, err
	}

	// GORM populates notif.ID with the returned id (either the newly-inserted
	// or the updated row's id). Re-fetch the canonical row so callers see the
	// final dedup_count / message.
	var fresh models.Notification
	if err := r.db.WithContext(ctx).
		Where("dedup_key = ?", *notif.DedupKey).
		Order("id DESC").
		First(&fresh).Error; err != nil {
		return nil, err
	}
	return &fresh, nil
}

// ExpireStaleDedupKeys NULLs out dedup_key on rows older than the given
// duration so a fresh occurrence starts a new row instead of incrementing the
// old one. Idempotent. Returns rows affected.
func (r *NotificationRepository) ExpireStaleDedupKeys(ctx context.Context, olderThan time.Duration) (int64, error) {
	cutoff := time.Now().Add(-olderThan)
	res := r.db.WithContext(ctx).
		Model(&models.Notification{}).
		Where("dedup_key IS NOT NULL AND created_at < ?", cutoff).
		Update("dedup_key", nil)
	return res.RowsAffected, res.Error
}

// MarkReadForUser records that userID has read notificationID.
// Idempotent: repeated calls do not create duplicates.
func (r *NotificationRepository) MarkReadForUser(ctx context.Context, notificationID int64, userID string) error {
	rec := &models.NotificationRead{
		NotificationID: notificationID,
		UserID:         userID,
		ReadAt:         time.Now(),
	}
	return r.db.WithContext(ctx).Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "notification_id"}, {Name: "user_id"}},
		DoNothing: true,
	}).Create(rec).Error
}

// MarkAllReadForUser bulk-inserts a read record for every existing notification
// the user has not already read. Returns number of rows inserted.
func (r *NotificationRepository) MarkAllReadForUser(ctx context.Context, userID string) (int64, error) {
	res := r.db.WithContext(ctx).Exec(`
		INSERT INTO notification_reads (notification_id, user_id, read_at)
		SELECT n.id, ?, now()
		FROM notifications n
		LEFT JOIN notification_reads r
		  ON r.notification_id = n.id AND r.user_id = ?
		WHERE r.notification_id IS NULL
	`, userID, userID)
	return res.RowsAffected, res.Error
}

// BulkMarkReadForUser marks the given IDs as read for a user. Returns number
// of rows inserted (skips IDs that are already read or do not exist).
func (r *NotificationRepository) BulkMarkReadForUser(ctx context.Context, userID string, ids []int64) (int64, error) {
	if len(ids) == 0 {
		return 0, nil
	}
	res := r.db.WithContext(ctx).Exec(`
		INSERT INTO notification_reads (notification_id, user_id, read_at)
		SELECT n.id, ?, now()
		FROM notifications n
		LEFT JOIN notification_reads r
		  ON r.notification_id = n.id AND r.user_id = ?
		WHERE n.id IN ? AND r.notification_id IS NULL
	`, userID, userID, ids)
	return res.RowsAffected, res.Error
}

// BulkDelete removes notifications by ID and returns rows affected.
func (r *NotificationRepository) BulkDelete(ctx context.Context, ids []int64) (int64, error) {
	if len(ids) == 0 {
		return 0, nil
	}
	res := r.db.WithContext(ctx).Where("id IN ?", ids).Delete(&models.Notification{})
	return res.RowsAffected, res.Error
}

// Snooze sets snoozed_until = until on a notification.
func (r *NotificationRepository) Snooze(ctx context.Context, id int64, until time.Time) error {
	return r.db.WithContext(ctx).Model(&models.Notification{}).
		Where("id = ?", id).
		Updates(map[string]any{"snoozed_until": until, "updated_at": time.Now()}).Error
}

// CountsForUser returns total, unread-for-user, and per-severity breakdown.
func (r *NotificationRepository) CountsForUser(ctx context.Context, userID string) (*Counts, error) {
	c := &Counts{BySeverity: map[int16]int64{}}

	if err := r.db.WithContext(ctx).Model(&models.Notification{}).Count(&c.Total).Error; err != nil {
		return nil, err
	}
	if err := r.db.WithContext(ctx).Raw(`
		SELECT COUNT(*)
		FROM notifications n
		LEFT JOIN notification_reads r
		  ON r.notification_id = n.id AND r.user_id = ?
		WHERE r.notification_id IS NULL
	`, userID).Scan(&c.Unread).Error; err != nil {
		return nil, err
	}

	rows, err := r.db.WithContext(ctx).Raw(`
		SELECT severity, COUNT(*) AS n
		FROM notifications
		GROUP BY severity
	`).Rows()
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var sev int16
		var n int64
		if err := rows.Scan(&sev, &n); err != nil {
			return nil, err
		}
		c.BySeverity[sev] = n
	}
	return c, nil
}

// ListActive returns notifications with V2 filtering and pagination.
// Snoozed notifications and expired notifications are excluded.
func (r *NotificationRepository) ListActive(ctx context.Context, f ListFilter) ([]models.Notification, error) {
	if f.Limit <= 0 || f.Limit > 200 {
		f.Limit = 50
	}

	q := r.db.WithContext(ctx).
		Model(&models.Notification{}).
		Where("(snoozed_until IS NULL OR snoozed_until <= now())").
		Where("(expires_at IS NULL OR expires_at > now())").
		Order("created_at DESC, id DESC").
		Limit(f.Limit)

	if f.SinceID > 0 {
		q = q.Where("id < ?", f.SinceID)
	}
	if f.Category != "" {
		q = q.Where("category = ?", f.Category)
	}
	if f.MinSeverity > 0 {
		q = q.Where("severity >= ?", f.MinSeverity)
	}
	if f.From != nil {
		q = q.Where("created_at >= ?", *f.From)
	}
	if f.To != nil {
		q = q.Where("created_at <= ?", *f.To)
	}
	if s := strings.TrimSpace(f.Search); s != "" {
		like := "%" + strings.ToLower(s) + "%"
		q = q.Where("LOWER(title) LIKE ? OR LOWER(message) LIKE ?", like, like)
	}
	if f.UnreadOnly {
		q = q.Where(`NOT EXISTS (
			SELECT 1 FROM notification_reads r
			WHERE r.notification_id = notifications.id AND r.user_id = ?
		)`, f.UserID)
	}

	var items []models.Notification
	if err := q.Find(&items).Error; err != nil {
		return nil, err
	}
	return items, nil
}

// EnsureDedupIndex creates the partial unique index used by UpsertDedup.
// Called by MigrateTenantDatabase so the DB always has the index the app
// expects, without depending on a separate migration runner.
//
// The predicate is intentionally time-free (Postgres forbids now() in index
// predicates). Active-window scoping is enforced at the application layer by
// ExpireStaleDedupKeys.
func (r *NotificationRepository) EnsureDedupIndex(ctx context.Context) error {
	return r.db.WithContext(ctx).Exec(`
		CREATE UNIQUE INDEX IF NOT EXISTS ix_notif_dedup_active
		  ON notifications (dedup_key)
		  WHERE dedup_key IS NOT NULL`).Error
}
