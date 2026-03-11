import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
import { useAuthStore } from "./authStore";
import type { User } from "@/types/auth";
import apiClient from "@/api/client";
import { logger } from "@/lib/logger";

vi.mock("@/lib/constants", () => ({
  STORAGE_KEYS: {
    AUTH_USER: "authUser",
    TENANT_ID: "tenantId",
  },
  API_BASE_URL: "http://localhost/api",
}));

vi.mock("@/api/client", () => ({
  default: {
    post: vi.fn(),
  },
}));

vi.mock("@/lib/logger", () => ({
  logger: {
    debug: vi.fn(),
    info: vi.fn(),
    warn: vi.fn(),
    error: vi.fn(),
  },
}));

const mockUser: User = {
  id: "user-001",
  username: "testuser",
  email: "test@example.com",
  role: "admin",
};

const mockApiPost = vi.mocked(apiClient.post);
const mockLogger = vi.mocked(logger);

function resetStore(): void {
  useAuthStore.setState({
    user: null,
    token: null,
    accessToken: null,
    isAuthenticated: false,
    tenantId: null,
    expiresAt: null,
  });
}

describe("authStore", () => {
  beforeEach(() => {
    vi.clearAllMocks();
    sessionStorage.clear();
    localStorage.clear();
    resetStore();
  });

  afterEach(() => {
    vi.unstubAllGlobals();
  });

  it("starts unauthenticated", () => {
    expect(useAuthStore.getState()).toMatchObject({
      user: null,
      token: null,
      accessToken: null,
      isAuthenticated: false,
      tenantId: null,
      expiresAt: null,
    });
  });

  it("setAuth stores user, tenant, token, and clears legacy localStorage", () => {
    localStorage.setItem("authToken", "old-token");
    localStorage.setItem("authUser", "old-user");
    localStorage.setItem("tenantId", "old-tenant");
    localStorage.setItem("userRole", "old-role");
    localStorage.setItem("userName", "old-name");

    useAuthStore.getState().setAuth({
      token: "legacy-token",
      access_token: "fresh-token",
      user: mockUser,
      tenant_id: "tenant-xyz",
      expires_in: 3600,
    });

    expect(useAuthStore.getState()).toMatchObject({
      user: mockUser,
      token: "fresh-token",
      accessToken: "fresh-token",
      isAuthenticated: true,
      tenantId: "tenant-xyz",
    });
    expect(sessionStorage.getItem("authUser")).toBe(JSON.stringify(mockUser));
    expect(sessionStorage.getItem("tenantId")).toBe("tenant-xyz");
    expect(localStorage.getItem("authToken")).toBeNull();
    expect(localStorage.getItem("authUser")).toBeNull();
    expect(localStorage.getItem("tenantId")).toBeNull();
    expect(localStorage.getItem("userRole")).toBeNull();
    expect(localStorage.getItem("userName")).toBeNull();
    expect(useAuthStore.getState().expiresAt).not.toBeNull();
  });

  it("clearAuth resets auth state and clears storage", () => {
    sessionStorage.setItem("authUser", JSON.stringify(mockUser));
    sessionStorage.setItem("tenantId", "tenant-xyz");
    localStorage.setItem("authToken", "old-token");

    useAuthStore.getState().setAuth({
      token: "token-1",
      user: mockUser,
      tenant_id: "tenant-xyz",
    });

    useAuthStore.getState().clearAuth();

    expect(useAuthStore.getState()).toMatchObject({
      user: null,
      token: null,
      accessToken: null,
      isAuthenticated: false,
      tenantId: null,
      expiresAt: null,
    });
    expect(sessionStorage.getItem("authUser")).toBeNull();
    expect(sessionStorage.getItem("tenantId")).toBeNull();
    expect(localStorage.getItem("authToken")).toBeNull();
  });

  it("refreshAccessToken updates token on success", async () => {
    vi.stubGlobal(
      "fetch",
      vi.fn().mockResolvedValue({
        ok: true,
        status: 200,
        json: async () => ({
          success: true,
          access_token: "new-token",
          expires_in: 900,
        }),
      }),
    );

    const result = await useAuthStore.getState().refreshAccessToken();

    expect(result).toBe(true);
    expect(useAuthStore.getState()).toMatchObject({
      token: "new-token",
      accessToken: "new-token",
      isAuthenticated: true,
    });
  });

  it("refreshAccessToken clears auth and logs on network error", async () => {
    vi.stubGlobal(
      "fetch",
      vi.fn().mockRejectedValue(new Error("Network down")),
    );
    useAuthStore.getState().setAuth({ token: "tok", user: mockUser });

    const result = await useAuthStore.getState().refreshAccessToken();

    expect(result).toBe(false);
    expect(useAuthStore.getState().isAuthenticated).toBe(false);
    expect(mockLogger.error).toHaveBeenCalledWith("Token refresh failed", {
      error: expect.any(Error),
    });
  });

  it("getValidToken returns current token when not expired", async () => {
    useAuthStore.setState({
      token: "valid-token",
      accessToken: "valid-token",
      expiresAt: Date.now() + 60_000,
      isAuthenticated: true,
      user: mockUser,
      tenantId: null,
    });

    await expect(useAuthStore.getState().getValidToken()).resolves.toBe(
      "valid-token",
    );
  });

  it("initializeAuth restores stored user and refreshes token", async () => {
    sessionStorage.setItem("authUser", JSON.stringify(mockUser));
    sessionStorage.setItem("tenantId", "tenant-xyz");
    vi.stubGlobal(
      "fetch",
      vi.fn().mockResolvedValue({
        ok: true,
        status: 200,
        json: async () => ({
          success: true,
          access_token: "refreshed-token",
          expires_in: 900,
        }),
      }),
    );

    const result = await useAuthStore.getState().initializeAuth();

    expect(result).toBe(true);
    expect(useAuthStore.getState()).toMatchObject({
      user: mockUser,
      tenantId: "tenant-xyz",
      accessToken: "refreshed-token",
      token: "refreshed-token",
      isAuthenticated: true,
    });
  });

  it("initializeAuth removes corrupted stored user", async () => {
    sessionStorage.setItem("authUser", "NOT VALID JSON{{{");

    const result = await useAuthStore.getState().initializeAuth();

    expect(result).toBe(false);
    expect(sessionStorage.getItem("authUser")).toBeNull();
    expect(mockLogger.warn).toHaveBeenCalledWith(
      "Failed to parse stored auth user",
      { error: expect.any(Error) },
    );
  });

  it("logout posts to auth endpoint and clears auth", async () => {
    mockApiPost.mockResolvedValue({ success: true });
    useAuthStore.getState().setAuth({
      token: "token-1",
      user: mockUser,
      tenant_id: "tenant-xyz",
    });

    await useAuthStore.getState().logout();

    expect(mockApiPost).toHaveBeenCalledWith("/auth/logout", undefined, {
      headers: {
        Authorization: "Bearer token-1",
        "x-tenant-id": "tenant-xyz",
      },
    });
    expect(useAuthStore.getState().isAuthenticated).toBe(false);
  });

  it("logout clears auth even when api client fails", async () => {
    mockApiPost.mockRejectedValue(new Error("Network error"));
    useAuthStore.getState().setAuth({ token: "token-1", user: mockUser });

    await useAuthStore.getState().logout();

    expect(useAuthStore.getState().isAuthenticated).toBe(false);
    expect(mockLogger.error).toHaveBeenCalledWith("Logout request failed", {
      error: expect.any(Error),
    });
  });
});
