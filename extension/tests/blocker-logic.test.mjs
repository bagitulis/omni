// Tests for blocker detection and the tri-state last-page verdict.
//
// These load the REAL content-shopee.js through a stubbed browser surface (see
// helpers/content-harness.mjs), so a regression in the shipped file fails here.
// The older test files in this directory mirror their logic instead; that pattern
// cannot detect a change in production code and is not extended.
//
// What these tests exist to prevent: a captcha interstitial has no pagination
// control, and isLastPage() returned `true` whenever no control was found. A
// blocked scrape therefore ended as a clean success and the operator was shown
// an empty catalogue rather than a challenge to solve.

import test from 'node:test';
import assert from 'node:assert/strict';
import { loadContentScript, makeDocument, makeElement } from './helpers/content-harness.mjs';

// Selectors the content script uses, restated here so a test can express
// "this control is present" without duplicating the detection logic itself.
const NEXT_ENABLED = 'button.shopee-mini-page-controller__next-btn';
const NEXT_DISABLED =
  'button.shopee-button-outline.shopee-mini-page-controller__next-btn.shopee-button-outline--disabled';
const GRID = 'ul.row.shopee-search-item-result__items li';

function productCard(name, href) {
  const card = makeElement({ text: name, href });
  card.querySelector = (sel) => (sel === 'div.line-clamp-2' ? makeElement({ text: name }) : null);
  card.closest = () => makeElement({ href });
  return card;
}

// ─── check_blocked ───────────────────────────────────────────────────────────

test('a captcha interstitial URL is reported as a captcha blocker', async () => {
  const cs = loadContentScript({
    locationHref: 'https://shopee.co.id/verify/traffic?next=%2Fsearch',
    document: makeDocument(),
  });

  const res = await cs.command('check_blocked');
  assert.equal(res.success, true);
  assert.equal(res.data.blocked, true);
  assert.equal(res.data.kind, 'captcha');
  assert.equal(res.data.url, 'https://shopee.co.id/verify/traffic?next=%2Fsearch');
});

test('a login wall is reported as login_required', async () => {
  const cs = loadContentScript({
    locationHref: 'https://shopee.co.id/buyer/login?next=%2Fsearch',
    document: makeDocument(),
  });

  const res = await cs.command('check_blocked');
  assert.equal(res.data.blocked, true);
  assert.equal(res.data.kind, 'login_required');
});

test('classification survives Shopee varying the casing of a path', async () => {
  const cs = loadContentScript({
    locationHref: 'https://Shopee.co.id/VERIFY/Traffic',
    document: makeDocument(),
  });

  const res = await cs.command('check_blocked');
  assert.equal(res.data.kind, 'captcha');
});

test('an ordinary search page is not blocked', async () => {
  const cs = loadContentScript({
    locationHref: 'https://shopee.co.id/search?keyword=kaos',
    document: makeDocument({ selectorMap: { [GRID]: [productCard('Kaos', '/kaos-i.1.2')] } }),
  });

  const res = await cs.command('check_blocked');
  assert.equal(res.data.blocked, false);
  assert.equal(res.data.kind, '');
});

test('a listing whose slug merely contains "login" is not a false positive', async () => {
  // Without anchoring on the real paths, every product named like this would
  // stop the scrape as a login wall.
  const cs = loadContentScript({
    locationHref: 'https://shopee.co.id/Kaos-Logino-Murah-i.123.456',
    document: makeDocument(),
  });

  const res = await cs.command('check_blocked');
  assert.equal(res.data.blocked, false);
});

test('a captcha DOM marker is a blocker even when the URL looks ordinary', async () => {
  const cs = loadContentScript({
    locationHref: 'https://shopee.co.id/search?keyword=kaos',
    document: makeDocument({ selectorMap: { 'iframe[src*="captcha"]': makeElement({}) } }),
  });

  const res = await cs.command('check_blocked');
  assert.equal(res.data.blocked, true);
  assert.equal(res.data.kind, 'anti_bot');
});

test('a login form is a blocker even when the URL looks ordinary', async () => {
  const cs = loadContentScript({
    locationHref: 'https://shopee.co.id/search?keyword=kaos',
    document: makeDocument({ selectorMap: { 'input[type="password"]': makeElement({}) } }),
  });

  const res = await cs.command('check_blocked');
  assert.equal(res.data.blocked, true);
  assert.equal(res.data.kind, 'login_required');
});

test('the blocker payload carries no page text or HTML', async () => {
  // The operator needs the kind and the location. Capturing the body would drag
  // Shopee's content into omni's database and logs for no operational gain.
  const cs = loadContentScript({
    locationHref: 'https://shopee.co.id/verify/traffic',
    document: makeDocument({ bodyText: 'Please verify you are human. Secret session data.' }),
  });

  const res = await cs.command('check_blocked');
  assert.deepEqual(Object.keys(res.data).sort(), ['blocked', 'kind', 'url']);
});

// ─── check_last_page is tri-state ────────────────────────────────────────────

test('a disabled next control proves the last page', async () => {
  const cs = loadContentScript({
    document: makeDocument({
      selectorMap: {
        [NEXT_DISABLED]: makeElement({ className: 'shopee-button-outline--disabled', disabled: true }),
        [GRID]: [productCard('Kaos', '/kaos-i.1.2')],
      },
    }),
  });

  const res = await cs.command('check_last_page');
  assert.equal(res.data.is_last, true);
  assert.equal(res.data.reason, 'next_disabled');
});

test('an enabled next control proves there is more to fetch', async () => {
  const cs = loadContentScript({
    document: makeDocument({
      selectorMap: {
        [NEXT_ENABLED]: makeElement({ className: '' }),
        [GRID]: [productCard('Kaos', '/kaos-i.1.2')],
      },
    }),
  });

  const res = await cs.command('check_last_page');
  assert.equal(res.data.is_last, false);
  assert.equal(res.data.reason, 'next_enabled');
});

test('no pagination WITH a product grid is a genuine single page', async () => {
  const cs = loadContentScript({
    document: makeDocument({ selectorMap: { [GRID]: [productCard('Kaos', '/kaos-i.1.2')] } }),
  });

  const res = await cs.command('check_last_page');
  assert.equal(res.data.is_last, true);
  assert.equal(res.data.reason, 'single_page');
});

test('no pagination and NO product grid is unknown, never "last page"', async () => {
  // This is the captcha / login-wall / block-page shape, and the original defect:
  // the scrape stopped here and was recorded as a completed run over an empty
  // catalogue.
  const cs = loadContentScript({ document: makeDocument() });

  const res = await cs.command('check_last_page');
  assert.equal(res.data.is_last, false, 'a page with no grid must not end the scrape as a success');
  assert.equal(res.data.reason, 'unknown');
});

test('every verdict carries a recognised reason', async () => {
  const scenarios = [
    { selectorMap: { [NEXT_DISABLED]: makeElement({ disabled: true }) } },
    { selectorMap: { [NEXT_ENABLED]: makeElement({}) } },
    { selectorMap: { [GRID]: [productCard('Kaos', '/kaos-i.1.2')] } },
    { selectorMap: {} },
  ];
  for (const scenario of scenarios) {
    const cs = loadContentScript({ document: makeDocument(scenario) });
    const res = await cs.command('check_last_page');
    assert.equal(typeof res.data.is_last, 'boolean');
    assert.ok(
      ['next_disabled', 'next_enabled', 'single_page', 'unknown'].includes(res.data.reason),
      `unexpected reason ${res.data.reason}`,
    );
  }
});

// ─── extract reports whether any selector set matched ────────────────────────

test('extract reports containers_matched when a grid was found', async () => {
  const cs = loadContentScript({
    document: makeDocument({ selectorMap: { [GRID]: [productCard('Kaos', '/kaos-i.1.2')] } }),
  });

  const res = await cs.command('extract');
  assert.equal(res.success, true);
  assert.equal(res.containers_matched, true);
  assert.equal(res.data.length, 1);
});

test('extract reports containers_matched:false when every selector set missed', async () => {
  // This is what separates a Shopee redesign from an empty shop. Without it the
  // backend cannot tell "no product grid exists" from "the grid is empty", and a
  // markup change is reported as a successful scrape of nothing.
  const cs = loadContentScript({ document: makeDocument() });

  const res = await cs.command('extract');
  assert.equal(res.success, true);
  assert.equal(res.containers_matched, false);
  assert.deepEqual(res.data, []);
});

test('extract distinguishes an empty grid from a missing one', async () => {
  // Containers are present but every card is empty: the page is genuinely empty
  // and the selectors are fine.
  const emptyCard = makeElement({ text: '', href: '' });
  const cs = loadContentScript({
    document: makeDocument({ selectorMap: { [GRID]: [emptyCard] } }),
  });

  const res = await cs.command('extract');
  assert.equal(res.containers_matched, true, 'containers existed, so the selectors are not broken');
  assert.deepEqual(res.data, []);
});

// ─── click_next verifies the page actually advanced ──────────────────────────

test('click_next reports a no-op when the URL did not change', async () => {
  // The old implementation slept 3s and returned success unconditionally, so a
  // dead click made the loop re-capture the same page.
  const cs = loadContentScript({
    locationHref: 'https://shopee.co.id/search?keyword=kaos',
    document: makeDocument({ selectorMap: { [NEXT_ENABLED]: makeElement({ className: '' }) } }),
  });

  const res = await cs.command('click_next');
  assert.equal(res.success, false);
  assert.match(res.error, /did not advance/i);
});

test('click_next succeeds when the click advanced the page', async () => {
  const win = { href: 'https://shopee.co.id/search?keyword=kaos' };
  const next = makeElement({
    className: '',
    onClick() {
      win.href = 'https://shopee.co.id/search?keyword=kaos&page=1';
    },
  });

  const cs = loadContentScript({
    locationHref: win.href,
    document: makeDocument({ selectorMap: { [NEXT_ENABLED]: next } }),
  });
  // Point the script's location at the same mutable object the click updates.
  Object.defineProperty(cs.window.location, 'href', {
    get: () => win.href,
    configurable: true,
  });

  const res = await cs.command('click_next');
  assert.equal(res.success, true);
});

test('click_next still reports a missing control as such', async () => {
  const cs = loadContentScript({ document: makeDocument() });

  const res = await cs.command('click_next');
  assert.equal(res.success, false);
  assert.match(res.error, /not found/i);
});
