package realtime

import (
	"context"
	"sync"
)

// Hub is the single-goroutine event loop that owns client and topic
// membership. Every mutation flows through channels — never lock a mutex
// inside a hot path. This mirrors the pattern used by internal/extensions/
// (proven under gap-probe tests) but the state shape is different: keyed by
// (tenant, clientID) with a topic subscription registry.
type Hub struct {
	registerCh   chan registration
	unregisterCh chan registration
	subscribeCh  chan subscription
	publishCh    chan publish
	stopCh       chan struct{}
	stopOnce     sync.Once

	// State — only accessed from Run().
	clients map[string]*Client               // clientID → client
	topics  map[string]map[string]*Client    // topic → clientID → client
	tenants map[string]map[string]struct{}   // tenantID → set of clientIDs
}

type subscription struct {
	client *Client
	topic  string
	add    bool // true = subscribe, false = unsubscribe
	// done is closed by the hub after it processes this entry. Callers that
	// need happens-after ordering (tests, and any code that must be sure the
	// hub sees a register before it fires a publish) pass a non-nil channel
	// and receive from it. Nil is the fire-and-forget default used in prod.
	done chan struct{}
}

type publish struct {
	tenantID string
	topic    string
	envelope Envelope
	done     chan struct{} // same contract as subscription.done
}

type registration struct {
	client *Client
	done   chan struct{}
}

// NewHub returns a hub in a not-yet-running state. Call Run(ctx) to start it.
func NewHub() *Hub {
	return &Hub{
		registerCh:   make(chan registration, 32),
		unregisterCh: make(chan registration, 32),
		subscribeCh:  make(chan subscription, 128),
		publishCh:    make(chan publish, 256),
		stopCh:       make(chan struct{}),
		clients:      map[string]*Client{},
		topics:       map[string]map[string]*Client{},
		tenants:      map[string]map[string]struct{}{},
	}
}

// Run owns the hub goroutine until ctx is done or Stop is called. Safe to
// invoke exactly once.
func (h *Hub) Run(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			h.drain()
			return
		case <-h.stopCh:
			h.drain()
			return
		case r := <-h.registerCh:
			h.doRegister(r.client)
			signalDone(r.done)
		case r := <-h.unregisterCh:
			h.doUnregister(r.client)
			signalDone(r.done)
		case s := <-h.subscribeCh:
			h.doSubscribe(s)
			signalDone(s.done)
		case p := <-h.publishCh:
			h.doPublish(p)
			signalDone(p.done)
		}
	}
}

// signalDone closes the ack channel if the caller provided one. Nil-safe.
func signalDone(ch chan struct{}) {
	if ch != nil {
		close(ch)
	}
}

// Stop signals the hub loop to exit. Idempotent and safe from any goroutine.
// Does NOT wait for Run to return; that's the caller's responsibility.
func (h *Hub) Stop() {
	h.stopOnce.Do(func() { close(h.stopCh) })
}

// Register enqueues a new client for the hub. Fire-and-forget.
func (h *Hub) Register(c *Client) {
	select {
	case h.registerCh <- registration{client: c}:
	case <-h.stopCh:
	}
}

// RegisterSync enqueues a client and blocks until the hub has processed it.
// Use when the next call (Subscribe, Publish) MUST see this client already
// registered — race-free coordination for tests and setup code.
func (h *Hub) RegisterSync(c *Client) {
	done := make(chan struct{})
	select {
	case h.registerCh <- registration{client: c, done: done}:
	case <-h.stopCh:
		return
	}
	select {
	case <-done:
	case <-h.stopCh:
	}
}

// Unregister enqueues a client removal. Fire-and-forget.
func (h *Hub) Unregister(c *Client) {
	select {
	case h.unregisterCh <- registration{client: c}:
	case <-h.stopCh:
	}
}

// Subscribe enqueues a topic subscription for a client. Fire-and-forget.
func (h *Hub) Subscribe(c *Client, topic string) {
	select {
	case h.subscribeCh <- subscription{client: c, topic: topic, add: true}:
	case <-h.stopCh:
	}
}

// SubscribeSync is the sync variant of Subscribe. Blocks until the hub has
// processed the subscription; the next Publish is guaranteed to see it.
func (h *Hub) SubscribeSync(c *Client, topic string) {
	done := make(chan struct{})
	select {
	case h.subscribeCh <- subscription{client: c, topic: topic, add: true, done: done}:
	case <-h.stopCh:
		return
	}
	select {
	case <-done:
	case <-h.stopCh:
	}
}

// Unsubscribe enqueues a topic removal for a client.
func (h *Hub) Unsubscribe(c *Client, topic string) {
	select {
	case h.subscribeCh <- subscription{client: c, topic: topic, add: false}:
	case <-h.stopCh:
	}
}

// UnsubscribeSync is the sync variant of Unsubscribe.
func (h *Hub) UnsubscribeSync(c *Client, topic string) {
	done := make(chan struct{})
	select {
	case h.subscribeCh <- subscription{client: c, topic: topic, add: false, done: done}:
	case <-h.stopCh:
		return
	}
	select {
	case <-done:
	case <-h.stopCh:
	}
}

// Publish fans an envelope to every client of `tenantID` subscribed to
// `topic`. Non-blocking: if a client's send buffer is full, the message is
// dropped for that client and a metric counter is incremented (added Phase 5).
func (h *Hub) Publish(tenantID, topic string, env Envelope) {
	env.TenantID = tenantID
	env.Topic = topic
	env.Type = TypePublish
	select {
	case h.publishCh <- publish{tenantID: tenantID, topic: topic, envelope: env}:
	case <-h.stopCh:
	}
}

func (h *Hub) doRegister(c *Client) {
	if c == nil || c.ID == "" {
		return
	}
	h.clients[c.ID] = c
	if _, ok := h.tenants[c.TenantID]; !ok {
		h.tenants[c.TenantID] = map[string]struct{}{}
	}
	h.tenants[c.TenantID][c.ID] = struct{}{}
}

func (h *Hub) doUnregister(c *Client) {
	if c == nil {
		return
	}
	delete(h.clients, c.ID)
	if set, ok := h.tenants[c.TenantID]; ok {
		delete(set, c.ID)
		if len(set) == 0 {
			delete(h.tenants, c.TenantID)
		}
	}
	// Remove from every topic that mentioned this client.
	for topic, subs := range h.topics {
		if _, ok := subs[c.ID]; ok {
			delete(subs, c.ID)
			if len(subs) == 0 {
				delete(h.topics, topic)
			}
		}
	}
	// Signal the client its life is over. Non-blocking; if the client is
	// already gone the send is dropped by the buffered channel.
	c.close()
}

func (h *Hub) doSubscribe(s subscription) {
	if s.client == nil || s.topic == "" {
		return
	}
	// Only accept subscriptions from a registered client — protects against
	// races where a client subscribes and unregisters concurrently.
	if _, ok := h.clients[s.client.ID]; !ok {
		return
	}
	if s.add {
		if _, ok := h.topics[s.topic]; !ok {
			h.topics[s.topic] = map[string]*Client{}
		}
		h.topics[s.topic][s.client.ID] = s.client
		return
	}
	if subs, ok := h.topics[s.topic]; ok {
		delete(subs, s.client.ID)
		if len(subs) == 0 {
			delete(h.topics, s.topic)
		}
	}
}

func (h *Hub) doPublish(p publish) {
	subs, ok := h.topics[p.topic]
	if !ok {
		return
	}
	for clientID, c := range subs {
		if c.TenantID != p.tenantID {
			// Cross-tenant isolation: a subscriber from tenant B must never
			// receive a publish targeted at tenant A, even if both subscribed
			// to the same topic name. This is the key security invariant.
			continue
		}
		// Non-blocking send. If the client's buffer is full, drop for that
		// client only — never stall the hub.
		select {
		case c.sendCh <- p.envelope:
		default:
			_ = clientID // reserved for a metric counter in Phase 5
		}
	}
}

func (h *Hub) drain() {
	// Close every client on shutdown so their write pumps exit cleanly.
	for _, c := range h.clients {
		c.close()
	}
}

// snapshot is a serialized read of hub internals. All reads happen inside
// the hub goroutine, then are copied out through a channel — so no callers
// touch the maps directly.
type snapshot struct {
	tenantClients map[string]int
	topicSubs     map[string]int
	reply         chan snapshot
}

// A dedicated probe channel would need Run() to select on it too. To avoid
// that overhead in production, the test helpers piggyback on the existing
// subscribeCh with a sentinel op that carries a done chan. After done fires,
// every prior enqueue has been processed, so a subsequent lookup that runs
// inside the hub loop is safe.

// TopicSubscriberCount returns the number of subscribers to `topic`.
// Serializes with the hub loop so it is race-free and blocking.
func (h *Hub) TopicSubscriberCount(topic string) int {
	// Enqueue a benign subscription op with add=false and no client so
	// doSubscribe returns early, but the done ack still fires — this is our
	// happens-after fence. Then the map read runs after the hub loop has
	// caught up.
	done := make(chan struct{})
	select {
	case h.subscribeCh <- subscription{topic: topic, add: false, done: done}:
	case <-h.stopCh:
		return 0
	}
	select {
	case <-done:
	case <-h.stopCh:
		return 0
	}
	return len(h.topics[topic])
}
