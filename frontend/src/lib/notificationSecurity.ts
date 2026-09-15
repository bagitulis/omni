/**
 * Security sanitizer for user-facing notification messages.
 *
 * Prevents internal/technical details from leaking to the UI.
 * Raw details are expected to be logged separately via `logger`.
 */

const BLOCKED_PATTERNS: RegExp[] = [
  /\bsql\b|query failed|database|relation|column "\w+"/i,
  /stack|traceback|at\s+\w+\.\w+\s*\(/i,
  /\/internal\/|\/pkg\/|\/src\//i,
  /token[=:]\s*\S|jwt[=:]\s*\S|bearer\s+\S|apikey|secret[=:]/i,
  /(?:password|passwd|pwd)[=:]\s*\S/i,
  /127\.0\.0\.1:\d{4,5}|localhost:\d{4,5}/i,
  /panic:|goroutine\s+\d/i,
  /ECONNREFUSED|ECONNRESET|EPIPE/i,
];

const SAFE_FALLBACK = "An error occurred. Please try again or contact support.";

/**
 * Sanitize an error message before showing it to the user.
 *
 * Returns the original message if it looks safe, otherwise returns
 * a generic fallback. This must NEVER be skipped for backend-sourced
 * error strings.
 */
export function sanitizeForUser(raw: string | undefined | null): string {
  if (!raw) return SAFE_FALLBACK;

  for (const pattern of BLOCKED_PATTERNS) {
    if (pattern.test(raw)) {
      return SAFE_FALLBACK;
    }
  }

  // Truncate overly long messages
  if (raw.length > 500) {
    return `${raw.substring(0, 497)}...`;
  }

  return raw;
}

// SPA-relative path allowlist — mirrors backend/internal/notify/safe_url.go.
// The single source of truth is the backend regex; keep this in sync.
const SAFE_PATH_RE = /^\/[A-Za-z0-9\-_./?=&%#]*$/;

/**
 * isSafeActionURL returns true when `url` is a SPA-relative path safe to
 * navigate to via react-router. Empty is treated as "no action" (true).
 * Any absolute URL, protocol-relative URL, or non-http scheme is rejected.
 */
export function isSafeActionURL(url: string | null | undefined): boolean {
  if (!url) return true;
  const trimmed = url.trim();
  if (trimmed === "") return false;
  if (trimmed.length > 500) return false;
  if (trimmed.startsWith("//")) return false;
  return SAFE_PATH_RE.test(trimmed);
}
