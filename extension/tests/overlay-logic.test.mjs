// Tests for the operator overlay's pure logic and its safety gates.
//
// The overlay is loaded the same way the content script is: through a stubbed
// browser surface, against the real shipped file. See helpers/content-harness.mjs.
//
// The safety gates are not decoration. The overlay runs inside a page the
// operator is logged into, so a module that reads cookies, submits a host-page
// form, or blocks on alert() would turn a debugging aid into a liability.

import test from 'node:test';
import assert from 'node:assert/strict';
import { readFileSync } from 'node:fs';
import { fileURLToPath } from 'node:url';
import { dirname, join } from 'node:path';

import { loadOverlay } from './helpers/content-harness.mjs';

const here = dirname(fileURLToPath(import.meta.url));
const overlaySource = readFileSync(join(here, '..', 'overlay.js'), 'utf8');

/**
 * The source with comments removed.
 *
 * The safety gates below assert on what the overlay DOES, so they must not fire
 * on a comment that names the forbidden thing in order to explain why it is
 * forbidden — that would push the reasoning out of the file to satisfy a test.
 */
const overlayCode = overlaySource
  .replace(/\/\*[\s\S]*?\*\//g, '')
  .replace(/(^|[^:])\/\/.*$/gm, '$1');

// ─── Safety gates ────────────────────────────────────────────────────────────

test('the overlay never reads document.cookie', () => {
  // The service worker owns cookie access. A content script reading them would
  // widen the blast radius of any injection on the host page.
  assert.ok(!/document\s*\.\s*cookie/.test(overlayCode));
});

test('the overlay never calls alert, confirm, or prompt', () => {
  // These block the page. During a scrape that means the run stalls behind a
  // dialog nobody is watching.
  assert.ok(!/(^|[^.\w])(alert|confirm|prompt)\s*\(/.test(overlayCode));
});

test('the overlay never submits or clicks host-page forms', () => {
  assert.ok(!/\.submit\s*\(/.test(overlayCode));
  assert.ok(!/querySelector\([^)]*form[^)]*\)\s*\.\s*click/.test(overlayCode));
});

test('the overlay keeps its double-load guard', () => {
  // The manifest and a programmatic injection can both load the file; without
  // the guard the operator gets two stacked panels and doubled listeners.
  assert.match(overlayCode, /window\.__omniOverlayLoaded/);
});

test('the overlay uses a shadow root rather than styling the host page', () => {
  // Shopee's own CSS would otherwise fight the panel, and the panel's CSS would
  // leak into the page the operator is looking at.
  assert.match(overlayCode, /attachShadow/);
});

test('the comment-stripping the gates rely on actually removes comments', () => {
  // Without this, a gate could silently pass because the stripper ate the code
  // it was supposed to inspect, or fail because it left a comment behind.
  const stripped = `x\n// document.cookie\n/* alert( */\ny`
    .replace(/\/\*[\s\S]*?\*\//g, '')
    .replace(/(^|[^:])\/\/.*$/gm, '$1');
  assert.ok(!/document\s*\.\s*cookie/.test(stripped));
  assert.ok(!/alert\s*\(/.test(stripped));
  assert.match(stripped, /x/);
  assert.match(stripped, /y/);
});

// ─── Position clamping ───────────────────────────────────────────────────────

test('a dragged panel is clamped inside the viewport', () => {
  const o = loadOverlay();
  const clamp = o.api.clampPosition;

  const viewport = { width: 1000, height: 800 };
  const size = { width: 320, height: 400 };

  assert.deepEqual(clamp({ x: -50, y: -50 }, size, viewport), { x: 0, y: 0 },
    'a panel dragged off the top-left must stay reachable');
  assert.deepEqual(clamp({ x: 5000, y: 5000 }, size, viewport), { x: 680, y: 400 },
    'a panel dragged off the bottom-right must stay reachable');
  assert.deepEqual(clamp({ x: 100, y: 100 }, size, viewport), { x: 100, y: 100 },
    'a position already inside the viewport is unchanged');
});

test('a panel larger than the viewport is pinned to the origin, not hidden', () => {
  const o = loadOverlay();
  const clamp = o.api.clampPosition;
  assert.deepEqual(clamp({ x: 50, y: 50 }, { width: 900, height: 900 }, { width: 400, height: 300 }), { x: 0, y: 0 });
});

// ─── Debug ring buffer ───────────────────────────────────────────────────────

test('the debug log is a bounded ring buffer', () => {
  // An unbounded log in a long-running tab is a memory leak that only shows up
  // during exactly the long scrapes it exists to diagnose.
  const o = loadOverlay();
  const log = o.api.makeRingBuffer(3);

  for (const entry of ['a', 'b', 'c', 'd', 'e']) log.push(entry);

  assert.deepEqual(log.entries(), ['c', 'd', 'e']);
  assert.equal(log.entries().length, 3);
});

test('the ring buffer keeps insertion order', () => {
  const o = loadOverlay();
  const log = o.api.makeRingBuffer(10);
  log.push('first');
  log.push('second');
  assert.deepEqual(log.entries(), ['first', 'second']);
});

// ─── Blocker panel state ─────────────────────────────────────────────────────

test('the resume action is disabled until the operator confirms', () => {
  // Resuming while the captcha is still up just burns another page and blocks
  // again, which reads to the operator as the resume being broken.
  const o = loadOverlay();
  const state = o.api.blockerPanelState({ blocked: true, kind: 'captcha', confirmed: false });
  assert.equal(state.resumeEnabled, false);
});

test('the resume action is enabled once the operator confirms', () => {
  const o = loadOverlay();
  const state = o.api.blockerPanelState({ blocked: true, kind: 'captcha', confirmed: true });
  assert.equal(state.resumeEnabled, true);
});

test('the blocker panel is hidden when nothing is blocked', () => {
  const o = loadOverlay();
  const state = o.api.blockerPanelState({ blocked: false, kind: '', confirmed: false });
  assert.equal(state.visible, false);
  assert.equal(state.resumeEnabled, false, 'confirmation cannot enable a resume with no block');
});

test('an unknown blocker kind is still shown to the operator', () => {
  const o = loadOverlay();
  const state = o.api.blockerPanelState({ blocked: true, kind: 'some_new_wall', confirmed: true });
  assert.equal(state.visible, true);
  assert.equal(state.label, 'some_new_wall');
});

// ─── Tab selection ───────────────────────────────────────────────────────────

test('the overlay exposes exactly the four operator tabs', () => {
  const o = loadOverlay();
  assert.deepEqual(o.api.tabIds(), ['scrape', 'blocker', 'cookie', 'debug']);
});

test('an unknown tab id falls back to scrape rather than rendering nothing', () => {
  const o = loadOverlay();
  assert.equal(o.api.resolveTab('nope'), 'scrape');
  assert.equal(o.api.resolveTab('debug'), 'debug');
});
