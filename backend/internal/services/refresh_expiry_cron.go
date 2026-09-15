// Phase 11.5 — daily cron that scans all tenants for expiring refresh
// tokens and delivers warnings via the existing notification stack.

package services

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	"github.com/omni/backend/internal/config"
	"github.com/omni/backend/internal/models"
	"github.com/omni/backend/internal/notify"
	"github.com/omni/backend/internal/repositories"
	"github.com/rs/zerolog/log"
	"gorm.io/gorm"
)

// DefaultRefreshExpiryScanInterval is how often the cron re-scans every
// tenant. Daily is enough because expiry granularity is days, not hours.
const DefaultRefreshExpiryScanInterval = 24 * time.Hour

// RefreshExpiryCron scans every tenant once per interval, fans a
// warning notification per expiring connection, and publishes a
// realtime `notifications/updated` event so the bell badge updates
// live (Phase 5 hub wire).
type RefreshExpiryCron struct {
	systemDB   *gorm.DB
	basePath   string
	interval   time.Duration
	windowDays int

	stopCh   chan struct{}
	stopOnce sync.Once
	running  atomic.Bool
}

// NewRefreshExpiryCron constructs a cron with sensible defaults.
// Passing 0 for interval/windowDays uses defaults (24h / 7 days).
func NewRefreshExpiryCron(systemDB *gorm.DB, basePath string, interval time.Duration, windowDays int) *RefreshExpiryCron {
	if interval <= 0 {
		interval = DefaultRefreshExpiryScanInterval
	}
	if windowDays <= 0 {
		windowDays = DefaultRefreshExpiryWindowDays
	}
	return &RefreshExpiryCron{
		systemDB:   systemDB,
		basePath:   basePath,
		interval:   interval,
		windowDays: windowDays,
		stopCh:     make(chan struct{}),
	}
}

// Start launches the loop in a background goroutine. Idempotent — a
// second Start on a running cron is a no-op.
func (c *RefreshExpiryCron) Start() {
	if !c.running.CompareAndSwap(false, true) {
		return
	}
	log.Info().Int("window_days", c.windowDays).
		Dur("interval", c.interval).
		Msg("🔔 Refresh-token expiry cron started")
	go c.loop()
}

// Stop signals the loop to exit. Idempotent.
func (c *RefreshExpiryCron) Stop() {
	c.stopOnce.Do(func() { close(c.stopCh) })
}

func (c *RefreshExpiryCron) loop() {
	// First scan runs immediately so startup coverage doesn't wait 24h.
	c.scanOnce()
	t := time.NewTicker(c.interval)
	defer t.Stop()
	for {
		select {
		case <-c.stopCh:
			return
		case <-t.C:
			c.scanOnce()
		}
	}
}

// scanOnce iterates every tenant schema, collects warnings, and fans
// notifications. Kept public for smoke tests / manual cron trigger.
func (c *RefreshExpiryCron) scanOnce() {
	ctx := context.Background()
	tenantIDs, err := c.listTenantIDs()
	if err != nil {
		log.Error().Err(err).Msg("refresh-expiry cron: list tenants failed")
		return
	}
	total := 0
	for _, tenantID := range tenantIDs {
		delivered := c.scanTenant(ctx, tenantID)
		total += delivered
	}
	if total > 0 {
		log.Info().Int("warnings", total).Int("tenants", len(tenantIDs)).
			Msg("🔔 Refresh-expiry cron: delivered warnings")
	}
}

func (c *RefreshExpiryCron) scanTenant(ctx context.Context, tenantID string) int {
	tenantDB, err := config.GetTenantDBByID(tenantID)
	if err != nil {
		log.Warn().Err(err).Str("tenant_id", tenantID).Msg("refresh-expiry cron: skip tenant (db unavailable)")
		return 0
	}
	repo := repositories.NewCredentialRepository(tenantDB)
	// Empty platform = all platforms for this tenant.
	conns, err := repo.ListConnections(ctx, tenantID, "")
	if err != nil {
		log.Warn().Err(err).Str("tenant_id", tenantID).Msg("refresh-expiry cron: list connections failed")
		return 0
	}
	warnings := ScanConnectionsForExpiryWarnings(tenantID, conns, c.windowDays)
	if len(warnings) == 0 {
		return 0
	}
	delivered := 0
	notifRepo := repositories.NewNotificationRepository(tenantDB)
	// V2: attach notify.Bus so EmitWithKey de-dupes storms from repeated cron
	// runs (one row per platform+store, dedup_count increments instead of a
	// new notification every hour). Bus fanout defaults to in-process; when
	// realtime.Publisher is wired, events also broadcast to /api/realtime/ws.
	bus := notify.NewBus(notifRepo, notify.NewInProcessFanout())
	notifSvc := NewNotificationService(notifRepo).WithTenant(tenantID).WithBus(bus)
	for _, w := range warnings {
		if err := c.deliver(ctx, notifSvc, w); err != nil {
			log.Warn().Err(err).Str("tenant_id", tenantID).Str("platform", w.Platform).
				Msg("refresh-expiry cron: deliver failed")
			continue
		}
		delivered++
	}
	return delivered
}

// deliver persists a notification row + publishes a compact realtime
// event. Failure to publish realtime is logged but not returned so the
// persisted row (source of truth) still counts as delivered.
//
// Uses EmitWithKey so repeated cron runs collapse into one row per
// platform+store (dedup_count increments) instead of flooding the feed.
func (c *RefreshExpiryCron) deliver(ctx context.Context, notifSvc *NotificationService, w RefreshExpiryWarning) error {
	title := FormatExpiryWarningTitle(w)
	message := FormatExpiryWarningMessage(w)
	notifType := models.NotifTypeWarning
	severity := models.SeverityMedium
	if w.AlreadyExpired {
		notifType = models.NotifTypeError
		severity = models.SeverityHigh
	}
	metaJSON, _ := json.Marshal(map[string]string{
		"platform":  w.Platform,
		"store":     w.StoreIdentifier,
		"days_left": fmt.Sprintf("%d", w.DaysLeft),
	})
	key := fmt.Sprintf("cron:refresh_expiry:%s:%s", w.Platform, w.StoreIdentifier)
	_, err := notifSvc.EmitWithKey(ctx, key, notify.Event{
		Type:      notifType,
		Category:  models.CatAuth,
		Severity:  severity,
		Title:     title,
		Message:   message,
		Metadata:  string(metaJSON),
		ActionURL: "/settings?tab=platforms",
		Source:    "cron:refresh_expiry",
	})
	if err != nil {
		return fmt.Errorf("notif emit: %w", err)
	}
	return nil
}

// listTenantIDs matches multi_tenant_scheduler's approach — cheap query
// against information_schema.
func (c *RefreshExpiryCron) listTenantIDs() ([]string, error) {
	var schemas []string
	err := c.systemDB.Raw(`
		SELECT schema_name
		FROM information_schema.schemata
		WHERE schema_name LIKE 'tenant_%'
		ORDER BY schema_name
	`).Scan(&schemas).Error
	if err != nil {
		return nil, err
	}
	tenants := make([]string, 0, len(schemas))
	for _, s := range schemas {
		if len(s) > 7 {
			tenants = append(tenants, s[7:])
		}
	}
	return tenants, nil
}
