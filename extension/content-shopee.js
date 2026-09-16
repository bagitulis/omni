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
   *
   * Returns the rows AND whether any selector set matched a product grid. The
   * flag is what separates "this shop is empty" from "Shopee changed its
   * markup": without it both look like a successful scrape of nothing, and a
   * redesign goes unnoticed until someone wonders where the catalogue went.
   */
  function extractProducts() {
    let containersMatched = false;

    for (const set of PRODUCT_SELECTOR_SETS) {
      let containers;
      try {
        containers = Array.from(document.querySelectorAll(set.container));
      } catch {
        continue;
      }
      if (containers.length === 0) continue;
      containersMatched = true;

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

      if (rows.length > 0) return { rows, containersMatched: true };
    }
    return { rows: [], containersMatched };
  }

  /**
   * Whether a product grid is present on the page at all.
   *
   * Used to decide whether "no pagination control" means a single page of
   * results or a page that is not a result list at all.
   */
  function hasProductGrid() {
    for (const set of PRODUCT_SELECTOR_SETS) {
      try {
        if (document.querySelectorAll(set.container).length > 0) return true;
      } catch {
        continue;
      }
    }
    return false;
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

  /**
   * Decide whether this is the last page, and say how that was decided.
   *
   * Tri-state on purpose. The previous version returned `true` whenever no
   * pagination control was found, which is exactly the shape of a captcha
   * interstitial, a login wall, and a block page — none of which have pagination
   * OR a product grid. A blocked scrape therefore ended as a clean success.
   *
   * `single_page` is only claimed when the product grid is actually present.
   * Otherwise the answer is `unknown`, which does not stop the run: the blocker
   * check is what decides that case.
   */
  function lastPageVerdict() {
    const disabled = findFirst(NEXT_DISABLED) || findFirst(NEXT_BIG_DISABLED);
    if (disabled && !isClickable(disabled)) {
      return { is_last: true, reason: 'next_disabled' };
    }

    const enabled = findFirst(NEXT_ENABLED) || findFirst(NEXT_BIG);
    if (enabled && isClickable(enabled)) {
      return { is_last: false, reason: 'next_enabled' };
    }

    if (hasProductGrid()) {
      return { is_last: true, reason: 'single_page' };
    }
    return { is_last: false, reason: 'unknown' };
  }

  function isClickable(el) {
    if (!el) return false;
    if (el.disabled) return false;
    const cls = el.className || '';
    return !String(cls).includes('disabled');
  }

  // ─── Blocker detection ────────────────────────────────────────────────────

  // Paths Shopee redirects to when it interposes a challenge. Matched against
  // the URL rather than the page text, because text is localised and changes far
  // more often than these routes do.
  const BLOCKER_URL_PATTERNS = [
    { kind: 'captcha', patterns: ['/verify/traffic', '/verify/captcha', '/verify/bot'] },
    { kind: 'login_required', patterns: ['/buyer/login', '/login?next=', '/buyer/signup'] },
    { kind: 'rate_limited', patterns: ['/verify/rate'] },
  ];

  // DOM markers that indicate a challenge even when the URL looks ordinary:
  // Shopee sometimes renders the captcha in place rather than redirecting.
  const CAPTCHA_MARKERS = [
    'iframe[src*="captcha"]',
    'iframe[src*="recaptcha"]',
    'div[class*="captcha"]',
    '#captcha',
  ];
  const LOGIN_FORM_MARKERS = ['input[type="password"]', 'form[action*="login"]'];

  /**
   * Classify a URL as a known blocker route, or '' when it is ordinary.
   *
   * The patterns are anchored on real paths rather than bare words: a product
   * slug containing "login" must not stop a scrape.
   */
  function classifyBlockerUrl(url) {
    const lower = String(url || '').toLowerCase();
    for (const { kind, patterns } of BLOCKER_URL_PATTERNS) {
      for (const pattern of patterns) {
        if (lower.includes(pattern)) return kind;
      }
    }
    return '';
  }

  /**
   * Report whether the page is a challenge the operator has to clear.
   *
   * Returns the kind and the location only. Page text and HTML are deliberately
   * never included: the operator needs to know what kind of wall it is and
   * where, and capturing the body would pull Shopee's content into omni's
   * database and logs for no operational gain.
   */
  function checkBlocked() {
    const url = window.location.href;

    let kind = classifyBlockerUrl(url);
    if (!kind && findFirst(CAPTCHA_MARKERS)) kind = 'anti_bot';
    if (!kind && findFirst(LOGIN_FORM_MARKERS)) kind = 'login_required';

    return { success: true, data: { blocked: Boolean(kind), kind, url } };
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

  /**
   * Advance to the next page and confirm it actually happened.
   *
   * The click is verified rather than assumed: Shopee paginates client-side, and
   * a control that is present but inert used to report success. The page loop
   * would then re-capture the same page, which its own deduplication hid — so a
   * scrape stopped early and still looked complete.
   */
  async function clickNext() {
    const el = findFirst(NEXT_BIG) || findFirst(NEXT_ENABLED);
    if (!el) return { success: false, error: 'next page control not found' };
    if (!isClickable(el)) return { success: false, error: 'next page control is disabled (last page)' };

    const before = window.location.href;

    el.scrollIntoView({ block: 'center' });
    await sleep(200);
    el.click();
    // Shopee paginates client-side; a short settle avoids reading the previous
    // page's DOM.
    await sleep(3000);

    const after = window.location.href;
    if (!after || after === before) {
      return { success: false, error: 'clicked next but the page did not advance' };
    }
    return { success: true, data: { url: after } };
  }

  function getUrl() {
    return { success: true, data: { url: window.location.href } };
  }

  /**
   * Extract the page's products.
   *
   * containers_matched rides alongside the rows rather than inside `data`,
   * because the backend unwraps the envelope's `data` member and would discard a
   * sibling placed there.
   */
  function extract() {
    const { rows, containersMatched } = extractProducts();
    return { success: true, data: rows, containers_matched: containersMatched };
  }

  function checkLastPage() {
    return { success: true, data: lastPageVerdict() };
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
    check_blocked: checkBlocked,
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
