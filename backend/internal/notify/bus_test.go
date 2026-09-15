package notify

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/omni/backend/internal/models"
	"github.com/omni/backend/internal/repositories"
	"github.com/omni/backend/internal/testutils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupBus(t *testing.T) (*Bus, *repositories.NotificationRepository, *InProcessFanout) {
	db := testutils.SetupTestPostgresWithModels(t,
		&models.Notification{},
		&models.NotificationRead{},
		&models.NotificationSettings{},
	)
	require.NoError(t, db.Exec(`
		CREATE UNIQUE INDEX IF NOT EXISTS ix_notif_dedup_active
		  ON notifications (dedup_key)
		  WHERE dedup_key IS NOT NULL`).Error)
	repo := repositories.NewNotificationRepository(db)
	f := NewInProcessFanout()
	return NewBus(repo, f), repo, f
}

func TestBus_EmitPersistsAndFansOut(t *testing.T) {
	bus, repo, f := setupBus(t)
	ch := f.Subscribe("t-1")
	defer f.Unsubscribe("t-1", ch)

	ev := Event{
		TenantID: "t-1",
		Type:     models.NotifTypeSuccess,
		Category: models.CatOrder,
		Severity: models.SeverityLow,
		Title:    "Order confirmed",
		Message:  "ok",
	}
	notif, err := bus.Emit(context.Background(), ev)
	require.NoError(t, err)
	require.NotNil(t, notif)
	assert.NotZero(t, notif.ID)

	// Verify persistence.
	fromDB, err := repo.GetByID(context.Background(), notif.ID)
	require.NoError(t, err)
	assert.Equal(t, ev.Title, fromDB.Title)

	// Verify fanout.
	select {
	case env := <-ch:
		var payload map[string]any
		require.NoError(t, json.Unmarshal(env.Payload, &payload))
		assert.EqualValues(t, notif.ID, int64(payload["id"].(float64)))
	case <-time.After(500 * time.Millisecond):
		t.Fatal("fanout did not deliver event")
	}
}

func TestBus_EmitSanitizesTitleAndMessage(t *testing.T) {
	bus, repo, _ := setupBus(t)

	ev := Event{
		TenantID: "t-1",
		Type:     models.NotifTypeError,
		Category: models.CatSync,
		Severity: models.SeverityHigh,
		Title:    "panic: runtime error boom",
		Message:  "SELECT * FROM secret_users",
	}
	notif, err := bus.Emit(context.Background(), ev)
	require.NoError(t, err)

	fromDB, err := repo.GetByID(context.Background(), notif.ID)
	require.NoError(t, err)
	assert.NotEqual(t, ev.Title, fromDB.Title, "title must be sanitized")
	assert.NotEqual(t, ev.Message, fromDB.Message, "message must be sanitized")
}

func TestBus_EmitRejectsInvalidActionURL(t *testing.T) {
	bus, _, _ := setupBus(t)

	_, err := bus.Emit(context.Background(), Event{
		TenantID:  "t-1",
		Type:      models.NotifTypeInfo,
		Category:  models.CatSystem,
		Severity:  models.SeverityLow,
		Title:     "x",
		ActionURL: "https://evil.example",
	})
	assert.Error(t, err, "invalid action_url must fail Emit")
}

func TestBus_EmitWithKeyDedupCollapses(t *testing.T) {
	bus, repo, _ := setupBus(t)
	ctx := context.Background()

	first, err := bus.EmitWithKey(ctx, "cron:refresh", Event{
		TenantID: "t-1",
		Type:     models.NotifTypeError,
		Category: models.CatSystem,
		Severity: models.SeverityMedium,
		Title:    "credential expiring",
		Message:  "1st",
	})
	require.NoError(t, err)

	second, err := bus.EmitWithKey(ctx, "cron:refresh", Event{
		TenantID: "t-1",
		Type:     models.NotifTypeError,
		Category: models.CatSystem,
		Severity: models.SeverityMedium,
		Title:    "credential expiring",
		Message:  "2nd",
	})
	require.NoError(t, err)
	assert.Equal(t, first.ID, second.ID, "same dedup key must collapse")
	assert.Equal(t, 2, second.DedupCount)

	// Fanout still fires for BOTH occurrences (client counts them).
	fromDB, err := repo.GetByID(ctx, first.ID)
	require.NoError(t, err)
	assert.Equal(t, "2nd", fromDB.Message, "latest message wins")
}

func TestBus_EmitSkipsFanoutForEmptyTenant(t *testing.T) {
	bus, _, f := setupBus(t)
	ch := f.Subscribe("t-1")
	defer f.Unsubscribe("t-1", ch)

	_, err := bus.Emit(context.Background(), Event{
		Type:     models.NotifTypeInfo,
		Category: models.CatSystem,
		Title:    "sysgen no tenant",
	})
	require.NoError(t, err) // still persists
	select {
	case <-ch:
		t.Fatal("must not fan out when tenant is empty")
	case <-time.After(150 * time.Millisecond):
	}
}
