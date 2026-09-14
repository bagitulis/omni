// popup.js — pairing and connection UI for the Omni extension.
//
// The popup is intentionally thin: it validates input, sends one message to the
// service worker, and renders whatever status the worker reports. All state
// lives in the worker/storage so the popup can be closed at any time without
// interrupting a connection.

const STATUS_LABELS = {
  connected: 'Connected',
  connecting: 'Connecting…',
  disconnected: 'Disconnected',
  unpaired: 'Not paired',
  paired: 'Paired',
  misconfigured: 'Misconfigured',
  // Terminal state: the server rejected the stored credential, so the operator
  // must pair again rather than wait for a reconnect that will never succeed.
  auth_failed: 'Not paired — pair again',
  error: 'Error',
};

const el = {
  dot: document.getElementById('dot'),
  statusText: document.getElementById('statusText'),
  serverUrl: document.getElementById('serverUrl'),
  code: document.getElementById('code'),
  pair: document.getElementById('pair'),
  reconnect: document.getElementById('reconnect'),
  msg: document.getElementById('msg'),
  extensionId: document.getElementById('extensionId'),
  debug: document.getElementById('debug'),
};

function showMessage(text, kind) {
  el.msg.textContent = text;
  el.msg.className = 'msg ' + (kind || '');
}

function clearMessage() {
  el.msg.textContent = '';
  el.msg.className = 'msg';
}

/**
 * Surface a rejected credential prominently.
 *
 * Without this the extension looks merely "disconnected", and the operator waits
 * for a reconnect that will never come because the token is dead.
 */
function warnIfRevoked(state) {
  if (state !== 'auth_failed') return;
  showMessage(
    'This extension is no longer paired (the credential was rejected). Generate a new pairing code in Omni and pair again.',
    'err',
  );
}

/** Normalize a typed pairing code: codes are upper-case and exclude lookalikes. */
export function normalizeCode(input) {
  return String(input || '').trim().toUpperCase();
}

/** Validate a pairing code before spending a round trip on it. */
export function validateCode(input) {
  const code = normalizeCode(input);
  if (!code) return { ok: false, error: 'Enter the pairing code shown in Omni' };
  if (code.length !== 8) return { ok: false, error: 'Pairing codes are 8 characters' };
  // Deliberately excludes 0/O and 1/I/L, matching the server's alphabet, so a
  // transcription slip is reported here rather than as a generic server error.
  if (!/^[ABCDEFGHJKMNPQRSTUVWXYZ23456789]+$/.test(code)) {
    return { ok: false, error: 'That code contains characters pairing codes never use' };
  }
  return { ok: true, code };
}

async function refreshStatus() {
  const status = await chrome.runtime.sendMessage({ type: 'omni_status' });
  const state = status?.connectionStatus || 'unpaired';

  el.dot.className = 'dot ' + state;
  el.statusText.textContent = STATUS_LABELS[state] || state;
  el.serverUrl.value = status?.serverUrl || '';
  el.extensionId.textContent = status?.extensionId ? 'ID: ' + status.extensionId : '';

  const detail = status?.connectionDetail;
  el.debug.textContent = detail ? detail : '';

  el.pair.disabled = state === 'connecting';

  // Show the terminal state prominently: otherwise the extension looks merely
  // disconnected and the operator waits for a reconnect that cannot happen.
  warnIfRevoked(state);
}

el.pair.addEventListener('click', async () => {
  clearMessage();

  const check = validateCode(el.code.value);
  if (!check.ok) {
    showMessage(check.error, 'err');
    return;
  }

  const serverUrl = el.serverUrl.value.trim();
  if (!serverUrl) {
    showMessage('Enter the Omni server URL', 'err');
    return;
  }

  el.pair.disabled = true;
  el.pair.textContent = 'Pairing…';

  try {
    const result = await chrome.runtime.sendMessage({
      type: 'omni_pair',
      code: check.code,
      serverUrl,
    });

    if (result?.success) {
      showMessage('Paired. Connecting…', 'ok');
      el.code.value = '';
      // Give the worker a moment to open the socket before reading status.
      setTimeout(refreshStatus, 500);
    } else {
      showMessage(result?.error || 'Pairing failed', 'err');
    }
  } catch (err) {
    showMessage(String(err), 'err');
  } finally {
    el.pair.disabled = false;
    el.pair.textContent = 'Pair';
  }
});

el.reconnect.addEventListener('click', async () => {
  clearMessage();
  await chrome.runtime.sendMessage({ type: 'omni_reconnect' });
  showMessage('Reconnecting…', 'ok');
  setTimeout(refreshStatus, 800);
});

// Refresh periodically so the status reflects a dropped or restored socket
// without the user having to reopen the popup.
refreshStatus();
const timer = setInterval(refreshStatus, 2000);
window.addEventListener('unload', () => clearInterval(timer));
