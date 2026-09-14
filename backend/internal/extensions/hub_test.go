package extensions

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

// The hub is the riskiest component: it multiplexes commands and correlates
// results across concurrent goroutines. These tests are pure unit tests (no
// network, no database) so they execute even without a container runtime.

const testSendTimeout = 2 * time.Second

func newTestConn(id string) *Client {
	return &Client{
		extensionID: id,
		send:        make(chan []byte, sendBufferSize),
		done:        make(chan struct{}),
	}
}

func TestHub_RegisterAndSendToExtension(t *testing.T) {
	h := NewHub(nil)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go h.Run(ctx)

	client := newTestConn("ext-1")
	h.RegisterSync(client)

	msg := WSMessage{ID: "m1", Type: "command", Action: "open_tab"}
	if err := h.SendToExtension("ext-1", msg); err != nil {
		t.Fatalf("SendToExtension error: %v", err)
	}

	select {
	case raw := <-client.send:
		var got WSMessage
		if err := json.Unmarshal(raw, &got); err != nil {
			t.Fatalf("unmarshal: %v", err)
		}
		if got.ID != "m1" || got.Action != "open_tab" {
			t.Errorf("got %+v, want ID=m1 Action=open_tab", got)
		}
	case <-time.After(testSendTimeout):
		t.Fatal("timed out waiting for message on client send channel")
	}
}

func TestHub_SendToUnknownExtensionErrors(t *testing.T) {
	h := NewHub(nil)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go h.Run(ctx)

	err := h.SendToExtension("nobody", WSMessage{ID: "x"})
	if err == nil {
		t.Fatal("sending to an unregistered extension must return an error, not silently drop")
	}
}

func TestHub_ResultCorrelation(t *testing.T) {
	h := NewHub(nil)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go h.Run(ctx)

	client := newTestConn("ext-c")
	h.RegisterSync(client)

	resultCh := make(chan WSMessage, 1)
	h.RegisterResultChannel("msg-42", resultCh)

	// Simulate the extension replying with the correlated ID.
	payload, _ := json.Marshal(map[string]any{"success": true})
	h.Route(client, WSMessage{ID: "msg-42", Type: "result", Payload: payload})

	select {
	case got := <-resultCh:
		if got.ID != "msg-42" {
			t.Errorf("correlated ID = %q, want msg-42", got.ID)
		}
	case <-time.After(testSendTimeout):
		t.Fatal("result was not routed to the registered channel")
	}
}

func TestHub_ResultForUnknownIDIsNotFatal(t *testing.T) {
	h := NewHub(nil)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go h.Run(ctx)

	client := newTestConn("ext-u")
	h.RegisterSync(client)

	// A late/unsolicited result must be dropped without panicking — the runner
	// may have already timed out and unregistered.
	h.Route(client, WSMessage{ID: "never-registered", Type: "result"})

	// Hub must still be usable afterwards. A short settle is needed because
	// Route is asynchronous.
	time.Sleep(50 * time.Millisecond)
	if err := h.SendToExtension("ext-u", WSMessage{ID: "after"}); err != nil {
		t.Fatalf("hub became unusable after unsolicited result: %v", err)
	}
}

func TestHub_UnregisterResultChannel(t *testing.T) {
	h := NewHub(nil)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go h.Run(ctx)

	client := newTestConn("ext-x")
	h.RegisterSync(client)

	resultCh := make(chan WSMessage, 1)
	h.RegisterResultChannel("msg-gone", resultCh)
	h.UnregisterResultChannel("msg-gone")

	// Registration and unregistration are queued operations, so a Route issued
	// immediately after could be processed before them. Barrier through the loop
	// to make ordering deterministic rather than sleeping and hoping.
	h.RegisterSync(newTestConn("barrier-1"))

	// After unregistering, a result must not be delivered (and must not block).
	h.Route(client, WSMessage{ID: "msg-gone", Type: "result"})
	h.RegisterSync(newTestConn("barrier-2"))

	select {
	case <-resultCh:
		t.Error("result delivered after channel was unregistered")
	case <-time.After(150 * time.Millisecond):
		// Expected: nothing delivered.
	}
}

func TestHub_InFlightLimit(t *testing.T) {
	h := NewHub(nil)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go h.Run(ctx)

	// A client whose send buffer never drains, so in-flight slots stay occupied.
	client := &Client{
		extensionID: "ext-full",
		send:        make(chan []byte, 0),
		done:        make(chan struct{}),
	}
	h.RegisterSync(client)

	// Fill the in-flight allowance. Sends are non-blocking on an unbuffered
	// channel, so each occupies a slot without a reader.
	var rejected bool
	for i := 0; i < maxInFlightPerExtension+5; i++ {
		if err := h.SendToExtension("ext-full", WSMessage{ID: "flood"}); err != nil {
			rejected = true
			break
		}
	}
	if !rejected {
		t.Errorf("expected the hub to reject sends beyond the in-flight limit of %d",
			maxInFlightPerExtension)
	}
}

func TestHub_ConcurrentSendsAreSerialised(t *testing.T) {
	h := NewHub(nil)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go h.Run(ctx)

	const n = 50
	client := &Client{
		extensionID: "ext-conc",
		send:        make(chan []byte, n*2),
		done:        make(chan struct{}),
	}
	h.RegisterSync(client)

	// Exactly one reader: a single consumer replies to every delivered command so
	// in-flight slots are released. Two competing readers would steal messages
	// from each other and make the count meaningless.
	var delivered atomic.Int64
	consumerDone := make(chan struct{})
	go func() {
		defer close(consumerDone)
		for raw := range client.send {
			delivered.Add(1)
			var m WSMessage
			_ = json.Unmarshal(raw, &m)
			h.Route(client, WSMessage{ID: m.ID, Type: MsgTypeResult})
		}
	}()

	var wg sync.WaitGroup
	var accepted atomic.Int64
	var rejected atomic.Int64
	for i := 0; i < n; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if err := h.SendToExtension("ext-conc", WSMessage{ID: "c", Type: MsgTypeCommand}); err != nil {
				rejected.Add(1)
			} else {
				accepted.Add(1)
			}
		}()
	}
	wg.Wait()

	// Contract under concurrency:
	//   * no accepted send is lost — accepted == delivered (the event loop is the
	//     sole writer, so a queued command cannot vanish)
	//   * every send gets a verdict — accepted + rejected == n
	//
	// The in-flight cap legitimately refuses the overflow when many goroutines
	// fire at once, so accepted may be well below n. That is the cap working as
	// designed; deterministic capacity behaviour is covered separately by
	// TestHub_ReplyReleasesCapacity.
	deadline := time.After(testSendTimeout)
	for {
		if delivered.Load() >= accepted.Load() {
			break
		}
		select {
		case <-deadline:
			t.Fatalf("accepted %d but only %d delivered: the hub dropped a queued command",
				accepted.Load(), delivered.Load())
		case <-time.After(5 * time.Millisecond):
		}
	}

	if accepted.Load()+rejected.Load() != n {
		t.Errorf("accepted(%d) + rejected(%d) != %d: a send returned no verdict",
			accepted.Load(), rejected.Load(), n)
	}
	if accepted.Load() == 0 {
		t.Error("no sends accepted at all; the hub is not accepting work")
	}

	close(client.send)
	<-consumerDone
}

func TestHub_BroadcastToDashboardsDoesNotPanic(t *testing.T) {
	h := NewHub(nil)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go h.Run(ctx)

	// No dashboard clients registered — must be a safe no-op.
	h.BroadcastToDashboards(WSMessage{Type: "task_progress"})
}

func TestHub_UnregisterRemovesClient(t *testing.T) {
	h := NewHub(nil)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go h.Run(ctx)

	client := newTestConn("ext-bye")
	h.RegisterSync(client)
	h.UnregisterSync(client)

	if err := h.SendToExtension("ext-bye", WSMessage{ID: "x"}); err == nil {
		t.Error("send after unregister must fail")
	}
	if h.IsConnected("ext-bye") {
		t.Error("IsConnected must report false after unregister")
	}
}

func TestHub_IsConnectedAndConnectedIDs(t *testing.T) {
	h := NewHub(nil)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go h.Run(ctx)

	if h.IsConnected("ext-none") {
		t.Error("unknown extension must not report connected")
	}

	a := newTestConn("ext-a")
	b := newTestConn("ext-b")
	h.RegisterSync(a)
	h.RegisterSync(b)

	if !h.IsConnected("ext-a") {
		t.Error("registered extension must report connected")
	}
	ids := h.ConnectedIDs()
	if len(ids) != 2 {
		t.Errorf("ConnectedIDs() returned %d ids, want 2 (%v)", len(ids), ids)
	}
}

// In-flight capacity must be released by a correlated reply, otherwise a
// long-lived extension would exhaust its allowance after 10 commands and stall
// forever.
//
// The reply must carry the SAME id the command was sent with: capacity is only
// released by a reply that matches an outstanding command, which is what stops
// unrelated traffic from resetting the counter and bypassing the cap.
func TestHub_ReplyReleasesCapacity(t *testing.T) {
	h := NewHub(nil)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go h.Run(ctx)

	client := &Client{
		extensionID: "ext-cap",
		send:        make(chan []byte, 1024),
		done:        make(chan struct{}),
	}
	h.RegisterSync(client)

	// Send/answer pairs well beyond the in-flight cap. Each correlated reply must
	// free a slot, so none of these should be rejected.
	for i := 0; i < maxInFlightPerExtension*3; i++ {
		msgID := fmt.Sprintf("cap-%d", i)

		// Register the waiter the way HubSender does, so the reply can be
		// correlated by id.
		resultCh := make(chan WSMessage, 1)
		h.RegisterResultChannel(msgID, resultCh)
		h.SetResultOwner(msgID, "ext-cap")

		if err := h.SendToExtension("ext-cap", WSMessage{ID: msgID, Type: MsgTypeCommand}); err != nil {
			t.Fatalf("send %d rejected despite replies freeing capacity: %v", i, err)
		}
		<-client.send // drain so the buffer cannot mask the in-flight state

		h.Route(client, WSMessage{ID: msgID, Type: MsgTypeResult})

		select {
		case <-resultCh:
		case <-time.After(time.Second):
			t.Fatalf("reply for %s was never correlated", msgID)
		}
		h.UnregisterResultChannel(msgID)
	}
}

// Unrelated inbound traffic must NOT release capacity, or a chatty extension
// could reset the counter and bypass the flood cap entirely.
func TestHub_UnrelatedTrafficDoesNotReleaseCapacity(t *testing.T) {
	h := NewHub(nil)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go h.Run(ctx)

	client := &Client{
		extensionID: "ext-noisy",
		send:        make(chan []byte, 4096),
		done:        make(chan struct{}),
	}
	h.RegisterSync(client)

	// Fill the cap with commands that will not be answered.
	for i := 0; i < maxInFlightPerExtension; i++ {
		if err := h.SendToExtension("ext-noisy", WSMessage{ID: fmt.Sprintf("c%d", i), Type: MsgTypeCommand}); err != nil {
			t.Fatalf("filling cap: %v", err)
		}
	}

	// Flush unrelated frames.
	for i := 0; i < 50; i++ {
		h.Route(client, WSMessage{ID: "noise", Type: "progress", Action: "update"})
	}
	time.Sleep(100 * time.Millisecond)

	if err := h.SendToExtension("ext-noisy", WSMessage{ID: "over", Type: MsgTypeCommand}); err == nil {
		t.Error("capacity was released by unrelated traffic, so the in-flight cap " +
			"can be bypassed by any extension that sends extra frames")
	}
}

func TestHub_StopIsIdempotent(t *testing.T) {
	h := NewHub(nil)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go h.Run(ctx)

	h.Stop()
	h.Stop() // second stop must not panic (stopOnce semantics)
}

func TestHub_RouteUpdatesLastSeenCallback(t *testing.T) {
	var mu sync.Mutex
	seen := 0

	h := NewHub(nil)
	h.OnActivity = func(extensionID string) {
		mu.Lock()
		defer mu.Unlock()
		seen++
	}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go h.Run(ctx)

	client := newTestConn("ext-act")
	h.RegisterSync(client)
	h.Route(client, WSMessage{ID: "a", Type: "result"})

	// The callback is invoked on the event loop; allow it to be observed.
	deadline := time.After(testSendTimeout)
	for {
		mu.Lock()
		n := seen
		mu.Unlock()
		if n > 0 {
			break
		}
		select {
		case <-deadline:
			t.Fatal("OnActivity was never invoked for a routed message")
		case <-time.After(10 * time.Millisecond):
		}
	}
}
