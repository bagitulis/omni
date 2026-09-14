package extensions

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gorilla/websocket"
)

// These tests drive a real WebSocket through httptest, so they exercise the
// handshake, read/write pumps, and hub integration together rather than in
// isolation.

type testConnEnv struct {
	server    *httptest.Server
	hub       *Hub
	authCalls int
	// authErr, when set, makes authentication fail (used for the reject path).
	authErr error
	// lastReq records the connect request the server received.
	lastReq ConnectRequest
}

func newTestConnEnv(t *testing.T, authErr error) *testConnEnv {
	t.Helper()

	env := &testConnEnv{hub: NewHub(nil), authErr: authErr}

	ctx, cancel := context.WithCancel(context.Background())
	go env.hub.Run(ctx)

	cfg := ConnConfig{
		Hub: env.hub,
		Authenticate: func(_ context.Context, req ConnectRequest) (*ConnectionIdentity, error) {
			env.authCalls++
			env.lastReq = req
			if env.authErr != nil {
				return nil, env.authErr
			}
			return &ConnectionIdentity{
				ExtensionID: req.ExtensionID,
				TenantID:    "tenant-a",
				UserID:      "user-1",
			}, nil
		},
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/api/extensions/ws", cfg.ServeUpgrade)
	env.server = httptest.NewServer(mux)

	t.Cleanup(func() {
		env.server.Close()
		cancel()
	})

	return env
}

func dialWS(t *testing.T, env *testConnEnv) *websocket.Conn {
	t.Helper()
	url := "ws" + strings.TrimPrefix(env.server.URL, "http") + "/api/extensions/ws"
	conn, _, err := websocket.DefaultDialer.Dial(url, nil)
	if err != nil {
		t.Fatalf("dial: %v", err)
	}
	return conn
}

func TestConn_HappyPathAuthAndReceive(t *testing.T) {
	env := newTestConnEnv(t, nil)
	conn := dialWS(t, env)
	defer conn.Close()

	if err := conn.WriteJSON(ConnectRequest{Token: "tok", ExtensionID: "ext-ws-1"}); err != nil {
		t.Fatalf("write connect: %v", err)
	}

	// Server must acknowledge only after the extension is routable.
	var ack WSMessage
	if err := conn.ReadJSON(&ack); err != nil {
		t.Fatalf("read auth_ok: %v", err)
	}
	if ack.Type != "auth_ok" {
		t.Fatalf("expected auth_ok, got %+v", ack)
	}
	if !env.hub.IsConnected("ext-ws-1") {
		t.Fatal("extension must be routable after auth_ok")
	}

	// A command sent by the server must arrive at the extension.
	if err := env.hub.SendToExtension("ext-ws-1", WSMessage{ID: "cmd-1", Type: MsgTypeCommand, Action: "open_tab"}); err != nil {
		t.Fatalf("SendToExtension: %v", err)
	}
	_ = conn.SetReadDeadline(time.Now().Add(3 * time.Second))
	var cmd WSMessage
	if err := conn.ReadJSON(&cmd); err != nil {
		t.Fatalf("read command: %v", err)
	}
	if cmd.ID != "cmd-1" || cmd.Action != "open_tab" {
		t.Errorf("got %+v, want ID=cmd-1 Action=open_tab", cmd)
	}
}

func TestConn_ResultCorrelationOverTheWire(t *testing.T) {
	env := newTestConnEnv(t, nil)
	conn := dialWS(t, env)
	defer conn.Close()

	if err := conn.WriteJSON(ConnectRequest{Token: "tok", ExtensionID: "ext-ws-2"}); err != nil {
		t.Fatalf("connect: %v", err)
	}
	var ack WSMessage
	if err := conn.ReadJSON(&ack); err != nil {
		t.Fatalf("ack: %v", err)
	}

	// Register the waiter, then have the extension reply with the same ID.
	resultCh := make(chan WSMessage, 1)
	env.hub.RegisterResultChannel("round-trip-1", resultCh)

	payload, _ := json.Marshal(map[string]any{"success": true, "data": []string{"a"}})
	if err := conn.WriteJSON(WSMessage{ID: "round-trip-1", Type: MsgTypeResult, Payload: payload}); err != nil {
		t.Fatalf("write result: %v", err)
	}

	select {
	case got := <-resultCh:
		if got.ID != "round-trip-1" {
			t.Errorf("correlated ID = %q, want round-trip-1", got.ID)
		}
	case <-time.After(3 * time.Second):
		t.Fatal("result never correlated over the wire")
	}
}

func TestConn_RejectedAuthClosesWithoutRegistering(t *testing.T) {
	env := newTestConnEnv(t, errors.New("invalid token"))
	conn := dialWS(t, env)
	defer conn.Close()

	if err := conn.WriteJSON(ConnectRequest{Token: "bad", ExtensionID: "ext-bad"}); err != nil {
		t.Fatalf("connect: %v", err)
	}

	// The server must close rather than acknowledge.
	_ = conn.SetReadDeadline(time.Now().Add(3 * time.Second))
	for {
		_, _, err := conn.ReadMessage()
		if err != nil {
			break
		}
	}

	if env.hub.IsConnected("ext-bad") {
		t.Fatal("a rejected handshake must not leave the extension connected")
	}
}

func TestConn_InvalidExtensionIDRejected(t *testing.T) {
	env := newTestConnEnv(t, nil)
	conn := dialWS(t, env)
	defer conn.Close()

	// A path-traversal style ID must be refused before registration.
	if err := conn.WriteJSON(ConnectRequest{Token: "tok", ExtensionID: "../../etc/passwd"}); err != nil {
		t.Fatalf("connect: %v", err)
	}

	_ = conn.SetReadDeadline(time.Now().Add(3 * time.Second))
	for {
		_, _, err := conn.ReadMessage()
		if err != nil {
			break
		}
	}
	if len(env.hub.ConnectedIDs()) != 0 {
		t.Errorf("invalid extension_id must not register, got %v", env.hub.ConnectedIDs())
	}
}

func TestConn_DisappearsFromHubOnClose(t *testing.T) {
	env := newTestConnEnv(t, nil)
	conn := dialWS(t, env)

	if err := conn.WriteJSON(ConnectRequest{Token: "tok", ExtensionID: "ext-close"}); err != nil {
		t.Fatalf("connect: %v", err)
	}
	var ack WSMessage
	if err := conn.ReadJSON(&ack); err != nil {
		t.Fatalf("ack: %v", err)
	}
	if !env.hub.IsConnected("ext-close") {
		t.Fatal("expected connected before close")
	}

	_ = conn.Close()

	// The hub must eventually drop it, otherwise the UI shows a phantom
	// connected extension and sends commands into the void.
	deadline := time.After(5 * time.Second)
	for {
		if !env.hub.IsConnected("ext-close") {
			break
		}
		select {
		case <-deadline:
			t.Fatal("extension still reported connected after socket close")
		case <-time.After(20 * time.Millisecond):
		}
	}
}

func TestConn_MalformedFrameKeepsConnectionAlive(t *testing.T) {
	env := newTestConnEnv(t, nil)
	conn := dialWS(t, env)
	defer conn.Close()

	if err := conn.WriteJSON(ConnectRequest{Token: "tok", ExtensionID: "ext-bad-json"}); err != nil {
		t.Fatalf("connect: %v", err)
	}
	var ack WSMessage
	if err := conn.ReadJSON(&ack); err != nil {
		t.Fatalf("ack: %v", err)
	}

	// Garbage input must be dropped, not fatal.
	if err := conn.WriteMessage(websocket.TextMessage, []byte("{not json")); err != nil {
		t.Fatalf("write garbage: %v", err)
	}

	// The connection must still work afterwards.
	deadline := time.After(3 * time.Second)
	for {
		if err := env.hub.SendToExtension("ext-bad-json", WSMessage{ID: "after-bad", Type: MsgTypeCommand}); err == nil {
			break
		}
		select {
		case <-deadline:
			t.Fatal("connection unusable after a malformed frame")
		case <-time.After(20 * time.Millisecond):
		}
	}

	_ = conn.SetReadDeadline(time.Now().Add(3 * time.Second))
	var cmd WSMessage
	if err := conn.ReadJSON(&cmd); err != nil {
		t.Fatalf("read after garbage: %v", err)
	}
	if cmd.ID != "after-bad" {
		t.Errorf("got %+v, want ID=after-bad", cmd)
	}
}

func TestUpgrader_CheckOrigin(t *testing.T) {
	up := newUpgrader([]string{"https://app.example.com"})

	cases := []struct {
		origin string
		want   bool
		why    string
	}{
		{"", true, "non-browser clients cannot be driven by a hostile page"},
		{"chrome-extension://abcdefghijklmnopabcdefghijklmnop", true, "extensions are the primary client"},
		{"https://app.example.com", true, "configured dashboard origin"},
		{"https://evil.example.net", false, "an unrelated origin must be refused"},
		{"http://localhost.evil.com", false, "prefix lookalike must not pass a suffix compare"},
	}

	for _, tc := range cases {
		t.Run(tc.origin, func(t *testing.T) {
			req, err := http.NewRequest(http.MethodGet, "http://example.com/ws", nil)
			if err != nil {
				t.Fatalf("new request: %v", err)
			}
			if tc.origin != "" {
				req.Header.Set("Origin", tc.origin)
			}
			if got := up.CheckOrigin(req); got != tc.want {
				t.Errorf("CheckOrigin(%q) = %v, want %v (%s)", tc.origin, got, tc.want, tc.why)
			}
		})
	}
}

func TestIsClosed(t *testing.T) {
	// Normal shutdowns must be recognised, so they are not logged as faults.
	if !IsClosed(io.EOF) {
		t.Error("io.EOF must be recognised as a normal close")
	}
	if !IsClosed(io.ErrUnexpectedEOF) {
		t.Error("io.ErrUnexpectedEOF must be recognised as a normal close")
	}
	if !IsClosed(&websocket.CloseError{Code: websocket.CloseNormalClosure}) {
		t.Error("a normal close frame must be recognised")
	}
	if !IsClosed(&websocket.CloseError{Code: websocket.CloseGoingAway}) {
		t.Error("a going-away close frame must be recognised")
	}
	if !IsClosed(&websocket.CloseError{Code: websocket.CloseAbnormalClosure}) {
		t.Error("1006 (no close handshake) is routine here — a closed tab or reloaded extension — and must not be logged as a fault")
	}

	// Anything else is a fault and must be reported as such.
	if IsClosed(errors.New("some other error")) {
		t.Error("a generic error must not be reported as a normal close")
	}
}
