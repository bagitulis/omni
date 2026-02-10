import api from "./client";
import {
  RouteData,
  RouteStatistics,
  BackendRoute,
  BackendAnalysis,
} from "@/types/routeMapping";

interface RouteSummary {
  endpoint: string;
  method?: string;
  is_dynamic?: boolean;
  category?: string;
}

export async function getRouteMappingDetailed(): Promise<RouteData> {
  // Use correct backend endpoint: /routes (returns { success: true, data: BackendRoute[], total: number })
  // Note: Backend returns 'total' at top level, but our client types map 'data' to T.
  // We can derive total from the array length.
  const response = await api.get<BackendRoute[]>("/routes");

  if (!response.success) {
    throw new Error(response.error || "Failed to fetch route mapping");
  }

  const routes = response.data || [];

  // Transform BackendRoute[] to RouteData structure
  const byCategory: Record<string, RouteSummary[]> = {};

  routes.forEach((route) => {
    const category =
      route.tags && route.tags.length > 0 ? route.tags[0] : "uncategorized";
    if (!byCategory[category]) {
      byCategory[category] = [];
    }

    byCategory[category].push({
      endpoint: route.path,
      method: route.method,
      is_dynamic: route.path.includes(":"),
      category: category,
    });
  });

  const categories = Object.keys(byCategory);
  const backendOnly = Object.values(byCategory).flat();

  // Construct minimal RouteData to satisfy interface
  // Note: Backend doesn't provide component mapping yet, so many fields are mocked/empty
  return {
    total_routes: routes.length,
    total_components: 0,
    total_categories: categories.length,
    total_dynamic_routes: routes.filter((r) => r.path.includes(":")).length,
    total_called_routes: 0,
    total_disconnected_routes: 0,
    total_unused_routes: 0,
    connection_rate: "0%",
    by_category: byCategory,
    categories: {
      connected: [],
      frontend_only: [],
      backend_only: backendOnly,
      unused: [],
    },
    category_labels: {},
    category_stats: {},
    components: {},
    disconnected_routes: {},
    backend_only_routes: {},
    unused_routes: {},
    button_to_endpoints: {},
    timestamp: new Date().toISOString(),
  };
}

export async function getRouteMappingStatistics(): Promise<RouteStatistics> {
  // Use correct backend endpoint: /routes/analyze (returns { success: true, data: BackendAnalysis })
  const response = await api.get<BackendAnalysis>("/routes/analyze");

  if (!response.success) {
    throw new Error(response.error || "Failed to fetch statistics");
  }

  // The backend returns the analysis object directly in data
  return { statistics: response.data || {} };
}
