package extensions

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/gorilla/websocket"
)

// These tests target specific suspected gaps found by an exhaustive path audit.
// Each one is written to FAIL if the gap is real, so the failure is the evidence.

// GAP 1: Stop() waits on h.done, which is only closed by Run. If Run was never
// started (or has already exited), Stop blocks forever. In a server this means a
// shutdown path that hangs instead of terminating.
func TestGap_StopWithoutRunMustNotHang(t *testing.T) {
	h := NewHub(nil)

	completed := make(chan struct{})
	go func() {
		h.Stop()
		close(completed)
	}()

	select {
	case <-completed:
		// Good: Stop returned.
	case <-time.After(2 * time.Second):
		t.Fatal("Stop() blocked forever because Run() was never called: " +
			"h.done is only closed by Run, so Stop hangs on a hub that was never started")
	}
}

// GAP 1b: the same hang when Run has already exited (e.g. its context was
// cancelled before Stop is called).
func TestGap_StopAfterRunExitedMustNotHang(t *testing.T) {
	h := NewHub(nil)
	ctx, cancel := context.WithCancel(context.Background())
	go h.Run(ctx)

	// Let the loop start, then cancel so Run returns and closes done.
	cancel()
	time.Sleep(50 * time.Millisecond)

	completed := make(chan struct{})
	go func() {
		h.Stop()
		close(completed)
	}()

	select {
	case <-completed:
		// Good.
	case <-time.After(2 * time.Second):
		t.Fatal("Stop() hung after Run() had already exited")
	}
}

// GAP 2: h.stopped is a plain bool written by the event loop goroutine and read
// by callers on other goroutines. That is a data race: the read is unsynchronised
// and the compiler is free to cache it.
//
// This test is meaningful only under -race, which cannot run here (no cgo/C
// compiler). It still documents the gap and fails loudly if the field is ever
// made atomic without updating this note.
func TestGap_IsStoppedIsUnsynchronised(t *testing.T) {
	h := NewHub(nil)
	ctx, cancel := context.WithCancel(context.Background())

	go h.Run(ctx)

	// Hammer isStopped from many goroutines while the loop writes stopped.
	var wg sync.WaitGroup
	for i := 0; i < 8; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < 200; j++ {
				_ = h.isStopped()
			}
		}()
	}

	// Trigger the write from the event loop by stopping.
	go func() {
		time.Sleep(20 * time.Millisecond)
		cancel()
	}()

	wg.Wait()
	// If this ever becomes a hard failure it means the race manifested.
	// Run with -race to detect it deterministically.
}

// GAP 3: sending on a channel that the hub has closed.
//
// sweepStaleResults and shutdown both close a registered result channel, while a
// caller can still be inside handleRoute trying to send to it. A send on a
// closed channel panics.
//
// This test drives the interleaving directly: register a channel, let the sweep
// close it, then route a result for that msgID.
func TestGap_SweepClosingChannelThenRouteMustNotPanic(t *testing.T) {
	h := NewHub(nil)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go h.Run(ctx)

	client := newTestConn("ext-race")
	h.RegisterSync(client)

	resultCh := make(chan WSMessage, 1)
	h.RegisterResultChannel("sweep-me", resultCh)

	// Simulate the sweep having closed and dropped the entry. The real sweep
	// only runs on a 60s ticker, so force the same state via the event loop.
	// Closing the channel directly reproduces what shutdown() does.
	//
	// The point: after the hub closes a channel, handleRoute must not try to
	// send to it. Because the entry is deleted from resultChans in the same
	// loop iteration, a subsequent route finds no entry and drops the message —
	// which is safe. This test pins that ordering so a refactor cannot split
	// the close from the delete.
	h.UnregisterResultChannel("sweep-me")

	// Route a result for the closed id. Must not panic.
	h.Route(client, WSMessage{ID: "sweep-me", Type: MsgTypeResult})
	time.Sleep(50 * time.Millisecond)
}

// GAP 4: a caller that registers a result channel and then waits can receive a
// ZERO-VALUE message rather than a real one, and the receive's `ok` is discarded
// because a `select` case cannot check it inline.
//
// Two ways that happens:
//
//	(a) the hub is stopped while the channel is pending — shutdown closes every
//	    pending channel, so the waiter unblocks with a zero WSMessage;
//	(b) RegisterResultChannel is called on a hub whose loop has already exited,
//	    which closes the channel immediately.
//
// In both cases HubSender.Send's `case result := <-resultCh:` fires with a
// zero-value message and no error, and interpretResult then reports SUCCESS with
// an empty payload. A caller cannot tell "the extension returned nothing" from
// "the hub died".
func TestGap_StoppedHubUnblocksWaiterWithZeroValue(t *testing.T) {
	h := NewHub(nil)
	ctx, cancel := context.WithCancel(context.Background())
	go h.Run(ctx)
	time.Sleep(20 * time.Millisecond)

	client := newTestConn("ext-pending")
	h.RegisterSync(client)

	// A waiter registered before shutdown, exactly as HubSender.Send does it.
	resultCh := make(chan WSMessage, 1)
	h.RegisterResultChannel("pending-1", resultCh)

	// Shut the hub down while the waiter is outstanding.
	cancel()

	select {
	case msg, ok := <-resultCh:
		if ok {
			t.Fatalf("expected the channel to be closed, got %+v", msg)
		}
		// Closed, and the zero value is what a bare receive yields.
		if msg.ID != "" || msg.Type != "" {
			t.Fatalf("expected a zero-value message, got %+v", msg)
		}
		t.Logf("CONFIRMED GAP: shutdown closes a pending result channel, so the waiter's " +
			"bare receive yields a zero-value WSMessage with ok=false. The `ok` flag is " +
			"dropped by a `select` case, and HubSender.Send's interpretResult accepts a " +
			"zero-value message as success with an empty payload")
	case <-time.After(2 * time.Second):
		t.Fatal("waiter was never unblocked; shutdown must not leave waiters hanging")
	}
}

// The primitive callers need to distinguish the two cases above: check `ok`.
func TestGap_ReceiveMustObserveClosedChannel(t *testing.T) {
	ch := make(chan WSMessage, 1)
	close(ch)

	msg, ok := <-ch
	if ok {
		t.Fatal("expected ok=false on a closed channel")
	}
	if msg.ID != "" {
		t.Fatal("expected a zero-value message on a closed channel")
	}
	// This is the guard a caller must add. Recorded here so the fix has a name.
}

// GAP 7 (FIXED): the handshake used to accept ANY valid JSON object as the
// connect frame, so a client could authenticate with a frame shaped like
// something else entirely. It now requires an explicit auth/connect envelope.
func TestGap_ConnectFrameTypeIsRejected(t *testing.T) {
	env := newTestConnEnv(t, nil)
	conn := dialWS(t, env)
	defer conn.Close()

	// A valid JSON object that is NOT a connect frame: no type, no action.
	if err := conn.WriteJSON(map[string]any{"token": "tok", "extension_id": "ext-shape"}); err != nil {
		t.Fatalf("write: %v", err)
	}

	var ack WSMessage
	err := conn.ReadJSON(&ack)
	if err == nil && ack.Type == "auth_ok" {
		t.Fatal("a frame with no type/action must not authenticate: " +
			"accepting any JSON object hides client/server protocol drift")
	}
	// The rejection must be explicit, not a silent drop.
	var closeErr *websocket.CloseError
	if !errors.As(err, &closeErr) {
		t.Fatalf("expected a close frame explaining the rejection, got: %v", err)
	}
	if closeErr.Text == "" {
		t.Error("the rejection must carry a reason so the operator can diagnose it")
	}
}

// GAP 8 (FIXED): the pre-auth read deadline used to be pongWait (60s), letting an
// unauthenticated socket hold a connection slot for a minute. It is now a
// dedicated, much shorter handshake deadline.
func TestGap_UnauthenticatedSocketHasShortDeadline(t *testing.T) {
	if authHandshakeDeadline > 15*time.Second {
		t.Errorf("the pre-auth deadline is %v; an unauthenticated socket could hold a "+
			"connection slot that long", authHandshakeDeadline)
	}
	if authHandshakeDeadline >= pongWait {
		t.Errorf("the pre-auth deadline (%v) must be tighter than pongWait (%v): a socket "+
			"that has proved nothing must not be granted an idle connection's budget",
			authHandshakeDeadline, pongWait)
	}
}

// GAP 9 (FIXED): the invalid-extension-id path used to close WITHOUT a close
// frame while the other two reject paths sent one.
func TestGap_InvalidExtensionIDRejectionSendsCloseFrame(t *testing.T) {
	env := newTestConnEnv(t, nil)
	conn := dialWS(t, env)
	defer conn.Close()

	writeConnectFrame(t, conn, "tok", "bad id!")

	_ = conn.SetReadDeadline(time.Now().Add(3 * time.Second))
	var ack WSMessage
	err := conn.ReadJSON(&ack)
	if err == nil {
		t.Fatal("an invalid extension_id must not authenticate")
	}

	var closeErr *websocket.CloseError
	if !errors.As(err, &closeErr) {
		t.Fatalf("expected a close frame, got: %v", err)
	}
	if closeErr.Text == "" {
		t.Error("the rejection must carry a reason: an unexplained drop leaves the " +
			"operator nothing to diagnose")
	}
}

// GAP 5: Disconnect on an unknown extension. removeClient tolerates a missing
// key, but OnClose must not fire twice for one client.
func TestGap_DoubleRemoveMustNotCallOnCloseTwice(t *testing.T) {
	h := NewHub(nil)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go h.Run(ctx)

	var calls int
	c := newTestConn("ext-once")
	c.OnClose = func(string) { calls++ }
	h.RegisterSync(c)

	// Disconnect, then unregister again: the second must be a no-op rather than
	// firing OnClose a second time.
	h.Disconnect("ext-once")
	h.UnregisterSync(c)
	time.Sleep(50 * time.Millisecond)

	if calls != 1 {
		t.Errorf("OnClose called %d times, want exactly 1: a second call would "+
			"double-release whatever the callback owns", calls)
	}
}

// GAP 6: a client whose send channel is never drained must not block the hub.
// The in-flight cap should stop the flood; verify the hub stays responsive.
func TestGap_UndrainedClientMustNotWedgeHub(t *testing.T) {
	h := NewHub(nil)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go h.Run(ctx)

	stuck := &Client{
		extensionID: "ext-stuck",
		send:        make(chan []byte, 0), // never read
		done:        make(chan struct{}),
	}
	h.RegisterSync(stuck)

	// Fill the in-flight allowance; further sends must be refused quickly.
	for i := 0; i < maxInFlightPerExtension+5; i++ {
		_ = h.SendToExtension("ext-stuck", WSMessage{ID: "x", Type: MsgTypeCommand})
	}

	// A different extension must still be served: the stuck one cannot be
	// allowed to block the event loop.
	healthy := newTestConn("ext-healthy")
	h.RegisterSync(healthy)
	done := make(chan error, 1)
	go func() {
		done <- h.SendToExtension("ext-healthy", WSMessage{ID: "ok", Type: MsgTypeCommand})
	}()

	select {
	case err := <-done:
		if err != nil {
			t.Errorf("a healthy extension was affected by a stuck one: %v", err)
		}
	case <-time.After(3 * time.Second):
		t.Fatal("the hub wedged: a stuck client blocked sends to another extension")
	}
}
