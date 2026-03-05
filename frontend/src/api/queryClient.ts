import { QueryClient } from "@tanstack/react-query";

/**
 * Shared TanStack Query client configuration.
 *
 * NOTE: There is intentionally NO global `mutations.onError` handler here.
 * Each mutation hook defines its own onError with a contextual message.
 * A global handler would cause *double* toasts for every mutation that
 * already shows its own error feedback.
 */
export const queryClient = new QueryClient({
  defaultOptions: {
    queries: {
      staleTime: 5 * 60 * 1000, // 5 minutes
      retry: 1, // Retry failed requests once
      refetchOnWindowFocus: false, // Don't refetch on window focus
    },
  },
});
