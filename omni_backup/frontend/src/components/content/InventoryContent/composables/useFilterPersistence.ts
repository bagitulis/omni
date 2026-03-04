import { ref } from "vue";
import cacheService from "@/services/cacheService";
import { getAuthHeaders } from "@/utils/apiHeaders";

/**
 * RESPONSIBILITY: API sync & cache persistence
 * - Handle all storage operations via cacheService
 * - Sync config to backend API
 * - Retry logic for network failures
 */

const STORAGE_KEY = {
  VISIBLE: "inventory_visible_columns",
  LOCKED: "inventory_locked_columns",
  FILTERS: "inventory_column_filters",
  SEARCH: "inventory_search_query",
  ORDER: "inventory_column_order",
} as const;

// Keys that should use direct localStorage (not cacheService)
const DIRECT_STORAGE_KEYS = [STORAGE_KEY.ORDER] as string[];

const CACHE_TTL = 24 * 60 * 60 * 1000; // 24 hours for filter preferences

function getApiBaseUrl() {
  const host = window.location.hostname;
  const protocol = window.location.protocol;
  // In production (non-localhost), use same origin without port
  // In development (localhost), use port 3000 for backend
  const isLocalhost = host === "localhost" || host === "127.0.0.1";
  return isLocalhost
    ? `${protocol}//${host}:3000/api`
    : `${protocol}//${host}/api`;
}

export function useFilterPersistence() {
  const lastSyncTime = ref<number>(0);
  const syncError = ref<string>("");

  /**
   * Load from storage - use direct localStorage for certain keys
   */
  function loadFromStorage(key: string, defaultValue: any): any {
    try {
      // For ORDER and other direct storage keys, use localStorage directly
      if (DIRECT_STORAGE_KEYS.includes(key)) {
        const stored = localStorage.getItem(key);
        console.log(`📖 [loadFromStorage] key=${key}, stored=`, stored);
        if (stored) {
          const parsed = JSON.parse(stored);
          console.log(`📖 [loadFromStorage] parsed=`, parsed);
          return parsed;
        }
        return defaultValue;
      }

      // For other keys, use cacheService
      const stored = cacheService.get(key, { ttl: CACHE_TTL });
      if (stored) {
        // Only convert to Set for VISIBLE and LOCKED keys
        if (
          stored instanceof Array &&
          (key === STORAGE_KEY.VISIBLE || key === STORAGE_KEY.LOCKED)
        ) {
          return new Set(stored);
        }
        return stored;
      }
    } catch (error) {
      console.error(`Error loading ${key}:`, error);
    }
    return defaultValue;
  }

  /**
   * Save to storage - use direct localStorage for certain keys
   */
  function saveToStorage(key: string, value: any) {
    try {
      const toSave = value instanceof Set ? Array.from(value) : value;

      // For ORDER and other direct storage keys, use localStorage directly
      if (DIRECT_STORAGE_KEYS.includes(key)) {
        console.log(`💾 [saveToStorage] key=${key}, value=`, toSave);
        localStorage.setItem(key, JSON.stringify(toSave));
        // Verify immediately
        const verify = localStorage.getItem(key);
        console.log(`💾 [saveToStorage] verified=`, verify);
        return;
      }

      // For other keys, use cacheService
      cacheService.set(key, toSave, { ttl: CACHE_TTL });
    } catch (error) {
      console.error(`Error saving ${key}:`, error);
    }
  }

  async function syncToAPI(payload: any, retries = 3): Promise<boolean> {
    try {
      const enrichedPayload = {
        platform: "inventory",
        page: "inventory",
        ...payload,
      };

      for (let attempt = 1; attempt <= retries; attempt++) {
        try {
          const url = `${getApiBaseUrl()}/filter-preferences`;

          const response = await fetch(url, {
            method: "POST",
            headers: getAuthHeaders(),
            body: JSON.stringify(enrichedPayload),
          });

          if (response.ok) {
            await response.json();
            lastSyncTime.value = Date.now();
            syncError.value = "";
            return true;
          } else if (attempt < retries) {
            await new Promise((r) => setTimeout(r, 500));
          } else {
            syncError.value = `Sync failed: ${response.status}`;

            return false;
          }
        } catch (err) {
          if (attempt < retries) {
            await new Promise((r) => setTimeout(r, 500));
          } else {
            throw err;
          }
        }
      }
      return false;
    } catch (error) {
      syncError.value = String(error);
      return false;
    }
  }

  async function loadFromAPI(): Promise<{
    lockedColumns: string[];
    visibleColumns: string[];
    columnFilters: Record<string, string>;
    searchQuery: string;
  } | null> {
    try {
      // Add cache-busting param to force fresh fetch
      const cacheBuster = `_cb=${Date.now()}`;
      const url = `${getApiBaseUrl()}/filter-preferences?platform=inventory&page=inventory&${cacheBuster}`;

      // Create abort controller for timeout
      const controller = new AbortController();
      const timeout = setTimeout(() => controller.abort(), 5000); // 5 second timeout

      const response = await fetch(url, {
        cache: "no-cache",
        headers: {
          ...getAuthHeaders(),
          "Cache-Control": "no-cache, must-revalidate",
        },
        signal: controller.signal,
      });
      clearTimeout(timeout);

      if (!response.ok) {
        return null;
      }

      const result = await response.json();

      if (result.success && result.data) {
        const data = result.data;

        // CRITICAL: Ensure visibleColumns and lockedColumns are always arrays
        // Backend might return them as strings (JSON stringified) or arrays
        let visibleColumns = data.visibleColumns;
        if (typeof visibleColumns === "string") {
          try {
            visibleColumns = JSON.parse(visibleColumns);
          } catch {
            visibleColumns = [];
          }
        }
        if (!Array.isArray(visibleColumns)) {
          visibleColumns = [];
        }

        let lockedColumns = data.lockedColumns;
        if (typeof lockedColumns === "string") {
          try {
            lockedColumns = JSON.parse(lockedColumns);
          } catch {
            lockedColumns = [];
          }
        }
        if (!Array.isArray(lockedColumns)) {
          lockedColumns = [];
        }

        // Return normalized data
        const normalizedData = {
          ...data,
          visibleColumns: visibleColumns as string[],
          lockedColumns: lockedColumns as string[],
        };

        return normalizedData;
      } else {
        return null;
      }
    } catch {
      return null;
    }
  }

  return {
    STORAGE_KEY,
    lastSyncTime,
    syncError,
    loadFromStorage,
    saveToStorage,
    syncToAPI,
    loadFromAPI,
  };
}
