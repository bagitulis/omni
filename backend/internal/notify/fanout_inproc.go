package notify

import "sync"

// InProcessFanout is a simple in-memory Fanout used for dev, tests, and any
// deployment that runs a single API replica. It is intentionally minimal —
// bounded per-subscriber buffer (64), best-effort drop on overflow.
type InProcessFanout struct {
	mu   sync.RWMutex
	subs map[string][]chan Envelope
}

// NewInProcessFanout returns a fresh in-memory fanout.
func NewInProcessFanout() *InProcessFanout {
	return &InProcessFanout{subs: map[string][]chan Envelope{}}
}

// Subscribe returns a channel that will receive Envelopes for tenantID.
// Callers MUST Unsubscribe on shutdown; otherwise the channel is leaked.
func (f *InProcessFanout) Subscribe(tenantID string) <-chan Envelope {
	ch := make(chan Envelope, 64)
	f.mu.Lock()
	f.subs[tenantID] = append(f.subs[tenantID], ch)
	f.mu.Unlock()
	return ch
}

// Unsubscribe removes ch from tenantID's subscriber set and closes it.
// Safe to call twice; the second call is a no-op.
func (f *InProcessFanout) Unsubscribe(tenantID string, target <-chan Envelope) {
	f.mu.Lock()
	defer f.mu.Unlock()
	list := f.subs[tenantID]
	for i, ch := range list {
		if readOnly(ch) == target {
			// Close and remove.
			close(ch)
			f.subs[tenantID] = append(list[:i], list[i+1:]...)
			return
		}
	}
}

// Publish delivers env to every subscriber of tenantID. Non-blocking: if a
// subscriber's buffer is full the event is dropped for that subscriber (the
// reconnect path will backfill via Last-Event-ID replay).
func (f *InProcessFanout) Publish(tenantID string, env Envelope) {
	f.mu.RLock()
	// Snapshot list so we can iterate without holding the lock while sending.
	list := f.subs[tenantID]
	snapshot := make([]chan Envelope, len(list))
	copy(snapshot, list)
	f.mu.RUnlock()

	for _, ch := range snapshot {
		f.trySend(ch, env)
	}
}

// trySend guards against sending on a closed channel (Unsubscribe race).
func (f *InProcessFanout) trySend(ch chan Envelope, env Envelope) {
	defer func() {
		// Recover from "send on closed channel" if Unsubscribe raced us.
		_ = recover()
	}()
	select {
	case ch <- env:
	default:
		// Buffer full — drop.
	}
}

// readOnly is a tiny helper to convert a bi-directional chan to a receive-only
// so the identity comparison in Unsubscribe works against the value handed to
// Subscribe callers.
func readOnly(ch chan Envelope) <-chan Envelope { return ch }
