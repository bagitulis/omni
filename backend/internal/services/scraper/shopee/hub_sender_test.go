package shopee

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/omni/backend/internal/extensions"
)

// HubSender is the bridge between the scraper's CommandSender interface and the
// real WebSocket hub. Its payload handling is worth testing independently of a
// browser, because the tab_id merge is what keeps a scrape pointed at its own
// page rather than whatever tab the user has focused.

// mergeTabID is a test-friendly wrapper over buildPayload.
func mergeTabID(t *testing.T, payload map[string]any, tabID int64) map[string]any {
	t.Helper()
	raw, err := buildPayload(payload, tabID)
	if err != nil {
		t.Fatalf("buildPayload: %v", err)
	}
	var out map[string]any
	if err := json.Unmarshal(raw, &out); err != nil {
		t.Fatalf("unmarshal merged payload: %v", err)
	}
	return out
}

func TestHubSender_MergesTabIDIntoPayload(t *testing.T) {
	// A payload's own keys must survive, and the tab id must be added.
	merged := mergeTabID(t, map[string]any{"filter": "search_items", "limit": 5}, 42)

	if merged["tab_id"] != float64(42) {
		t.Errorf("tab_id = %v, want 42", merged["tab_id"])
	}
	if merged["filter"] != "search_items" {
		t.Errorf("original payload key lost: %v", merged)
	}
	if merged["limit"] != float64(5) {
		t.Errorf("original payload key lost: %v", merged)
	}
}

func TestHubSender_OmitsTabIDWhenUnset(t *testing.T) {
	// Before a tab is opened there is nothing to target; sending tab_id 0 would
	// be worse than omitting it, since the extension would treat 0 as a real id.
	merged := mergeTabID(t, map[string]any{"url": "https://shopee.co.id"}, 0)

	if _, present := merged["tab_id"]; present {
		t.Errorf("tab_id must not be sent when no tab is open: %v", merged)
	}
}

func TestHubSender_TabIDOverridesCallerValue(t *testing.T) {
	// The sender owns the target tab. A caller-supplied tab_id would let one
	// scrape act on another's page, so the sender's value wins.
	merged := mergeTabID(t, map[string]any{"tab_id": 1}, 99)

	if merged["tab_id"] != float64(99) {
		t.Errorf("tab_id = %v, want the sender's 99", merged["tab_id"])
	}
}

func TestHubSender_RejectsNonObjectPayload(t *testing.T) {
	// A JSON array or scalar cannot carry a tab id, so it is a programming error
	// rather than something to silently send.
	s := NewHubSender(nil, "ext-1")
	_, err := s.Send(context.Background(), "extract", []string{"a", "b"})
	if err == nil {
		t.Error("a non-object payload must be rejected")
	}
}

func TestHubSender_NilHubIsAnError(t *testing.T) {
	s := NewHubSender(nil, "ext-1")
	if _, err := s.Send(context.Background(), "extract", map[string]any{}); err == nil {
		t.Error("a sender with no hub must error rather than panic")
	}
}

func TestHubSender_OpenTabParsesEnvelope(t *testing.T) {
	// open_tab replies are wrapped, and the tab id is the only thing we need
	// from them; failing to unwrap would look like tab_id 0.
	raw := json.RawMessage(`{"success":true,"data":{"tab_id":77,"timed_out":false}}`)

	var out struct {
		TabID int64 `json:"tab_id"`
	}
	if err := json.Unmarshal(unwrapEnvelope(raw), &out); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if out.TabID != 77 {
		t.Errorf("TabID = %d, want 77", out.TabID)
	}
}

func TestInterpretResult(t *testing.T) {
	t.Run("success returns payload", func(t *testing.T) {
		payload, err := interpretResult("extract", extensions.WSMessage{
			Type:    extensions.MsgTypeResult,
			Payload: json.RawMessage(`{"success":true}`),
		})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(payload) == 0 {
			t.Error("payload must be returned")
		}
	})

	t.Run("error type becomes an error", func(t *testing.T) {
		_, err := interpretResult("extract", extensions.WSMessage{
			Type:    extensions.MsgTypeError,
			Payload: json.RawMessage(`{"message":"element not found"}`),
		})
		if err == nil {
			t.Fatal("an error reply must become an error")
		}
		if got := err.Error(); got != "extract failed: element not found" {
			t.Errorf("error = %q", got)
		}
	})

	t.Run("error with only error field", func(t *testing.T) {
		_, err := interpretResult("goto", extensions.WSMessage{
			Type:    extensions.MsgTypeError,
			Payload: json.RawMessage(`{"error":"navigation failed"}`),
		})
		if err == nil || err.Error() != "goto failed: navigation failed" {
			t.Errorf("error = %v", err)
		}
	})

	t.Run("error with no detail still errors", func(t *testing.T) {
		_, err := interpretResult("reload", extensions.WSMessage{Type: extensions.MsgTypeError})
		if err == nil {
			t.Fatal("an error reply with no payload must still be an error")
		}
	})
}

func TestHubSender_TimeoutIsBounded(t *testing.T) {
	// A command must not wait forever. Uses a real hub with no extension
	// connected, so SendToExtension fails fast rather than waiting on the timer.
	hub := extensions.NewHub(nil)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go hub.Run(ctx)

	s := NewHubSender(hub, "ext-absent")
	s.timeout = 200 * time.Millisecond

	start := time.Now()
	_, err := s.Send(context.Background(), "extract", map[string]any{})
	elapsed := time.Since(start)

	if err == nil {
		t.Fatal("sending to a disconnected extension must fail")
	}
	if elapsed > 5*time.Second {
		t.Errorf("send blocked for %v; it must fail fast, not wait out the timer", elapsed)
	}
}
