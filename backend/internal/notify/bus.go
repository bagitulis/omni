package notify

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/omni/backend/internal/models"
	"github.com/omni/backend/internal/repositories"
	"github.com/rs/zerolog/log"
)

// Event is the input contract for Bus.Emit. Fields are named so producers do
// not accidentally pass a positional argument in the wrong slot (a common bug
// with the legacy service.Push signature).
type Event struct {
	TenantID        string     // routing key for fanout; if empty, persistence still happens but no fanout
	Type            string     // success | error | warning | info
	Category        string     // sync | order | product | ...
	Severity        int16      // 0 = derive from Type via models.SeverityFromType
	Title           string     // short subject
	Message         string     // detail (may be structured JSON parsed by FE)
	Metadata        string     // JSON string; "" is normalized to "{}"
	ActionURL       string     // SPA-relative path (see SafeActionURL); "" = no action
	Source          string     // "cron:refresh_expiry", "handler:inventory.bulk", "security"
	ActorUserID     *string    // who triggered this notification (nil = system); UUID from auth claims
	RecipientUserID *string    // nil = tenant-wide, non-nil = target user only (UUID)
	ExpiresAt       *time.Time // notification hides after this time
}

// Bus is the canonical emit path.
type Bus struct {
	repo   *repositories.NotificationRepository
	fanout Fanout
}

// NewBus constructs a Bus.
func NewBus(repo *repositories.NotificationRepository, fanout Fanout) *Bus {
	if fanout == nil {
		fanout = NewInProcessFanout()
	}
	return &Bus{repo: repo, fanout: fanout}
}

// Emit persists and (if TenantID is set) fans out an event.
//
// Order of operations:
//   1. SafeActionURL — invalid input fails fast BEFORE any DB write.
//   2. SanitizeForUser — mutates Title/Message to strip leaks. This makes
//      leaks a strictly local bug (the writer must fix it) instead of a
//      user-visible incident.
//   3. UpsertDedup — insert or bump dedup_count for the row.
//   4. fanout.Publish — best-effort; failure does not roll back the row.
func (b *Bus) Emit(ctx context.Context, ev Event) (*models.Notification, error) {
	if b == nil || b.repo == nil {
		return nil, errors.New("notify.Bus not initialized")
	}

	url, err := SafeActionURL(ev.ActionURL)
	if err != nil {
		return nil, fmt.Errorf("notify: %w", err)
	}

	// Sanitize user-visible text.
	title := SanitizeForUser(ev.Title)
	message := SanitizeForUser(ev.Message)

	severity := ev.Severity
	if severity == 0 {
		severity = models.SeverityFromType(ev.Type)
	}

	notif := &models.Notification{
		Type:            ev.Type,
		Category:        ev.Category,
		Severity:        severity,
		Title:           title,
		Message:         message,
		Metadata:        ev.Metadata,
		ActionURL:       url,
		Source:          ev.Source,
		ActorUserID:     ev.ActorUserID,
		RecipientUserID: ev.RecipientUserID,
		ExpiresAt:       ev.ExpiresAt,
		CreatedAt:       time.Now(),
	}

	saved, err := b.repo.UpsertDedup(ctx, notif)
	if err != nil {
		return nil, fmt.Errorf("notify: persist: %w", err)
	}

	if ev.TenantID != "" {
		b.publish(ev.TenantID, saved)
	}
	return saved, nil
}

// EmitWithKey is Emit with a dedup key set. Repeated occurrences within the
// key's active window collapse into one row with dedup_count incremented.
func (b *Bus) EmitWithKey(ctx context.Context, key string, ev Event) (*models.Notification, error) {
	// We shadow the input rather than mutate the caller's Event.
	ev2 := ev
	// Attach the key by round-tripping through UpsertDedup indirectly: build
	// the notification here so we can set DedupKey directly.
	url, err := SafeActionURL(ev2.ActionURL)
	if err != nil {
		return nil, fmt.Errorf("notify: %w", err)
	}
	title := SanitizeForUser(ev2.Title)
	message := SanitizeForUser(ev2.Message)
	severity := ev2.Severity
	if severity == 0 {
		severity = models.SeverityFromType(ev2.Type)
	}

	dk := key
	notif := &models.Notification{
		Type:            ev2.Type,
		Category:        ev2.Category,
		Severity:        severity,
		Title:           title,
		Message:         message,
		Metadata:        ev2.Metadata,
		ActionURL:       url,
		DedupKey:        &dk,
		Source:          ev2.Source,
		ActorUserID:     ev2.ActorUserID,
		RecipientUserID: ev2.RecipientUserID,
		ExpiresAt:       ev2.ExpiresAt,
		CreatedAt:       time.Now(),
	}
	saved, err := b.repo.UpsertDedup(ctx, notif)
	if err != nil {
		return nil, fmt.Errorf("notify: persist: %w", err)
	}
	if ev2.TenantID != "" {
		b.publish(ev2.TenantID, saved)
	}
	return saved, nil
}

// publish is the fanout adapter — marshals a compact payload for the FE.
// It publishes to BOTH the local Fanout (used by direct in-process subscribers
// like tests) AND to the shared realtime.Publisher singleton when it has been
// wired. This lets Bus.Emit deliver over the existing dashboard WebSocket at
// /api/realtime/ws under TopicNotifications without introducing a second
// transport.
func (b *Bus) publish(tenantID string, notif *models.Notification) {
	payload := map[string]any{
		"id":          notif.ID,
		"type":        notif.Type,
		"category":    notif.Category,
		"severity":    notif.Severity,
		"title":       notif.Title,
		"message":     notif.Message,
		"metadata":    notif.Metadata,
		"action_url":  notif.ActionURL,
		"dedup_count": notif.DedupCount,
		"created_at":  notif.CreatedAt,
	}
	blob, err := json.Marshal(payload)
	if err != nil {
		log.Error().Err(err).Msg("notify: marshal payload")
		return
	}
	b.fanout.Publish(tenantID, Envelope{Payload: blob})

	// Also fan out via the shared realtime publisher so any consumer connected
	// to /api/realtime/ws receives the event under TopicNotifications. This is
	// a no-op if the hub has not been wired yet.
	if realtimePublisher != nil {
		realtimePublisher(tenantID, payload)
	}
}

// realtimePublisher is a package-level indirection so notify does not depend
// on the realtime package (which would create an import cycle via services).
// Wired by main.go via SetRealtimePublisher during startup.
var realtimePublisher func(tenantID string, payload any)

// SetRealtimePublisher wires the shared realtime fanout. Passing nil disables
// forwarding.
func SetRealtimePublisher(fn func(tenantID string, payload any)) {
	realtimePublisher = fn
}
