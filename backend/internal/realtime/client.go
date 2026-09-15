package realtime

import "sync"

// Client is one WebSocket connection tied to a tenant + user.
//
// The client owns two channels: sendCh (server → client), and an internal
// closeCh that the hub uses to signal "we are done with you, stop your read
// pump". A read pump inside the connection handler pushes decoded envelopes
// back into the hub via Subscribe/Unsubscribe/Publish; the client struct
// itself never touches the network directly — that lives in conn.go so the
// hub is testable without a real socket.
type Client struct {
	ID       string
	TenantID string
	UserID   string
	Role     string

	sendCh    chan Envelope
	closeCh   chan struct{}
	closeOnce sync.Once
}

// NewClient constructs a Client with a bounded send buffer. bufSize must be
// positive; the hub uses 64 in production.
func NewClient(id, tenantID, userID, role string, bufSize int) *Client {
	if bufSize <= 0 {
		bufSize = 1
	}
	return &Client{
		ID:       id,
		TenantID: tenantID,
		UserID:   userID,
		Role:     role,
		sendCh:   make(chan Envelope, bufSize),
		closeCh:  make(chan struct{}),
	}
}

// Send returns the receive-only channel a write pump ranges over.
func (c *Client) Send() <-chan Envelope { return c.sendCh }

// Closed returns a channel that is closed once the hub has evicted this
// client. Write / read pumps must select on it and exit.
func (c *Client) Closed() <-chan struct{} { return c.closeCh }

// close is called by the hub during unregister / drain. Idempotent.
func (c *Client) close() {
	c.closeOnce.Do(func() { close(c.closeCh) })
}
