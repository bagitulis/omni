import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import {
  getAllTokenStatus,
  getTokenStatus,
  refreshToken,
  refreshAllTokens,
} from "@/api/tokens";
import { message } from "antd";
import { logger } from "@/lib/logger";

/**
 * Hook to fetch status of all platform tokens
 * Polls every minute to keep status up to date
 */
export function useAllTokenStatus() {
  return useQuery({
    queryKey: ["tokens", "status"],
    queryFn: getAllTokenStatus,
    refetchInterval: 60_000, // Poll every minute
  });
}

/**
 * Hook to fetch status of a specific platform token
 */
export function useTokenStatus(platform: string | null) {
  return useQuery({
    queryKey: ["tokens", "status", platform],
    queryFn: () => {
      if (!platform) throw new Error("Platform is required");
      return getTokenStatus(platform);
    },
    enabled: !!platform,
  });
}

/**
 * Hook to refresh a platform token
 */
export function useRefreshToken() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (platform: string) => refreshToken(platform),
    onSuccess: (_, platform) => {
      message.success(
        `${
          platform.charAt(0).toUpperCase() + platform.slice(1)
        } token refreshed successfully`,
      );
      // Invalidate both all-status and specific-platform-status queries
      queryClient.invalidateQueries({ queryKey: ["tokens", "status"] });
    },
    onError: (error: Error) => {
      message.error(error.message || "Failed to refresh token");
    },
  });
}

/**
 * Hook to refresh all platform tokens
 */
export function useRefreshAllTokens() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (force?: boolean) => refreshAllTokens(force),
    onSuccess: (data) => {
      // Check if any failed
      const errors = Object.entries(data)
        .filter(([_, result]) => !result.success)
        .map(([platform, _]) => platform);

      if (errors.length > 0) {
        message.warning(
          `Refreshed with errors on: ${errors.join(", ")}. Check console for details.`,
        );
        logger.warn("Token refresh errors:", data);
      } else {
        message.success("All tokens refreshed successfully");
      }
      queryClient.invalidateQueries({ queryKey: ["tokens", "status"] });
    },
    onError: (error: Error) => {
      message.error(error.message || "Failed to refresh tokens");
    },
  });
}
