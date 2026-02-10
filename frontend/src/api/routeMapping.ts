import api from "./client";
import { RouteData, RouteStatistics } from "@/types/routeMapping";

export async function getRouteMappingDetailed(): Promise<RouteData> {
  const response = await api.get<RouteData>("/route-mapping/detailed-mapping");

  if (!response.success) {
    throw new Error(response.error || "Failed to fetch route mapping");
  }

  if (!response.data) {
    throw new Error("No data returned from route mapping API");
  }

  return response.data;
}

export async function getRouteMappingStatistics(): Promise<RouteStatistics> {
  // Cast to any because the backend returns { success: true, statistics: {...} }
  // which doesn't match standard ApiResponse<T> where data is in 'data' field.
  const response = await api.get<any>("/route-mapping/statistics");

  if (!response.success) {
    throw new Error(response.error || "Failed to fetch statistics");
  }

  // Check for statistics property on the response object itself (non-standard)
  // or inside data property (standard)
  const stats = (response as any).statistics || response.data?.statistics;

  return { statistics: stats || {} };
}
