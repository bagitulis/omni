// main-network-observer.js — MAIN-world network observer for Shopee.
//
// WHY MAIN WORLD: content scripts run in an isolated world, so patching
// window.fetch there would only observe the content script's own requests, never
// the page's. The observer must be injected with world:'MAIN' to see Shopee's
// own API calls.
//
// SAFETY CONTRACT (read-only):
//   - Never aborts, blocks, redirects, or rewrites a request.
//   - Never suppresses a request or alters its response.
//   - Observation failures are swallowed: this file must never break the host
//     page, even if the page later replaces fetch or XHR itself.
//   - Bounded memory: body capped, buffer capped, so a long scraping session
//     cannot grow without limit.
//
// Network-first is the primary capture path because Shopee's own responses are
// structured JSON, which is far more robust than DOM selectors. The content
// script falls back to DOM extraction when nothing is captured.

(function () {
  if (window.__omniObserverInstalled) return;
  window.__omniObserverInstalled = true;

  // Body cap: a full listing response is far smaller than this, so the cap only
  // ever truncates something unexpected (for example an HTML error page).
  var BODY_CAP = 200000;
  var BUFFER_MAX = 30;

  // URL fragments that identify Shopee's product/search JSON endpoints.
  var INTERESTING = [
    '/api/v4/search/search_items',
    '/api/v4/recommend/recommend',
    '/api/v4/shop/rcmd_items',
    '/api/v4/pdp/get_pc',
    '/api/v4/item/get',
    'search_items',
    'get_pc',
  ];

  if (!Array.isArray(window.__omniObservedResponses)) {
    window.__omniObservedResponses = [];
  }

  function isInteresting(url) {
    if (!url) return false;
    var lower = String(url).toLowerCase();
    for (var i = 0; i < INTERESTING.length; i++) {
      if (lower.indexOf(INTERESTING[i].toLowerCase()) !== -1) return true;
    }
    return false;
  }

  function record(status, url, body) {
    try {
      if (!isInteresting(url)) return;
      var entry = {
        status: status,
        url: url || '',
        body: (body || '').slice(0, BODY_CAP),
        ts: Date.now(),
      };
      window.__omniObservedResponses.push(entry);
      if (window.__omniObservedResponses.length > BUFFER_MAX) {
        window.__omniObservedResponses.shift();
      }
      window.__omniObservedResponse = entry;
    } catch (e) {
      /* observation must never break the page */
    }
  }

  // ─── Hook fetch ───────────────────────────────────────────────────────────

  var origFetch = window.fetch;
  if (typeof origFetch === 'function') {
    window.fetch = function () {
      var args = arguments;
      var self = this;
      return origFetch.apply(self, args).then(function (res) {
        try {
          var url = (res && res.url) || (args[0] && args[0].url) || String(args[0] || '');
          if (isInteresting(url)) {
            // Clone so reading the body does not consume the page's own copy.
            var clone = res.clone();
            clone.text().then(function (body) {
              record(res.status, url, body);
            }).catch(function () {});
          }
        } catch (e) { /* ignore */ }
        return res;
      });
    };
  }

  // ─── Hook XMLHttpRequest ──────────────────────────────────────────────────

  try {
    var origOpen = XMLHttpRequest.prototype.open;
    var origSend = XMLHttpRequest.prototype.send;

    XMLHttpRequest.prototype.open = function (method, url) {
      this.__omniUrl = url;
      return origOpen.apply(this, arguments);
    };

    XMLHttpRequest.prototype.send = function () {
      var xhr = this;
      xhr.addEventListener('loadend', function () {
        try {
          if (!isInteresting(xhr.__omniUrl)) return;
          var body = '';
          // Only read the body when it was returned as text; touching a blob or
          // arraybuffer response would throw or force an unwanted conversion.
          if (xhr.responseType === '' || xhr.responseType === 'text') {
            body = xhr.responseText || '';
          }
          record(xhr.status, xhr.__omniUrl, body);
        } catch (e) { /* ignore */ }
      });
      return origSend.apply(this, arguments);
    };
  } catch (e) {
    /* if XHR cannot be hooked, fetch capture still works */
  }
})();
