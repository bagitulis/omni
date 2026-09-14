// Tests for pairing-code validation in popup.js.
//
// The popup validates before spending a round trip, and its alphabet must match
// the server's exactly: the server generates codes that exclude 0/O and 1/I/L
// (visually ambiguous), so a client that accepted them would report a generic
// server rejection instead of the actual mis-typing.
//
// These mirror normalizeCode and validateCode from popup.js, which cannot be
// imported directly because it touches the DOM at module scope.

import test from 'node:test';
import assert from 'node:assert/strict';

const AMBIGUOUS = /[01OIL]/;

function normalizeCode(input) {
  return String(input || '').trim().toUpperCase();
}

function validateCode(input) {
  const code = normalizeCode(input);
  if (!code) return { ok: false, error: 'Enter the pairing code shown in Omni' };
  if (code.length !== 8) return { ok: false, error: 'Pairing codes are 8 characters' };
  if (!/^[ABCDEFGHJKMNPQRSTUVWXYZ23456789]+$/.test(code)) {
    return { ok: false, error: 'That code contains characters pairing codes never use' };
  }
  return { ok: true, code };
}

test('normalizeCode trims and upper-cases', () => {
  assert.equal(normalizeCode('  abcd2345  '), 'ABCD2345');
});

test('normalizeCode handles non-string input', () => {
  assert.equal(normalizeCode(undefined), '');
  assert.equal(normalizeCode(null), '');
  assert.equal(normalizeCode(12345678), '12345678');
});

test('validateCode accepts a well-formed code', () => {
  const result = validateCode('abcd2345');
  assert.equal(result.ok, true);
  assert.equal(result.code, 'ABCD2345');
});

test('validateCode rejects empty input', () => {
  assert.equal(validateCode('').ok, false);
  assert.equal(validateCode('   ').ok, false);
});

test('validateCode rejects wrong length', () => {
  assert.equal(validateCode('ABC').ok, false);
  assert.equal(validateCode('ABCD23456').ok, false);
});

test('validateCode rejects ambiguous characters', () => {
  // These are the characters the server never emits, so seeing one means a
  // transcription error; catching it locally gives a clearer message.
  for (const raw of ['ABCD2340', 'ABCD234O', 'ABCD2341', 'ABCD234I', 'ABCD234L']) {
    const result = validateCode(raw);
    assert.equal(result.ok, false, `${raw} should be rejected`);
    assert.match(result.error, /never use/);
  }
});

test('validateCode alphabet matches the server', () => {
  // The server's alphabet is "ABCDEFGHJKMNPQRSTUVWXYZ23456789" — no I, L, O, 0, 1.
  const serverAlphabet = 'ABCDEFGHJKMNPQRSTUVWXYZ23456789';
  for (const ch of serverAlphabet) {
    assert.equal(AMBIGUOUS.test(ch), false, `${ch} must not be ambiguous`);
  }
  // Every generated code must validate.
  for (const ch of serverAlphabet) {
    const code = ('ABCD2345'.slice(0, 7) + ch);
    assert.equal(validateCode(code).ok, true, `code containing ${ch} should validate`);
  }
});

test('validateCode rejects symbols and whitespace inside the code', () => {
  assert.equal(validateCode('ABCD-234').ok, false);
  assert.equal(validateCode('ABCD 234').ok, false);
  assert.equal(validateCode('ABCD@345').ok, false);
});
