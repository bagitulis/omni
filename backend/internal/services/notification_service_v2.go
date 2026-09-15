package services

import (
	"context"
	"time"

	"github.com/omni/backend/internal/models"
	"github.com/omni/backend/internal/notify"
	"github.com/omni/backend/internal/repositories"
)

// V2 methods on the existing NotificationService: per-user reads, dedup emit,
// bulk ops, counts, snooze, richer list. These delegate to the repository and,
// where applicable, to a notify.Bus. They do NOT touch the legacy `read`
// column — new callers should use these methods exclusively.
//
// Backwards compat: legacy methods (Push, MarkAsRead, MarkAllAsRead,
// UnreadCount) remain on NotificationService until Phase Z of the rollout so
// existing call sites keep compiling while producers are migrated one by one.

// Bus is the notify pipeline this service uses when producers call Emit.
// Wiring lives in app_handlers; the service stays constructable without a bus
// so tests and legacy callers do not force a Redis/fanout wire.
func (s *NotificationService) WithBus(bus *notify.Bus) *NotificationService {
	s.bus = bus
	return s
}

// Emit forwards to the underlying bus. Legacy fallback: if no bus is wired,
// use Push so behaviour degrades safely to the pre-V2 path.
func (s *NotificationService) Emit(ctx context.Context, ev notify.Event) (*models.Notification, error) {
	if s.bus != nil {
		ev.TenantID = s.tenantID
		return s.bus.Emit(ctx, ev)
	}
	// Fallback preserves the old behaviour so migration is incremental.
	return s.Push(ctx, ev.Type, ev.Category, ev.Title, ev.Message, ev.ActionURL, ev.Metadata)
}

// EmitWithKey forwards to the underlying bus with a dedup key.
func (s *NotificationService) EmitWithKey(ctx context.Context, key string, ev notify.Event) (*models.Notification, error) {
	if s.bus != nil {
		ev.TenantID = s.tenantID
		return s.bus.EmitWithKey(ctx, key, ev)
	}
	return s.Push(ctx, ev.Type, ev.Category, ev.Title, ev.Message, ev.ActionURL, ev.Metadata)
}

// ListActive returns notifications with V2 filters (search, category,
// severity, unread-per-user, snooze/expiry excluded).
func (s *NotificationService) ListActive(ctx context.Context, f repositories.ListFilter) ([]models.Notification, error) {
	return s.repo.ListActive(ctx, f)
}

// CountsForUser returns total/unread/by-severity for a user.
func (s *NotificationService) CountsForUser(ctx context.Context, userID int64) (*repositories.Counts, error) {
	return s.repo.CountsForUser(ctx, userID)
}

// MarkReadForUser records that a specific user has read a notification.
func (s *NotificationService) MarkReadForUser(ctx context.Context, id, userID int64) error {
	return s.repo.MarkReadForUser(ctx, id, userID)
}

// MarkAllReadForUser bulk-marks every currently-unread notification as read
// for the user. Returns the number of newly-inserted read rows.
func (s *NotificationService) MarkAllReadForUser(ctx context.Context, userID int64) (int64, error) {
	return s.repo.MarkAllReadForUser(ctx, userID)
}

// BulkMarkReadForUser marks the specified IDs as read for the user.
func (s *NotificationService) BulkMarkReadForUser(ctx context.Context, userID int64, ids []int64) (int64, error) {
	return s.repo.BulkMarkReadForUser(ctx, userID, ids)
}

// BulkDelete removes the given notification IDs.
func (s *NotificationService) BulkDelete(ctx context.Context, ids []int64) (int64, error) {
	return s.repo.BulkDelete(ctx, ids)
}

// Snooze hides a notification until `until`.
func (s *NotificationService) Snooze(ctx context.Context, id int64, until time.Time) error {
	return s.repo.Snooze(ctx, id, until)
}

// ExpireStaleDedupKeys rotates dedup keys older than the given duration so a
// new occurrence starts a fresh row instead of forever-incrementing an old one.
// Intended to be called by the daily maintenance cron.
func (s *NotificationService) ExpireStaleDedupKeys(ctx context.Context, olderThan time.Duration) (int64, error) {
	return s.repo.ExpireStaleDedupKeys(ctx, olderThan)
}
