package notify

import "encoding/json"

// Envelope is the payload shape carried on the fanout channel. Keeping it a
// raw JSON blob lets the transport layer (WS/SSE) forward without re-marshal.
type Envelope struct {
	Payload json.RawMessage
}

// Fanout abstracts the "publish a tenant-scoped event to any listening
// replica" concern. Two implementations exist:
//   - InProcessFanout for dev / single-replica.
//   - RedisFanout (stub in this repo — activated when REDIS_URL is set).
type Fanout interface {
	Publish(tenantID string, env Envelope)
	Subscribe(tenantID string) <-chan Envelope
	Unsubscribe(tenantID string, ch <-chan Envelope)
}
