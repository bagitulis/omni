package shopee

import (
	"context"
	"encoding/json"
	"fmt"
	"sync/atomic"
	"time"

	"github.com/google/uuid"
	"github.com/omni/backend/internal/extensions"
)

// HubSender adapts the WebSocket hub to the scraper's CommandSender interface.
//
// Kept as an adapter in this package so the scraper does not depend on the hub:
// the page loop is then testable with a fake, which is the only way to exercise
// it here since the real transport needs a browser.
type HubSender struct {
	hub         *extensions.Hub
	extensionID string
	// tabID is the tab commands act on. Set by the caller after opening a tab,
	// so concurrent scrapes on one browser cannot act on each other's pages.
	tabID atomic.Int64

	// timeout bounds a single command. Commands are network round trips through
	// a browser, so the default is generous.
	timeout time.Duration
}

// NewHubSender creates a HubSender for one extension.
func NewHubSender(hub *extensions.Hub, extensionID string) *HubSender {
	return &HubSender{hub: hub, extensionID: extensionID, timeout: 60 * time.Second}
}

// SetTabID records the tab commands should target.
func (s *HubSender) SetTabID(tabID int64) { s.tabID.Store(tabID) }

// TabID returns the current tab, or 0 when none is set.
func (s *HubSender) TabID() int64 { return s.tabID.Load() }

// Send runs one command and waits for its correlated result.
//
// A tab_id is attached to every command. Without it the extension would fall
// back to the active tab, so a user switching tabs mid-scrape would silently
// redirect the scrape to whatever page they opened.
func (s *HubSender) Send(ctx context.Context, action string, payload any) (json.RawMessage, error) {
	if s.hub == nil {
		return nil, fmt.Errorf("hub sender: no hub configured")
	}

	finalPayload, err := buildPayload(payload, s.tabID.Load())
	if err != nil {
		return nil, fmt.Errorf("hub sender: %s: %w", action, err)
	}

	msgID := uuid.New().String()
	resultCh := make(chan extensions.WSMessage, 1)
	// Register with the target extension so a reply that never arrives can
	// release this extension's in-flight slot during the hub's sweep. Without
	// the owner, one lost reply would hold capacity indefinitely.
	s.hub.RegisterResultChannelFor(msgID, s.extensionID, resultCh)
	defer s.hub.UnregisterResultChannel(msgID)

	msg := extensions.WSMessage{
		ID:        msgID,
		Type:      extensions.MsgTypeCommand,
		Action:    action,
		Payload:   finalPayload,
		Timestamp: time.Now().UnixMilli(),
	}

	if err := s.hub.SendToExtension(s.extensionID, msg); err != nil {
		return nil, fmt.Errorf("hub sender: send %s: %w", action, err)
	}

	timeout := s.timeout
	if timeout <= 0 {
		timeout = 60 * time.Second
	}
	timer := time.NewTimer(timeout)
	defer timer.Stop()

	select {
	case <-ctx.Done():
		return nil, fmt.Errorf("hub sender: cancelled waiting for %s: %w", action, ctx.Err())
	case <-timer.C:
		return nil, fmt.Errorf("hub sender: timeout after %s waiting for %s", timeout, action)
	case result, ok := <-resultCh:
		// ok == false means the hub closed this channel instead of delivering a
		// reply — which it does for every pending channel on shutdown, and
		// immediately if the loop has already exited.
		//
		// Without this check the receive yields a zero-value WSMessage, which
		// interpretResult accepts as a SUCCESS with an empty payload. That would
		// report a scrape as having captured nothing when in fact the hub died,
		// so the two cases must be distinguished here.
		if !ok {
			return nil, fmt.Errorf("hub sender: hub closed the result channel while waiting for %s (hub stopped)", action)
		}
		return interpretResult(action, result)
	}
}

// buildPayload serialises a command payload with the target tab id merged in.
//
// The payload must be a JSON object. A caller-supplied tab_id is overwritten by
// the sender's own value: one scrape must not be able to direct commands at
// another scrape's tab.
func buildPayload(payload any, tabID int64) (json.RawMessage, error) {
	raw, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("marshal payload: %w", err)
	}

	merged := map[string]any{}
	if len(raw) > 0 && string(raw) != "null" {
		if err := json.Unmarshal(raw, &merged); err != nil {
			return nil, fmt.Errorf("payload must be a JSON object: %w", err)
		}
	}

	if tabID > 0 {
		merged["tab_id"] = tabID
	}

	out, err := json.Marshal(merged)
	if err != nil {
		return nil, fmt.Errorf("marshal merged payload: %w", err)
	}
	return out, nil
}

// interpretResult converts an extension reply into a payload or an error.
//
// A reply of type "error" is a command failure, so it becomes an error rather
// than a payload the caller would have to inspect.
func interpretResult(action string, result extensions.WSMessage) (json.RawMessage, error) {
	if result.Type == extensions.MsgTypeError {
		var errPayload struct {
			Error   string `json:"error"`
			Message string `json:"message"`
		}
		if len(result.Payload) > 0 {
			_ = json.Unmarshal(result.Payload, &errPayload)
		}
		msg := errPayload.Message
		if msg == "" {
			msg = errPayload.Error
		}
		if msg == "" {
			msg = "extension reported an error"
		}
		return nil, fmt.Errorf("%s failed: %s", action, msg)
	}
	return result.Payload, nil
}

// OpenTab opens a background tab for scraping and records it for later commands.
//
// The tab is created non-active and then activated by the extension: Chrome
// throttles background tabs hard enough to stall lazy-loading, so remaining
// hidden is not an option for a scrape.
func (s *HubSender) OpenTab(ctx context.Context, url string) (int64, error) {
	raw, err := s.Send(ctx, "open_tab", map[string]any{"url": url, "active": false})
	if err != nil {
		return 0, err
	}

	var out struct {
		TabID int64 `json:"tab_id"`
	}
	if err := json.Unmarshal(unwrapEnvelope(raw), &out); err != nil {
		return 0, fmt.Errorf("hub sender: parse open_tab result: %w", err)
	}
	if out.TabID == 0 {
		return 0, fmt.Errorf("hub sender: open_tab returned tab_id 0")
	}

	s.SetTabID(out.TabID)
	return out.TabID, nil
}

// CloseTab closes the tab recorded by OpenTab. Errors are ignored: the tab may
// already be gone, and failing to close it must not fail an otherwise complete
// scrape.
func (s *HubSender) CloseTab(ctx context.Context) {
	tabID := s.tabID.Load()
	if tabID == 0 {
		return
	}
	_, _ = s.Send(ctx, "close_tab", map[string]any{"tab_id": tabID})
	s.tabID.Store(0)
}
