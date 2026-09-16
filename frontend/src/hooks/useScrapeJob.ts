// TanStack Query hooks for a single scrape job's lifecycle.
//
// Kept apart from useExtensions because this is per-job state with its own
// polling rules: it must stop on its own once the job stops moving, which the
// list queries never need to do.

import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import {
  cancelScrapeJob,
  extractBlocker,
  getScrapeJob,
  resumeScrape,
  shouldStopPolling,
  type ScrapeJob,
} from "@/api/scrapeJob";

const SCRAPE_JOB_KEY = ["extensions", "scrape-job"] as const;
const PRODUCTS_KEY = ["extensions", "scraped-products"] as const;

/** How often a live job is re-read while it is still moving. */
const POLL_INTERVAL_MS = 3000;

export function scrapeJobKey(jobId: string) {
  return [...SCRAPE_JOB_KEY, jobId] as const;
}

/**
 * Poll a scrape job until it stops moving.
 *
 * The interval is a function of the last payload rather than a fixed number so
 * polling ends at the job's resting state instead of running for as long as the
 * page is open.
 */
export function useScrapeJob(jobId: string | undefined | null) {
  const safeJobId = typeof jobId === "string" ? jobId.trim() : "";

  const query = useQuery({
    queryKey: scrapeJobKey(safeJobId),
    queryFn: ({ signal }) => getScrapeJob(safeJobId, signal),
    enabled: safeJobId.length > 0,
    refetchInterval: (query) => {
      const status = query.state.data?.status;
      return shouldStopPolling(status) ? false : POLL_INTERVAL_MS;
    },
    refetchOnWindowFocus: false,
  });

  return query;
}

export function useCancelScrapeJob(jobId: string) {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: () => cancelScrapeJob(jobId),
    onSuccess: () => {
      void queryClient.invalidateQueries({ queryKey: scrapeJobKey(jobId) });
    },
  });
}

export function useResumeScrape(jobId: string) {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: () => resumeScrape(jobId),
    onSuccess: () => {
      void queryClient.invalidateQueries({ queryKey: scrapeJobKey(jobId) });
      // A resumed run writes more rows, so any cached result page is stale.
      void queryClient.invalidateQueries({ queryKey: PRODUCTS_KEY });
    },
  });
}

/** The blocker for a job, or null when it is not blocked. */
export function jobBlocker(jobDetail: ScrapeJob | undefined | null) {
  return extractBlocker(jobDetail ?? undefined);
}
