/**
 * Route Management Logic Composable
 * Centralize all route management business logic
 * Single Responsibility: Handle all route operations
 */

import { ref, computed, Ref } from "vue";
import axios from "axios";
import { useRouteCacheManager } from "@/composables/useRouteCacheManager";
import type { RouteConfig, PresetType } from "../types/routeManagement";

export function useRouteManagementLogic() {
  const cacheManager = useRouteCacheManager();

  // State
  const isSaving = ref(false);
  const error = ref("");
  const allRoutes: Ref<RouteConfig[]> = ref([]);
  const categories: Ref<string[]> = ref([]);

  // Computed
  const isLoading = computed(() => cacheManager.isFetching.value);
  const enabledCount = computed(
    () => allRoutes.value.filter((r) => r.enabled).length,
  );
  const cachingEnabledCount = computed(
    () => allRoutes.value.filter((r) => r.caching_enabled).length,
  );
  const queueEnabledCount = computed(
    () => allRoutes.value.filter((r) => r.queue_enabled).length,
  );

  /**
   * Fetch routes from cache or API
   */
  const fetchRouteConfigs = async (forceRefresh = false): Promise<void> => {
    error.value = cacheManager.fetchError.value || "";
    try {
      const routes = await cacheManager.fetchRouteConfigs(forceRefresh);
      allRoutes.value = routes;
      extractCategories();
    } catch (err: any) {
      error.value = err.message || "Failed to load routes";
      console.error("Error loading routes:", err);
    }
  };

  /**
   * Extract unique categories from routes
   */
  const extractCategories = (): void => {
    const cats = new Set<string>();
    allRoutes.value.forEach((route) => {
      if (route.category) cats.add(route.category);
    });
    categories.value = Array.from(cats).sort();
  };

  /**
   * Toggle route enabled state
   */
  const toggleRouteEnabled = async (
    routeId: string,
    currentState: boolean,
  ): Promise<void> => {
    try {
      await axios.patch(`/api/routes-config/${routeId}`, {
        enabled: !currentState,
      });
      const route = allRoutes.value.find((r) => r.id === routeId);
      if (route) route.enabled = !currentState;
    } catch (err: any) {
      error.value = "Failed to update route";
      console.error("Error toggling route:", err);
    }
  };

  /**
   * Bulk update routes
   */
  const bulkUpdateEnabled = async (
    routeIds: string[],
    enabled: boolean,
  ): Promise<void> => {
    try {
      await axios.post("/api/routes-config/bulk-update", {
        routeIds,
        updates: { enabled },
      });

      routeIds.forEach((id) => {
        const route = allRoutes.value.find((r) => r.id === id);
        if (route) route.enabled = enabled;
      });
    } catch (err: any) {
      error.value = "Failed to update routes";
      console.error("Error bulk updating:", err);
    }
  };

  /**
   * Save route (create or update)
   */
  const saveRoute = async (route: RouteConfig): Promise<void> => {
    if (!route) return;

    isSaving.value = true;
    try {
      if (route.id) {
        const response = await axios.patch(
          `/api/routes-config/${route.id}`,
          route,
        );
        const index = allRoutes.value.findIndex((r) => r.id === route.id);
        if (index > -1) allRoutes.value[index] = response.data.data;
      } else {
        const response = await axios.post("/api/routes-config", route);
        allRoutes.value.push(response.data.data);
      }
      extractCategories();
    } catch (err: any) {
      error.value = err.response?.data?.error || "Failed to save route";
      console.error("Error saving route:", err);
    } finally {
      isSaving.value = false;
    }
  };

  /**
   * Delete route
   */
  const deleteRoute = async (routeId: string): Promise<void> => {
    if (!confirm("Delete this route?")) return;

    try {
      await axios.delete(`/api/routes-config/${routeId}`);
      allRoutes.value = allRoutes.value.filter((r) => r.id !== routeId);
    } catch (err: any) {
      error.value = "Failed to delete route";
      console.error("Error deleting route:", err);
    }
  };

  /**
   * Bulk delete routes
   */
  const bulkDelete = async (routeIds: string[]): Promise<void> => {
    if (!confirm(`Delete ${routeIds.length} routes?`)) return;

    try {
      await Promise.all(
        routeIds.map((id) => axios.delete(`/api/routes-config/${id}`)),
      );
      allRoutes.value = allRoutes.value.filter(
        (r) => !routeIds.includes(r.id || ""),
      );
    } catch (err: any) {
      error.value = "Failed to delete routes";
      console.error("Error deleting routes:", err);
    }
  };

  /**
   * Apply preset to routes
   */
  const applyPreset = async (
    routeIds: string[],
    preset: PresetType,
  ): Promise<void> => {
    try {
      await axios.post(`/api/routes-config/apply-preset/${preset}`, {
        routeIds,
      });
      // Force refresh to update cache
      await fetchRouteConfigs(true);
    } catch (err: any) {
      error.value = `Failed to apply preset: ${err.response?.data?.error || err.message}`;
      console.error("Error applying preset:", err);
    }
  };

  /**
   * Cache configuration methods
   */
  const updateCacheConfig = (setting: string, value: any): void => {
    const newConfig: any = {};
    newConfig[setting] = value;
    cacheManager.updateCacheConfig(newConfig);
  };

  const handleToggleCaching = (enabled: boolean): void => {
    updateCacheConfig("enabled", enabled);
  };

  const handleCacheTTLChange = (newTTL: number): void => {
    updateCacheConfig("ttlSeconds", newTTL);
  };

  const handleAutoRefreshChange = (enabled: boolean): void => {
    const interval = enabled ? 30000 : 0;
    updateCacheConfig("autoRefreshMs", interval);
  };

  const handleExcludeRoutesChange = (routes: string[]): void => {
    updateCacheConfig("excludeRoutes", routes);
    console.log("🚫 Cache exclusion list updated:", routes);
  };

  const clearCache = (): void => {
    cacheManager.clearCache();
    error.value = "";
  };

  return {
    // State
    isSaving,
    error,
    allRoutes,
    categories,
    isLoading,
    cacheManager,

    // Computed
    enabledCount,
    cachingEnabledCount,
    queueEnabledCount,

    // Methods
    fetchRouteConfigs,
    toggleRouteEnabled,
    bulkUpdateEnabled,
    saveRoute,
    deleteRoute,
    bulkDelete,
    applyPreset,
    handleToggleCaching,
    handleCacheTTLChange,
    handleAutoRefreshChange,
    handleExcludeRoutesChange,
    clearCache,
  };
}
