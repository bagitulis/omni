// TanStack Query hooks for the extensions feature.
//
// Server state lives here rather than in component state so that repeated
// navigation does not refetch everything, and so a mutation can invalidate the
// exact queries it affects.

import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import {
  generatePairingCode,
  listExtensions,
  listScrapedProducts,
  startScrape,
  unpairExtension,
  type ScrapeRequest,
} from "@/api/extensions";
import { pollingFailureNotice } from "@/lib/queryFailure";

const EXTENSIONS_KEY = ["extensions"] as const;
const PRODUCTS_KEY = ["extensions", "scraped-products"] as const;

/**
 * Poll the paired extensions.
 *
 * `failureNotice` is returned alongside the query so every consumer sees a
 * repeated polling failure. Without it the interval keeps firing silently and
 * the page goes on presenting stale connection states as current.
 */
export function useExtensions() {
  const query = useQuery({
    queryKey: EXTENSIONS_KEY,
    queryFn: ({ signal }) => listExtensions(signal),
    // Extensions connect and disconnect on their own schedule, so a periodic
    // refresh keeps the status honest without the user reloading.
    refetchInterval: 15000,
  });

  return {
    ...query,
    failureNotice: pollingFailureNotice({
      subject: "extensions",
      failureCount: query.failureCount,
      hasData: query.data !== undefined,
      message: (query.error as Error | null)?.message,
    }),
  };
}

export function useGeneratePairingCode() {
  return useMutation({ mutationFn: () => generatePairingCode() });
}

export function useUnpairExtension() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: (extensionId: string) => unpairExtension(extensionId),
    onSuccess: () => {
      void queryClient.invalidateQueries({ queryKey: EXTENSIONS_KEY });
    },
  });
}

export function useStartScrape() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: (req: ScrapeRequest) => startScrape(req),
    onSuccess: () => {
      // New rows may have been written, so any cached result page is stale.
      void queryClient.invalidateQueries({ queryKey: PRODUCTS_KEY });
    },
  });
}

/**
 * Fetch scraped products for a job.
 *
 * `jobId` is normalised defensively rather than trusted: it reaches here from a
 * URL query parameter, and `undefined.length` throws. A malformed or absent id
 * must disable the query, not crash the page.
 */
export function useScrapedProducts(
  jobId: string | undefined | null,
  page = 1,
  pageSize = 50,
) {
  const safeJobId = typeof jobId === "string" ? jobId.trim() : "";

  return useQuery({
    queryKey: [...PRODUCTS_KEY, safeJobId, page, pageSize],
    queryFn: ({ signal }) =>
      listScrapedProducts(safeJobId, page, pageSize, signal),
    // Nothing to fetch until a run has produced rows.
    enabled: safeJobId.length > 0,
    // A finished run's rows do not change, so there is no value in refetching on
    // every focus; an explicit refresh button covers the live case.
    refetchOnWindowFocus: false,
  });
}
