package realtime

import (
	"context"
	"encoding/json"
	"testing"
	"time"
)

// TestPublisher_NoHubIsNoop is the fast-path safety net: services can hold a
// Publisher and start publishing before Wire() has attached a hub. The call
// must not panic and must not deliver.
func TestPublisher_NoHubIsNoop(t *testing.T) {
	p := NewPublisher()
	p.PublishOrderUpdated("tenant-a", map[string]string{"order_sn": "X"})
	pub, skip := p.Stats()
	if pub != 0 {
		t.Errorf("published = %d, want 0 (no hub wired)", pub)
	}
	if skip != 1 {
		t.Errorf("skipped = %d, want 1", skip)
	}
}

// TestPublisher_WireThenPublishReachesSubscriber — the happy path.
func TestPublisher_WireThenPublishReachesSubscriber(t *testing.T) {
	h := NewHub()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go h.Run(ctx)
	defer h.Stop()

	c := NewClient("c", "tenant-a", "u", "user", 4)
	h.RegisterSync(c)
	h.SubscribeSync(c, TopicOrdersUpdated)

	p := NewPublisher()
	p.Wire(h)
	p.PublishOrderUpdated("tenant-a", map[string]string{
		"order_sn": "ORD-1",
		"platform": "shopee",
	})

	select {
	case env := <-c.Send():
		if env.Topic != TopicOrdersUpdated {
			t.Errorf("Topic = %q, want %q", env.Topic, TopicOrdersUpdated)
		}
		if env.TenantID != "tenant-a" {
			t.Errorf("TenantID = %q, want tenant-a", env.TenantID)
		}
		var payload map[string]string
		if err := json.Unmarshal(env.Payload, &payload); err != nil {
			t.Fatalf("payload unmarshal: %v", err)
		}
		if payload["order_sn"] != "ORD-1" {
			t.Errorf("payload order_sn = %q, want ORD-1", payload["order_sn"])
		}
	case <-time.After(1 * time.Second):
		t.Fatal("subscriber did not receive order update")
	}

	pub, skip := p.Stats()
	if pub != 1 || skip != 0 {
		t.Errorf("Stats = (%d,%d), want (1,0)", pub, skip)
	}
}

// TestPublisher_EmptyTenantIsIgnored — negative path. Publishing with a
// missing tenant would leak to no one (topic-first fan-out never matches),
// but explicitly guard against it so a caller-side bug is easy to detect.
func TestPublisher_EmptyTenantIsIgnored(t *testing.T) {
	h := NewHub()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go h.Run(ctx)
	defer h.Stop()

	p := NewPublisher()
	p.Wire(h)
	p.PublishOrderUpdated("", map[string]int{"x": 1})
	pub, skip := p.Stats()
	if pub != 0 {
		t.Errorf("published = %d, want 0 for empty tenant", pub)
	}
	if skip != 0 {
		t.Errorf("skipped = %d, want 0 for empty tenant", skip)
	}
}

// TestPublisher_CrossTenantIsolationHolds — publish for tenant-a must NOT
// arrive at a subscriber from tenant-b. This is a redundant check of the
// hub's invariant through the Publisher path.
func TestPublisher_CrossTenantIsolationHolds(t *testing.T) {
	h := NewHub()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go h.Run(ctx)
	defer h.Stop()

	a := NewClient("a", "tenant-a", "u", "user", 4)
	b := NewClient("b", "tenant-b", "u", "user", 4)
	h.RegisterSync(a)
	h.RegisterSync(b)
	h.SubscribeSync(a, TopicOrdersUpdated)
	h.SubscribeSync(b, TopicOrdersUpdated)

	p := NewPublisher()
	p.Wire(h)
	p.PublishOrderUpdated("tenant-a", map[string]string{"order_sn": "leak-check"})

	// a must receive.
	select {
	case env := <-a.Send():
		if env.TenantID != "tenant-a" {
			t.Errorf("a got wrong tenant: %q", env.TenantID)
		}
	case <-time.After(1 * time.Second):
		t.Fatal("a did not receive its own tenant's publish")
	}
	// b must NOT.
	select {
	case env := <-b.Send():
		t.Fatalf("b received cross-tenant leak via publisher: %+v", env)
	case <-time.After(300 * time.Millisecond):
	}
}

// TestPublisher_SingletonGetIsSingleInstance — Get() returns the same instance.
func TestPublisher_SingletonGetIsSingleInstance(t *testing.T) {
	ResetForTests()
	a := Get()
	b := Get()
	if a != b {
		t.Errorf("Get() returned different instances")
	}
}

// TestPublisher_AllTopicsSameShape — smoke that each PublishX method
// routes to a distinct topic without panicking.
func TestPublisher_AllTopicsSameShape(t *testing.T) {
	h := NewHub()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go h.Run(ctx)
	defer h.Stop()

	p := NewPublisher()
	p.Wire(h)

	// Subscribe one client per topic so each publish reaches exactly one
	// receiver — proves the topic name is the exact one the publisher uses.
	topics := []struct {
		name string
		fn   func()
	}{
		{TopicOrdersUpdated, func() { p.PublishOrderUpdated("t", "x") }},
		{TopicInventoryUpdated, func() { p.PublishInventoryUpdated("t", "x") }},
		{TopicSyncProgress, func() { p.PublishSyncProgress("t", "x") }},
		{TopicChatHint, func() { p.PublishChatHint("t", "x") }},
		{TopicNotifications, func() { p.PublishNotification("t", "x") }},
	}
	for i, tc := range topics {
		c := NewClient(tc.name, "t", "u", "user", 4)
		h.RegisterSync(c)
		h.SubscribeSync(c, tc.name)
		tc.fn()
		select {
		case env := <-c.Send():
			if env.Topic != tc.name {
				t.Errorf("topic[%d] delivered on %q, want %q", i, env.Topic, tc.name)
			}
		case <-time.After(500 * time.Millisecond):
			t.Errorf("topic[%d] %q did not deliver", i, tc.name)
		}
	}
}
