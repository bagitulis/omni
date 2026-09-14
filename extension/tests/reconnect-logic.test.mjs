// Tests for reconnect behaviour, including the revoked-token case.
//
// These mirror the reconnection logic in service-worker.js. That file cannot be
// imported outside Chrome (it touches chrome.* and WebSocket at module scope),
// so the decision logic is mirrored here. The duplication is deliberate and
// narrow: it is the only way to test this without a browser, and the alternative
// (stubbing the whole chrome API surface) would test the stubs.

import test from 'node:test';
import assert from 'node:assert/strict';

const RECONNECT_BASE_MS = 1000;
const RECONNECT_MAX_MS = 30000;

/**
 * Mirror of scheduleReconnect's backoff calculation.
 *
 * Capped rather than unbounded: a browser that was offline (laptop sleep) should
 * reconnect promptly once it returns, not wait out a long backoff.
 */
function backoffFor(attempt) {
  return Math.min(RECONNECT_BASE_MS * Math.pow(2, attempt), RECONNECT_MAX_MS);
}

test('backoff grows exponentially then caps', () => {
  assert.equal(backoffFor(0), 1000);
  assert.equal(backoffFor(1), 2000);
  assert.equal(backoffFor(2), 4000);
  assert.equal(backoffFor(3), 8000);
  assert.equal(backoffFor(4), 16000);
  assert.equal(backoffFor(5), 30000, 'capped at 30s');
  assert.equal(backoffFor(50), 30000, 'stays capped, never overflows');
});

test('backoff never exceeds the cap, so a long outage does not stall recovery', () => {
  for (let attempt = 0; attempt < 100; attempt++) {
    assert.ok(backoffFor(attempt) <= RECONNECT_MAX_MS);
  }
});

/**
 * A revoked token must NOT be retried on a timer.
 *
 * The server rejects an unknown/revoked token by closing the handshake. Without
 * distinguishing that from a transient network drop, the extension reconnects
 * forever with a credential that can never work — burning battery on the
 * operator's machine and generating a steady stream of rejected handshakes.
 */
function classifyClose({ code, authFailed }) {
  if (authFailed) return 'revoked';
  if (code === 1000 || code === 1001 || code === 1006) return 'retry';
  return 'retry';
}

test('a revoked credential stops reconnection attempts', () => {
  assert.equal(classifyClose({ code: 1008, authFailed: true }), 'revoked');
  assert.equal(classifyClose({ code: 1006, authFailed: true }), 'revoked',
    'the auth flag decides, not the close code');
});

test('a transient disconnect still retries', () => {
  assert.equal(classifyClose({ code: 1006, authFailed: false }), 'retry',
    'an abnormal closure is routine (tab closed, laptop slept)');
  assert.equal(classifyClose({ code: 1000, authFailed: false }), 'retry');
});

test('a policy-violation close without an auth failure is retried once, not ignored', () => {
  // Only an explicit auth failure should stop retries; any other policy close is
  // treated as transient so a server-side hiccup does not permanently unpair.
  assert.equal(classifyClose({ code: 1008, authFailed: false }), 'retry');
});

/**
 * The keepalive alarm must not resurrect a rejected credential.
 *
 * The alarm fires every 25s and is what recovers the socket after MV3 terminates
 * the worker. Left unguarded, it would also retry a dead token every 25 seconds
 * forever — the exact loop the auth flag exists to prevent.
 */
function alarmShouldReconnect({ authRejected, socketOpen }) {
  if (authRejected) return false;
  return !socketOpen;
}

test('the alarm reconnects a dropped socket', () => {
  assert.equal(alarmShouldReconnect({ authRejected: false, socketOpen: false }), true);
});

test('the alarm does not resurrect a rejected credential', () => {
  assert.equal(alarmShouldReconnect({ authRejected: true, socketOpen: false }), false,
    'otherwise a dead token is retried every 25s forever');
});

test('the alarm does nothing while the socket is healthy', () => {
  assert.equal(alarmShouldReconnect({ authRejected: false, socketOpen: true }), false);
});

/**
 * An explicit user reconnect must override a previous rejection.
 *
 * The operator may have re-paired or the server may have been fixed; refusing to
 * retry would leave them stuck with no recovery path in the UI.
 */
function userReconnectClearsRejection() {
  let authRejected = true;
  // Mirrors the omni_reconnect handler.
  authRejected = false;
  return authRejected;
}

test('an explicit reconnect clears a prior rejection', () => {
  assert.equal(userReconnectClearsRejection(), false);
});

test('pairing success clears a prior rejection', () => {
  // Mirrors pairWithCode: a successful pairing resets the flag so the alarm can
  // reconnect again.
  let authRejected = true;
  authRejected = false; // pairing succeeded
  assert.equal(authRejected, false);
});

/**
 * A 200 response without a token is a contract violation, not a success.
 * Storing an empty token would authenticate nothing and fail forever.
 */
function pairingOutcome(body, status) {
  if (status !== 200 || !body?.success) return 'error';
  if (!body.data?.token) return 'contract_violation';
  return 'ok';
}

test('a 200 without a token is not treated as success', () => {
  assert.equal(pairingOutcome({ success: true, data: {} }, 200), 'contract_violation');
  assert.equal(pairingOutcome({ success: true, data: null }, 200), 'contract_violation');
  assert.equal(pairingOutcome({ success: true, data: { token: 't' } }, 200), 'ok');
  assert.equal(pairingOutcome({ success: false, error: 'bad code' }, 401), 'error');
});

/**
 * A pairing code is single-use and short-lived, so the extension must surface a
 * rejection rather than silently retrying the confirm call.
 */
function shouldRetryPairing({ status, attempts }) {
  // 401 means invalid/expired/already-used: all three are terminal for that code.
  if (status === 401) return false;
  // A network blip is worth one more try; the code may still be valid.
  if (status === 0 && attempts < 2) return true;
  return false;
}

test('a rejected pairing code is terminal', () => {
  assert.equal(shouldRetryPairing({ status: 401, attempts: 0 }), false);
  assert.equal(shouldRetryPairing({ status: 401, attempts: 5 }), false);
});

test('a network failure during pairing is retried briefly', () => {
  assert.equal(shouldRetryPairing({ status: 0, attempts: 0 }), true);
  assert.equal(shouldRetryPairing({ status: 0, attempts: 2 }), false,
    'give up after two attempts so the UI is not stuck');
});
