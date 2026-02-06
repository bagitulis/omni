/**
 * Centralized Cache Service for React
 * SRP: Manage all caching with consistent invalidation strategy
 * Similar to Vue frontend's cacheService.ts
 */

interface CacheEntry<T = unknown> {
  data: T;
  timestamp: number;
  version?: string;
}

interface CacheOptions {
  ttl?: number; // milliseconds
  version?: string;
  skipCache?: boolean;
}

const DEFAULT_TTL = 5 * 60 * 1000; // 5 minutes default
const CACHE_VERSION_KEY = "_cache_version";

class CacheService {
  private memoryCache = new Map<string, CacheEntry>();
  private cacheVersion: string;

  constructor() {
    this.cacheVersion = this.getCurrentVersion();
    this.validateAllCache();
  }

  /**
   * Get current cache version from localStorage
   * Version changes invalidate all cache
   */
  private getCurrentVersion(): string {
    const stored = localStorage.getItem(CACHE_VERSION_KEY);
    if (!stored) {
      const newVersion = Date.now().toString();
      localStorage.setItem(CACHE_VERSION_KEY, newVersion);
      return newVersion;
    }
    return stored;
  }

  /**
   * Invalidate all cache by changing version
   */
  invalidateAll(): void {
    const newVersion = Date.now().toString();
    localStorage.setItem(CACHE_VERSION_KEY, newVersion);
    this.cacheVersion = newVersion;
    this.memoryCache.clear();
    this.clearLocalStorageCache();
  }

  /**
   * Clear cache entries from localStorage
   */
  private clearLocalStorageCache(): void {
    const keysToRemove: string[] = [];
    for (let i = 0; i < localStorage.length; i++) {
      const key = localStorage.key(i);
      if (key && this.isCacheKey(key)) {
        keysToRemove.push(key);
      }
    }
    keysToRemove.forEach((key) => localStorage.removeItem(key));
  }

  /**
   * Check if key is a cache key
   */
  private isCacheKey(key: string): boolean {
    const cacheKeys = [
      "api_cache_",
      "route_configs_cache",
      "inventory_filter_",
      "_cache_entry",
    ];
    return cacheKeys.some((prefix) => key.startsWith(prefix));
  }

  /**
   * Validate cache entry is not expired and version matches
   */
  private isValid<T>(entry: CacheEntry<T>, ttl: number): boolean {
    if (entry.version && entry.version !== this.cacheVersion) {
      return false;
    }
    const now = Date.now();
    return now - entry.timestamp < ttl;
  }

  /**
   * Validate all cache on initialization
   */
  private validateAllCache(): void {
    for (const [key, entry] of this.memoryCache) {
      if (!this.isValid(entry, DEFAULT_TTL)) {
        this.memoryCache.delete(key);
      }
    }
  }

  /**
   * Get data from cache (memory first, then localStorage)
   */
  get<T>(key: string, options: CacheOptions = {}): T | null {
    const { ttl = DEFAULT_TTL, skipCache = false } = options;

    if (skipCache) return null;

    // Check memory cache first
    const memEntry = this.memoryCache.get(key);
    if (memEntry && this.isValid(memEntry, ttl)) {
      return memEntry.data as T;
    }

    // Check localStorage
    try {
      const stored = localStorage.getItem(`api_cache_${key}`);
      if (!stored) return null;

      const entry: CacheEntry<T> = JSON.parse(stored);
      if (this.isValid(entry, ttl)) {
        // Restore to memory cache
        this.memoryCache.set(key, entry);
        return entry.data;
      } else {
        // Expired, remove
        localStorage.removeItem(`api_cache_${key}`);
      }
    } catch {
      // Invalid data, remove
      localStorage.removeItem(`api_cache_${key}`);
    }

    return null;
  }

  /**
   * Set data in cache (both memory and localStorage)
   */
  set<T>(key: string, data: T, options: CacheOptions = {}): void {
    const { version = this.cacheVersion } = options;

    const entry: CacheEntry<T> = {
      data,
      timestamp: Date.now(),
      version,
    };

    // Set in memory
    this.memoryCache.set(key, entry);

    // Set in localStorage (with error handling)
    try {
      localStorage.setItem(`api_cache_${key}`, JSON.stringify(entry));
    } catch (error) {
      // localStorage full or disabled, continue with memory cache only
      console.warn("localStorage unavailable:", error);
    }
  }

  /**
   * Remove specific cache entry
   */
  remove(key: string): void {
    this.memoryCache.delete(key);
    localStorage.removeItem(`api_cache_${key}`);
  }

  /**
   * Remove multiple cache entries by prefix
   */
  removeByPrefix(prefix: string): void {
    // Memory cache
    for (const key of this.memoryCache.keys()) {
      if (key.startsWith(prefix)) {
        this.memoryCache.delete(key);
      }
    }

    // localStorage
    const keysToRemove: string[] = [];
    for (let i = 0; i < localStorage.length; i++) {
      const key = localStorage.key(i);
      if (key && key.startsWith(`api_cache_${prefix}`)) {
        keysToRemove.push(key);
      }
    }
    keysToRemove.forEach((key) => localStorage.removeItem(key));
  }

  /**
   * Get cache with auto-fetch if miss
   */
  async getOrFetch<T>(
    key: string,
    fetcher: () => Promise<T>,
    options: CacheOptions = {},
  ): Promise<T> {
    const cached = this.get<T>(key, options);
    if (cached !== null) {
      return cached;
    }

    const data = await fetcher();
    this.set(key, data, options);
    return data;
  }

  /**
   * Get cache stats
   */
  getStats(): { memoryEntries: number; storageEntries: number } {
    let storageEntries = 0;
    for (let i = 0; i < localStorage.length; i++) {
      const key = localStorage.key(i);
      if (key && this.isCacheKey(key)) {
        storageEntries++;
      }
    }

    return {
      memoryEntries: this.memoryCache.size,
      storageEntries,
    };
  }

  /**
   * Clear all cache for specific type
   */
  clearType(type: "inventory" | "orders" | "products" | "all"): void {
    const prefixes: Record<string, string[]> = {
      inventory: ["inventory_"],
      orders: ["orders_"],
      products: ["products_", "master-products_"],
      all: [""],
    };

    const targetPrefixes = prefixes[type] || [];
    targetPrefixes.forEach((prefix) => {
      if (prefix === "") {
        this.invalidateAll();
      } else {
        this.removeByPrefix(prefix);
      }
    });
  }
}

// Singleton instance
export const cacheService = new CacheService();
export default cacheService;
