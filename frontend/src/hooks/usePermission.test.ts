import { beforeEach, describe, expect, it, vi } from "vitest";
import { renderHook } from "@testing-library/react";
import type { AuthState } from "@/stores/authStore";

// Mock authStore before importing hooks
vi.mock("@/stores/authStore", () => ({
  useAuthStore: vi.fn(),
}));

vi.mock("@/lib/permissionService", () => ({
  hasPermission: vi.fn(),
  hasRole: vi.fn(),
  canAccessAdmin: vi.fn(),
  canManageUsers: vi.fn(),
  canSwitchTenant: vi.fn(),
  getAllPermissions: vi.fn(),
  getRoleDisplayName: vi.fn(),
}));

import { usePermission, useHasPermission, useHasRole } from "./usePermission";
import { useAuthStore } from "@/stores/authStore";
import * as permissionService from "@/lib/permissionService";

const mockUseAuthStore = vi.mocked(useAuthStore);

/** Build a minimal AuthState-compatible mock object for selector calls */
function makeState(user: AuthState["user"]): AuthState {
  return {
    user,
    token: null,
    accessToken: null,
    isAuthenticated: false,
    isInitializing: false,
    tenantId: null,
    expiresAt: null,
    setAuth: vi.fn(),
    logout: vi.fn(),
    initializeAuth: vi.fn(),
    getValidToken: vi.fn(),
    refreshAccessToken: vi.fn(),
    clearAuth: vi.fn(),
  };
}

describe("usePermission", () => {
  beforeEach(() => {
    vi.clearAllMocks();
    // Default: no user (guest)
    mockUseAuthStore.mockImplementation(
      (selector: (state: AuthState) => unknown) => selector(makeState(null)),
    );
    vi.mocked(permissionService.hasPermission).mockReturnValue(false);
    vi.mocked(permissionService.hasRole).mockReturnValue(false);
    vi.mocked(permissionService.canAccessAdmin).mockReturnValue(false);
    vi.mocked(permissionService.canManageUsers).mockReturnValue(false);
    vi.mocked(permissionService.canSwitchTenant).mockReturnValue(false);
    vi.mocked(permissionService.getAllPermissions).mockReturnValue([]);
    vi.mocked(permissionService.getRoleDisplayName).mockReturnValue("Guest");
  });

  it("returns userRole as empty string when no user", () => {
    const { result } = renderHook(() => usePermission());
    expect(result.current.userRole).toBe("");
  });

  it("returns userRole from authenticated user", () => {
    mockUseAuthStore.mockImplementation(
      (selector: (state: AuthState) => unknown) =>
        selector(makeState({ role: "admin" } as AuthState["user"])),
    );
    const { result } = renderHook(() => usePermission());
    expect(result.current.userRole).toBe("admin");
  });

  it("calls hasPermission with userRole and given permission", () => {
    mockUseAuthStore.mockImplementation(
      (selector: (state: AuthState) => unknown) =>
        selector(makeState({ role: "owner" } as AuthState["user"])),
    );
    vi.mocked(permissionService.hasPermission).mockReturnValue(true);

    const { result } = renderHook(() => usePermission());
    const perm = result.current.hasPermission("read:orders");

    expect(permissionService.hasPermission).toHaveBeenCalledWith(
      "owner",
      "read:orders",
    );
    expect(perm).toBe(true);
  });

  it("calls hasRole with userRole and given role", () => {
    mockUseAuthStore.mockImplementation(
      (selector: (state: AuthState) => unknown) =>
        selector(makeState({ role: "admin" } as AuthState["user"])),
    );
    vi.mocked(permissionService.hasRole).mockReturnValue(true);

    const { result } = renderHook(() => usePermission());
    const check = result.current.hasRole("admin");

    expect(permissionService.hasRole).toHaveBeenCalledWith("admin", "admin");
    expect(check).toBe(true);
  });

  it("sets isDeveloper based on role", () => {
    mockUseAuthStore.mockImplementation(
      (selector: (state: AuthState) => unknown) =>
        selector(makeState({ role: "developer" } as AuthState["user"])),
    );
    const { result } = renderHook(() => usePermission());
    expect(result.current.isDeveloper).toBe(true);
    expect(result.current.isOwner).toBe(false);
    expect(result.current.isAdmin).toBe(false);
  });

  it("sets isOwner based on role", () => {
    mockUseAuthStore.mockImplementation(
      (selector: (state: AuthState) => unknown) =>
        selector(makeState({ role: "owner" } as AuthState["user"])),
    );
    const { result } = renderHook(() => usePermission());
    expect(result.current.isOwner).toBe(true);
    expect(result.current.isDeveloper).toBe(false);
  });

  it("sets isAdmin based on role", () => {
    mockUseAuthStore.mockImplementation(
      (selector: (state: AuthState) => unknown) =>
        selector(makeState({ role: "admin" } as AuthState["user"])),
    );
    const { result } = renderHook(() => usePermission());
    expect(result.current.isAdmin).toBe(true);
  });

  it("calls canAccessAdmin with userRole", () => {
    mockUseAuthStore.mockImplementation(
      (selector: (state: AuthState) => unknown) =>
        selector(makeState({ role: "admin" } as AuthState["user"])),
    );
    vi.mocked(permissionService.canAccessAdmin).mockReturnValue(true);

    const { result } = renderHook(() => usePermission());
    expect(result.current.canAccessAdmin).toBe(true);
    expect(permissionService.canAccessAdmin).toHaveBeenCalledWith("admin");
  });

  it("calls canManageUsers with userRole", () => {
    mockUseAuthStore.mockImplementation(
      (selector: (state: AuthState) => unknown) =>
        selector(makeState({ role: "owner" } as AuthState["user"])),
    );
    vi.mocked(permissionService.canManageUsers).mockReturnValue(true);

    const { result } = renderHook(() => usePermission());
    expect(result.current.canManageUsers).toBe(true);
    expect(permissionService.canManageUsers).toHaveBeenCalledWith("owner");
  });

  it("calls canSwitchTenant with userRole", () => {
    mockUseAuthStore.mockImplementation(
      (selector: (state: AuthState) => unknown) =>
        selector(makeState({ role: "developer" } as AuthState["user"])),
    );
    vi.mocked(permissionService.canSwitchTenant).mockReturnValue(true);

    const { result } = renderHook(() => usePermission());
    expect(result.current.canSwitchTenant).toBe(true);
    expect(permissionService.canSwitchTenant).toHaveBeenCalledWith("developer");
  });

  it("calls getAllPermissions with userRole", () => {
    mockUseAuthStore.mockImplementation(
      (selector: (state: AuthState) => unknown) =>
        selector(makeState({ role: "admin" } as AuthState["user"])),
    );
    vi.mocked(permissionService.getAllPermissions).mockReturnValue([
      "read:orders",
      "write:orders",
    ]);

    const { result } = renderHook(() => usePermission());
    expect(result.current.permissions).toEqual(["read:orders", "write:orders"]);
    expect(permissionService.getAllPermissions).toHaveBeenCalledWith("admin");
  });

  it("calls getRoleDisplayName with userRole", () => {
    mockUseAuthStore.mockImplementation(
      (selector: (state: AuthState) => unknown) =>
        selector(makeState({ role: "admin" } as AuthState["user"])),
    );
    vi.mocked(permissionService.getRoleDisplayName).mockReturnValue("Admin");

    const { result } = renderHook(() => usePermission());
    expect(result.current.roleDisplayName).toBe("Admin");
    expect(permissionService.getRoleDisplayName).toHaveBeenCalledWith("admin");
  });
});

describe("useHasPermission", () => {
  beforeEach(() => {
    vi.clearAllMocks();
    mockUseAuthStore.mockImplementation(
      (selector: (state: AuthState) => unknown) =>
        selector(makeState({ role: "admin" } as AuthState["user"])),
    );
  });

  it("returns true when user has the permission", () => {
    vi.mocked(permissionService.hasPermission).mockReturnValue(true);
    const { result } = renderHook(() => useHasPermission("read:orders"));
    expect(result.current).toBe(true);
    expect(permissionService.hasPermission).toHaveBeenCalledWith(
      "admin",
      "read:orders",
    );
  });

  it("returns false when user lacks the permission", () => {
    vi.mocked(permissionService.hasPermission).mockReturnValue(false);
    const { result } = renderHook(() => useHasPermission("delete:users"));
    expect(result.current).toBe(false);
  });

  it("uses empty role when no user is authenticated", () => {
    mockUseAuthStore.mockImplementation(
      (selector: (state: AuthState) => unknown) => selector(makeState(null)),
    );
    vi.mocked(permissionService.hasPermission).mockReturnValue(false);
    const { result } = renderHook(() => useHasPermission("read:orders"));
    expect(result.current).toBe(false);
    expect(permissionService.hasPermission).toHaveBeenCalledWith(
      "",
      "read:orders",
    );
  });
});

describe("useHasRole", () => {
  beforeEach(() => {
    vi.clearAllMocks();
    mockUseAuthStore.mockImplementation(
      (selector: (state: AuthState) => unknown) =>
        selector(makeState({ role: "owner" } as AuthState["user"])),
    );
  });

  it("returns true when user has the role", () => {
    vi.mocked(permissionService.hasRole).mockReturnValue(true);
    const { result } = renderHook(() => useHasRole("owner"));
    expect(result.current).toBe(true);
    expect(permissionService.hasRole).toHaveBeenCalledWith("owner", "owner");
  });

  it("returns false when user lacks the role", () => {
    vi.mocked(permissionService.hasRole).mockReturnValue(false);
    const { result } = renderHook(() => useHasRole("developer"));
    expect(result.current).toBe(false);
  });
});
