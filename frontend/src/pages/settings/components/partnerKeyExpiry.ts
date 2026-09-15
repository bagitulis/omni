// Phase 8 — Shopee partner-key expiry helper.
//
// Extracted out of PlatformCard.tsx so React Fast Refresh works cleanly (a
// TSX file that also exports plain functions breaks HMR). Behaviour mirrors
// backend `pkg/shopee/expiry.go` — keep them in sync.

/** Default warning window in days for Shopee partner-key rotation. */
export const PARTNER_KEY_WARNING_DAYS = 7;

/**
 * Returns whether a partner key ISO timestamp is inside the warning window,
 * plus the friendly "days left" number the UI renders.
 *
 * Rules:
 * - Empty / malformed input → not expiring (no banner).
 * - Past date → `expired: true, expiring: true, days_left: 0`.
 * - Within `warningDays` of now → `expiring: true`.
 * - Negative `warningDays` clamped to 0 (defensive — matches backend).
 */
export function isPartnerKeyExpiring(
  iso: string | undefined,
  warningDays: number = PARTNER_KEY_WARNING_DAYS,
): { expired: boolean; expiring: boolean; days_left: number } {
  if (!iso) return { expired: false, expiring: false, days_left: 0 };
  const t = new Date(iso).getTime();
  if (Number.isNaN(t)) return { expired: false, expiring: false, days_left: 0 };
  const now = Date.now();
  const diffMs = t - now;
  const daysLeft = Math.ceil(diffMs / 86_400_000);
  if (diffMs <= 0) return { expired: true, expiring: true, days_left: 0 };
  const window = warningDays < 0 ? 0 : warningDays;
  const expiring = daysLeft <= window;
  return { expired: false, expiring, days_left: daysLeft };
}
