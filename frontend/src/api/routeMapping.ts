import api from "./client";
import { RouteData, RouteStatistics } from "@/types/routeMapping";

export async function getRouteMappingDetailed(): Promise<RouteData> {
  const response = await api.get<RouteData>("/routes/mapping");

  if (!response.success || !response.data) {
    throw new Error(response.error || "Failed to fetch route mapping");
  }

  return response.data;
}

export async function getRouteMappingStatistics(): Promise<RouteStatistics> {
  const response = await api.get<{ statistics?: Record<string, unknown> }>(
    "/routes/mapping/stats",
  );

  if (!response.success) {
    throw new Error(response.error || "Failed to fetch statistics");
  }

  return { statistics: response.data?.statistics || {} };
}
