// Turning a polling query's failure state into an operator-facing notice.
//
// Separated from the hooks so the threshold decision — when a transient blip
// becomes a reportable failure — is testable and applied consistently.

export interface PollingFailureState {
  /** What is being polled, used in the message (e.g. "extensions"). */
  subject: string;
  /** TanStack Query's consecutive failure count for the query. */
  failureCount: number;
  /** Whether cached data is still being shown. */
  hasData: boolean;
  /** The error message, if the error carried one. */
  message?: string;
}

/**
 * Failures below this are not reported.
 *
 * A single failure between polls is routine — a token refresh, a redeploy, a
 * dropped connection — and TanStack Query retries it. Only a repeat means the
 * displayed data can no longer be trusted.
 */
const DEFAULT_FAILURE_THRESHOLD = 2;

/**
 * Describe a repeated polling failure, or null when there is nothing to report.
 *
 * The two cases are worded differently on purpose: stale-but-present data is a
 * trust problem (what is on screen may be wrong), while no data at all is a
 * plain load failure.
 */
export function pollingFailureNotice(
  state: PollingFailureState,
  threshold: number = DEFAULT_FAILURE_THRESHOLD,
): string | null {
  if (state.failureCount < threshold) return null;

  const reason = state.message?.trim() ? state.message.trim() : "the request failed";

  if (state.hasData) {
    return `Live updates for ${state.subject} have failed ${state.failureCount} times in a row, so what is shown may be out of date — ${reason}`;
  }
  return `Failed to load ${state.subject} after ${state.failureCount} attempts — ${reason}`;
}
