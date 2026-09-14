package extensions

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/gorilla/websocket"
)

// Connection tuning.
const (
	// writeWait bounds a single write; a peer that cannot accept a frame this
	// fast is treated as wedged rather than allowed to block the writer.
	writeWait = 10 * time.Second

	// pongWait is how long we tolerate silence before considering the peer gone.
	// The extension pings more often than this; a missed window means a dead
	// socket that TCP has not yet reported.
	pongWait = 60 * time.Second

	// pingPeriod must be shorter than pongWait so a ping always precedes the
	// deadline it refreshes.
	pingPeriod = 30 * time.Second

	// maxMessageSize caps an inbound frame. Scraped payloads are chunked by the
	// extension, so a single frame stays small; this exists to bound memory.
	maxMessageSize = 1 << 20 // 1 MiB

	// handshakeTimeout bounds the upgrade itself.
	handshakeTimeout = 10 * time.Second
)

// upgrader is configured for extensions (chrome-extension:// origins) and
// dashboards served from localhost or the deployed domain.
//
// CheckOrigin matters: without a check, gorilla accepts any origin, which would
// let a web page in another tab open an authenticated socket by riding the
// user's cookies. Extensions do not send cookies, but the token is still a
// bearer credential, so origins are restricted rather than left open.
func newUpgrader(allowedOrigins []string) *websocket.Upgrader {
	return &websocket.Upgrader{
		HandshakeTimeout: handshakeTimeout,
		ReadBufferSize:   4096,
		WriteBufferSize:  4096,
		CheckOrigin: func(r *http.Request) bool {
			origin := r.Header.Get("Origin")
			// Non-browser clients (tests, CLI) send no Origin. Allow that: they
			// cannot be driven by a hostile web page.
			if origin == "" {
				return true
			}
			if strings.HasPrefix(origin, "chrome-extension://") {
				return true
			}
			for _, allowed := range allowedOrigins {
				if strings.EqualFold(origin, allowed) {
					return true
				}
			}
			return false
		},
	}
}

// Connection owns one extension's WebSocket lifecycle.
type Connection struct {
	conn        *websocket.Conn
	client      *Client
	hub         *Hub
	extensionID string
}

// ConnConfig supplies the dependencies a connection needs.
type ConnConfig struct {
	Hub            *Hub
	AllowedOrigins []string
	// Authenticate resolves an inbound handshake to a tenant-bound identity.
	// Returning an error closes the socket without registering it.
	Authenticate func(ctx context.Context, req ConnectRequest) (*ConnectionIdentity, error)
	// OnDisconnect is called once when the socket closes.
	OnDisconnect func(extensionID string)
}

// ConnectionIdentity is the server-side result of authenticating a handshake.
//
// TenantID comes from the token lookup, never from the client, which is what
// keeps one tenant's extension from acting on another's data.
type ConnectionIdentity struct {
	ExtensionID string
	TenantID    string
	UserID      string
	DB          any // tenant-scoped *gorm.DB, opaque to this package
}

// ServeUpgrade handles the HTTP→WebSocket upgrade for an extension.
//
// The first message MUST be an auth/connect frame within the handshake window.
// Authenticating on a message rather than a query parameter keeps the token out
// of URLs, where it would be captured by access logs and proxies.
func (cfg ConnConfig) ServeUpgrade(w http.ResponseWriter, r *http.Request) {
	if cfg.Hub == nil || cfg.Authenticate == nil {
		http.Error(w, "extensions: not configured", http.StatusServiceUnavailable)
		return
	}

	up := newUpgrader(cfg.AllowedOrigins)
	conn, err := up.Upgrade(w, r, nil)
	if err != nil {
		// Upgrade already wrote a response.
		log.Printf("extensions: upgrade failed: %v", err)
		return
	}

	conn.SetReadLimit(maxMessageSize)
	_ = conn.SetReadDeadline(time.Now().Add(pongWait))

	// Expect the connect frame promptly; an idle socket must not hold a
	// connection slot open.
	var req ConnectRequest
	if err := conn.ReadJSON(&req); err != nil {
		_ = conn.WriteControl(websocket.CloseMessage,
			websocket.FormatCloseMessage(websocket.ClosePolicyViolation, "expected connect frame"),
			time.Now().Add(writeWait))
		_ = conn.Close()
		return
	}

	identity, err := cfg.Authenticate(r.Context(), req)
	if err != nil {
		// Deliberately generic: distinguishing "unknown token" from "expired"
		// would let a caller probe for valid tokens.
		_ = conn.WriteControl(websocket.CloseMessage,
			websocket.FormatCloseMessage(websocket.ClosePolicyViolation, "unauthorized"),
			time.Now().Add(writeWait))
		_ = conn.Close()
		log.Printf("extensions: handshake rejected: %v", err)
		return
	}

	if err := ValidateExtensionID(identity.ExtensionID); err != nil {
		_ = conn.Close()
		return
	}

	client := &Client{
		extensionID: identity.ExtensionID,
		userID:      identity.UserID,
		send:        make(chan []byte, sendBufferSize),
		done:        make(chan struct{}),
	}
	client.OnClose = func(id string) {
		if cfg.OnDisconnect != nil {
			cfg.OnDisconnect(id)
		}
	}

	// Register before announcing success so no command can be accepted for a
	// socket that is not yet routable.
	cfg.Hub.Register(client)

	c := &Connection{
		conn:        conn,
		client:      client,
		hub:         cfg.Hub,
		extensionID: identity.ExtensionID,
	}

	// Acknowledge only after registration.
	if err := conn.WriteJSON(WSMessage{Type: "auth_ok", Action: "connected"}); err != nil {
		cfg.Hub.Unregister(client)
		_ = conn.Close()
		return
	}

	go c.writePump()
	c.readPump()
}

// readPump forwards inbound frames to the hub until the socket closes.
//
// Every read must complete before the next, so this runs on its own goroutine;
// gorilla permits only one concurrent reader.
func (c *Connection) readPump() {
	defer func() {
		c.hub.Unregister(c.client)
		_ = c.conn.Close()
	}()

	c.conn.SetPongHandler(func(string) error {
		return c.conn.SetReadDeadline(time.Now().Add(pongWait))
	})

	for {
		_, data, err := c.conn.ReadMessage()
		if err != nil {
			// Routine disconnects (tab closed, extension reloaded) are expected
			// for this feature and must not be logged as faults, or real
			// problems would be buried in noise.
			if !IsClosed(err) {
				log.Printf("extensions: read error for %s: %v", c.extensionID, err)
			}
			return
		}

		// Refresh the deadline on any traffic, not just pongs: a busy extension
		// that is clearly alive must not be disconnected for missing a ping.
		_ = c.conn.SetReadDeadline(time.Now().Add(pongWait))

		var msg WSMessage
		if err := json.Unmarshal(data, &msg); err != nil {
			// Malformed input drops the frame, not the connection.
			log.Printf("extensions: malformed frame from %s: %v", c.extensionID, err)
			continue
		}
		c.hub.Route(c.client, msg)
	}
}

// writePump drains the client's outbound queue and emits periodic pings.
//
// Only this goroutine writes to the socket, which is required by gorilla.
func (c *Connection) writePump() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		_ = c.conn.Close()
	}()

	for {
		select {
		case data, ok := <-c.client.SendChan():
			_ = c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				_ = c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}
			if err := c.conn.WriteMessage(websocket.TextMessage, data); err != nil {
				return
			}

		case <-ticker.C:
			_ = c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}

		case <-c.client.Done():
			return
		}
	}
}

// IsClosed reports whether a websocket error indicates a normal shutdown, so
// expected disconnects are not logged as faults.
//
// CloseAbnormalClosure is included deliberately: a client that drops the TCP
// connection without a close handshake (a browser tab closing, a laptop
// sleeping, an extension being reloaded) surfaces as 1006 and is routine for
// this feature. Logging it as a fault would bury genuinely abnormal closes.
func IsClosed(err error) bool {
	if err == nil {
		return true
	}
	if errors.Is(err, io.EOF) || errors.Is(err, io.ErrUnexpectedEOF) {
		return true
	}
	return websocket.IsCloseError(err,
		websocket.CloseNormalClosure,
		websocket.CloseGoingAway,
		websocket.CloseAbnormalClosure,
		websocket.CloseNoStatusReceived,
	)
}
