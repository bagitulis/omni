package extensions

import (
	"errors"
	"fmt"
	"strings"
)

// Hub errors.
//
// These are distinct sentinel-wrapped values rather than bare strings so callers
// (and tests) can classify failures — most importantly, to tell "extension is
// offline" (retryable once it reconnects) from "hub is shutting down" (terminal).

// ErrExtensionNotConnected means no live WebSocket exists for that extension.
var ErrExtensionNotConnected = errors.New("extension is not connected")

// ErrInFlightLimit means the extension already has the maximum concurrent
// commands outstanding.
var ErrInFlightLimit = errors.New("extension reached the in-flight limit")

// ErrSendBufferFull means the extension's outbound queue is saturated, usually
// because it is slow or wedged.
var ErrSendBufferFull = errors.New("extension send buffer is full")

// ErrHubStopped means the hub has been shut down and accepts no more work.
var ErrHubStopped = errors.New("hub is stopped")

// ErrEventLoopBusy means the hub event loop did not respond in time.
var ErrEventLoopBusy = errors.New("hub event loop is not responding")

func errNotConnected(extensionID string) error {
	return fmt.Errorf("extension %q: %w", extensionID, ErrExtensionNotConnected)
}

func errInFlightLimit(extensionID string, limit int) error {
	return fmt.Errorf("extension %q (limit %d): %w", extensionID, limit, ErrInFlightLimit)
}

func errSendBufferFull(extensionID string) error {
	return fmt.Errorf("extension %q: %w", extensionID, ErrSendBufferFull)
}

func errHubStopped() error {
	return ErrHubStopped
}

func errEventLoopBusy(extensionID string) error {
	return fmt.Errorf("while sending to %q: %w", extensionID, ErrEventLoopBusy)
}

// IsNotConnected reports whether err means the extension is offline, so a caller
// can wait for a reconnect instead of treating it as a permanent failure.
func IsNotConnected(err error) bool {
	return errors.Is(err, ErrExtensionNotConnected)
}

// describeBufferState gives an operator-facing hint for a saturated buffer,
// since the usual cause is a stuck browser rather than an Omni fault.
func describeBufferState(extensionID string, queued int) string {
	if queued >= sendBufferSize {
		return fmt.Sprintf("extension %s outbound queue full (%d); the browser may be stuck", extensionID, queued)
	}
	return strings.TrimSpace(fmt.Sprintf("extension %s queue depth %d", extensionID, queued))
}
