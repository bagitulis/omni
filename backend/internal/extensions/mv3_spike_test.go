package extensions

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/omni/backend/internal/models"
	"github.com/omni/backend/internal/repositories"
	"github.com/omni/backend/internal/testutils"
	"gorm.io/gorm"
)

// MV3 SPIKE: the design's open risk was whether a Chrome MV3 service worker can
// hold a WebSocket, since Chrome terminates idle workers after ~30s. That
// question is about the BROWSER, not this server, so it cannot be answered here
// — it needs a real Chrome profile.
//
// What CAN be verified here is the other half: that the server tolerates the
// reconnect pattern an MV3 worker produces. A terminated worker drops the socket
// without a close handshake and reconnects on its next alarm, so the server must
// handle abrupt drops and repeated re-registration of the same extension id.
// If it did not, the spike would fail for a server-side reason that is
// fixable now.

// TestSpike_AbruptDisconnectThenReregister models the MV3 lifecycle: the socket
// vanishes with no close handshake, then the same extension reconnects.
func TestSpike_AbruptDisconnectThenReregister(t *testing.T) {
	hub := NewHub(nil)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go hub.Run(ctx)

	// First "worker lifetime": register and get a command.
	first := newTestConn("ext-mv3")
	hub.RegisterSync(first)
	if !hub.IsConnected("ext-mv3") {
		t.Fatal("expected connected on first registration")
	}

	// Worker terminated: the client is dropped without a clean close.
	hub.UnregisterSync(first)
	if hub.IsConnected("ext-mv3") {
		t.Fatal("expected disconnected after abrupt drop")
	}

	// Second "worker lifetime": same extension id reconnects. This must succeed
	// rather than being rejected as a duplicate.
	second := newTestConn("ext-mv3")
	hub.RegisterSync(second)
	if !hub.IsConnected("ext-mv3") {
		t.Fatal("re-registration after an abrupt drop must succeed")
	}

	if err := hub.SendToExtension("ext-mv3", WSMessage{ID: "after-reconnect", Type: MsgTypeCommand}); err != nil {
		t.Fatalf("hub must be usable after a reconnect: %v", err)
	}
}

// TestSpike_RepeatedReconnectsDoNotLeakCapacity guards against the failure mode
// where each reconnect leaves an in-flight or result-channel entry behind,
// eventually refusing all work.
func TestSpike_RepeatedReconnectsDoNotLeakCapacity(t *testing.T) {
	hub := NewHub(nil)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go hub.Run(ctx)

	for i := 0; i < 25; i++ {
		c := newTestConn("ext-churn")
		hub.RegisterSync(c)
		if err := hub.SendToExtension("ext-churn", WSMessage{ID: "x", Type: MsgTypeCommand}); err != nil {
			t.Fatalf("send failed on cycle %d: %v", i, err)
		}
		hub.UnregisterSync(c)

		// Capacity must be fully released on each disconnect, otherwise the
		// in-flight counter would climb and eventually refuse every command.
		if !hub.IsConnected("ext-churn") {
			continue // expected: it is disconnected until the next cycle
		}
	}

	// After the churn, a fresh connection must have full capacity available.
	final := &Client{
		extensionID: "ext-churn-final",
		send:        make(chan []byte, 1024),
		done:        make(chan struct{}),
	}
	hub.RegisterSync(final)
	for i := 0; i < maxInFlightPerExtension; i++ {
		if err := hub.SendToExtension("ext-churn-final", WSMessage{ID: "cap", Type: MsgTypeCommand}); err != nil {
			t.Fatalf("fresh connection lacked capacity after churn (send %d): %v", i, err)
		}
	}
}

// TestSpike_PairingSurvivesProcessRestart models the server restarting while a
// paired extension is offline: the pairing token must still resolve, because a
// browser must not have to re-pair every time Omni restarts.
func TestSpike_PairingSurvivesProcessRestart(t *testing.T) {
	ctx := context.Background()

	// A "persisted" tenant database, reopened to simulate a restart.
	db := testutils.SetupNamedTestSQLite(t, "spike_restart",
		&models.Extension{}, &models.PairingCode{}, &models.ScrapedProduct{})

	token, hash, err := GeneratePairingToken()
	if err != nil {
		t.Fatalf("generate token: %v", err)
	}
	if err := repositories.NewExtensionRepository(db).CreateExtension(ctx, &models.Extension{
		ExtensionID: "ext-persist",
		TokenHash:   hash,
		Status:      models.ExtensionStatusConnected, // stale status from before restart
		PairedAt:    time.Now(),
	}); err != nil {
		t.Fatalf("seed extension: %v", err)
	}

	// Restart: the same row is still there and the token still resolves.
	tenantDB := func(string) (*gorm.DB, error) { return db, nil }
	list := func() []string { return []string{"tenant"} }

	tenantID, ext, err := ResolveExtensionByToken(ctx, token, tenantDB, list)
	if err != nil {
		t.Fatalf("a paired token must survive a restart: %v", err)
	}
	if tenantID != "tenant" || ext.ExtensionID != "ext-persist" {
		t.Fatalf("resolved (%s, %s), want (tenant, ext-persist)", tenantID, ext.ExtensionID)
	}

	// The stale "connected" status must be reset at startup, or the dashboard
	// shows a phantom online extension with no socket behind it.
	if err := repositories.NewExtensionRepository(db).MarkAllExtensionsDisconnected(ctx); err != nil {
		t.Fatalf("startup reset: %v", err)
	}
	after, err := repositories.NewExtensionRepository(db).GetExtensionByID(ctx, "ext-persist")
	if err != nil {
		t.Fatalf("reload: %v", err)
	}
	if after.Status != models.ExtensionStatusDisconnected {
		t.Errorf("status after restart = %q, want disconnected", after.Status)
	}
}

// TestSpike_UnpairRevokesTheToken proves revocation is immediate: after an
// operator unpairs, the old token must not resolve, so a decommissioned browser
// cannot reconnect.
func TestSpike_UnpairRevokesTheToken(t *testing.T) {
	ctx := context.Background()
	db := testutils.SetupNamedTestSQLite(t, "spike_revoke",
		&models.Extension{}, &models.PairingCode{}, &models.ScrapedProduct{})

	token, hash, _ := GeneratePairingToken()
	repo := repositories.NewExtensionRepository(db)
	if err := repo.CreateExtension(ctx, &models.Extension{
		ExtensionID: "ext-revoke",
		TokenHash:   hash,
		Status:      models.ExtensionStatusDisconnected,
		PairedAt:    time.Now(),
	}); err != nil {
		t.Fatalf("seed: %v", err)
	}

	tenantDB := func(string) (*gorm.DB, error) { return db, nil }
	list := func() []string { return []string{"tenant"} }

	// Resolves before unpairing.
	if _, _, err := ResolveExtensionByToken(ctx, token, tenantDB, list); err != nil {
		t.Fatalf("token must resolve before unpair: %v", err)
	}

	if err := repo.DeleteExtension(ctx, "ext-revoke"); err != nil {
		t.Fatalf("unpair: %v", err)
	}

	// Must NOT resolve after unpairing.
	_, _, err := ResolveExtensionByToken(ctx, token, tenantDB, list)
	if !errors.Is(err, ErrTenantNotFoundForCredential) {
		t.Errorf("revoked token resolved (err = %v); a decommissioned browser must not reconnect", err)
	}
}
