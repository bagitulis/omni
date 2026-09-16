// Scrape-job status logic and API client.
//
// Types are snake_case to match the Go backend, per the frontend conventions.
// The pure helpers here (terminal detection, summary/blocker parsing, tag
// mapping) are the load-bearing decisions the scraper UI makes, so they are kept
// separable from the components and unit-tested directly.

import apiClient from "./client";
import { API_TIMEOUT } from "@/lib/constants";

/** A job's lifecycle status. Mirrors backend `models.JobStatus`. */
export type ScrapeJobStatus =
  | "pending"
  | "running"
  | "completed"
  | "failed"
  | "cancelled"
  | "blocked";

/** Why a scrape stopped. Mirrors backend `StopReason`. */
export type StopReason =
  | "last_page"
  | "max_pages"
  | "max_products"
  | "cancelled"
  | "blocked"
  | "selectors_broken"
  | "empty_pages"
  | "page_error";

/** What stopped a blocked run. Mirrors backend `BlockerKind`. */
export type BlockerKind =
  | "captcha"
  | "login_required"
  | "rate_limited"
  | "anti_bot";

/** An obstacle a human must clear before a blocked job can resume. */
export interface ScrapeBlocker {
  kind: BlockerKind | string;
  url: string;
  /** 1-based page the block appeared on; doubles as the resume cursor. */
  blocked_page: number;
}

/** The `result_data` summary a finished scrape writes. */
export interface ScrapeSummary {
  job_id?: string;
  products?: number;
  pages?: number;
  source?: string;
  extension_id?: string;
  mode?: string;
  reason?: StopReason | string;
  blocker?: ScrapeBlocker;
  resume_from_page?: number;
}

/** A job as returned by GET /api/jobs/:id. */
export interface ScrapeJob {
  id: string;
  type: string;
  status: ScrapeJobStatus | string;
  progress_percent?: number;
  progress_message?: string;
  total_items?: number;
  processed_items?: number;
  result_data?: string;
  error_message?: string;
  created_at: string;
  updated_at: string;
  started_at?: string;
  completed_at?: string;
}

/**
 * The response of resuming a blocked job.
 *
 * `resumed` is false when the job was not actually blocked, so the caller can
 * tell a re-queue from a no-op without inspecting status text.
 */
export interface ResumeScrapeResponse {
  job_id: string;
  status: string;
  start_page?: number;
  resumed?: boolean;
}

const TERMINAL_STATUSES: ReadonlySet<string> = new Set([
  "completed",
  "failed",
  "cancelled",
]);

/**
 * A status from which the job will not move on its own.
 *
 * `blocked` is deliberately excluded: it is resumable, so it is not terminal in
 * the lifecycle sense even though polling should stop there.
 */
export function isTerminalScrapeStatus(status: string | undefined): boolean {
  return status !== undefined && TERMINAL_STATUSES.has(status);
}

/**
 * Whether polling should stop.
 *
 * Broader than terminal: a blocked job rests until a human acts, so there is no
 * state change to poll for and continuing would spin forever.
 */
export function shouldStopPolling(status: string | undefined): boolean {
  return isTerminalScrapeStatus(status) || status === "blocked";
}

/** Parse a job's `result_data` summary, tolerating absence and malformed JSON. */
export function parseScrapeSummary(
  resultData: string | undefined,
): ScrapeSummary | null {
  if (!resultData) return null;
  try {
    const parsed = JSON.parse(resultData) as unknown;
    if (parsed && typeof parsed === "object") {
      return parsed as ScrapeSummary;
    }
    return null;
  } catch {
    // A summary that cannot be parsed must not take the page down; the operator
    // still has the raw status to act on.
    return null;
  }
}

/**
 * Extract a usable blocker from a job.
 *
 * Returns null unless the summary carries a blocker with a URL: the panel links
 * to that URL, so a blocker without one is not actionable and is treated as
 * absent.
 */
export function extractBlocker(
  jobDetail: ScrapeJob | undefined,
): ScrapeBlocker | null {
  if (!jobDetail) return null;
  const summary = parseScrapeSummary(jobDetail.result_data);
  const blocker = summary?.blocker;
  if (!blocker || !blocker.url) return null;
  return blocker;
}

interface StatusTag {
  color: string;
  label: string;
}

const STATUS_TAGS: Record<string, StatusTag> = {
  pending: { color: "default", label: "Pending" },
  running: { color: "processing", label: "Running" },
  completed: { color: "success", label: "Completed" },
  failed: { color: "error", label: "Failed" },
  cancelled: { color: "warning", label: "Cancelled" },
  // Distinct from failed on purpose: a blocked run is recoverable, and colouring
  // it like a failure would tell the operator to give up on resumable work.
  blocked: { color: "gold", label: "Blocked" },
};

/** Map a scrape status to a Tag colour and label. */
export function scrapeStatusTag(status: string): StatusTag {
  return STATUS_TAGS[status] ?? { color: "default", label: status };
}

const BLOCKER_KIND_LABELS: Record<string, string> = {
  captcha: "Captcha",
  login_required: "Login required",
  rate_limited: "Rate limited",
  anti_bot: "Anti-bot check",
};

/** Humanise a blocker kind for display. */
export function blockerKindLabel(kind: string): string {
  return BLOCKER_KIND_LABELS[kind] ?? kind;
}

/** Fetch a single job by id. Returns null when the job is not found. */
export async function getScrapeJob(
  jobId: string,
  signal?: AbortSignal,
): Promise<ScrapeJob | null> {
  const response = await apiClient.get<never>(
    `/jobs/${encodeURIComponent(jobId)}`,
    { signal, timeout: API_TIMEOUT.SHORT },
  );
  if (!response.success) {
    throw new Error(response.error || "Failed to load job status");
  }
  // GetJob returns the job under a top-level `job` key rather than `data`, so it
  // is read from the raw envelope here.
  const job = (response as { job?: ScrapeJob }).job;
  return job ?? null;
}

/** Cancel a queued or running job via DELETE /api/jobs/:id. */
export async function cancelScrapeJob(
  jobId: string,
  signal?: AbortSignal,
): Promise<void> {
  const response = await apiClient.delete<never>(
    `/jobs/${encodeURIComponent(jobId)}`,
    { signal, timeout: API_TIMEOUT.SHORT },
  );
  if (!response.success) {
    throw new Error(response.error || "Failed to cancel the job");
  }
}

/** Resume a blocked scrape via POST /api/extensions/scrape/:job_id/resume. */
export async function resumeScrape(
  jobId: string,
  signal?: AbortSignal,
): Promise<ResumeScrapeResponse> {
  const response = await apiClient.post<ResumeScrapeResponse>(
    `/extensions/scrape/${encodeURIComponent(jobId)}/resume`,
    undefined,
    { signal, timeout: API_TIMEOUT.DEFAULT },
  );
  if (!response.success || !response.data) {
    throw new Error(response.error || "Failed to resume the scrape");
  }
  return response.data;
}
