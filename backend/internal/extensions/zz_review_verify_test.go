package extensions

import (
	"context"
	"testing"
	"time"
)

// Verification of claims from an external adversarial review. Each test either
// confirms the claim (fails, showing the bug) or refutes it. I do not accept a
// finding without reproducing it.

// CLAIM 5: RegisterSync does not actually order after Register, because both
// channels are buffered and the loop's select picks pseudo-randomly. If true, the
// documented guarantee is false.
func TestClaim5_RegisterSyncOrdersAfterRegister(t *testing.T) {
	h := NewHub(nil)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go h.Run(ctx)

	// Repeated because a select bias would show up statistically, not once.
	for i := 0; i < 200; i++ {
		id := "ext-order-" + string(rune('A'+i%26))
		c := newTestConn(id)
		h.RegisterSync(c)

		if !h.IsConnected(id) {
			t.Fatalf("iteration %d: RegisterSync returned but %s was not registered, "+
				"so a subsequent send would fail with 'not connected'", i, id)
		}
		h.UnregisterSync(c)
	}
}

// CLAIM 4: handleRoute decrements inFlight on EVERY inbound frame, including
// frames that are not replies, so the in-flight cap can be bypassed.
func TestClaim4_InFlightNotBypassedByNonReplyFrames(t *testing.T) {
	h := NewHub(nil)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go h.Run(ctx)

	// A client that accepts commands but never replies.
	c := &Client{
		extensionID: "ext-bypass",
		send:        make(chan []byte, 4096),
		done:        make(chan struct{}),
	}
	h.RegisterSync(c)

	// Saturate the cap.
	accepted := 0
	for i := 0; i < maxInFlightPerExtension; i++ {
		if err := h.SendToExtension("ext-bypass", WSMessage{ID: "cmd", Type: MsgTypeCommand}); err == nil {
			accepted++
		}
	}
	if accepted != maxInFlightPerExtension {
		t.Fatalf("expected to fill the cap, accepted %d", accepted)
	}

	// Now send NON-reply frames, as a chatty extension might.
	for i := 0; i < maxInFlightPerExtension*2; i++ {
		h.Route(c, WSMessage{ID: "noise", Type: "progress", Action: "update"})
	}
	time.Sleep(100 * time.Millisecond)

	// The cap must still be in force: these should be refused.
	over := 0
	for i := 0; i < 20; i++ {
		if err := h.SendToExtension("ext-bypass", WSMessage{ID: "more", Type: MsgTypeCommand}); err == nil {
			over++
		}
	}
	if over > 0 {
		t.Errorf("in-flight cap bypassed: %d extra commands accepted after %d non-reply "+
			"frames were routed, so a chatty extension defeats the flood protection",
			over, maxInFlightPerExtension*2)
	}
}

// CLAIM 11: a RegisterResultChannel that lands in the buffered resultReg channel
// just before shutdown is never processed, so the waiter's channel is never
// closed and a caller waiting only on it hangs forever.
func TestClaim11_ResultChannelRegisteredDuringShutdownIsClosed(t *testing.T) {
	h := NewHub(nil)
	ctx, cancel := context.WithCancel(context.Background())
	go h.Run(ctx)
	time.Sleep(20 * time.Millisecond)

	// Register many channels so some are still buffered when shutdown runs.
	const n = 40
	chans := make([]chan WSMessage, n)
	for i := 0; i < n; i++ {
		chans[i] = make(chan WSMessage, 1)
		h.RegisterResultChannel("bulk-"+string(rune('a'+i%26)), chans[i])
	}

	// Shut down immediately.
	cancel()
	time.Sleep(200 * time.Millisecond)
	h.Stop()

	// Every registered channel must have been closed, or a waiter hangs.
	hung := 0
	for i, ch := range chans {
		select {
		case _, ok := <-ch:
			if ok {
				// A message arrived (unlikely); treat as not-closed-but-not-hung.
			}
		case <-time.After(200 * time.Millisecond):
			hung++
			_ = i
		}
	}
	if hung > 0 {
		t.Errorf("%d of %d result channels were never closed by shutdown, so a caller "+
			"waiting only on its channel would hang forever", hung, n)
	}
}

// CLAIM 6: Stop can return before shutdown completes when called concurrently
// with a starting Run.
func TestClaim6_StopWaitsForShutdownCompletion(t *testing.T) {
	for attempt := 0; attempt < 50; attempt++ {
		h := NewHub(nil)
		ctx, cancel := context.WithCancel(context.Background())

		runDone := make(chan struct{})
		go func() {
			defer close(runDone)
			h.Run(ctx)
		}()

		stopReturned := make(chan struct{})
		go func() {
			defer close(stopReturned)
			h.Stop()
		}()

		select {
		case <-stopReturned:
			// Stop returned. If the loop is still running, a caller could free
			// resources the hub still uses.
			select {
			case <-runDone:
				// Good: the loop finished before Stop returned.
			case <-time.After(50 * time.Millisecond):
				t.Fatalf("attempt %d: Stop returned while Run was still executing, so a "+
					"caller could release resources the hub still holds", attempt)
			}
		case <-time.After(2 * time.Second):
			t.Fatalf("attempt %d: Stop never returned", attempt)
		}
		cancel()
	}
}

// CLAIM 7: OnClose is invoked on the event loop, so a blocking OnClose stalls the
// whole hub.
func TestClaim7_BlockingOnCloseStallsHub(t *testing.T) {
	h := NewHub(nil)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go h.Run(ctx)

	release := make(chan struct{})
	c := newTestConn("ext-slowclose")
	c.OnClose = func(string) {
		<-release // deliberately blocking
	}
	h.RegisterSync(c)

	// Trigger removal on the event loop.
	h.Disconnect("ext-slowclose")

	// A different extension must still be serviceable. If OnClose runs on the
	// loop, this blocks.
	other := newTestConn("ext-other")
	done := make(chan struct{})
	go func() {
		h.RegisterSync(other)
		close(done)
	}()

	select {
	case <-done:
		// The hub kept working.
	case <-time.After(2 * time.Second):
		t.Errorf("a blocking OnClose stalled the hub event loop: no other extension " +
			"could be registered while the callback was running")
	}
	close(release)
}
