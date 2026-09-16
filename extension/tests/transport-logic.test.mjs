// Tests for transport resilience helpers in the service worker.
//
// These import the real functions from service-worker.js. That file guards its
// browser-only work behind `typeof chrome !== 'undefined'`, so the helpers below
// are importable under Node while the socket and alarm wiring stays inert.
//
// Each block corresponds to a defect found in the audit: a result dropped during
// a reconnect, a content command that could hang until the backend's own 60s
// bound expired, and a navigation failure that resolved exactly like a success.

import test from 'node:test';
import assert from 'node:assert/strict';

import {
  makeResultQueue,
  withTimeout,
  classifyLoadOutcome,
  flushQueue,
  PENDING_RESULT_LIMIT,
} from '../service-worker.js';

// ─── Results survive a socket that is not OPEN ───────────────────────────────

test('a result produced while the socket is closed is queued, not lost', () => {
  const q = makeResultQueue();
  q.enqueue({ id: 'cmd-1', type: 'result', action: 'extract' });
  assert.equal(q.size(), 1);
});

test('the queue drains in order, and only once', () => {
  const q = makeResultQueue();
  q.enqueue({ id: 'cmd-1' });
  q.enqueue({ id: 'cmd-2' });
  q.enqueue({ id: 'cmd-3' });

  assert.deepEqual(q.drain().map((e) => e.id), ['cmd-1', 'cmd-2', 'cmd-3']);
  assert.equal(q.size(), 0, 'a drained result must not be sent twice');
});

test('the queue is bounded and drops the oldest first', () => {
  // An extension left offline must not grow this without limit. The oldest are
  // dropped because the backend has long since timed out on them, while the
  // newest may still have a caller waiting.
  const q = makeResultQueue(3);
  for (const id of ['a', 'b', 'c', 'd', 'e']) q.enqueue({ id });

  assert.deepEqual(q.drain().map((e) => e.id), ['c', 'd', 'e']);
});

test('the default bound is a real number, so the queue cannot grow unbounded', () => {
  assert.ok(Number.isInteger(PENDING_RESULT_LIMIT) && PENDING_RESULT_LIMIT > 0);
});

test('draining an empty queue is a no-op rather than an error', () => {
  const q = makeResultQueue();
  assert.deepEqual(q.drain(), []);
});

test('flushQueue sends every envelope when the socket accepts them all', () => {
  const q = makeResultQueue();
  for (const id of ['a', 'b', 'c']) q.enqueue({ id });
  const sent = [];
  const count = flushQueue(q, (e) => { sent.push(e.id); return true; });
  assert.equal(count, 3);
  assert.deepEqual(sent, ['a', 'b', 'c']);
  assert.equal(q.size(), 0);
});

test('flushQueue re-queues the failing envelope AND everything after it', () => {
  // The old implementation broke out of the loop after re-queuing only the
  // failing envelope, so 'c' and every later result were pulled from the queue
  // and never returned. A mid-flush socket death used to lose the tail.
  const q = makeResultQueue();
  for (const id of ['a', 'b', 'c', 'd']) q.enqueue({ id });
  let calls = 0;
  const sent = [];
  const count = flushQueue(q, (e) => {
    calls++;
    if (calls === 3) return false;
    sent.push(e.id);
    return true;
  });
  assert.equal(count, 2, 'two envelopes went out before the mid-flush failure');
  assert.deepEqual(sent, ['a', 'b']);
  assert.deepEqual(q.drain().map((e) => e.id), ['c', 'd'],
    'the failing envelope AND everything after it must be re-queued');
});

test('flushQueue on an empty queue is a no-op that does not touch send', () => {
  const q = makeResultQueue();
  let called = 0;
  flushQueue(q, () => { called++; return true; });
  assert.equal(called, 0);
});

// ─── Every content command is bounded ────────────────────────────────────────

test('a content script that never answers yields a timeout error, not a hang', async () => {
  const result = await withTimeout(new Promise(() => {}), 20, 'extract');

  assert.equal(result.success, false);
  assert.match(result.error, /timeout/, 'the backend must see a real error class');
  assert.match(result.error, /extract/, 'the error names the action that hung');
});

test('a prompt answer is returned untouched', async () => {
  const result = await withTimeout(Promise.resolve({ success: true, data: [1, 2] }), 1000, 'extract');
  assert.deepEqual(result, { success: true, data: [1, 2] });
});

test('a rejection becomes a failed result rather than an unhandled rejection', async () => {
  const result = await withTimeout(
    Promise.reject(new Error('receiving end does not exist')),
    1000,
    'extract',
  );
  assert.equal(result.success, false);
  assert.match(result.error, /receiving end/);
});

test('a late answer does not overwrite an already-returned timeout', async () => {
  let resolveLate;
  const late = new Promise((r) => { resolveLate = r; });

  const result = await withTimeout(late, 10, 'extract');
  assert.equal(result.success, false);

  resolveLate({ success: true, data: 'late' });
  await new Promise((r) => setTimeout(r, 20));
  assert.equal(result.success, false, 'the settled result is final');
});

// ─── Navigation failures are not silent ──────────────────────────────────────

test('a completed load is reported as loaded', () => {
  assert.equal(classifyLoadOutcome({ completeEventFired: true }), 'loaded');
});

test('a tab that disappeared is reported as gone, not as loaded', () => {
  // The old code routed chrome.tabs.get(...).catch(finish) into the same
  // resolution as a successful load, so a closed or crashed tab was
  // indistinguishable from a page that finished rendering.
  assert.equal(
    classifyLoadOutcome({ completeEventFired: false, tabLookupFailed: true }),
    'gone',
  );
});

test('a load that never completed is reported as a timeout', () => {
  assert.equal(classifyLoadOutcome({ completeEventFired: false, timedOut: true }), 'timeout');
});

test('every outcome is one of the three known values', () => {
  const inputs = [
    { completeEventFired: true },
    { completeEventFired: false, tabLookupFailed: true },
    { completeEventFired: false, timedOut: true },
    { completeEventFired: false },
  ];
  for (const input of inputs) {
    assert.ok(['loaded', 'gone', 'timeout'].includes(classifyLoadOutcome(input)));
  }
});
