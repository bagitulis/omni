import { useState, useMemo } from "react";
import { useQuery } from "@tanstack/react-query";
import {
  getRouteMappingDetailed,
  getRouteMappingStatistics,
} from "@/api/routeMapping";
import {
  ViewMode,
  RouteData,
  RouteCategoriesUi,
  RouteSummary,
  FrontendOnlySummary,
} from "@/types/routeMapping";

export function useRouteMapping() {
  const [viewMode, setViewMode] = useState<ViewMode>("categories");
  const [searchQuery, setSearchQuery] = useState("");

  const { data, isLoading, error, refetch, isFetching } = useQuery<RouteData>({
    queryKey: ["route-mapping"],
    queryFn: async () => {
      const [mapping, stats] = await Promise.all([
        getRouteMappingDetailed(),
        getRouteMappingStatistics(),
      ]);

      // Merge statistics
      const fullData = { ...mapping, statistics: stats.statistics };

      // Apply Vue-like normalization if needed
      // Vue: parseCategories(result.categories, result.by_category)
      // We'll assume the API returns 'categories' populated, or we use 'by_category' as fallback
      // But since we can't easily replicate the exact Vue logic without knowing the raw API response shape perfectly,
      // we'll rely on the defined types.
      // However, let's ensure 'categories' has the 4 keys

      const categories: RouteCategoriesUi = {
        connected:
          fullData.categories?.connected ||
          fullData.by_category?.connected ||
          [],
        frontend_only:
          fullData.categories?.frontend_only ||
          fullData.by_category?.frontend_only ||
          [],
        backend_only:
          fullData.categories?.backend_only ||
          fullData.by_category?.backend_only ||
          [],
        unused:
          fullData.categories?.unused || fullData.by_category?.unused || [],
      };

      return { ...fullData, categories };
    },
  });

  const categoryStats = useMemo(() => {
    if (!data)
      return { connected: 0, frontend_only: 0, backend_only: 0, unused: 0 };
    return {
      connected: data.categories.connected.length,
      frontend_only: data.categories.frontend_only.length,
      backend_only: data.categories.backend_only.length,
      unused: data.categories.unused.length,
    };
  }, [data]);

  const filteredComponents = useMemo(() => {
    if (!data?.components) return {};
    const query = searchQuery.toLowerCase();
    if (!query) return data.components;

    const filtered: Record<string, any> = {};
    Object.entries(data.components).forEach(([name, component]) => {
      if (!component.routes_called?.length) return;

      const matches =
        name.toLowerCase().includes(query) ||
        (component.path || "").toLowerCase().includes(query) ||
        component.routes_called?.some((r: string) =>
          r.toLowerCase().includes(query),
        ) ||
        component.buttons?.some((b: string) => b.toLowerCase().includes(query));

      if (matches) filtered[name] = component;
    });
    return filtered;
  }, [data, searchQuery]);

  const filteredLists = useMemo(() => {
    if (!data)
      return { connected: [], frontendOnly: [], backendOnly: [], unused: [] };
    const query = searchQuery.toLowerCase();

    const filterRoute = (r: RouteSummary) =>
      !query || r.endpoint.toLowerCase().includes(query);

    const filterFrontend = (r: FrontendOnlySummary) =>
      !query ||
      r.endpoint.toLowerCase().includes(query) ||
      r.components.some((c) => c.toLowerCase().includes(query));

    return {
      connected: data.categories.connected.filter(filterRoute),
      frontendOnly: data.categories.frontend_only.filter(filterFrontend),
      backendOnly: data.categories.backend_only.filter(filterRoute),
      unused: data.categories.unused.filter(filterRoute),
    };
  }, [data, searchQuery]);

  return {
    data,
    isLoading,
    error,
    refetch,
    isFetching,
    viewMode,
    setViewMode,
    searchQuery,
    setSearchQuery,
    categoryStats,
    filteredComponents,
    filteredLists,
  };
}
