import { QueryClient } from "@tanstack/react-query";
import { message } from "antd";

/**
 * Shared TanStack Query client configuration
 * Used across all API queries and mutations
 */
export const queryClient = new QueryClient({
  defaultOptions: {
    queries: {
      staleTime: 5 * 60 * 1000, // 5 minutes
      retry: 1, // Retry failed requests once
      refetchOnWindowFocus: false, // Don't refetch on window focus
    },
    mutations: {
      onError: (error: Error) => {
        // Show error toast for mutations on error
        message.error(error.message || "Operation failed");
      },
    },
  },
});
