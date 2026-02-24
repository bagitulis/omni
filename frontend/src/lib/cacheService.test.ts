import { describe, it, expect, vi, beforeEach, afterEach } from "vitest";

// ---- localStorage mock ----
const localStorageMock = (() => {
  let store: Record<string, string> = {};
  return {
    getItem: vi.fn((key: string) => store[key] ?? null),
    setItem: vi.fn((key: string, value: string) => {
      store[key] = value;
    }),
    removeItem: vi.fn((key: string) => {
      delete store[key];
    }),
    clear: vi.fn(() => {
      store = {};
    }),
    get length() {
      return Object.keys(store).length;
    },
    key: vi.fn((index: number) => Object.keys(store)[index] ?? null),
  };
})();

Object.defineProperty(window, "localStorage", {
  value: localStorageMock,
  writable: true,
});

// Import AFTER mocking localStorage (so CacheService constructor uses the mock)
// We use a factory function to create fresh instances per test
async function createFreshCacheService() {
  // Reset local mock store
  localStorageMock.clear();
  vi.clearAllMocks();

  // Dynamic re-import doesn't work in Vitest due to module caching,
  // so we test the singleton directly but reset its state via its methods.
  const module = await import("./cacheService");
  return module.cacheService;
}

describe("cacheService", () => {
  let cacheService: Awaited<ReturnType<typeof createFreshCacheService>>;

  beforeEach(async () => {
    localStorageMock.clear();
    vi.clearAllMocks();
    cacheService = await createFreshCacheService();
    // Invalidate any leftover state
    cacheService.invalidateAll();
    localStorageMock.clear();
    vi.clearAllMocks();
  });

  afterEach(() => {
    vi.clearAllMocks();
  });

  describe("set and get", () => {
    it("stores and retrieves a value from memory cache", () => {
      cacheService.set("test-key", { name: "hello" });
      const result = cacheService.get("test-key");
      expect(result).toEqual({ name: "hello" });
    });

    it("returns null when key does not exist", () => {
      const result = cacheService.get("nonexistent");
      expect(result).toBeNull();
    });

    it("returns null when skipCache is true", () => {
      cacheService.set("key1", "value1");
      const result = cacheService.get("key1", { skipCache: true });
      expect(result).toBeNull();
    });

    it("stores primitive values", () => {
      cacheService.set("num", 42);
      expect(cacheService.get("num")).toBe(42);
    });

    it("stores arrays", () => {
      cacheService.set("arr", [1, 2, 3]);
      expect(cacheService.get("arr")).toEqual([1, 2, 3]);
    });

    it("writes to localStorage with api_cache_ prefix", () => {
      cacheService.set("my-key", "my-value");
      expect(localStorageMock.setItem).toHaveBeenCalledWith(
        "api_cache_my-key",
        expect.stringContaining("my-value"),
      );
    });
  });

  describe("TTL expiry", () => {
    it("returns null for expired entries", () => {
      vi.useFakeTimers();
      cacheService.set("expiring", "data", { ttl: 1000 });
      vi.advanceTimersByTime(2000);
      const result = cacheService.get("expiring", { ttl: 1000 });
      expect(result).toBeNull();
      vi.useRealTimers();
    });

    it("returns value for non-expired entries", () => {
      vi.useFakeTimers();
      cacheService.set("fresh", "data", { ttl: 5000 });
      vi.advanceTimersByTime(1000);
      const result = cacheService.get("fresh", { ttl: 5000 });
      expect(result).toBe("data");
      vi.useRealTimers();
    });
  });

  describe("remove", () => {
    it("removes a specific key from memory and localStorage", () => {
      cacheService.set("to-remove", "value");
      cacheService.remove("to-remove");
      expect(cacheService.get("to-remove")).toBeNull();
      expect(localStorageMock.removeItem).toHaveBeenCalledWith(
        "api_cache_to-remove",
      );
    });
  });

  describe("removeByPrefix", () => {
    it("removes all keys matching prefix from memory cache", () => {
      cacheService.set("orders_1", "order1");
      cacheService.set("orders_2", "order2");
      cacheService.set("products_1", "product1");
      cacheService.removeByPrefix("orders_");
      expect(cacheService.get("orders_1")).toBeNull();
      expect(cacheService.get("orders_2")).toBeNull();
      expect(cacheService.get("products_1")).toBe("product1");
    });
  });

  describe("invalidateAll", () => {
    it("clears all memory cache entries", () => {
      cacheService.set("key-a", "val-a");
      cacheService.set("key-b", "val-b");
      cacheService.invalidateAll();
      expect(cacheService.get("key-a")).toBeNull();
      expect(cacheService.get("key-b")).toBeNull();
    });

    it("updates cacheVersion in localStorage", () => {
      cacheService.invalidateAll();
      expect(localStorageMock.setItem).toHaveBeenCalledWith(
        "_cache_version",
        expect.any(String),
      );
    });
  });

  describe("getOrFetch", () => {
    it("returns cached value without calling fetcher", async () => {
      cacheService.set("fetch-key", "cached");
      const fetcher = vi.fn().mockResolvedValue("fetched");
      const result = await cacheService.getOrFetch("fetch-key", fetcher);
      expect(result).toBe("cached");
      expect(fetcher).not.toHaveBeenCalled();
    });

    it("calls fetcher on cache miss and stores result", async () => {
      const fetcher = vi.fn().mockResolvedValue("fresh-data");
      const result = await cacheService.getOrFetch("miss-key", fetcher);
      expect(result).toBe("fresh-data");
      expect(fetcher).toHaveBeenCalledTimes(1);
      // Should now be cached
      expect(cacheService.get("miss-key")).toBe("fresh-data");
    });

    it("calls fetcher when skipCache is true", async () => {
      cacheService.set("skip-key", "old");
      const fetcher = vi.fn().mockResolvedValue("new-data");
      const result = await cacheService.getOrFetch("skip-key", fetcher, {
        skipCache: true,
      });
      expect(result).toBe("new-data");
      expect(fetcher).toHaveBeenCalledTimes(1);
    });
  });

  describe("getStats", () => {
    it("returns memory entry count", () => {
      cacheService.invalidateAll();
      cacheService.set("stat-key-1", 1);
      cacheService.set("stat-key-2", 2);
      const stats = cacheService.getStats();
      expect(stats.memoryEntries).toBeGreaterThanOrEqual(2);
    });
  });

  describe("clearType", () => {
    it('clears all when type is "all"', () => {
      cacheService.set("inventory_abc", "x");
      cacheService.set("orders_xyz", "y");
      cacheService.clearType("all");
      expect(cacheService.get("inventory_abc")).toBeNull();
      expect(cacheService.get("orders_xyz")).toBeNull();
    });

    it('clears inventory entries when type is "inventory"', () => {
      cacheService.set("inventory_1", "inv");
      cacheService.set("orders_1", "ord");
      cacheService.clearType("inventory");
      expect(cacheService.get("inventory_1")).toBeNull();
      // orders should still be there
      expect(cacheService.get("orders_1")).toBe("ord");
    });

    it('clears product entries when type is "products"', () => {
      cacheService.set("products_1", "prod");
      cacheService.set("master-products_1", "master");
      cacheService.clearType("products");
      expect(cacheService.get("products_1")).toBeNull();
      expect(cacheService.get("master-products_1")).toBeNull();
    });
  });
});
