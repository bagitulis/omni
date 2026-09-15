package realtime

import (
	"sync"
	"sync/atomic"
)

// Publisher is the service-layer entry point for pushing updates to
// subscribed dashboard clients. Any code that needs to notify the frontend
// (order upserts, inventory changes, background job progress) calls a method
// on the singleton `publisher` from `Get()`.
//
// Design:
//   - Publisher owns a reference to the hub, not a copy of its channels.
//     If the hub is nil (e.g. WebSocket disabled), every publish becomes a
//     no-op and is counted in `SkippedNoHub` so tests can observe it.
//   - Zero-arg fast path: constructing a Publisher without a hub is legal;
//     the service layer can safely hold a `*realtime.Publisher` from
//     process start even before Wire() has been called.
type Publisher struct {
	hub          *Hub
	mu           sync.RWMutex
	skippedNoHub atomic.Int64
	published    atomic.Int64
}

// Topic names — kept as consts so publisher and subscriber never drift.
const (
	TopicOrdersUpdated    = "orders/updated"
	TopicInventoryUpdated = "inventory/updated"
	TopicSyncProgress     = "sync/progress"
	TopicChatHint         = "chat/hint"
	TopicNotifications    = "notifications/updated"
)

// NewPublisher constructs an unwired publisher. Use Wire() to attach a hub.
func NewPublisher() *Publisher {
	return &Publisher{}
}

// Wire attaches (or replaces) the hub that this publisher will fan out to.
// Passing nil detaches — subsequent publishes become no-ops.
func (p *Publisher) Wire(h *Hub) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.hub = h
}

// PublishOrderUpdated notifies subscribers of tenant `tenantID` that an
// order changed. Payload is opaque to the hub; typical shape:
//
//	{"order_sn":"...","platform":"shopee","status":"SHIPPED","fields":["status","tracking_number"]}
func (p *Publisher) PublishOrderUpdated(tenantID string, payload any) {
	p.publish(tenantID, TopicOrdersUpdated, payload)
}

// PublishInventoryUpdated notifies subscribers of tenant `tenantID` that a
// SKU stock/price changed.
func (p *Publisher) PublishInventoryUpdated(tenantID string, payload any) {
	p.publish(tenantID, TopicInventoryUpdated, payload)
}

// PublishSyncProgress notifies subscribers of a background sync job's
// progress. Typical payload: {"job_id","percent","message"}.
func (p *Publisher) PublishSyncProgress(tenantID string, payload any) {
	p.publish(tenantID, TopicSyncProgress, payload)
}

// PublishChatHint fans a lightweight chat notification (typing, new message
// pending) to the dashboard. Not for message bodies.
func (p *Publisher) PublishChatHint(tenantID string, payload any) {
	p.publish(tenantID, TopicChatHint, payload)
}

// PublishNotification fans a notification-updated event so the bell / list
// can invalidate without polling. Payload usually is just `{"unread_count":N}`.
func (p *Publisher) PublishNotification(tenantID string, payload any) {
	p.publish(tenantID, TopicNotifications, payload)
}

func (p *Publisher) publish(tenantID, topic string, payload any) {
	if tenantID == "" || topic == "" {
		return
	}
	p.mu.RLock()
	h := p.hub
	p.mu.RUnlock()
	if h == nil {
		p.skippedNoHub.Add(1)
		return
	}
	env := Envelope{Payload: Marshal(payload)}
	h.Publish(tenantID, topic, env)
	p.published.Add(1)
}

// Stats returns internal counters. Test-only; production reads these via
// Prometheus later.
func (p *Publisher) Stats() (published, skipped int64) {
	return p.published.Load(), p.skippedNoHub.Load()
}

// ---- Process-wide singleton ----

var (
	singleton     *Publisher
	singletonOnce sync.Once
)

// Get returns the process-wide Publisher, creating it once.
func Get() *Publisher {
	singletonOnce.Do(func() {
		singleton = NewPublisher()
	})
	return singleton
}

// ResetForTests replaces the singleton with a fresh, unwired instance.
// Test-only; do not call from production code.
func ResetForTests() {
	singleton = NewPublisher()
}
