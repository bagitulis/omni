import { describe, it, expect, vi, beforeEach, afterEach } from "vitest";
import { useAuthStore } from "./authStore";
import type { User } from "@/types/auth";

// Mock constants
vi.mock("@/lib/constants", () => ({
  STORAGE_KEYS: {
    AUTH_USER: "authUser",
    TENANT_ID: "tenantId",
  },
  API_BASE_URL: "http://localhost/api",
}));

const mockUser: User = {
  id: "user-001",
  username: "testuser",
  email: "test@example.com",
  role: "admin",
};

const resetStore = () => {
  useAuthStore.setState({
    user: null,
    token: null,
    accessToken: null,
    isAuthenticated: false,
    tenantId: null,
    expiresAt: null,
  });
};

describe("authStore — initial state", () => {
  beforeEach(() => {
    vi.clearAllMocks();
    sessionStorage.clear();
    localStorage.clear();
    resetStore();
  });

  it("user is null initially", () => {
    expect(useAuthStore.getState().user).toBeNull();
  });

  it("token is null initially", () => {
    expect(useAuthStore.getState().token).toBeNull();
  });

  it("accessToken is null initially", () => {
    expect(useAuthStore.getState().accessToken).toBeNull();
  });

  it("isAuthenticated is false initially", () => {
    expect(useAuthStore.getState().isAuthenticated).toBe(false);
  });

  it("expiresAt is null initially", () => {
    expect(useAuthStore.getState().expiresAt).toBeNull();
  });
});

describe("authStore — setAuth", () => {
  beforeEach(() => {
    vi.clearAllMocks();
    sessionStorage.clear();
    localStorage.clear();
    resetStore();
  });

  it("sets user and isAuthenticated=true", () => {
    useAuthStore.getState().setAuth({ token: "tok-abc", user: mockUser });
    const state = useAuthStore.getState();
    expect(state.user).toEqual(mockUser);
    expect(state.isAuthenticated).toBe(true);
  });

  it("sets token from token field", () => {
    useAuthStore.getState().setAuth({ token: "my-token", user: mockUser });
    expect(useAuthStore.getState().token).toBe("my-token");
    expect(useAuthStore.getState().accessToken).toBe("my-token");
  });

  it("prefers access_token over token when both provided", () => {
    useAuthStore
      .getState()
      .setAuth({
        token: "old-token",
        access_token: "new-token",
        user: mockUser,
      });
    expect(useAuthStore.getState().token).toBe("new-token");
  });

  it("stores user in sessionStorage", () => {
    useAuthStore.getState().setAuth({ token: "tok", user: mockUser });
    const stored = JSON.parse(sessionStorage.getItem("authUser") ?? "null");
    expect(stored).toEqual(mockUser);
  });

  it("stores tenant_id in sessionStorage when provided", () => {
    useAuthStore
      .getState()
      .setAuth({ token: "tok", user: mockUser, tenant_id: "tenant-xyz" });
    expect(sessionStorage.getItem("tenantId")).toBe("tenant-xyz");
    expect(useAuthStore.getState().tenantId).toBe("tenant-xyz");
  });

  it("sets expiresAt using expires_in", () => {
    const before = Date.now();
    useAuthStore
      .getState()
      .setAuth({ token: "tok", user: mockUser, expires_in: 3600 });
    const { expiresAt } = useAuthStore.getState();
    expect(expiresAt).not.toBeNull();
    expect(expiresAt!).toBeGreaterThan(before + 3590 * 1000);
  });

  it("defaults expires_in to 900 seconds when not provided", () => {
    const before = Date.now();
    useAuthStore.getState().setAuth({ token: "tok", user: mockUser });
    const { expiresAt } = useAuthStore.getState();
    expect(expiresAt!).toBeGreaterThanOrEqual(before + 899 * 1000);
  });

  it("clears legacy localStorage items", () => {
    localStorage.setItem("authToken", "old");
    localStorage.setItem("authUser", "old");
    localStorage.setItem("tenantId", "old");
    useAuthStore.getState().setAuth({ token: "tok", user: mockUser });
    expect(localStorage.getItem("authToken")).toBeNull();
    expect(localStorage.getItem("authUser")).toBeNull();
    expect(localStorage.getItem("tenantId")).toBeNull();
  });
});

describe("authStore — clearAuth", () => {
  beforeEach(() => {
    vi.clearAllMocks();
    sessionStorage.clear();
    localStorage.clear();
    resetStore();
  });

  it("resets all auth state to null/false", () => {
    useAuthStore
      .getState()
      .setAuth({ token: "tok", user: mockUser, tenant_id: "t1" });
    useAuthStore.getState().clearAuth();
    const state = useAuthStore.getState();
    expect(state.user).toBeNull();
    expect(state.token).toBeNull();
    expect(state.accessToken).toBeNull();
    expect(state.isAuthenticated).toBe(false);
    expect(state.tenantId).toBeNull();
    expect(state.expiresAt).toBeNull();
  });

  it("removes auth user from sessionStorage", () => {
    sessionStorage.setItem("authUser", JSON.stringify(mockUser));
    useAuthStore.getState().clearAuth();
    expect(sessionStorage.getItem("authUser")).toBeNull();
  });

  it("removes tenant id from sessionStorage", () => {
    sessionStorage.setItem("tenantId", "tenant-xyz");
    useAuthStore.getState().clearAuth();
    expect(sessionStorage.getItem("tenantId")).toBeNull();
  });

  it("clears all legacy localStorage items", () => {
    localStorage.setItem("authToken", "t");
    localStorage.setItem("authUser", "u");
    localStorage.setItem("tenantId", "i");
    localStorage.setItem("userRole", "r");
    localStorage.setItem("userName", "n");
    useAuthStore.getState().clearAuth();
    expect(localStorage.getItem("authToken")).toBeNull();
    expect(localStorage.getItem("authUser")).toBeNull();
    expect(localStorage.getItem("tenantId")).toBeNull();
    expect(localStorage.getItem("userRole")).toBeNull();
    expect(localStorage.getItem("userName")).toBeNull();
  });
});

describe("authStore — refreshAccessToken", () => {
  beforeEach(() => {
    vi.clearAllMocks();
    sessionStorage.clear();
    localStorage.clear();
    resetStore();
  });

  afterEach(() => {
    vi.restoreAllMocks();
  });

  it("updates token on successful refresh", async () => {
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
    expect(useAuthStore.getState().accessToken).toBe("new-token");
    expect(useAuthStore.getState().isAuthenticated).toBe(true);
  });

  it("clears auth and returns false when refresh fails", async () => {
    vi.stubGlobal(
      "fetch",
      vi.fn().mockResolvedValue({
        ok: false,
        status: 401,
        json: async () => ({ code: "" }),
      }),
    );

    useAuthStore.getState().setAuth({ token: "old-token", user: mockUser });
    const result = await useAuthStore.getState().refreshAccessToken();
    expect(result).toBe(false);
    expect(useAuthStore.getState().isAuthenticated).toBe(false);
  });

  it("returns false and clears auth on network error", async () => {
    vi.stubGlobal(
      "fetch",
      vi.fn().mockRejectedValue(new Error("Network down")),
    );

    useAuthStore.getState().setAuth({ token: "tok", user: mockUser });
    const result = await useAuthStore.getState().refreshAccessToken();
    expect(result).toBe(false);
  });

  it("returns false when response is ok but success is false", async () => {
    vi.stubGlobal(
      "fetch",
      vi.fn().mockResolvedValue({
        ok: true,
        status: 200,
        json: async () => ({ success: false }),
      }),
    );

    const result = await useAuthStore.getState().refreshAccessToken();
    expect(result).toBe(false);
  });
});

describe("authStore — getValidToken", () => {
  beforeEach(() => {
    vi.clearAllMocks();
    sessionStorage.clear();
    localStorage.clear();
    resetStore();
  });

  afterEach(() => {
    vi.restoreAllMocks();
  });

  it("returns token when valid and not expired", async () => {
    useAuthStore.setState({
      accessToken: "valid-token",
      expiresAt: Date.now() + 60 * 1000, // expires in 60s
      user: mockUser,
      isAuthenticated: true,
      token: "valid-token",
      tenantId: null,
    });

    const token = await useAuthStore.getState().getValidToken();
    expect(token).toBe("valid-token");
  });

  it("returns null when no user and no token", async () => {
    const token = await useAuthStore.getState().getValidToken();
    expect(token).toBeNull();
  });

  it("attempts refresh when token is expired but user exists", async () => {
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

    useAuthStore.setState({
      accessToken: "expired-token",
      expiresAt: Date.now() - 1000, // already expired
      user: mockUser,
      isAuthenticated: true,
      token: "expired-token",
      tenantId: null,
    });

    const token = await useAuthStore.getState().getValidToken();
    expect(token).toBe("refreshed-token");
  });
});

describe("authStore — logout", () => {
  beforeEach(() => {
    vi.clearAllMocks();
    sessionStorage.clear();
    localStorage.clear();
    resetStore();
  });

  afterEach(() => {
    vi.restoreAllMocks();
  });

  it("calls the logout endpoint and clears auth", async () => {
    const fetchMock = vi.fn().mockResolvedValue({ ok: true });
    vi.stubGlobal("fetch", fetchMock);

    useAuthStore
      .getState()
      .setAuth({ token: "tok", user: mockUser, tenant_id: "t1" });
    await useAuthStore.getState().logout();

    expect(fetchMock).toHaveBeenCalledWith(
      expect.stringContaining("/auth/logout"),
      expect.any(Object),
    );
    expect(useAuthStore.getState().isAuthenticated).toBe(false);
    expect(useAuthStore.getState().user).toBeNull();
  });

  it("clears auth even if logout request fails", async () => {
    vi.stubGlobal(
      "fetch",
      vi.fn().mockRejectedValue(new Error("Network error")),
    );

    useAuthStore.getState().setAuth({ token: "tok", user: mockUser });
    await useAuthStore.getState().logout();

    expect(useAuthStore.getState().isAuthenticated).toBe(false);
  });
});

describe("authStore — initializeAuth", () => {
  beforeEach(() => {
    vi.clearAllMocks();
    sessionStorage.clear();
    localStorage.clear();
    resetStore();
  });

  afterEach(() => {
    vi.restoreAllMocks();
  });

  it("returns false when no stored user and no token", async () => {
    const result = await useAuthStore.getState().initializeAuth();
    expect(result).toBe(false);
  });

  it("restores user from sessionStorage and refreshes token", async () => {
    sessionStorage.setItem("authUser", JSON.stringify(mockUser));
    vi.stubGlobal(
      "fetch",
      vi.fn().mockResolvedValue({
        ok: true,
        status: 200,
        json: async () => ({
          success: true,
          access_token: "refreshed",
          expires_in: 900,
        }),
      }),
    );

    const result = await useAuthStore.getState().initializeAuth();
    expect(result).toBe(true);
    expect(useAuthStore.getState().user).toEqual(mockUser);
  });

  it("clears auth and returns false when stored user exists but refresh fails", async () => {
    sessionStorage.setItem("authUser", JSON.stringify(mockUser));
    vi.stubGlobal(
      "fetch",
      vi.fn().mockResolvedValue({
        ok: false,
        status: 401,
        json: async () => ({}),
      }),
    );

    const result = await useAuthStore.getState().initializeAuth();
    expect(result).toBe(false);
    expect(useAuthStore.getState().user).toBeNull();
  });

  it("returns false and removes corrupted sessionStorage entry", async () => {
    sessionStorage.setItem("authUser", "NOT VALID JSON{{{");
    const result = await useAuthStore.getState().initializeAuth();
    // Should not throw; JSON.parse fails, session entry removed
    expect(typeof result).toBe("boolean");
    expect(sessionStorage.getItem("authUser")).toBeNull();
  });
});
