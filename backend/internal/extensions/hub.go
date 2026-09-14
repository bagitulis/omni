// Package extensions implements the browser-extension WebSocket hub.
//
// The hub multiplexes commands from Omni to paired Chrome extensions and
// correlates their responses. It is ported from AutoFlow's hub, which is proven
// at scale, but trimmed to Omni's needs: Omni already has an authenticated SSE
// channel for server→browser updates, so the hub carries only
// extension→server traffic.
//
// Concurrency model: net/http serves each WebSocket on its own goroutine, so
// the hub must be safe under concurrent callers. Rather than lock shared maps,
// all state lives on a single event-loop goroutine (Run) and callers
// communicate over channels. This is what makes the correlation map safe
// without a mutex.
package extensions

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"sync"
	"sync/atomic"
	"time"
)

const (
	// sendBufferSize is the per-client outbound queue depth. A slow or stalled
	// extension fills this, after which sends are rejected rather than blocking
	// the hub.
	sendBufferSize = 256

	// maxInFlightPerExtension caps concurrent commands per extension. Without a
	// cap, a flood of commands would consume unbounded memory and could stall a
	// single browser.
	maxInFlightPerExtension = 10

	// eventLoopTimeout bounds how long a caller waits for the event loop to
	// accept work before giving up. It exists so a wedged loop surfaces as an
	// error instead of a hang.
	eventLoopTimeout = 5 * time.Second

	// resultChanTTL is how long an unanswered result channel is retained.
	// Correlating an abandoned channel leaks memory, so stale entries are swept.
	resultChanTTL = 5 * time.Minute

	// cleanupInterval is how often the stale-result sweep runs.
	cleanupInterval = 60 * time.Second
)

// WSMessage is the wire format exchanged with an extension.
type WSMessage struct {
	ID        string          `json:"id,omitempty"`
	Type      string          `json:"type,omitempty"`
	Action    string          `json:"action,omitempty"`
	Payload   json.RawMessage `json:"payload,omitempty"`
	Timestamp int64           `json:"timestamp,omitempty"`
}

// Message type values.
const (
	// MsgTypeCommand is server→extension work.
	MsgTypeCommand = "command"
	// MsgTypeResult is the correlated extension→server reply.
	MsgTypeResult = "result"
	// MsgTypeError is an extension→server failure reply.
	MsgTypeError = "error"
)

// Client is a single connected extension.
//
// send is written only by the hub event loop and read only by that client's
// write pump, which is what makes a plain channel sufficient — no mutex.
type Client struct {
	extensionID string
	userID      string
	send        chan []byte
	done        chan struct{}
	closeOnce   sync.Once

	// OnClose is invoked once when the client is removed, so callers can release
	// per-connection resources without the hub knowing about them.
	//
	// It MUST NOT BLOCK: it is invoked on the event loop, so a callback that
	// waits on anything (a database write, a mutex held by a caller currently
	// blocked in a hub send) stalls the entire hub and every other extension.
	// Long work belongs in a goroutine started from here.
	OnClose func(extensionID string)
}

// ExtensionID returns the client's extension identifier.
func (c *Client) ExtensionID() string { return c.extensionID }

// UserID returns the paired user identifier, if any.
func (c *Client) UserID() string { return c.userID }

// SendChan exposes the outbound queue for the connection's write pump.
func (c *Client) SendChan() <-chan []byte { return c.send }

// Done is closed when the client is removed from the hub.
func (c *Client) Done() <-chan struct{} { return c.done }

// close signals the client's pump to stop. Safe to call more than once.
//
// OnClose is dispatched on its own goroutine. It is invoked from the event loop,
// so calling it inline would let a slow callback — a database write, a lock held
// by a caller blocked in a hub send — stall the hub and therefore every other
// extension. Running it detached means the hub cannot be held hostage by a
// callback, and the contract no longer relies on callers being careful.
func (c *Client) close() {
	c.closeOnce.Do(func() {
		close(c.done)
		if c.OnClose != nil {
			cb := c.OnClose
			id := c.extensionID
			go func() {
				defer func() {
					if r := recover(); r != nil {
						// A panicking callback must not take the process down:
						// it is running detached, so the recover is the only
						// thing between it and a fatal crash.
						log.Printf("extensions: OnClose panic for %s: %v", id, r)
					}
				}()
				cb(id)
			}()
		}
	})
}

// sendCmd asks the event loop to deliver raw bytes to one extension.
type sendCmd struct {
	extensionID string
	data        []byte
	err         chan<- error
}

// resultReg registers (ch != nil) or removes (ch == nil) a result channel.
type resultReg struct {
	msgID string
	ch    chan WSMessage
	// extensionID is the extension the command targets, recorded so a lost reply
	// can release that extension's in-flight slot during a sweep.
	extensionID string
}

// resultEntry tracks a pending result channel and when it was registered.
type resultEntry struct {
	ch        chan WSMessage
	createdAt time.Time

	// extensionID records which extension the command was sent to, so the
	// sweep can release that extension's in-flight slot when a reply never
	// arrives. Without it, a lost reply would consume capacity permanently.
	extensionID string
}

// Hub routes commands to extensions and correlates their replies.
type Hub struct {
	// clients and resultChans are touched ONLY by the Run goroutine.
	clients     map[string]*Client
	resultChans map[string]resultEntry
	inFlight    map[string]int

	register   chan *Client
	unregister chan *Client
	route      chan *routedMessage
	send       chan sendCmd
	resultReg  chan resultReg

	// connectedQ and listQ carry read-only membership questions to the event
	// loop, so callers never read the maps concurrently with the loop.
	connectedQ  chan connectedQuery
	listQ       chan listQuery
	syncQ       chan syncRequest
	disconnectQ chan disconnectRequest
	ownerQ      chan ownerUpdate

	stop     chan struct{}
	done     chan struct{}
	stopOnce sync.Once
	// started records that Run has begun, so Stop can tell "the loop is running
	// and will close done" from "the loop was never started, so nothing will
	// ever close done". Without it, Stop on a never-started hub blocks forever.
	started atomic.Bool

	// stopped is written by the event loop and read by callers on other
	// goroutines, so it must be atomic. A plain bool here was a data race: the
	// read was unsynchronised and unsynchronised reads can be cached
	// indefinitely, so a caller could keep seeing false after shutdown.
	stopped atomic.Bool

	// OnActivity is called on the event loop whenever any message arrives from
	// an extension, so the caller can refresh last_seen without the hub knowing
	// about the database. Must not block.
	OnActivity func(extensionID string)
}

// routedMessage is an inbound message from a client, tagged with its origin.
type routedMessage struct {
	client *Client
	data   []byte
}

// NewHub creates a hub. It does nothing until Run is called.
func NewHub(_ any) *Hub {
	return &Hub{
		clients:     make(map[string]*Client),
		resultChans: make(map[string]resultEntry),
		inFlight:    make(map[string]int),
		register:    make(chan *Client, 64),
		unregister:  make(chan *Client, 64),
		route:       make(chan *routedMessage, 256),
		send:        make(chan sendCmd, 256),
		resultReg:   make(chan resultReg, 64),
		connectedQ:  make(chan connectedQuery, 32),
		listQ:       make(chan listQuery, 32),
		syncQ:       make(chan syncRequest, 32),
		disconnectQ: make(chan disconnectRequest, 16),
		ownerQ:      make(chan ownerUpdate, 64),
		stop:        make(chan struct{}),
		done:        make(chan struct{}),
	}
}

// Run is the single event loop. Call it in its own goroutine.
func (h *Hub) Run(ctx context.Context) {
	// Record that the loop is live before anything else, so a concurrent Stop
	// knows h.done will eventually close. Guarded by CompareAndSwap so a second
	// Run call is rejected rather than racing to close done twice.
	if !h.started.CompareAndSwap(false, true) {
		log.Printf("extensions: Run called on an already-started hub; ignoring")
		return
	}
	defer close(h.done)

	ticker := time.NewTicker(cleanupInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			h.shutdown()
			return

		case <-h.stop:
			h.shutdown()
			return

		case <-ticker.C:
			h.sweepStaleResults()

		case c := <-h.register:
			h.clients[c.extensionID] = c

		case c := <-h.unregister:
			h.removeClient(c.extensionID)

		case msg := <-h.route:
			h.handleRoute(msg)

		case cmd := <-h.send:
			h.handleSend(cmd)

		case reg := <-h.resultReg:
			if reg.ch == nil {
				delete(h.resultChans, reg.msgID)
			} else {
				h.resultChans[reg.msgID] = resultEntry{
					ch:          reg.ch,
					createdAt:   time.Now(),
					extensionID: reg.extensionID,
				}
			}

		case q := <-h.connectedQ:
			_, ok := h.clients[q.extensionID]
			q.result <- ok

		case u := <-h.ownerQ:
			if entry, ok := h.resultChans[u.msgID]; ok {
				entry.extensionID = u.extensionID
				h.resultChans[u.msgID] = entry
			}

		case q := <-h.listQ:
			ids := make([]string, 0, len(h.clients))
			for id := range h.clients {
				ids = append(ids, id)
			}
			q.result <- ids

		case s := <-h.syncQ:
			// Ack after all earlier cases have been processed, giving the caller
			// a happens-before edge on any register/unregister they queued.
			close(s.ack)

		case d := <-h.disconnectQ:
			h.removeClient(d.extensionID)
			close(d.done)
		}
	}
}

// shutdown closes every client and pending result channel exactly once.
//
// It also drains the buffered control channels. A caller can have queued work
// that the loop never reached — including a result-channel registration whose
// waiter would otherwise never be told. Leaving those queued means a caller
// waiting only on its channel hangs forever, which is worse than the shutdown
// itself.
func (h *Hub) shutdown() {
	h.stopped.Store(true)

	// Drain and answer anything already queued, so no caller is left waiting.
	h.drainPending()

	for id := range h.clients {
		h.removeClient(id)
	}
	// Closing pending channels unblocks any waiter rather than leaking it.
	for msgID, entry := range h.resultChans {
		close(entry.ch)
		delete(h.resultChans, msgID)
	}
}

// drainPending empties the buffered request channels, answering each caller.
//
// Non-blocking by construction: the channels are drained with a default branch so
// shutdown cannot itself block on a caller that has already given up.
func (h *Hub) drainPending() {
	for {
		select {
		case c := <-h.register:
			// Admitted after shutdown: close it so its socket teardown runs.
			c.close()
		case c := <-h.unregister:
			c.close()
		case cmd := <-h.send:
			select {
			case cmd.err <- errHubStopped():
			default:
			}
		case reg := <-h.resultReg:
			if reg.ch != nil {
				// Registered too late to ever be answered: close it so the
				// waiter is unblocked rather than hanging.
				close(reg.ch)
			}
		case d := <-h.disconnectQ:
			close(d.done)
		case _ = <-h.route:
			// Inbound traffic after shutdown has nowhere to go.
		case q := <-h.connectedQ:
			select {
			case q.result <- false:
			default:
			}
		case q := <-h.listQ:
			select {
			case q.result <- nil:
			default:
			}
		case s := <-h.syncQ:
			// A barrier waiting for a loop that is shutting down: ack it once so
			// the caller is not left blocked. Guarded by a recover-free approach:
			// the ack channel is only ever closed here, and syncRequest holders
			// do not close it themselves.
			closeSyncAck(s)
		default:
			return
		}
	}
}

// closeSyncAck acks a barrier exactly once, tolerating an already-closed
// channel. A double close would panic on the loop goroutine and take the server
// down, so it is guarded rather than assumed.
func closeSyncAck(s syncRequest) {
	defer func() {
		if r := recover(); r != nil {
			// Already closed by the normal path; nothing to do.
			_ = r
		}
	}()
	close(s.ack)
}

// Stop halts the event loop. Safe to call more than once, and safe to call on a
// hub whose Run was never started.
//
// The started check matters: h.done is closed only by Run, so waiting on it
// before the loop has begun would block forever. That turned a shutdown path
// into a hang, which is far worse than a shutdown that reports nothing to do.
func (h *Hub) Stop() {
	h.stopOnce.Do(func() {
		close(h.stop)
	})

	if !h.started.Load() {
		// No loop is running, so nothing will ever close done. Mark stopped so
		// callers observe the terminal state, and return rather than block.
		h.stopped.Store(true)
		return
	}
	<-h.done
}

// removeClient drops a client and its in-flight accounting.
func (h *Hub) removeClient(extensionID string) {
	if c, ok := h.clients[extensionID]; ok {
		c.close()
		delete(h.clients, extensionID)
	}
	delete(h.inFlight, extensionID)
}

// sweepStaleResults closes and drops result channels nobody answered, releasing
// the in-flight slot each one held.
//
// Releasing matters: handleRoute only decrements on a correlated reply, so
// without this a reply lost in transit (dropped frame, extension crash) would
// consume that extension's capacity permanently and eventually starve it.
func (h *Hub) sweepStaleResults() {
	now := time.Now()
	for msgID, entry := range h.resultChans {
		if now.Sub(entry.createdAt) > resultChanTTL {
			close(entry.ch)
			if entry.extensionID != "" && h.inFlight[entry.extensionID] > 0 {
				h.inFlight[entry.extensionID]--
			}
			delete(h.resultChans, msgID)
		}
	}
}

// handleSend delivers raw bytes to one extension, enforcing the in-flight cap.
//
// In-flight counts commands handed to the socket but not yet answered. It is
// decremented on reply (handleRoute) AND when the command is dropped, so a
// reply that never arrives cannot permanently consume capacity. The cap exists
// to stop one extension being flooded, not to limit sustained throughput: with
// a well-behaved extension the slot frees as soon as it responds.
func (h *Hub) handleSend(cmd sendCmd) {
	c, ok := h.clients[cmd.extensionID]
	if !ok {
		cmd.err <- errNotConnected(cmd.extensionID)
		return
	}
	if h.inFlight[cmd.extensionID] >= maxInFlightPerExtension {
		cmd.err <- errInFlightLimit(cmd.extensionID, maxInFlightPerExtension)
		return
	}
	if !c.tryQueue(cmd.data) {
		cmd.err <- errSendBufferFull(cmd.extensionID)
		return
	}
	h.inFlight[cmd.extensionID]++
	cmd.err <- nil
}

// handleRoute dispatches an inbound extension message, first to any waiting
// result channel, then to the registered activity callback.
func (h *Hub) handleRoute(msg *routedMessage) {
	if h.OnActivity != nil {
		h.OnActivity(msg.client.extensionID)
	}

	var wsm WSMessage
	if err := json.Unmarshal(msg.data, &wsm); err != nil {
		// Malformed input must not kill the connection or the loop.
		log.Printf("extensions: route unmarshal error from %s: %v", msg.client.extensionID, err)
		return
	}

	if wsm.Type != MsgTypeResult && wsm.Type != MsgTypeError {
		// Not a reply, so it must not consume in-flight capacity. Decrementing on
		// every inbound frame let a chatty extension reset the counter with
		// unrelated traffic (progress events, pings) and bypass the cap entirely,
		// which is the one thing the cap exists to prevent.
		return
	}

	// Only a reply that actually matches an outstanding command frees its slot.
	// An unsolicited or stale reply must not, or the same bypass applies.
	entry, ok := h.resultChans[wsm.ID]
	if !ok {
		// Late or unsolicited result: the waiter already gave up. Dropping it is
		// correct — it must not be delivered to a stale channel.
		return
	}

	if h.inFlight[msg.client.extensionID] > 0 {
		h.inFlight[msg.client.extensionID]--
	}

	select {
	case entry.ch <- wsm:
	default:
		log.Printf("extensions: result channel full for msgID=%s, dropping", wsm.ID)
	}
}

// tryQueue attempts a non-blocking send. A full buffer returns false rather
// than blocking the caller.
func (c *Client) tryQueue(data []byte) bool {
	select {
	case c.send <- data:
		return true
	case <-c.done:
		return false
	default:
		return false
	}
}

// Register adds a client. Safe from any goroutine.
//
// Registration is asynchronous: this returns once the request is queued, not
// once it is visible to the loop. Callers that must observe the client
// immediately (tests, or admitting a connection before sending it work) should
// use RegisterSync, which round-trips through the loop.
func (h *Hub) Register(c *Client) {
	select {
	case h.register <- c:
	case <-h.done:
	case <-time.After(eventLoopTimeout):
		log.Printf("extensions: Register timed out for %s", c.extensionID)
	}
}

// RegisterSync adds a client and waits until the event loop has admitted it, so
// a subsequent SendToExtension is guaranteed to see it.
func (h *Hub) RegisterSync(c *Client) {
	h.Register(c)
	ack := make(chan struct{}, 1)
	select {
	case h.syncQ <- syncRequest{ack: ack}:
	case <-h.done:
		return
	case <-time.After(eventLoopTimeout):
		log.Printf("extensions: RegisterSync ack timed out for %s", c.extensionID)
		return
	}
	select {
	case <-ack:
	case <-h.done:
	case <-time.After(eventLoopTimeout):
	}
}

// UnregisterSync removes a client and waits until the loop has processed it, so
// a subsequent send is guaranteed to fail as "not connected".
func (h *Hub) UnregisterSync(c *Client) {
	h.Unregister(c)
	ack := make(chan struct{}, 1)
	select {
	case h.syncQ <- syncRequest{ack: ack}:
	case <-h.done:
		return
	case <-time.After(eventLoopTimeout):
		return
	}
	select {
	case <-ack:
	case <-h.done:
	case <-time.After(eventLoopTimeout):
	}
}

// Unregister removes a client. Safe from any goroutine.
func (h *Hub) Unregister(c *Client) {
	select {
	case h.unregister <- c:
	case <-h.done:
	case <-time.After(eventLoopTimeout):
		log.Printf("extensions: Unregister timed out for %s", c.extensionID)
	}
}

// SendToExtension queues msg for one extension.
//
// Returns an error when the extension is absent, its buffer is full, or it has
// reached the in-flight cap — never a silent drop, because a silently discarded
// command would surface later as an unexplained timeout.
func (h *Hub) SendToExtension(extensionID string, msg WSMessage) error {
	if h.isStopped() {
		return errHubStopped()
	}
	data, err := json.Marshal(msg)
	if err != nil {
		return err
	}
	errCh := make(chan error, 1)
	select {
	case h.send <- sendCmd{extensionID: extensionID, data: data, err: errCh}:
	case <-h.done:
		return errHubStopped()
	case <-time.After(eventLoopTimeout):
		return errEventLoopBusy(extensionID)
	}
	select {
	case err := <-errCh:
		return err
	case <-h.done:
		return errHubStopped()
	case <-time.After(eventLoopTimeout):
		return errEventLoopBusy(extensionID)
	}
}

// RegisterResultChannel registers ch to receive the reply whose ID matches
// msgID. Call before sending so the reply cannot be missed.
func (h *Hub) RegisterResultChannel(msgID string, ch chan WSMessage) {
	h.RegisterResultChannelFor(msgID, "", ch)
}

// RegisterResultChannelFor registers a result channel and records which
// extension the command targets.
//
// The owner matters for capacity accounting: a reply that never arrives would
// otherwise hold the extension's in-flight slot forever, so the sweep needs to
// know whose slot to release.
func (h *Hub) RegisterResultChannelFor(msgID, extensionID string, ch chan WSMessage) {
	select {
	case h.resultReg <- resultReg{msgID: msgID, ch: ch, extensionID: extensionID}:
	case <-h.done:
		close(ch)
	case <-time.After(eventLoopTimeout):
		log.Printf("extensions: RegisterResultChannel timed out for %s", msgID)
	}
}

// SetResultOwner records which extension an already-registered result channel
// belongs to. Provided for callers that register before knowing the target.
func (h *Hub) SetResultOwner(msgID, extensionID string) {
	select {
	case h.ownerQ <- ownerUpdate{msgID: msgID, extensionID: extensionID}:
	case <-h.done:
	case <-time.After(eventLoopTimeout):
	}
}

// UnregisterResultChannel removes msgID's channel. Always call it (typically via
// defer) to avoid leaking an entry for a reply that never arrives.
func (h *Hub) UnregisterResultChannel(msgID string) {
	select {
	case h.resultReg <- resultReg{msgID: msgID, ch: nil}:
	case <-h.done:
	case <-time.After(eventLoopTimeout):
		log.Printf("extensions: UnregisterResultChannel timed out for %s", msgID)
	}
}

// Route feeds an inbound message into the hub. Exported so the WebSocket read
// pump (and tests) can inject traffic without touching internal channels.
func (h *Hub) Route(c *Client, msg WSMessage) {
	data, err := json.Marshal(msg)
	if err != nil {
		return
	}
	select {
	case h.route <- &routedMessage{client: c, data: data}:
	case <-h.done:
	case <-time.After(eventLoopTimeout):
		log.Printf("extensions: Route timed out for %s", c.extensionID)
	}
}

// BroadcastToDashboards exists for interface parity with AutoFlow's hub.
//
// Omni fans progress out over its existing authenticated SSE channel, so there
// are no dashboard clients on this hub. It is retained as an explicit no-op so
// call sites read the same as AutoFlow's and the intent is documented rather
// than looking like a missing implementation.
func (h *Hub) BroadcastToDashboards(_ WSMessage) {}

// IsConnected reports whether an extension currently has a live connection.
//
// Membership is owned by the event loop, so the question is asked there rather
// than read from the map (which would be a data race).
func (h *Hub) IsConnected(extensionID string) bool {
	if h.isStopped() {
		return false
	}
	probe := make(chan bool, 1)
	select {
	case h.connectedQ <- connectedQuery{extensionID: extensionID, result: probe}:
	case <-h.done:
		return false
	case <-time.After(eventLoopTimeout):
		return false
	}
	select {
	case connected := <-probe:
		return connected
	case <-h.done:
		return false
	case <-time.After(eventLoopTimeout):
		return false
	}
}

// ConnectedIDs returns a snapshot of all connected extension IDs.
func (h *Hub) ConnectedIDs() []string {
	if h.isStopped() {
		return nil
	}
	out := make(chan []string, 1)
	select {
	case h.listQ <- listQuery{result: out}:
	case <-h.done:
		return nil
	case <-time.After(eventLoopTimeout):
		return nil
	}
	select {
	case ids := <-out:
		return ids
	case <-h.done:
		return nil
	case <-time.After(eventLoopTimeout):
		return nil
	}
}

// Disconnect removes and closes an extension's connection, so an operator who
// unpairs or revokes a browser cannot have it keep acting on live commands.
func (h *Hub) Disconnect(extensionID string) {
	probe := make(chan struct{}, 1)
	select {
	case h.disconnectQ <- disconnectRequest{extensionID: extensionID, done: probe}:
	case <-h.done:
		return
	case <-time.After(eventLoopTimeout):
		log.Printf("extensions: Disconnect timed out for %s", extensionID)
		return
	}
	select {
	case <-probe:
	case <-h.done:
	case <-time.After(eventLoopTimeout):
	}
}

// RegisterTestClient registers a connection that accepts commands and never
// replies, for tests that need a pending command to stay outstanding.
//
// Exported so a test in another package can drive the hub without building a
// real socket. The client's queue is never drained, so a command is accepted and
// then simply not answered.
func (h *Hub) RegisterTestClient(extensionID string) error {
	if extensionID == "" {
		return fmt.Errorf("extensions: extensionID is required")
	}
	c := &Client{
		extensionID: extensionID,
		send:        make(chan []byte, sendBufferSize),
		done:        make(chan struct{}),
	}
	h.RegisterSync(c)
	return nil
}

func (h *Hub) isStopped() bool { return h.stopped.Load() }
