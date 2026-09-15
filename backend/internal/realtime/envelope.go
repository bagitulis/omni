package realtime

import "encoding/json"

// Package realtime — dashboard-facing WebSocket hub (Phase 4 of the roadmap
// in docs/superpowers/specs/2026-09-15-realtime-e2e-platform-drift-design.md).
//
// This is a separate hub from `internal/extensions/`. The extensions hub is
// keyed by extension_id and models command-and-reply RPC for a browser
// extension; its `BroadcastToDashboards` is an explicit no-op. This package
// is JWT-authed, keyed by (tenant_id, client_id), and supports topic
// subscribe + fan-out — the shape dashboards need for live order / inventory
// / sync updates.
//
// Phase 4 delivers: hub, per-client, topic registry, JWT auth, and a single
// mount point (RealtimePath). Phase 5 wires publishers in the service layer.

// Envelope is the wire format for every WebSocket frame in either direction.
// All fields are optional except Type; the meaning of each field depends on
// Type.
type Envelope struct {
	// Type is one of: "auth" | "subscribe" | "unsubscribe" | "publish" | "ack" | "error" | "ping" | "pong".
	Type string `json:"type"`

	// Topic is used for subscribe / unsubscribe / publish frames.
	Topic string `json:"topic,omitempty"`

	// Token is the JWT presented in the first frame (Type == "auth"). Never
	// echoed back to the client.
	Token string `json:"token,omitempty"`

	// TenantID is set by the server on outbound envelopes so a multi-tenant
	// consumer can double-check. Never trusted from the client.
	TenantID string `json:"tenant_id,omitempty"`

	// ID is a client-generated correlation ID; the server echoes it in ack
	// or error responses. Optional.
	ID string `json:"id,omitempty"`

	// Payload carries topic-specific data on publish. Opaque bytes so the
	// hub does not need to know the shape.
	Payload json.RawMessage `json:"payload,omitempty"`

	// Error carries a human-readable message on Type == "error".
	Error string `json:"error,omitempty"`
}

// Frame type constants.
const (
	TypeAuth        = "auth"
	TypeAck         = "ack"
	TypeError       = "error"
	TypeSubscribe   = "subscribe"
	TypeUnsubscribe = "unsubscribe"
	TypePublish     = "publish"
	TypePing        = "ping"
	TypePong        = "pong"
)
