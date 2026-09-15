package realtime

import (
	"context"
	"encoding/json"
	"testing"
	"time"
)

// newRunningHub starts a hub in the background and returns a cleanup that
// stops it. Uses a bounded timeout so a hung test cannot leak the goroutine.
func newRunningHub(t *testing.T) (*Hub, func()) {
	t.Helper()
	h := NewHub()
	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan struct{})
	go func() {
		h.Run(ctx)
		close(done)
	}()
	return h, func() {
		h.Stop()
		cancel()
		select {
		case <-done:
		case <-time.After(2 * time.Second):
			t.Fatalf("hub Run did not exit within 2s")
		}
	}
}

// TestHub_PublishRoutesToSubscribedClient — happy path. Uses RegisterSync /
// SubscribeSync so the publish is guaranteed to arrive after subscribe.
func TestHub_PublishRoutesToSubscribedClient(t *testing.T) {
	h, cleanup := newRunningHub(t)
	defer cleanup()

	c := NewClient("client-1", "tenant-a", "user-1", "admin", 4)
	h.RegisterSync(c)
	h.SubscribeSync(c, "orders/updated")

	h.Publish("tenant-a", "orders/updated", Envelope{
		Payload: json.RawMessage(`{"order_sn":"ABC123"}`),
	})

	select {
	case env := <-c.Send():
		if env.Topic != "orders/updated" {
			t.Errorf("Topic = %q, want orders/updated", env.Topic)
		}
		if env.TenantID != "tenant-a" {
			t.Errorf("TenantID = %q, want tenant-a", env.TenantID)
		}
		if env.Type != TypePublish {
			t.Errorf("Type = %q, want %q", env.Type, TypePublish)
		}
		if string(env.Payload) != `{"order_sn":"ABC123"}` {
			t.Errorf("Payload = %s, want order_sn=ABC123", string(env.Payload))
		}
	case <-time.After(1 * time.Second):
		t.Fatal("did not receive publish within 1s")
	}
}

// TestHub_CrossTenantIsolation — the security invariant.
func TestHub_CrossTenantIsolation(t *testing.T) {
	h, cleanup := newRunningHub(t)
	defer cleanup()

	a := NewClient("a", "tenant-a", "u-a", "user", 4)
	b := NewClient("b", "tenant-b", "u-b", "user", 4)
	h.RegisterSync(a)
	h.RegisterSync(b)
	h.SubscribeSync(a, "orders/updated")
	h.SubscribeSync(b, "orders/updated")

	h.Publish("tenant-a", "orders/updated", Envelope{
		Payload: json.RawMessage(`{"secret":"tenant-a-only"}`),
	})

	// Expect a to receive.
	select {
	case env := <-a.Send():
		if env.TenantID != "tenant-a" {
			t.Errorf("a received wrong tenant: %q", env.TenantID)
		}
	case <-time.After(1 * time.Second):
		t.Fatal("a did not receive tenant-a publish")
	}

	// Expect b to NOT receive within a reasonable window.
	select {
	case env := <-b.Send():
		t.Fatalf("b received cross-tenant leak: %+v", env)
	case <-time.After(300 * time.Millisecond):
		// good — no delivery
	}
}

// TestHub_PublishToNoSubscribers is a no-op that must not panic.
func TestHub_PublishToNoSubscribers(t *testing.T) {
	h, cleanup := newRunningHub(t)
	defer cleanup()

	h.Publish("tenant-a", "orders/updated", Envelope{Payload: json.RawMessage(`{}`)})
	if h.TopicSubscriberCount("orders/updated") != 0 {
		t.Errorf("expected zero subscribers on a topic nobody subscribed to")
	}
}

// TestHub_UnsubscribeStopsDelivery.
func TestHub_UnsubscribeStopsDelivery(t *testing.T) {
	h, cleanup := newRunningHub(t)
	defer cleanup()

	c := NewClient("c", "tenant-a", "u", "user", 4)
	h.RegisterSync(c)
	h.SubscribeSync(c, "orders/updated")
	h.UnsubscribeSync(c, "orders/updated")

	h.Publish("tenant-a", "orders/updated", Envelope{Payload: json.RawMessage(`{}`)})
	select {
	case env := <-c.Send():
		t.Fatalf("received after unsubscribe: %+v", env)
	case <-time.After(300 * time.Millisecond):
		// good
	}
}

// TestHub_UnregisterEvictsAndClosesClient.
func TestHub_UnregisterEvictsAndClosesClient(t *testing.T) {
	h, cleanup := newRunningHub(t)
	defer cleanup()

	c := NewClient("c", "tenant-a", "u", "user", 4)
	h.RegisterSync(c)
	h.SubscribeSync(c, "orders/updated")

	// Use fire-and-forget Unregister + wait on Closed(), which the hub
	// closes as part of processing the unregister op.
	h.Unregister(c)

	select {
	case <-c.Closed():
		// good — hub signalled close
	case <-time.After(1 * time.Second):
		t.Fatal("client Closed() channel not closed after unregister")
	}

	// Subsequent publish must not deliver.
	h.Publish("tenant-a", "orders/updated", Envelope{Payload: json.RawMessage(`{}`)})
	select {
	case env := <-c.Send():
		t.Fatalf("received after unregister: %+v", env)
	case <-time.After(300 * time.Millisecond):
	}
}

// TestHub_SendBufferFullDropsWithoutStalling — noisy client must not stall
// the hub for others.
func TestHub_SendBufferFullDropsWithoutStalling(t *testing.T) {
	h, cleanup := newRunningHub(t)
	defer cleanup()

	// Small buffer so we can overflow easily.
	slow := NewClient("slow", "tenant-a", "u", "user", 1)
	fast := NewClient("fast", "tenant-a", "u", "user", 4)
	h.RegisterSync(slow)
	h.RegisterSync(fast)
	h.SubscribeSync(slow, "orders/updated")
	h.SubscribeSync(fast, "orders/updated")

	// Fire 20 publishes; slow's buffer is 1 so ~19 will drop for it.
	// fast's buffer is 4 — some may drop too but the hub must not deadlock.
	for i := 0; i < 20; i++ {
		h.Publish("tenant-a", "orders/updated", Envelope{Payload: json.RawMessage(`{"i":0}`)})
	}

	// Fence: wait until every publish has been processed. A dummy
	// SubscribeSync on a throwaway topic serves as a happens-after barrier
	// because the hub processes ops in the order they appear on each channel,
	// and Go's select is fair enough that alternating channels drain.
	// Using SubscribeSync here is deterministic; a Sleep would not be.
	throwaway := NewClient("throw", "tenant-a", "u", "user", 1)
	h.RegisterSync(throwaway)
	h.SubscribeSync(throwaway, "__fence__")

	// slow drained partially; expect at least 1 message on its buffer.
	drainedSlow := 0
DrainSlow:
	for {
		select {
		case <-slow.Send():
			drainedSlow++
		default:
			break DrainSlow
		}
	}
	if drainedSlow == 0 {
		t.Errorf("slow client received nothing; expected at least one delivery")
	}

	drainedFast := 0
DrainFast:
	for {
		select {
		case <-fast.Send():
			drainedFast++
		default:
			break DrainFast
		}
	}
	if drainedFast == 0 {
		t.Errorf("fast client received nothing; hub appears to have stalled on slow")
	}
}

// TestHub_StopIsIdempotentAndSafeBeforeRun — negative path.
func TestHub_StopIsIdempotentAndSafeBeforeRun(t *testing.T) {
	h := NewHub()
	// Stop twice before Run: must not panic.
	h.Stop()
	h.Stop()
}

// TestHub_StopReleasesRun — the loop must exit within a bounded time.
func TestHub_StopReleasesRun(t *testing.T) {
	h := NewHub()
	done := make(chan struct{})
	go func() { h.Run(context.Background()); close(done) }()
	// Small sleep to make sure Run has entered its select — pure timing here
	// is unavoidable because the hub does not expose a "started" signal.
	time.Sleep(20 * time.Millisecond)
	h.Stop()
	select {
	case <-done:
	case <-time.After(500 * time.Millisecond):
		t.Fatal("Run did not exit within 500ms after Stop")
	}
}

// TestHub_SubscribeFromUnregisteredClientIgnored — negative path.
func TestHub_SubscribeFromUnregisteredClientIgnored(t *testing.T) {
	h, cleanup := newRunningHub(t)
	defer cleanup()
	stray := NewClient("stray", "tenant-a", "u", "user", 4)
	// NOT registered.
	h.SubscribeSync(stray, "orders/updated")
	if got := h.TopicSubscriberCount("orders/updated"); got != 0 {
		t.Errorf("topics has stray subscriber: got %d, want 0", got)
	}
}

// TestHub_TopicSubscriberCount_HappyPath — sanity check for the test helper.
func TestHub_TopicSubscriberCount_HappyPath(t *testing.T) {
	h, cleanup := newRunningHub(t)
	defer cleanup()
	c1 := NewClient("c1", "tenant-a", "u", "user", 4)
	c2 := NewClient("c2", "tenant-a", "u", "user", 4)
	h.RegisterSync(c1)
	h.RegisterSync(c2)
	h.SubscribeSync(c1, "topic-x")
	h.SubscribeSync(c2, "topic-x")
	if got := h.TopicSubscriberCount("topic-x"); got != 2 {
		t.Errorf("TopicSubscriberCount = %d, want 2", got)
	}
	if got := h.TopicSubscriberCount("nope"); got != 0 {
		t.Errorf("TopicSubscriberCount(nope) = %d, want 0", got)
	}
}
