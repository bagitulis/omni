// overlay.js — operator panel for Shopee scrape runs.
//
// A Shadow-DOM panel injected into Shopee pages so the person running a scrape
// can see what it is doing and intervene when Shopee interposes a challenge.
// Before it, a blocked run was invisible from the page it was blocked on: the
// operator had to notice an empty table in the dashboard and guess why.
//
// Boundaries this file keeps, deliberately:
//   - It never reads document.cookie. The service worker owns that permission,
//     and a content script holding it would widen the blast radius of any
//     injection on the host page.
//   - It never submits or clicks the host page's forms. The panel reports and
//     relays; Shopee's own UI is the operator's to drive.
//   - It never calls alert/confirm/prompt. Those block the page, which during a
//     scrape means the run stalls behind a dialog nobody is watching.
//   - Its styles live in a shadow root, so Shopee's CSS cannot fight the panel
//     and the panel's CSS cannot leak into the page.

(function () {
  if (window.__omniOverlayLoaded) return;
  window.__omniOverlayLoaded = true;

  const PANEL_ID = 'omni-operator-overlay';
  const POSITION_KEY = 'overlayPosition';
  const DEBUG_LOG_LIMIT = 50;

  const TABS = [
    { id: 'scrape', label: 'Scrape' },
    { id: 'blocker', label: 'Blocker' },
    { id: 'cookie', label: 'Cookie' },
    { id: 'debug', label: 'Debug' },
  ];

  // ─── Pure logic ───────────────────────────────────────────────────────────

  /**
   * Keep a dragged panel inside the viewport.
   *
   * A panel dragged past an edge would otherwise be unreachable: there is no
   * way to grab something that is off-screen, and the stored position would
   * reproduce it on every page load.
   */
  function clampPosition(pos, size, viewport) {
    const maxX = Math.max(0, viewport.width - size.width);
    const maxY = Math.max(0, viewport.height - size.height);
    return {
      x: Math.min(Math.max(pos.x, 0), maxX),
      y: Math.min(Math.max(pos.y, 0), maxY),
    };
  }

  /**
   * A bounded FIFO of recent commands and results.
   *
   * Bounded because the panel lives in a tab that stays open for the length of a
   * scrape: an unbounded log would leak memory during exactly the long runs it
   * exists to diagnose.
   */
  function makeRingBuffer(limit = DEBUG_LOG_LIMIT) {
    const items = [];
    return {
      push(entry) {
        items.push(entry);
        while (items.length > limit) items.shift();
      },
      entries() {
        return items.slice();
      },
    };
  }

  /**
   * What the blocker tab should show.
   *
   * Resume stays disabled until the operator says they cleared the challenge.
   * Resuming while the captcha is still up just burns another page and blocks
   * again, which reads as the resume being broken.
   */
  function blockerPanelState({ blocked, kind, confirmed }) {
    if (!blocked) {
      return { visible: false, label: '', resumeEnabled: false };
    }
    return {
      visible: true,
      // The kind is shown as reported: the extension may learn a new wall before
      // this list does, and hiding an unrecognised one leaves the operator with
      // a blocked run and no explanation.
      label: kind || 'unknown',
      resumeEnabled: Boolean(confirmed),
    };
  }

  function tabIds() {
    return TABS.map((t) => t.id);
  }

  /** Fall back to the first tab rather than rendering an empty panel. */
  function resolveTab(id) {
    return tabIds().includes(id) ? id : 'scrape';
  }

  // Exposed for tests: these are the parts worth pinning, and they need no DOM.
  window.__omniOverlayApi = {
    clampPosition,
    makeRingBuffer,
    blockerPanelState,
    tabIds,
    resolveTab,
  };

  // ─── Panel ────────────────────────────────────────────────────────────────

  const debugLog = makeRingBuffer();
  let activeTab = 'scrape';
  let lastState = { job: null, blocker: null, confirmed: false };

  // Styles are inlined rather than loaded from a file so no entry has to be
  // added to web_accessible_resources: an entry there makes the asset fetchable
  // by the host page, which is a wider grant than a panel needs.
  const PANEL_STYLE = `
    :host { all: initial; }
    .panel {
      position: fixed; z-index: 2147483647; width: 320px;
      font: 12px/1.5 system-ui, sans-serif; color: #1f1f1f;
      background: #fff; border: 1px solid #d9d9d9; border-radius: 8px;
      box-shadow: 0 4px 16px rgba(0,0,0,.15); overflow: hidden;
    }
    .drag { padding: 8px 10px; background: #fafafa; border-bottom: 1px solid #eee; cursor: move; font-weight: 600; }
    .tabs { display: flex; border-bottom: 1px solid #eee; }
    .tab { flex: 1; padding: 6px 4px; text-align: center; cursor: pointer; background: #fff; border: 0; font: inherit; }
    .tab[data-active="true"] { background: #f0f5ff; font-weight: 600; }
    .body { padding: 10px; max-height: 260px; overflow: auto; }
    .row { display: flex; justify-content: space-between; gap: 8px; padding: 2px 0; }
    .muted { color: #8c8c8c; }
    .warn { color: #ad4e00; }
    button.action { padding: 4px 10px; font: inherit; cursor: pointer; }
    button.action[disabled] { cursor: not-allowed; opacity: .5; }
    .log { font-family: ui-monospace, monospace; font-size: 11px; white-space: pre-wrap; }
  `;

  function buildPanel() {
    const host = document.createElement('div');
    host.setAttribute('id', PANEL_ID);
    const root = host.attachShadow({ mode: 'open' });

    const style = document.createElement('style');
    style.textContent = PANEL_STYLE;
    root.appendChild(style);

    const panel = document.createElement('div');
    panel.classList.add('panel');

    const handle = document.createElement('div');
    handle.classList.add('drag');
    handle.textContent = 'Omni — scrape operator';
    panel.appendChild(handle);

    const tabs = document.createElement('div');
    tabs.classList.add('tabs');
    for (const tab of TABS) {
      const btn = document.createElement('button');
      btn.classList.add('tab');
      btn.textContent = tab.label;
      btn.dataset.tab = tab.id;
      btn.addEventListener('click', () => {
        activeTab = resolveTab(tab.id);
        render();
      });
      tabs.appendChild(btn);
    }
    panel.appendChild(tabs);

    const body = document.createElement('div');
    body.classList.add('body');
    panel.appendChild(body);

    root.appendChild(panel);
    return { host, panel, handle, body };
  }

  function line(parent, label, value, className) {
    const row = document.createElement('div');
    row.classList.add('row');
    const l = document.createElement('span');
    l.classList.add('muted');
    l.textContent = label;
    const v = document.createElement('span');
    if (className) v.classList.add(className);
    v.textContent = value;
    row.appendChild(l);
    row.appendChild(v);
    parent.appendChild(row);
    return row;
  }

  let ui = null;

  function render() {
    if (!ui) return;
    const { body } = ui;
    body.textContent = '';

    if (activeTab === 'scrape') {
      const job = lastState.job;
      line(body, 'Job', job?.job_id || 'none');
      line(body, 'Status', job?.status || 'idle');
      line(body, 'Page', String(job?.pages ?? 0));
      line(body, 'Products', String(job?.products ?? 0));
      return;
    }

    if (activeTab === 'blocker') {
      const state = blockerPanelState({
        blocked: Boolean(lastState.blocker),
        kind: lastState.blocker?.kind || '',
        confirmed: lastState.confirmed,
      });
      if (!state.visible) {
        line(body, 'Blocker', 'none');
        return;
      }
      line(body, 'Kind', state.label, 'warn');
      line(body, 'URL', lastState.blocker?.url || '');
      line(body, 'Page', String(lastState.blocker?.blocked_page ?? ''));

      const confirm = document.createElement('button');
      confirm.classList.add('action');
      confirm.textContent = lastState.confirmed ? 'Challenge marked as solved' : 'I solved the challenge';
      confirm.addEventListener('click', () => {
        lastState.confirmed = true;
        render();
      });
      body.appendChild(confirm);

      const resume = document.createElement('button');
      resume.classList.add('action');
      resume.textContent = 'Resume scrape';
      if (!state.resumeEnabled) resume.setAttribute('disabled', 'true');
      resume.addEventListener('click', () => {
        if (!state.resumeEnabled) return;
        chrome.runtime.sendMessage({ type: 'omni_overlay_resume', job_id: lastState.job?.job_id });
        debugLog.push(`resume requested for ${lastState.job?.job_id || 'unknown job'}`);
      });
      body.appendChild(resume);
      return;
    }

    if (activeTab === 'cookie') {
      // Status only. The worker owns cookie access; this panel reports what it
      // was told and never reads document.cookie itself.
      line(body, 'Shopee session', lastState.job ? 'active scrape' : 'unknown');
      return;
    }

    const log = document.createElement('div');
    log.classList.add('log');
    log.textContent = debugLog.entries().join('\n') || 'no activity yet';
    body.appendChild(log);
  }

  function makeDraggable(panel, handle) {
    let dragging = false;
    let offset = { x: 0, y: 0 };

    handle.addEventListener('mousedown', (e) => {
      dragging = true;
      const rect = panel.getBoundingClientRect();
      offset = { x: e.clientX - rect.left, y: e.clientY - rect.top };
    });

    window.addEventListener('mousemove', (e) => {
      if (!dragging) return;
      const rect = panel.getBoundingClientRect();
      const next = clampPosition(
        { x: e.clientX - offset.x, y: e.clientY - offset.y },
        { width: rect.width, height: rect.height },
        { width: window.innerWidth, height: window.innerHeight },
      );
      panel.style.left = `${next.x}px`;
      panel.style.top = `${next.y}px`;
    });

    window.addEventListener('mouseup', () => {
      if (!dragging) return;
      dragging = false;
      chrome.storage.local.set({
        [POSITION_KEY]: { x: parseInt(panel.style.left, 10) || 0, y: parseInt(panel.style.top, 10) || 0 },
      });
    });
  }

  async function mount() {
    ui = buildPanel();
    const stored = await chrome.storage.local.get([POSITION_KEY]);
    const saved = stored?.[POSITION_KEY] || { x: 16, y: 16 };
    const start = clampPosition(
      saved,
      { width: 320, height: 300 },
      { width: window.innerWidth, height: window.innerHeight },
    );
    ui.panel.style.left = `${start.x}px`;
    ui.panel.style.top = `${start.y}px`;

    makeDraggable(ui.panel, ui.handle);
    document.body.appendChild(ui.host);
    render();
  }

  chrome.runtime.onMessage.addListener((message) => {
    if (message?.type !== 'omni_overlay_state') return false;

    lastState = {
      job: message.job || null,
      blocker: message.blocker || null,
      // A new blocker invalidates a previous confirmation: the operator has not
      // seen this one yet, and carrying the flag over would let a resume fire
      // straight into a challenge that is still up.
      confirmed: message.blocker && message.blocker.url !== lastState.blocker?.url
        ? false
        : lastState.confirmed,
    };
    debugLog.push(`${new Date().toISOString()} state ${message.job?.status || 'unknown'}`);
    render();
    return false;
  });

  // Only mount where a scrape can actually run, so the panel does not appear on
  // unrelated pages the operator happens to have open.
  if (document.body) {
    mount();
  } else {
    document.addEventListener('DOMContentLoaded', mount);
  }
})();
