package realtime

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"github.com/gorilla/websocket"
	"github.com/rs/zerolog/log"
)

// Connection tuning — matches extensions/conn.go for consistency.
const (
	writeWait             = 10 * time.Second
	pongWait              = 60 * time.Second
	pingPeriod            = 30 * time.Second
	maxMessageSize        = 1 << 20 // 1 MiB
	handshakeTimeout      = 10 * time.Second
	authHandshakeDeadline = 10 * time.Second
	sendBufferSize        = 64
)

// ConnConfig wires the WebSocket handler to a running hub and auth resolver.
type ConnConfig struct {
	Hub            *Hub
	Authenticate   Authenticate
	AllowedOrigins []string
}

// newUpgrader — dashboard origins are the app's own domain(s); no
// chrome-extension:// support here (that's the extensions hub's job).
func newUpgrader(allowedOrigins []string) *websocket.Upgrader {
	return &websocket.Upgrader{
		HandshakeTimeout: handshakeTimeout,
		ReadBufferSize:   4096,
		WriteBufferSize:  4096,
		CheckOrigin: func(r *http.Request) bool {
			origin := r.Header.Get("Origin")
			if origin == "" {
				// Non-browser (tests, curl-ws) — no CSRF surface.
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

// ServeUpgrade is the http.HandlerFunc entry point. Accepts token via either
// the auth frame OR ?token= query (parity with SSE + Auth middleware).
func (cfg *ConnConfig) ServeUpgrade(w http.ResponseWriter, r *http.Request) {
	up := newUpgrader(cfg.AllowedOrigins)
	conn, err := up.Upgrade(w, r, nil)
	if err != nil {
		log.Warn().Err(err).Msg("realtime: upgrade failed")
		return
	}

	conn.SetReadLimit(maxMessageSize)
	_ = conn.SetReadDeadline(time.Now().Add(authHandshakeDeadline))

	// Prefer query-string token (matches how the browser attaches JWT to WS
	// URLs, since the WS handshake cannot set custom headers reliably).
	token := r.URL.Query().Get("token")
	if token == "" {
		var envelope Envelope
		if err := conn.ReadJSON(&envelope); err != nil {
			_ = conn.WriteJSON(Envelope{Type: TypeError, Error: "handshake decode failed"})
			_ = conn.Close()
			return
		}
		if envelope.Type != TypeAuth {
			_ = conn.WriteJSON(Envelope{Type: TypeError, Error: "expected auth frame"})
			_ = conn.Close()
			return
		}
		token = envelope.Token
	}

	tenantID, userID, role, err := cfg.Authenticate(token)
	if err != nil || tenantID == "" {
		_ = conn.WriteJSON(Envelope{Type: TypeError, Error: "unauthorized"})
		_ = conn.Close()
		return
	}

	// Return to healthy read deadline (refreshed by pong handler).
	_ = conn.SetReadDeadline(time.Now().Add(pongWait))
	conn.SetPongHandler(func(string) error {
		return conn.SetReadDeadline(time.Now().Add(pongWait))
	})

	client := NewClient(newClientID(), tenantID, userID, role, sendBufferSize)
	cfg.Hub.Register(client)
	if err := conn.WriteJSON(Envelope{Type: TypeAck, TenantID: tenantID}); err != nil {
		cfg.Hub.Unregister(client)
		_ = conn.Close()
		return
	}

	go writePump(conn, client)
	readPump(conn, client, cfg.Hub)
}

func writePump(conn *websocket.Conn, c *Client) {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		_ = conn.Close()
	}()
	for {
		select {
		case env, ok := <-c.Send():
			_ = conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				_ = conn.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, "hub closed"))
				return
			}
			if err := conn.WriteJSON(env); err != nil {
				return
			}
		case <-ticker.C:
			_ = conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		case <-c.Closed():
			return
		}
	}
}

func readPump(conn *websocket.Conn, c *Client, hub *Hub) {
	defer func() {
		hub.Unregister(c)
		_ = conn.Close()
	}()
	for {
		var env Envelope
		if err := conn.ReadJSON(&env); err != nil {
			return
		}
		switch env.Type {
		case TypeSubscribe:
			if env.Topic != "" {
				hub.Subscribe(c, env.Topic)
				_ = writeAck(conn, c, env.ID)
			}
		case TypeUnsubscribe:
			if env.Topic != "" {
				hub.Unsubscribe(c, env.Topic)
				_ = writeAck(conn, c, env.ID)
			}
		case TypePing:
			_ = conn.SetWriteDeadline(time.Now().Add(writeWait))
			_ = conn.WriteJSON(Envelope{Type: TypePong, ID: env.ID})
		default:
			// Ignore unknown frames — clients may include vendor extensions
			// we do not know about yet. Do not close the connection over it.
		}
	}
}

func writeAck(conn *websocket.Conn, c *Client, id string) error {
	_ = conn.SetWriteDeadline(time.Now().Add(writeWait))
	return conn.WriteJSON(Envelope{Type: TypeAck, ID: id, TenantID: c.TenantID})
}

func newClientID() string {
	var b [8]byte
	_, _ = rand.Read(b[:])
	return hex.EncodeToString(b[:])
}

// Marshal is a shortcut used by publishers that already have a Go object.
// Kept here so callers do not need to import encoding/json AND realtime.
func Marshal(v any) json.RawMessage {
	b, err := json.Marshal(v)
	if err != nil {
		return json.RawMessage("null")
	}
	return b
}
