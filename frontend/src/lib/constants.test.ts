import { describe, it, expect, vi, beforeEach, afterEach } from "vitest";

// We need to control import.meta.env.DEV and window.location before importing constants.
// Use vi.hoisted to set up globals, then dynamic import inside each describe to get fresh module.

describe("API_TIMEOUT", () => {
  it("has correct HEALTH timeout", async () => {
    const { API_TIMEOUT } = await import("./constants");
    expect(API_TIMEOUT.HEALTH).toBe(10000);
  });

  it("has correct SHORT timeout", async () => {
    const { API_TIMEOUT } = await import("./constants");
    expect(API_TIMEOUT.SHORT).toBe(15000);
  });

  it("has correct DEFAULT timeout", async () => {
    const { API_TIMEOUT } = await import("./constants");
    expect(API_TIMEOUT.DEFAULT).toBe(30000);
  });

  it("has correct LONG timeout", async () => {
    const { API_TIMEOUT } = await import("./constants");
    expect(API_TIMEOUT.LONG).toBe(60000);
  });

  it("has correct EXTRA_LONG timeout", async () => {
    const { API_TIMEOUT } = await import("./constants");
    expect(API_TIMEOUT.EXTRA_LONG).toBe(180000);
  });
});

describe("STORAGE_KEYS", () => {
  it("has AUTH_USER key", async () => {
    const { STORAGE_KEYS } = await import("./constants");
    expect(STORAGE_KEYS.AUTH_USER).toBe("authUser");
  });

  it("has TENANT_ID key", async () => {
    const { STORAGE_KEYS } = await import("./constants");
    expect(STORAGE_KEYS.TENANT_ID).toBe("tenantId");
  });

  it("has USER_ROLE key", async () => {
    const { STORAGE_KEYS } = await import("./constants");
    expect(STORAGE_KEYS.USER_ROLE).toBe("userRole");
  });

  it("has USER_NAME key", async () => {
    const { STORAGE_KEYS } = await import("./constants");
    expect(STORAGE_KEYS.USER_NAME).toBe("userName");
  });

  it("has exactly 4 keys", async () => {
    const { STORAGE_KEYS } = await import("./constants");
    expect(Object.keys(STORAGE_KEYS)).toHaveLength(4);
  });
});

describe("COOKIE_NAMES", () => {
  it("is an empty object", async () => {
    const { COOKIE_NAMES } = await import("./constants");
    expect(COOKIE_NAMES).toEqual({});
  });
});

describe("getBackendUrl", () => {
  const originalLocation = window.location;

  beforeEach(() => {
    vi.resetModules();
  });

  afterEach(() => {
    // Restore window.location
    Object.defineProperty(window, "location", {
      value: originalLocation,
      writable: true,
      configurable: true,
    });
  });

  it("returns /api in production (DEV=false, no VITE_API_URL)", async () => {
    vi.stubEnv("DEV", false);
    vi.stubEnv("VITE_API_URL", "");
    const { getBackendUrl } = await import("./constants");
    expect(getBackendUrl()).toBe("/api");
    vi.unstubAllEnvs();
  });

  it("returns VITE_API_URL in production when set", async () => {
    vi.stubEnv("DEV", false);
    vi.stubEnv("VITE_API_URL", "https://api.example.com/api");
    const { getBackendUrl } = await import("./constants");
    expect(getBackendUrl()).toBe("https://api.example.com/api");
    vi.unstubAllEnvs();
  });

  it("returns /api for non-localhost in DEV mode", async () => {
    vi.stubEnv("DEV", true);
    Object.defineProperty(window, "location", {
      value: {
        origin: "https://staging.example.com",
        hostname: "staging.example.com",
        href: "https://staging.example.com/",
      },
      writable: true,
      configurable: true,
    });
    const { getBackendUrl } = await import("./constants");
    expect(getBackendUrl()).toBe("/api");
    vi.unstubAllEnvs();
  });

  it("returns /api for localhost in DEV mode", async () => {
    vi.stubEnv("DEV", true);
    Object.defineProperty(window, "location", {
      value: {
        origin: "http://localhost:5173",
        hostname: "localhost",
        href: "http://localhost:5173/",
      },
      writable: true,
      configurable: true,
    });
    const { getBackendUrl } = await import("./constants");
    expect(getBackendUrl()).toBe("/api");
    vi.unstubAllEnvs();
  });

  it("returns /api for 127.0.0.1 in DEV mode", async () => {
    vi.stubEnv("DEV", true);
    Object.defineProperty(window, "location", {
      value: {
        origin: "http://127.0.0.1:5173",
        hostname: "127.0.0.1",
        href: "http://127.0.0.1:5173/",
      },
      writable: true,
      configurable: true,
    });
    const { getBackendUrl } = await import("./constants");
    expect(getBackendUrl()).toBe("/api");
    vi.unstubAllEnvs();
  });
});
