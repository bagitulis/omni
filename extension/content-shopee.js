// content-shopee.js — in-page Shopee automation.
//
// Runs in the isolated content-script world. Two things follow from that:
//
//   1. Navigation, tab lifecycle, and `wait` are handled by the service worker,
//      because this script is destroyed whenever the page navigates.
//   2. The page's own fetch/XHR cannot be hooked from here. Network capture is
//      delegated to main-network-observer.js, injected into the MAIN world by
//      the service worker; this script only READS the buffer that observer
//      writes.
//
// Double-injection guard: the manifest and a programmatic executeScript can both
// load this file, and registering the listener twice would double-handle every
// command.

(function () {
  if (window.__omniShopeeLoaded) return;
  window.__omniShopeeLoaded = true;

  const sleep = (ms) => new Promise((r) => setTimeout(r, ms));

  // ─── Element resolution ───────────────────────────────────────────────────

  /**
   * Find the first element matching any selector.
   *
   * Selectors are tried in order because Shopee changes its markup without
   * notice; a single selector would break the whole scrape.
   */
  function findFirst(selectors, root) {
    const scope = root || document;
    for (const sel of selectors || []) {
      if (!sel) continue;
      try {
        const el = scope.querySelector(sel);
        if (el) return el;
      } catch {
        // An invalid selector must not abort the remaining candidates.
        continue;
      }
    }
    return null;
  }

  function textOf(el) {
    return el ? (el.textContent || '').trim() : '';
  }

  // ─── Product extraction ───────────────────────────────────────────────────

  // Fallback selector sets for search-result cards. Ordered most-specific first;
  // the first set that yields any rows wins.
  const PRODUCT_SELECTOR_SETS = [
    {
      container: 'ul.row.shopee-search-item-result__items li',
      name: ['div.line-clamp-2'],
      price: ['span.truncate.text-base\\/5.font-medium'],
      sold: ['div.truncate.text-shopee-black87.text-xs'],
      link: ['a'],
      image: ['img'],
    },
    {
      container: "div[data-sqe='item'] a",
      name: ["div[data-sqe='name']"],
      price: ['span[aria-label]'],
      sold: ["div[class*='sold']"],
      link: ['a'],
      image: ['img'],
    },
    {
      container: 'li.shopee-search-item-result__item',
      name: ["div[data-sqe='name']"],
      price: ['span.font-medium'],
      sold: ['div.text-xs'],
      link: ['a'],
      image: ['img'],
    },
  ];

  /**
   * Extract product rows from the current document.
   *
   * Every field is best-effort: a card missing its price still yields a row, so
   * one markup change does not silently discard the whole page.
   */
  function extractProducts() {
    for (const set of PRODUCT_SELECTOR_SETS) {
      let containers;
      try {
        containers = Array.from(document.querySelectorAll(set.container));
      } catch {
        continue;
      }
      if (containers.length === 0) continue;

      const rows = [];
      for (const card of containers) {
        const nameEl = findFirst(set.name, card);
        const priceEl = findFirst(set.price, card);
        const soldEl = findFirst(set.sold, card);
        const linkEl = findFirst(set.link, card) || card.closest('a[href]');
        const imgEl = findFirst(set.image, card);

        const name = textOf(nameEl);
        const link = linkEl ? linkEl.getAttribute('href') || '' : '';
        if (!name && !link) continue; // an empty card is not a product

        rows.push({
          name,
          price: textOf(priceEl),
          sold: textOf(soldEl),
          link: link.startsWith('http') ? link : link ? 'https://shopee.co.id' + link : '',
          image_url: imgEl ? imgEl.getAttribute('src') || '' : '',
        });
      }

      if (rows.length > 0) return rows;
    }
    return [];
  }

  // ─── Pagination detection ─────────────────────────────────────────────────

  // Mini-pagination next button, and its disabled variant. The disabled variant
  // is what proves we are on the last page.
  const NEXT_ENABLED = [
    'button.shopee-mini-page-controller__next-btn',
    'a.shopee-mini-page-controller__next-btn',
  ];
  const NEXT_DISABLED = [
    'button.shopee-button-outline.shopee-mini-page-controller__next-btn.shopee-button-outline--disabled',
    'a.shopee-button-outline.shopee-button-outline--disabled.shopee-mini-page-controller__next-btn',
  ];
  const NEXT_BIG = ['a.shopee-icon-button.shopee-icon-button--right'];
  const NEXT_BIG_DISABLED = ['a.shopee-icon-button.shopee-icon-button--right.shopee-icon-button--disabled'];

  function isLastPage() {
    const disabled = findFirst(NEXT_DISABLED) || findFirst(NEXT_BIG_DISABLED);
    if (disabled && !isClickable(disabled)) return true;

    // No pagination control at all means a single page of results.
    const enabled = findFirst(NEXT_ENABLED) || findFirst(NEXT_BIG);
    if (!enabled) return true;

    return false;
  }

  function isClickable(el) {
    if (!el) return false;
    if (el.disabled) return false;
    const cls = el.className || '';
    return !String(cls).includes('disabled');
  }

  // ─── Commands ─────────────────────────────────────────────────────────────

  /**
   * Incremental scroll until the bottom is reached.
   *
   * Shopee lazy-loads product images and cards, so extracting before scrolling
   * yields a partial page. The loop stops when the document stops growing,
   * rather than after a fixed count, so a short page is not scrolled twenty
   * times and a long one is not cut off.
   */
  async function smartScroll({ step = 800, waitTime = 900, maxSteps = 25 } = {}) {
    let lastHeight = 0;
    let stable = 0;

    for (let i = 0; i < maxSteps; i++) {
      window.scrollTo({ top: window.scrollY + step, behavior: 'smooth' });
      await sleep(waitTime);

      const height = document.documentElement.scrollHeight;
      if (height === lastHeight) {
        stable++;
        // Two consecutive non-growth checks: one is often just a slow network.
        if (stable >= 2) break;
      } else {
        stable = 0;
      }
      lastHeight = height;

      if (window.scrollY + window.innerHeight >= height - 2) break;
    }

    window.scrollTo(0, document.documentElement.scrollHeight);
    await sleep(500);

    return { success: true, data: { finalHeight: document.documentElement.scrollHeight } };
  }

  async function clickNext() {
    const el = findFirst(NEXT_BIG) || findFirst(NEXT_ENABLED);
    if (!el) return { success: false, error: 'next page control not found' };
    if (!isClickable(el)) return { success: false, error: 'next page control is disabled (last page)' };

    el.scrollIntoView({ block: 'center' });
    await sleep(200);
    el.click();
    // Shopee paginates client-side; a short settle avoids reading the previous
    // page's DOM.
    await sleep(3000);
    return { success: true, data: null };
  }

  function getUrl() {
    return { success: true, data: { url: window.location.href } };
  }

  function extract() {
    return { success: true, data: extractProducts() };
  }

  function checkLastPage() {
    return { success: true, data: { is_last: isLastPage() } };
  }

  function checkText({ text = '' } = {}) {
    return { success: true, data: { found: (document.body.innerText || '').includes(text) } };
  }

  // ─── Network capture (reads the MAIN-world buffer) ─────────────────────────

  /**
   * Arm the MAIN-world observer.
   *
   * Must be called BEFORE the request whose response we want: the observer hooks
   * fetch/XHR when it loads, so anything already in flight is missed. The service
   * worker injects it with world:'MAIN' because an isolated-world hook would only
   * see this script's own requests.
   */
  async function installObserver() {
    try {
      const resp = await chrome.runtime.sendMessage({ type: 'install_main_observer' });
      if (resp && resp.ok === false) {
        return { success: false, error: 'observer install failed: ' + resp.error };
      }
      return { success: true, data: { ok: true } };
    } catch (err) {
      return { success: false, error: 'observer install unavailable: ' + String(err) };
    }
  }

  /**
   * Read captured responses matching a URL substring.
   *
   * Returns every match, oldest first, so the caller can join a paginated
   * sequence. Read-only: it never modifies, blocks, or delays a request.
   */
  function observeNetwork({ filter = '', limit = 10 } = {}) {
    const buffer = Array.isArray(window.__omniObservedResponses) ? window.__omniObservedResponses : [];
    let matches = buffer;
    if (filter) {
      const needle = String(filter).toLowerCase();
      matches = buffer.filter((r) => String(r.url || '').toLowerCase().includes(needle));
    }
    return {
      success: true,
      data: {
        count: matches.length,
        responses: matches.slice(-limit).map((r) => ({
          status: r.status || 0,
          url: r.url || '',
          body: r.body || '',
        })),
      },
    };
  }

  // ─── Dispatch ─────────────────────────────────────────────────────────────

  const COMMANDS = {
    extract: extract,
    smart_scroll: smartScroll,
    click_next: clickNext,
    get_url: getUrl,
    check_last_page: checkLastPage,
    check_text: checkText,
    install_observer: installObserver,
    observe_network: observeNetwork,
  };

  chrome.runtime.onMessage.addListener((message, _sender, sendResponse) => {
    if (message?.type !== 'command') return false;

    const handler = COMMANDS[message.action];
    if (!handler) {
      sendResponse({ success: false, error: 'unsupported action: ' + message.action });
      return false;
    }

    Promise.resolve()
      .then(() => handler(message.payload || {}))
      .then((result) => sendResponse(result))
      .catch((err) => sendResponse({ success: false, error: String(err) }));

    return true; // async response
  });
})();
