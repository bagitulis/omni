package extensions

// Query types handled on the event loop.
//
// Membership lives on the loop goroutine, so reads must be requested there
// rather than performed directly. These channels are typed to their answers so
// the loop can serve them without reflection or type switches on `any`.
//
// Both carry a buffered result channel: the answer is sent non-blockingly, so a
// caller that has already given up cannot wedge the loop.

// connectedQuery asks whether one extension is connected.
type connectedQuery struct {
	extensionID string
	result      chan<- bool
}

// listQuery asks for a snapshot of every connected extension ID.
type listQuery struct {
	result chan<- []string
}

// syncRequest is a barrier: the loop closes ack after processing everything
// queued ahead of it. Used by RegisterSync/UnregisterSync so a caller can rely
// on a registration having taken effect.
type syncRequest struct {
	ack chan<- struct{}
}

// disconnectRequest asks the loop to close one extension's connection and
// signals completion on done.
type disconnectRequest struct {
	extensionID string
	done        chan<- struct{}
}
