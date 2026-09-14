# MV3 WebSocket Spike — Findings

> Date: 2026-09-14
> Risk: #1 in `docs/superpowers/specs/2026-09-14-extensions-shopee-scraper-design.md`
> Status: **partially resolved — browser-side verification still required**

## The question

Chrome MV3 service workers are terminated after roughly 30 seconds of idle time.
A WebSocket held by the worker dies with it. AutoFlow solved this with an
**offscreen document** surrogate; this design drops that layer and relies on
`chrome.alarms` plus application-level pings instead.

The risk was recorded as unresolved in the spec, so it was spiked before building
on top of it.

## What was verified (server side) — PASS

These were proven by `backend/internal/extensions/mv3_spike_test.go`, which runs
without a browser and therefore always executes:

| Property | Result | Why it matters for MV3 |
|---|---|---|
| Abrupt drop, no close handshake | PASS | A terminated worker drops the TCP connection without a close frame. The hub must not treat `1006` as a fault or leak state. |
| Same extension reconnects | PASS | The worker reconnects on its next alarm with the same `extension_id`; re-registration must not be rejected as a duplicate. |
| Repeated reconnects (25 cycles) | PASS | Each cycle must release in-flight capacity, or the 10-command cap would eventually refuse all work. |
| Pairing survives a restart | PASS | A paired browser must not have to re-pair every time Omni restarts. |
| Stale `connected` status reset | PASS | Without the startup reset the dashboard shows phantom online extensions with no socket behind them. |
| Unpair revokes immediately | PASS | A decommissioned browser must not reconnect with an old token. |

**Conclusion:** the reconnect pattern an MV3 worker produces is fully supported.
Nothing on the server side blocks the design.

## What was NOT verified — browser side, still open

The actual question — *does a terminated service worker reliably re-establish
its WebSocket?* — is a Chrome behaviour and **cannot be tested in this
environment**. It needs a real Chrome profile with the extension loaded.

Evidence recorded here so it is not mistaken for settled:

- No browser automation was run against `extension/`.
- The extension's JavaScript was syntax-checked and its pure logic unit-tested
  (20 tests), but **none of that exercises `chrome.*` or `WebSocket` at runtime**.
- The keepalive strategy in `service-worker.js` (alarm + ping) is a design
  hypothesis, not a measurement.

### How to close it

1. Load `extension/` unpacked in Chrome (`chrome://extensions`, Developer mode).
2. Pair it against a running Omni backend using a generated code.
3. Confirm the dashboard shows the extension as connected.
4. Leave the browser idle for **at least 5 minutes** — well past the ~30s worker
   timeout — without interacting with the popup.
5. Verify the extension is *still* connected, then dispatch a command and
   confirm a result comes back.

**Pass:** the socket survives idle periods and a command works after idling.
**Fail:** the socket is dropped and not restored, or commands time out after idle.

### If it fails

Add the offscreen-document surrogate that AutoFlow uses:

- An `offscreen` permission and an `offscreen.html` that owns the WebSocket.
- The service worker relays to and from the offscreen document.
- The worker itself becomes stateless, which is where MV3 wants it anyway.

That is a contained change: the wire protocol and the server are unaffected, so
the backend work already merged stays valid either way. This is the main reason
the partial spike was worth doing first — the server side is de-risked and the
fallback does not invalidate it.

## Design mitigations already in place

Chosen so the failure mode, if it occurs, degrades rather than breaks:

- **`chrome.alarms` every 25s** reconnects if the socket dropped. Alarms wake a
  terminated worker, so recovery does not require user interaction.
- **Ping every 20s** on the live socket, which is also traffic that resets
  Chrome's idle timer.
- **Exponential backoff capped at 30s** so a laptop returning from sleep
  reconnects promptly instead of waiting out a long backoff.
- **State re-read from `chrome.storage`** rather than trusted from memory, since
  any handler may run after a wake-up.
- **Server tolerates abrupt drops and re-registration** (proven above), so
  reconnect churn is not an error condition.
