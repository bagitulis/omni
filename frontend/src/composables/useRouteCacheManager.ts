import { ref, computed, Ref } from "vue";
import axios from "axios";

/**
 * Route Cache Manager
 * Manage frontend caching dengan smart invalidation
 * Single Responsibility: Cache management & validation
 */

interface CacheConfig {
  enabled: boolean; // Enable/disable caching
  ttlSeconds: number; // Cache TTL in seconds
  maxRetries: number; // Retry attempts jika fetch fail
  autoRefreshMs: number; // Auto refresh interval (0 = disabled)
  excludeRoutes?: string[]; // Routes to exclude from caching
  includeOnlyRoutes?: string[]; // If set, only cache these routes
}

interface CacheEntry<T> {
  data: T;
  timestamp: number;
  version: string | null; // ETag/version for smart invalidation
}

interface RouteConfig {
  id?: string;
  routePath: string;
  routeMethod: string;
  routeName?: string;
  description?: string;
  category?: string;
  enabled: boolean;
  cachingEnabled: boolean;
  cacheTTL: number;
  cacheStrategy: string;
  queueEnabled: boolean;
  queueMaxSize: number;
  queuePriority: string;
  maxConcurrent: number;
  rateLimitEnabled: boolean;
  rateLimitWindow: number;
  rateLimitMax: number;
  minIntervalMs: number;
  timeout: number;
  retryEnabled: boolean;
  maxRetries: number;
  retryDelayMs: number;
  customConfig?: string;
}

const CACHE_STORAGE_KEY = "route_configs_cache";
const CACHE_CONFIG_STORAGE_KEY = "route_cache_config";

// Default cache config
const DEFAULT_CACHE_CONFIG: CacheConfig = {
  enabled: true,
  ttlSeconds: 300, // 5 menit default
  maxRetries: 2,
  autoRefreshMs: 0, // No auto refresh by default
  excludeRoutes: ["/api/orders", "/api/inventory"], // Exclude order & inventory from caching
  includeOnlyRoutes: undefined, // If set, only these routes are cached
};

export function useRouteCacheManager() {
  // Cache state
  const routeCache: Ref<CacheEntry<RouteConfig[]> | null> = ref(null);
  const cacheConfig: Ref<CacheConfig> = ref(DEFAULT_CACHE_CONFIG);
  const isFetching = ref(false);
  const fetchError: Ref<string | null> = ref(null);
  const lastFetchTime: Ref<number | null> = ref(null);
  let autoRefreshInterval: number | null = null;

  /**
   * Load cache config from localStorage
   */
  const loadCacheConfig = (): void => {
    const stored = localStorage.getItem(CACHE_CONFIG_STORAGE_KEY);
    if (stored) {
      try {
        cacheConfig.value = { ...DEFAULT_CACHE_CONFIG, ...JSON.parse(stored) };
      } catch {
        cacheConfig.value = DEFAULT_CACHE_CONFIG;
      }
    }
  };

  /**
   * Save cache config to localStorage
   */
  const saveCacheConfig = (): void => {
    localStorage.setItem(
      CACHE_CONFIG_STORAGE_KEY,
      JSON.stringify(cacheConfig.value)
    );
  };

  /**
   * Load cache from localStorage
   */
  const loadCacheFromStorage = (): CacheEntry<RouteConfig[]> | null => {
    const stored = localStorage.getItem(CACHE_STORAGE_KEY);
    if (!stored) return null;

    try {
      return JSON.parse(stored);
    } catch {
      console.warn("Failed to parse cached route configs");
      return null;
    }
  };

  /**
   * Save cache to localStorage
   */
  const saveCacheToStorage = (entry: CacheEntry<RouteConfig[]>): void => {
    try {
      localStorage.setItem(CACHE_STORAGE_KEY, JSON.stringify(entry));
    } catch {
      console.warn("Failed to save route configs to cache");
    }
  };

  /**
   * Check if a specific route should be cached based on exclude/include rules
   */
  const shouldCacheRoute = (routePath: string): boolean => {
    // If caching disabled globally, don't cache anything
    if (!cacheConfig.value.enabled) return false;

    // Check exclude list first
    if (
      cacheConfig.value.excludeRoutes &&
      cacheConfig.value.excludeRoutes.length > 0
    ) {
      for (const excludePath of cacheConfig.value.excludeRoutes) {
        if (routePath.includes(excludePath)) {
          return false;
        }
      }
    }

    // If include list specified, only cache those
    if (
      cacheConfig.value.includeOnlyRoutes &&
      cacheConfig.value.includeOnlyRoutes.length > 0
    ) {
      for (const includePath of cacheConfig.value.includeOnlyRoutes) {
        if (routePath.includes(includePath)) {
          return true;
        }
      }
      return false;
    }

    // Default: cache if enabled
    return true;
  };

  /**
   * Check if cache is still valid
   */
  const isCacheValid = (): boolean => {
    if (!cacheConfig.value.enabled || !routeCache.value) return false;
    const age = Date.now() - routeCache.value.timestamp;
    return age < cacheConfig.value.ttlSeconds * 1000;
  };

  /**
   * Clear cache
   */
  const clearCache = (): void => {
    routeCache.value = null;
    localStorage.removeItem(CACHE_STORAGE_KEY);
  };

  /**
   * Update cache config (user configurable)
   */
  const updateCacheConfig = (newConfig: Partial<CacheConfig>): void => {
    cacheConfig.value = { ...cacheConfig.value, ...newConfig };
    saveCacheConfig();

    // Restart auto-refresh jika interval berubah
    if (newConfig.autoRefreshMs !== undefined) {
      stopAutoRefresh();
      if (newConfig.autoRefreshMs > 0) {
        startAutoRefresh();
      }
    }
  };

  /**
   * Fetch with retry logic
   */
  const fetchWithRetry = async (url: string, retryCount = 0): Promise<any> => {
    try {
      const config: any = {};

      // Add ETag header if we have cached version
      if (routeCache.value?.version) {
        config.headers = {
          "If-None-Match": routeCache.value.version,
        };
      }

      const response = await axios.get(url, config);

      // Store new version if available
      const version = response.headers["etag"] || null;

      return { data: response.data, version, status: response.status };
    } catch (error: any) {
      // 304 Not Modified - use cached data
      if (error.response?.status === 304) {
        return { data: null, version: null, status: 304 };
      }

      // Retry logic
      if (retryCount < cacheConfig.value.maxRetries) {
        await new Promise(
          (resolve) => setTimeout(resolve, Math.pow(2, retryCount) * 1000) // Exponential backoff
        );
        return fetchWithRetry(url, retryCount + 1);
      }

      throw error;
    }
  };

  /**
   * Fetch route configs dengan smart caching
   */
  const fetchRouteConfigs = async (force = false): Promise<RouteConfig[]> => {
    // If cache valid and not force refresh, return cached
    if (!force && isCacheValid() && routeCache.value) {
      return routeCache.value.data;
    }

    isFetching.value = true;
    fetchError.value = null;

    try {
      const response = await fetchWithRetry("/api/routes-config/all");

      // 304 Not Modified
      if (response.status === 304 && routeCache.value) {
        routeCache.value.timestamp = Date.now(); // Update timestamp
        saveCacheToStorage(routeCache.value);
        return routeCache.value.data;
      }

      // Success - update cache
      if (response.data?.success) {
        const newEntry: CacheEntry<RouteConfig[]> = {
          data: response.data.data.routes,
          timestamp: Date.now(),
          version: response.version,
        };

        routeCache.value = newEntry;
        saveCacheToStorage(newEntry);
        lastFetchTime.value = Date.now();

        return response.data.data.routes;
      }

      throw new Error("Invalid response format");
    } catch (error: any) {
      // If fetch fail, fallback to cache even if expired
      if (routeCache.value) {
        fetchError.value = "Using cached data (network error)";
        return routeCache.value.data;
      }

      fetchError.value = error.message || "Failed to fetch route configs";
      console.error("❌ Failed to fetch route configs:", error);
      return [];
    } finally {
      isFetching.value = false;
    }
  };

  /**
   * Start auto-refresh
   */
  const startAutoRefresh = (): void => {
    if (cacheConfig.value.autoRefreshMs <= 0) return;

    autoRefreshInterval = window.setInterval(async () => {
      await fetchRouteConfigs(true); // Force refresh
    }, cacheConfig.value.autoRefreshMs);
  };

  /**
   * Stop auto-refresh
   */
  const stopAutoRefresh = (): void => {
    if (autoRefreshInterval) {
      clearInterval(autoRefreshInterval);
      autoRefreshInterval = null;
    }
  };

  /**
   * Get cache stats
   */
  const getCacheStats = computed(() => {
    if (!routeCache.value) {
      return {
        cached: false,
        size: 0,
        age: 0,
        ageSeconds: 0,
        isValid: false,
        hitRate: 0,
      };
    }

    const age = Date.now() - routeCache.value.timestamp;
    return {
      cached: true,
      size: routeCache.value.data.length,
      age,
      ageSeconds: Math.round(age / 1000),
      isValid: isCacheValid(),
      hitRate: 0, // Can be enhanced to track hit rate
    };
  });

  // Load initial cache config
  loadCacheConfig();

  // Load cache from storage on init
  routeCache.value = loadCacheFromStorage();

  return {
    // State
    isFetching,
    fetchError,
    lastFetchTime,
    cacheConfig,
    cacheStats: getCacheStats,

    // Methods
    fetchRouteConfigs,
    updateCacheConfig,
    clearCache,
    startAutoRefresh,
    stopAutoRefresh,
    isCacheValid,
    shouldCacheRoute,
  };
}
