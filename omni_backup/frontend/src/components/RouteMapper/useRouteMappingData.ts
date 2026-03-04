import { computed, onMounted, ref, watch } from "vue";
import { useRoute, useRouter } from "vue-router";
import { getAuthHeaders, getApiBaseUrl } from "@/utils/apiHeaders";

// Get base URL for route-mapping API
const API_BASE_URL = getApiBaseUrl("/route-mapping");

export type ViewMode =
  | "categories"
  | "component"
  | "disconnected"
  | "backend"
  | "unused";

export interface RouteSummary {
  endpoint: string;
  method?: string;
  is_dynamic?: boolean;
  category?: string;
}

export interface FrontendOnlySummary {
  endpoint: string;
  components: string[];
  status?: string;
}

export interface RouteCategoryLabels {
  [key: string]: { title: string; detail: string };
}

export interface RouteCategoriesUi {
  connected: RouteSummary[];
  frontendOnly: FrontendOnlySummary[];
  backendOnly: RouteSummary[];
  unused: RouteSummary[];
}

export interface RouteData {
  total_routes: number;
  total_components: number;
  total_categories: number;
  total_dynamic_routes: number;
  total_called_routes: number;
  total_disconnected_routes: number;
  total_unused_routes: number;
  connection_rate: string;
  by_category?: Record<string, any[]>;
  categories?: RouteCategoriesUi;
  category_labels?: RouteCategoryLabels;
  category_stats?: Record<string, number>;
  components: Record<string, any>;
  disconnected_routes: Record<string, any>;
  backend_only_routes?: Record<string, any>;
  unused_routes?: Record<string, any>;
  button_to_endpoints?: Record<string, Record<string, string[]>>;
  statistics?: Record<string, any>;
  timestamp: string;
}

const defaultCategories: RouteCategoriesUi = {
  connected: [],
  frontendOnly: [],
  backendOnly: [],
  unused: [],
};

const toRouteSummary = (route: any): RouteSummary => ({
  endpoint: route?.endpoint || "",
  method: route?.method || "GET",
  is_dynamic: Boolean(route?.is_dynamic ?? route?.isDynamic),
  category: route?.category,
});

const parseCategories = (
  rawCategories?: any,
  fallback?: Record<string, any[]>
): RouteCategoriesUi => {
  const categories = rawCategories || {};
  const fallbackCategories = fallback || {};

  const connected = (
    categories.connected ||
    fallbackCategories.connected ||
    []
  ).map(toRouteSummary);

  const frontendOnly =
    categories.frontendOnly ||
    categories.frontend_only ||
    fallbackCategories.frontend_only ||
    [];

  const backendOnly = (
    categories.backendOnly ||
    categories.backend_only ||
    fallbackCategories.backend_only ||
    []
  ).map(toRouteSummary);

  const unused = (categories.unused || fallbackCategories.unused || []).map(
    toRouteSummary
  );

  return { connected, frontendOnly, backendOnly, unused };
};

export function useRouteMappingData() {
  const router = useRouter();
  const route = useRoute();
  const viewMode = ref<ViewMode>("categories");
  const searchQuery = ref("");
  const loading = ref(false);
  const error = ref("");
  const data = ref<RouteData | null>(null);

  const normalizedCategories = computed<RouteCategoriesUi>(() => {
    if (!data.value) return defaultCategories;
    return parseCategories(data.value.categories, data.value.by_category);
  });

  const categoryStats = computed<Record<string, number>>(() => {
    if (data.value?.category_stats) return data.value.category_stats;
    const categories = normalizedCategories.value;
    return {
      connected: categories.connected.length,
      frontend_only: categories.frontendOnly.length,
      backend_only: categories.backendOnly.length,
      unused: categories.unused.length,
    };
  });

  const updateURLQuery = (mode: ViewMode): void => {
    const query = { ...route.query, view: mode };
    router.push({ path: route.path, query }).catch(() => {});
  };

  const fetchData = async () => {
    loading.value = true;
    error.value = "";

    try {
      const response = await fetch(`${API_BASE_URL}/detailed-mapping`, {
        method: "GET",
        headers: getAuthHeaders(),
      });
      if (!response.ok) throw new Error("Failed to fetch route mapping");

      const result = await response.json();
      if (!result.success) {
        throw new Error(result.error || "No data returned");
      }

      const statsResponse = await fetch(`${API_BASE_URL}/statistics`, {
        method: "GET",
        headers: getAuthHeaders(),
      });
      const statsResult = await statsResponse.json();

      const categories = parseCategories(result.categories, result.by_category);
      const categoryStatsValue = result.category_stats || {
        connected: categories.connected.length,
        frontend_only: categories.frontendOnly.length,
        backend_only: categories.backendOnly.length,
        unused: categories.unused.length,
      };

      data.value = {
        total_routes: result.total_routes || 0,
        total_components: result.total_components || 0,
        total_categories: result.total_categories || 0,
        total_dynamic_routes: result.total_dynamic_routes || 0,
        total_called_routes: result.total_called_routes || 0,
        total_disconnected_routes: result.total_disconnected_routes || 0,
        total_unused_routes: result.total_unused_routes || 0,
        connection_rate: result.connection_rate || "0%",
        by_category: result.by_category || {},
        categories,
        category_labels: result.category_labels || {},
        category_stats: categoryStatsValue,
        components: result.components || {},
        disconnected_routes: result.disconnected_routes || {},
        backend_only_routes: result.backend_only_routes || {},
        unused_routes: result.unused_routes || {},
        button_to_endpoints: result.button_to_endpoints || {},
        timestamp: result.timestamp || new Date().toISOString(),
        statistics: statsResult.success ? statsResult.statistics : {},
      };
    } catch (err: any) {
      error.value = err.message || "Failed to load route mapping";
    } finally {
      loading.value = false;
    }
  };

  const filteredComponents = computed(() => {
    if (!data.value?.components) return {};

    const query = searchQuery.value.toLowerCase();
    const filtered: Record<string, any> = {};

    Object.entries(data.value.components).forEach(
      ([name, component]: [string, any]) => {
        if (!component.routes_called?.length) return;

        const matches =
          !query ||
          name.toLowerCase().includes(query) ||
          (component.path || "").toLowerCase().includes(query) ||
          component.routes_called?.some((r: string) =>
            r.toLowerCase().includes(query)
          ) ||
          component.buttons?.some((b: string) =>
            b.toLowerCase().includes(query)
          );

        if (matches) filtered[name] = component;
      }
    );

    return filtered;
  });

  const filteredDisconnected = computed(() => {
    const query = searchQuery.value.toLowerCase();
    return normalizedCategories.value.frontendOnly.filter(
      ({ endpoint, components }) => {
        if (!query) return true;
        return (
          endpoint.toLowerCase().includes(query) ||
          components.some((c: string) => c.toLowerCase().includes(query))
        );
      }
    );
  });

  const filteredBackendOnly = computed(() => {
    const query = searchQuery.value.toLowerCase();
    return normalizedCategories.value.backendOnly.filter((route) => {
      if (!query) return true;
      return route.endpoint.toLowerCase().includes(query);
    });
  });

  const filteredUnused = computed(() => {
    const query = searchQuery.value.toLowerCase();
    return normalizedCategories.value.unused.filter((route) => {
      if (!query) return true;
      return route.endpoint.toLowerCase().includes(query);
    });
  });

  const formatDate = (timestamp?: string): string => {
    if (!timestamp) return "Never";
    return new Date(timestamp).toLocaleString();
  };

  const refreshData = () => fetchData();

  onMounted(() => {
    const viewFromQuery = route.query.view as string;
    const validViews: ViewMode[] = [
      "categories",
      "component",
      "disconnected",
      "backend",
      "unused",
    ];

    if (viewFromQuery && validViews.includes(viewFromQuery as ViewMode)) {
      viewMode.value = viewFromQuery as ViewMode;
    }

    fetchData();
  });

  // Watch for URL query changes (browser back/forward, direct URL access)
  watch(
    () => route.query.view,
    (newView) => {
      const validViews: ViewMode[] = [
        "categories",
        "component",
        "disconnected",
        "backend",
        "unused",
      ];
      const viewValue = newView as string;

      if (
        viewValue &&
        validViews.includes(viewValue as ViewMode) &&
        viewValue !== viewMode.value
      ) {
        viewMode.value = viewValue as ViewMode;
      }
    }
  );

  return {
    data,
    loading,
    error,
    viewMode,
    searchQuery,
    normalizedCategories,
    categoryStats,
    filteredComponents,
    filteredDisconnected,
    filteredBackendOnly,
    filteredUnused,
    refreshData,
    formatDate,
    updateURLQuery,
  };
}
