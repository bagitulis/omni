import apiClient from "./client";
import type { RouteConfig } from "@/types/routeConfig";

/**
 * Get all route configurations
 */
export async function getRouteConfigs(): Promise<RouteConfig[]> {
  const response = await apiClient.get<RouteConfig[]>("/routes-config");
  if (!response.success) {
    throw new Error(response.error || "Failed to fetch route configs");
  }
  return response.data || [];
}

/**
 * Update a single route configuration
 */
export async function patchRouteConfig(
  id: number,
  data: Partial<RouteConfig>,
): Promise<void> {
  const response = await apiClient.patch(`/routes-config/${id}`, data);
  if (!response.success) {
    throw new Error(response.error || "Failed to update route config");
  }
}

/**
 * Bulk update multiple route configurations
 */
export async function bulkUpdateRouteConfigs(
  ids: number[],
  data: Partial<RouteConfig>,
): Promise<void> {
  const response = await apiClient.post("/routes-config/bulk-update", {
    ids,
    ...data,
  });
  if (!response.success) {
    throw new Error(response.error || "Failed to bulk update route configs");
  }
}
