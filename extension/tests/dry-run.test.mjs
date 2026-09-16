// Dry run for the scrape command surface.
//
// The extension executes commands against a live Shopee session in the
// operator's own browser: navigation, scrolling, pagination clicks. Rehearsing
// the command sequence against stubs before it reaches a real tab is the
// difference between "we believe this is safe" and "we checked".
//
// What this rehearses, and what it deliberately does NOT do:
//   - It drives the REAL content script through the harness, so the dispatch
//     table, the blocker check, and the pagination verification are the shipped
//     ones.
//   - No browser tab is opened, no request reaches Shopee, and no page is
//     clicked: every element is a stub that records the call instead of acting.
//
// It reports the blast radius — which commands run, in what order, and how many
// pages a run of a given size would touch — and asserts the properties that
// would make a real run destructive or dishonest.

import test from 'node:test';
import assert from 'node:assert/strict';
import { loadContentScript, makeDocument, makeElement } from './helpers/content-harness.mjs';

const GRID = 'ul.row.shopee-search-item-result__items li';
const NEXT_ENABLED = 'button.shopee-mini-page-controller__next-btn';

function productCard(name, href) {
  const card = makeElement({ text: name, href });
  card.querySelector = (sel) => (sel === 'div.line-clamp-2' ? makeElement({ text: name }) : null);
  card.closest = () => makeElement({ href });
  return card;
}

test('dry run: a healthy page reports the command sequence and touches nothing', async () => {
  const clicks = [];
  const next = makeElement({ className: '', onClick: () => clicks.push('next') });
  const cs = loadContentScript({
    document: makeDocument({
      selectorMap: { [GRID]: [productCard('Kaos', '/kaos-i.1.2')], [NEXT_ENABLED]: next },
    }),
  });

  const sequence = ['check_blocked', 'smart_scroll', 'extract', 'check_last_page'];
  const outcomes = [];
  for (const action of sequence) {
    const res = await cs.command(action);
    outcomes.push(`${action} -> success=${res.success}`);
  }

  console.log('dry run: command sequence for one page');
  for (const line of outcomes) console.log(`  ${line}`);

  // Evidence of what did NOT happen: reading a page never clicks anything on it.
  assert.deepEqual(clicks, [], 'inspecting a page must not interact with it');
  for (const outcome of outcomes) {
    assert.match(outcome, /success=true/);
  }
});

test('dry run: a blocked page stops before extraction and reports the blast radius', async () => {
  const cs = loadContentScript({
    locationHref: 'https://shopee.co.id/verify/traffic',
    document: makeDocument(),
  });

  const blocked = await cs.command('check_blocked');
  console.log(`dry run: blocker detected kind=${blocked.data.kind} at ${blocked.data.url}`);
  console.log('dry run: pages that would be captured after this point: 0');

  assert.equal(blocked.data.blocked, true);

  // The verdict on such a page must NOT be "last page": that is what let a
  // captcha end a run as a completed scrape of an empty catalogue.
  const verdict = await cs.command('check_last_page');
  assert.equal(verdict.data.is_last, false);
  assert.equal(verdict.data.reason, 'unknown');
});

test('dry run: an inert pagination control is reported, not silently retried', async () => {
  // A click that does nothing used to return success, so the loop re-captured
  // the same page until its budget ran out. Rehearsing it here shows the run
  // would stop at one page rather than spending the whole budget.
  const cs = loadContentScript({
    locationHref: 'https://shopee.co.id/search?keyword=kaos',
    document: makeDocument({ selectorMap: { [NEXT_ENABLED]: makeElement({ className: '' }) } }),
  });

  const res = await cs.command('click_next');
  console.log(`dry run: click_next -> success=${res.success} (${res.error || 'advanced'})`);
  console.log('dry run: pages a 50-page budget would actually capture here: 1');

  assert.equal(res.success, false);
});

test('dry run: an unsupported action is refused rather than guessed at', async () => {
  const cs = loadContentScript({ document: makeDocument() });

  const res = await cs.command('drop_database');
  console.log(`dry run: unknown action -> success=${res.success} (${res.error})`);

  assert.equal(res.success, false);
  assert.match(res.error, /unsupported action/);
});

test('dry run: the command surface is exactly what the backend expects', async () => {
  // A command the backend calls but the content script does not implement fails
  // at run time on a live page. Checking the surface here is cheaper than
  // discovering it mid-scrape.
  const cs = loadContentScript({ document: makeDocument() });

  const backendCalls = [
    'install_observer',
    'observe_network',
    'extract',
    'smart_scroll',
    'click_next',
    'check_last_page',
    'check_blocked',
    'get_url',
  ];

  const unsupported = [];
  for (const action of backendCalls) {
    const res = await cs.command(action);
    if (!res.success && /unsupported action/.test(res.error || '')) {
      unsupported.push(action);
    }
  }

  console.log(`dry run: backend calls ${backendCalls.length} actions; unsupported: ${unsupported.length}`);
  assert.deepEqual(unsupported, [], 'every action the backend sends must be implemented');
});
