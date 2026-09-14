// Pure-logic tests for the extension's URL handling and domain allow-listing.
//
// These are extracted from service-worker.js as standalone functions with the
// same bodies, because the worker itself cannot be imported outside Chrome (it
// references `chrome.*` and `WebSocket` at module scope). Keeping these as a
// separate, tested pair is the honest trade-off: it verifies the logic without
// pretending to verify the worker's browser integration.
//
// If you change the logic in service-worker.js, change it here too. The
// duplication is deliberate and small; a shared module would need the chrome
// APIs stubbed, which is a larger surface than it removes.

import test from 'node:test';
import assert from 'node:assert/strict';

/** Mirror of buildWsUrl in service-worker.js. */
function buildWsUrl(serverUrl) {
  const trimmed = (serverUrl || '').trim();
  if (!trimmed) return '';

  let base = trimmed.replace(/\/+$/, '');
  base = base.replace(/^http:\/\//i, 'ws://').replace(/^https:\/\//i, 'wss://');
  if (!/^wss?:\/\//i.test(base)) {
    base = 'wss://' + base;
  }
  return base + '/api/extensions/ws';
}

/** Mirror of isSupportedUrl in service-worker.js. */
const SUPPORTED_HOSTS = new Set(['shopee.co.id', 'www.shopee.co.id']);

function isSupportedUrl(url) {
  if (typeof url !== 'string' || !url) return false;

  let parsed;
  try {
    parsed = new URL(url);
  } catch {
    return false;
  }

  if (parsed.protocol === 'https:' && SUPPORTED_HOSTS.has(parsed.hostname)) {
    return true;
  }
  return parsed.protocol === 'http:' && parsed.hostname === 'localhost';
}

test('buildWsUrl converts https to wss', () => {
  assert.equal(buildWsUrl('https://app.example.com'), 'wss://app.example.com/api/extensions/ws');
});

test('buildWsUrl converts http to ws for local development', () => {
  assert.equal(buildWsUrl('http://localhost:8080'), 'ws://localhost:8080/api/extensions/ws');
});

test('buildWsUrl strips trailing slashes', () => {
  assert.equal(buildWsUrl('https://app.example.com///'), 'wss://app.example.com/api/extensions/ws');
});

test('buildWsUrl tolerates surrounding whitespace', () => {
  assert.equal(buildWsUrl('  https://app.example.com  '), 'wss://app.example.com/api/extensions/ws');
});

test('buildWsUrl assumes wss for a bare host', () => {
  // A bare host must not silently become an insecure ws:// connection.
  assert.equal(buildWsUrl('app.example.com'), 'wss://app.example.com/api/extensions/ws');
});

test('buildWsUrl returns empty for missing or empty input', () => {
  assert.equal(buildWsUrl(''), '');
  assert.equal(buildWsUrl(undefined), '');
  assert.equal(buildWsUrl(null), '');
  assert.equal(buildWsUrl('   '), '');
});

test('buildWsUrl preserves an explicit ws:// scheme', () => {
  assert.equal(buildWsUrl('ws://localhost:8080'), 'ws://localhost:8080/api/extensions/ws');
});

test('isSupportedUrl accepts Shopee and localhost', () => {
  assert.ok(isSupportedUrl('https://shopee.co.id/search?keyword=x'));
  assert.ok(isSupportedUrl('https://www.shopee.co.id/product/1'));
  assert.ok(isSupportedUrl('http://localhost:5173/dev'));
});

test('isSupportedUrl rejects unrelated sites', () => {
  // The extension must not be drivable against whatever the user has open.
  assert.equal(isSupportedUrl('https://example.com'), false);
  assert.equal(isSupportedUrl('https://evil.com/shopee.co.id'), false);
  assert.equal(isSupportedUrl('https://tokopedia.com'), false);
});

test('isSupportedUrl rejects lookalike domains', () => {
  // A prefix compare alone would accept these; the trailing dot/hyphen cases
  // are the ones worth pinning.
  assert.equal(isSupportedUrl('https://shopee.co.id.evil.com'), false);
  assert.equal(isSupportedUrl('https://notshopee.co.id'), false);
  assert.equal(isSupportedUrl('https://www.shopee.co.id.evil.com'), false);
});

test('isSupportedUrl rejects non-string input', () => {
  assert.equal(isSupportedUrl(undefined), false);
  assert.equal(isSupportedUrl(null), false);
  assert.equal(isSupportedUrl(123), false);
  assert.equal(isSupportedUrl(''), false);
});

test('isSupportedUrl does not accept plain http for Shopee', () => {
  // Shopee is https-only in practice; accepting http would allow a downgrade.
  assert.equal(isSupportedUrl('http://shopee.co.id'), false);
});
