// service-worker.js — Omni Extensions background worker (MV3).
//
// MV3 service workers are terminated after ~30s idle, which would kill a plain
// WebSocket. Three mechanisms keep the socket alive, in order of reliability:
//
//   1. An alarm every 25s reconnects if the socket dropped (chrome.alarms wakes
//      the worker even after termination).
//   2. A ping every 20s on the live socket, which is also traffic that resets
//      Chrome's idle timer.
//   3. Inbound server traffic resets the idle timer whenever the server pushes
//      a command or progress event.
//
// The worker treats every socket event as potentially happening after a wake-up,
// so state is re-read from chrome.storage.session rather than trusted from
// memory. No offscreen document is used; whether one is needed is the open
// question this file exists to answer empirically.

const KEEPALIVE_ALARM = 'omni_keepalive';
const PING_INTERVAL_MS = 20000;
const ALARM_PERIOD_MINUTES = 25 / 60; // 25 seconds
const RECONNECT_BASE_MS = 1000;
const RECONNECT_MAX_MS = 30000;

// Hosts the extension will act on. An exact-match set, checked against the
// parsed hostname, so a lookalike like "shopee.co.id.evil.com" cannot pass.
const SUPPORTED_HOSTS = new Set(['shopee.co.id', 'www.shopee.co.id']);

// In-memory socket. Deliberately NOT persisted: a socket cannot be restored
// across worker restarts, only re-established.
let ws = null;
let wsReady = false;
let pingTimer = null;
let reconnectAttempt = 0;
let reconnectTimer = null;
let activeTaskId = null;

// ─── Storage helpers ────────────────────────────────────────────────────────

async function getConfig() {
  const stored = await chrome.storage.local.get(['serverUrl', 'token', 'extensionId']);
  return {
    serverUrl: stored.serverUrl || '',
    token: stored.token || '',
    extensionId: stored.extensionId || '',
  };
}

async function setConnectionStatus(status, detail) {
  await chrome.storage.local.set({
    connectionStatus: status,
    connectionDetail: detail || '',
    connectionUpdatedAt: Date.now(),
  });
}

// ─── WebSocket lifecycle ────────────────────────────────────────────────────

/**
 * Build the WebSocket URL from the configured server.
 *
 * Derived from the stored server URL rather than hardcoded, so a deployment
 * change does not require a new extension build.
 */
export function buildWsUrl(serverUrl) {
  // Trim first: a whitespace-only value is falsy-checked too late otherwise and
  // yields a malformed "wss:///api/extensions/ws".
  const trimmed = (serverUrl || '').trim();
  if (!trimmed) return '';

  let base = trimmed.replace(/\/+$/, '');
  base = base.replace(/^http:\/\//i, 'ws://').replace(/^https:\/\//i, 'wss://');
  if (!/^wss?:\/\//i.test(base)) {
    // A bare host is assumed to be https in production.
    base = 'wss://' + base;
  }
  return base + '/api/extensions/ws';
}

async function connect() {
  if (ws && (ws.readyState === WebSocket.OPEN || ws.readyState === WebSocket.CONNECTING)) {
    return;
  }

  const { serverUrl, token, extensionId } = await getConfig();
  if (!token || !extensionId) {
    await setConnectionStatus('unpaired', 'Pair this extension to continue');
    return;
  }

  const url = buildWsUrl(serverUrl);
  if (!url) {
    await setConnectionStatus('misconfigured', 'Server URL is not set');
    return;
  }

  await setConnectionStatus('connecting');

  try {
    ws = new WebSocket(url);
  } catch (err) {
    await setConnectionStatus('error', String(err));
    scheduleReconnect();
    return;
  }

  ws.addEventListener('open', async () => {
    wsReady = false;
    reconnectAttempt = 0;
    // Authenticate on the first frame. The server expects this before anything
    // else and will close the socket if it does not arrive.
    send({ type: 'auth', action: 'connect', payload: { token, extension_id: extensionId } });
    startPing();
    await setConnectionStatus('connected');
  });

  ws.addEventListener('message', (event) => {
    let msg;
    try {
      msg = JSON.parse(event.data);
    } catch {
      return; // malformed frame must not kill the socket
    }
    handleServerMessage(msg);
  });

  ws.addEventListener('close', async () => {
    wsReady = false;
    stopPing();
    await setConnectionStatus('disconnected');
    scheduleReconnect();
  });

  ws.addEventListener('error', async () => {
    // 'close' follows, which owns the reconnect. Recording here aids debugging.
    await setConnectionStatus('error', 'WebSocket error');
  });
}

function disconnect() {
  stopPing();
  if (reconnectTimer) {
    clearTimeout(reconnectTimer);
    reconnectTimer = null;
  }
  if (ws) {
    try {
      ws.close();
    } catch {
      /* already closing */
    }
    ws = null;
  }
  wsReady = false;
}

/**
 * Schedule a reconnect with exponential backoff capped at 30s.
 *
 * Capped rather than unbounded because a browser that was offline (laptop
 * sleep) should reconnect promptly once it returns, not after a long backoff.
 */
function scheduleReconnect() {
  if (reconnectTimer) return;
  const delay = Math.min(RECONNECT_BASE_MS * Math.pow(2, reconnectAttempt), RECONNECT_MAX_MS);
  reconnectAttempt += 1;
  reconnectTimer = setTimeout(() => {
    reconnectTimer = null;
    connect();
  }, delay);
}

function startPing() {
  stopPing();
  pingTimer = setInterval(() => {
    send({ type: 'ping' });
  }, PING_INTERVAL_MS);
}

function stopPing() {
  if (pingTimer) {
    clearInterval(pingTimer);
    pingTimer = null;
  }
}

function send(msg) {
  if (!ws || ws.readyState !== WebSocket.OPEN) return false;
  try {
    ws.send(JSON.stringify(msg));
    return true;
  } catch {
    return false;
  }
}

// ─── Server message handling ────────────────────────────────────────────────

async function handleServerMessage(msg) {
  const { type, action, id, payload } = msg || {};

  if (type === 'auth_ok') {
    wsReady = true;
    await setConnectionStatus('connected');
    return;
  }

  if (type === 'ping' || action === 'pong') return;

  if (type === 'command') {
    const result = await runCommand(action, payload || {});
    send({
      id,
      type: result.success ? 'result' : 'error',
      action,
      payload: result,
    });
  }
}

// ─── Command dispatch ───────────────────────────────────────────────────────

/**
 * Execute a server command.
 *
 * Tab-scoped commands run in the content script; navigation and tab lifecycle
 * are handled here because content scripts are destroyed on navigation.
 */
async function runCommand(action, payload) {
  try {
    switch (action) {
      case 'open_tab':
        return await openTab(payload);
      case 'close_tab':
        return await closeTab(payload);
      case 'goto':
        return await goto(payload);
      case 'reload':
        return await reload(payload);
      case 'wait':
        return await waitMs(payload);
      default:
        return await forwardToContent(action, payload);
    }
  } catch (err) {
    return { success: false, error: String(err) };
  }
}

async function openTab({ url, active = false }) {
  const tab = await chrome.tabs.create({ url: url || 'about:blank', active });
  // Activate so Chrome does not throttle the tab: background tabs are heavily
  // throttled, which stalls scrolling and lazy-loading during a scrape.
  try {
    await chrome.tabs.update(tab.id, { active: true });
  } catch {
    /* activation is best effort */
  }
  await waitForTabLoad(tab.id, 30000);
  return { success: true, data: { tab_id: tab.id } };
}

async function closeTab({ tab_id }) {
  if (tab_id) {
    try {
      await chrome.tabs.remove(tab_id);
    } catch {
      /* already closed */
    }
  }
  return { success: true, data: null };
}

async function goto({ tab_id, url }) {
  if (!tab_id) return { success: false, error: 'tab_id is required' };
  await chrome.tabs.update(tab_id, { url });
  await waitForTabLoad(tab_id, 20000);
  return { success: true, data: { tab_id, url } };
}

async function reload({ tab_id }) {
  if (tab_id) {
    await chrome.tabs.reload(tab_id);
    await waitForTabLoad(tab_id, 20000);
  }
  return { success: true, data: null };
}

async function waitMs({ ms = 0 }) {
  // Clamp so a malformed payload cannot park the worker for minutes.
  await new Promise((r) => setTimeout(r, Math.min(ms, 30000)));
  return { success: true, data: null };
}

function waitForTabLoad(tabId, timeoutMs) {
  return new Promise((resolve) => {
    let settled = false;
    const finish = () => {
      if (settled) return;
      settled = true;
      chrome.tabs.onUpdated.removeListener(listener);
      clearTimeout(timer);
      resolve();
    };

    const listener = (id, info) => {
      if (id === tabId && info.status === 'complete') finish();
    };

    chrome.tabs.onUpdated.addListener(listener);
    const timer = setTimeout(finish, timeoutMs);

    // A tab that is already complete never fires onUpdated; resolve immediately
    // so the caller does not block for the full timeout.
    chrome.tabs.get(tabId).then((tab) => {
      if (tab && tab.status === 'complete') finish();
    }).catch(finish);
  });
}

async function forwardToContent(action, payload) {
  const tabId = payload?.tab_id;
  if (!tabId) return { success: false, error: `tab_id is required for ${action}` };

  try {
    const tab = await chrome.tabs.get(tabId);
    if (!isSupportedUrl(tab.url)) {
      return { success: false, error: `unsupported domain for ${action}: ${tab.url}` };
    }
  } catch (err) {
    return { success: false, error: `tab lookup failed: ${String(err)}` };
  }

  try {
    const response = await chrome.tabs.sendMessage(tabId, { type: 'command', action, payload });
    return response || { success: false, error: 'no response from content script' };
  } catch (err) {
    return { success: false, error: String(err) };
  }
}

/**
 * Domains this extension will act on.
 *
 * Narrow on purpose: a command must not be executable against an arbitrary site
 * the user happens to have open.
 *
 * The host is parsed rather than prefix-compared. A startsWith check would
 * accept "https://shopee.co.id.evil.com", which an attacker can register — the
 * extension would then run automation commands against their page.
 */
export function isSupportedUrl(url) {
  if (typeof url !== 'string' || !url) return false;

  let parsed;
  try {
    parsed = new URL(url);
  } catch {
    return false;
  }

  // Shopee is https-only: accepting http would allow a downgrade.
  if (parsed.protocol === 'https:' && SUPPORTED_HOSTS.has(parsed.hostname)) {
    return true;
  }
  // Localhost is allowed over http for development only.
  return parsed.protocol === 'http:' && parsed.hostname === 'localhost';
}

// ─── Lifecycle ──────────────────────────────────────────────────────────────

chrome.runtime.onInstalled.addListener(async ({ reason }) => {
  if (reason === 'install') {
    const existing = await chrome.storage.local.get(['extensionId']);
    if (!existing.extensionId) {
      await chrome.storage.local.set({ extensionId: crypto.randomUUID() });
    }
  }
  chrome.alarms.create(KEEPALIVE_ALARM, { periodInMinutes: ALARM_PERIOD_MINUTES });
  connect();
});

chrome.runtime.onStartup.addListener(async () => {
  const alarm = await chrome.alarms.get(KEEPALIVE_ALARM);
  if (!alarm) {
    chrome.alarms.create(KEEPALIVE_ALARM, { periodInMinutes: ALARM_PERIOD_MINUTES });
  }
  connect();
});

// The alarm is the primary survival mechanism: it fires even after the worker
// was terminated, which is what re-establishes the socket.
chrome.alarms.onAlarm.addListener(async (alarm) => {
  if (alarm.name !== KEEPALIVE_ALARM) return;
  if (!ws || ws.readyState !== WebSocket.OPEN) {
    connect();
  } else {
    send({ type: 'ping' });
  }
});

// Popup and content-script requests.
chrome.runtime.onMessage.addListener((message, sender, sendResponse) => {
  // The MAIN-world observer must be injected by the worker, not the content
  // script: an isolated-world hook would only observe the content script's own
  // requests, never Shopee's page requests.
  if (message?.type === 'install_main_observer') {
    const tabId = sender?.tab?.id;
    if (typeof tabId !== 'number') {
      sendResponse({ ok: false, error: 'no tab id' });
      return false;
    }
    chrome.scripting
      .executeScript({
        target: { tabId },
        world: 'MAIN',
        files: ['main-network-observer.js'],
      })
      .then(() => sendResponse({ ok: true }))
      .catch((err) => sendResponse({ ok: false, error: String(err) }));
    return true; // async response
  }

  if (message?.type === 'omni_reconnect') {
    reconnectAttempt = 0;
    disconnect();
    connect();
    sendResponse({ ok: true });
    return true;
  }
  if (message?.type === 'omni_status') {
    chrome.storage.local
      .get(['connectionStatus', 'connectionDetail', 'extensionId', 'serverUrl'])
      .then((data) => sendResponse({ ...data, socketState: ws ? ws.readyState : -1, authenticated: wsReady }));
    return true;
  }
  if (message?.type === 'omni_pair') {
    pairWithCode(message.code, message.serverUrl)
      .then(sendResponse)
      .catch((err) => sendResponse({ success: false, error: String(err) }));
    return true;
  }
  return false;
});

// ─── Pairing ────────────────────────────────────────────────────────────────

/**
 * Redeem a pairing code for a token.
 *
 * The extension never sends a tenant: the tenant is resolved server-side from
 * the code, which is what keeps one tenant's browser from acting as another's.
 */
export async function pairWithCode(code, serverUrl) {
  if (!code) return { success: false, error: 'Pairing code is required' };

  const base = (serverUrl || '').trim().replace(/\/+$/, '');
  if (!base) return { success: false, error: 'Server URL is required' };
  const httpBase = base.replace(/^ws(s?):\/\//i, 'http$1://');

  const { extensionId } = await getConfig();
  const deviceId = extensionId || crypto.randomUUID();

  const resp = await fetch(httpBase + '/api/extensions/pairing/confirm', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({
      code: String(code).trim().toUpperCase(),
      extension_id: deviceId,
      hostname: 'browser-extension',
      browser_info: navigator.userAgent,
      chrome_version: (navigator.userAgent.match(/Chrome\/([\d.]+)/) || [])[1] || 'unknown',
      extension_version: chrome.runtime.getManifest().version,
      protocol_version: '1',
      capabilities: ['shopee_scrape', 'tab_discovery'],
    }),
  });

  const body = await resp.json().catch(() => ({}));
  if (!resp.ok || !body?.success) {
    // The server deliberately returns one generic message for unknown, expired,
    // and already-used codes; surface it rather than guessing which it was.
    return { success: false, error: body?.error || 'Pairing failed' };
  }

  await chrome.storage.local.set({
    serverUrl: httpBase,
    token: body.data.token,
    extensionId: deviceId,
    connectionStatus: 'paired',
  });

  connect();
  return { success: true, extension_id: deviceId };
}

// Exported for the test harness.
export { connect, disconnect };
