// Test harness that loads the real content script.
//
// The other test files in this directory mirror the logic under test, because
// content-shopee.js is an IIFE that touches `window`, `document`, and `chrome`
// at load time. A mirror cannot fail when production code is wrong — it contains
// its own copy of the answer — so it proves nothing about the shipped file.
//
// This harness stubs the browser surface the script actually uses and then loads
// the real file, capturing the message listener it registers. Commands are then
// dispatched through that listener, exactly as the service worker does.

import { readFileSync } from 'node:fs';
import { fileURLToPath } from 'node:url';
import { dirname, join } from 'node:path';

const here = dirname(fileURLToPath(import.meta.url));
const extensionRoot = join(here, '..', '..');

/**
 * Build a minimal DOM stub.
 *
 * selectorMap maps a CSS selector string to the element (or list) it should
 * return. Anything not present in the map returns null / empty, which is what
 * makes "no pagination control" and "no product grid" expressible.
 */
export function makeDocument({ selectorMap = {}, bodyText = '' } = {}) {
  const resolve = (sel) => selectorMap[sel];

  const doc = {
    body: { innerText: bodyText },
    documentElement: { scrollHeight: 1000 },
    querySelector(sel) {
      const hit = resolve(sel);
      if (!hit) return null;
      return Array.isArray(hit) ? hit[0] || null : hit;
    },
    querySelectorAll(sel) {
      const hit = resolve(sel);
      if (!hit) return [];
      return Array.isArray(hit) ? hit : [hit];
    },
  };
  return doc;
}

/**
 * Build an element stub usable as a selector-map value.
 */
export function makeElement({ text = '', href = '', src = '', className = '', disabled = false, onClick } = {}) {
  const el = {
    textContent: text,
    className,
    disabled,
    getAttribute(name) {
      if (name === 'href') return href;
      if (name === 'src') return src;
      return null;
    },
    querySelector: () => null,
    querySelectorAll: () => [],
    closest: () => null,
    scrollIntoView() {},
    click() {
      if (onClick) onClick();
    },
  };
  return el;
}

/**
 * Load content-shopee.js against stubbed globals and return a command runner.
 *
 * The script is evaluated with `new Function` rather than imported so each test
 * gets a fresh instance: the file's double-load guard would otherwise make every
 * test after the first a no-op.
 */
export function loadContentScript({ document: doc, locationHref = 'https://shopee.co.id/search?keyword=kaos', observedResponses } = {}) {
  const source = readFileSync(join(extensionRoot, 'content-shopee.js'), 'utf8');

  let listener = null;
  const win = {
    __omniObservedResponses: observedResponses,
    location: { href: locationHref },
    scrollY: 0,
    innerHeight: 800,
    scrollTo() {},
  };
  const chromeStub = {
    runtime: {
      onMessage: {
        addListener(fn) {
          listener = fn;
        },
      },
      sendMessage: async () => ({ ok: true }),
    },
  };

  const documentStub = doc || makeDocument();

  // setTimeout is replaced with an immediate callback so the script's settle
  // delays (3s after a pagination click, 900ms between scroll steps) do not make
  // the suite take minutes.
  const fastSetTimeout = (fn) => {
    fn();
    return 0;
  };

  const run = new Function(
    'window', 'document', 'chrome', 'setTimeout',
    `${source}\n`,
  );
  run(win, documentStub, chromeStub, fastSetTimeout);

  if (!listener) {
    throw new Error('content script did not register a message listener');
  }

  return {
    window: win,
    document: documentStub,
    /**
     * Dispatch a command the way the service worker does and resolve its result.
     */
    command(action, payload = {}) {
      return new Promise((resolve, reject) => {
        const returned = listener({ type: 'command', action, payload }, {}, resolve);
        // The listener returns true for an async response; a synchronous handler
        // has already called sendResponse by now.
        if (returned !== true && returned !== false) {
          reject(new Error(`listener returned ${returned}`));
        }
      });
    },
  };
}

/** An element stub with the surface the overlay's DOM building touches. */
function makeOverlayElement() {
  return {
    style: {},
    dataset: {},
    classList: { add() {}, remove() {}, toggle() {} },
    children: [],
    textContent: '',
    shadowRoot: null,
    attachShadow() {
      this.shadowRoot = {
        appendChild() {},
        querySelector: () => null,
        querySelectorAll: () => [],
      };
      return this.shadowRoot;
    },
    appendChild(child) {
      this.children.push(child);
      return child;
    },
    setAttribute() {},
    getAttribute: () => null,
    addEventListener() {},
    removeEventListener() {},
    querySelector: () => null,
    querySelectorAll: () => [],
    remove() {},
    getBoundingClientRect: () => ({ width: 320, height: 400, top: 0, left: 0 }),
  };
}

/**
 * Load overlay.js against stubbed globals and return its test API.
 *
 * The overlay exposes its pure logic on `window.__omniOverlayApi` precisely so
 * it can be exercised without a DOM: the panel needs a real shadow root, but
 * clamping, the ring buffer, and the blocker state machine do not.
 */
export function loadOverlay({ document: doc } = {}) {
  const source = readFileSync(join(extensionRoot, 'overlay.js'), 'utf8');

  const win = {
    location: { href: 'https://shopee.co.id/search?keyword=kaos' },
    innerWidth: 1280,
    innerHeight: 900,
    addEventListener() {},
    removeEventListener() {},
  };

  const documentStub = doc || {
    createElement: () => makeOverlayElement(),
    documentElement: makeOverlayElement(),
    body: makeOverlayElement(),
    addEventListener() {},
    querySelector: () => null,
    querySelectorAll: () => [],
  };

  const chromeStub = {
    runtime: {
      onMessage: { addListener() {} },
      sendMessage: async () => ({ ok: true }),
    },
    storage: {
      local: { get: async () => ({}), set: async () => {} },
    },
  };

  const run = new Function('window', 'document', 'chrome', `${source}\n`);
  run(win, documentStub, chromeStub);

  if (!win.__omniOverlayApi) {
    throw new Error('overlay did not expose __omniOverlayApi');
  }
  return { window: win, api: win.__omniOverlayApi };
}
