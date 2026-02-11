import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import { message } from "antd";
import {
  getRouteConfigs,
  patchRouteConfig,
  bulkUpdateRouteConfigs,
} from "@/api/routeConfig";
import type { RouteConfig } from "@/types/routeConfig";

/**
 * Query hook for fetching route configurations
 */
export function useRouteConfigs() {
  return useQuery({
    queryKey: ["route-configs"],
    queryFn: getRouteConfigs,
  });
}

/**
 * Mutation hook for updating a single route config
 */
export function useUpdateRouteConfig() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: ({ id, data }: { id: number; data: Partial<RouteConfig> }) =>
      patchRouteConfig(id, data),
    onSuccess: () => {
      message.success("Route config updated");
      queryClient.invalidateQueries({ queryKey: ["route-configs"] });
    },
    onError: (e: Error) => message.error(e.message),
  });
}

/**
 * Mutation hook for bulk updating route configs
 */
export function useBulkUpdateRouteConfigs() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: ({
      ids,
      data,
    }: {
      ids: number[];
      data: Partial<RouteConfig>;
    }) => bulkUpdateRouteConfigs(ids, data),
    onSuccess: () => {
      message.success("Route configs updated");
      queryClient.invalidateQueries({ queryKey: ["route-configs"] });
    },
    onError: (e: Error) => message.error(e.message),
  });
}
