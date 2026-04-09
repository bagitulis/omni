import { create } from "zustand";
import { STORAGE_KEYS, API_BASE_URL } from "@/lib/constants";
import type { User } from "@/types/auth";
import { logger } from "@/lib/logger";
import apiClient from "@/api/client";

export interface AuthState {
  user: User | null;
  /** @deprecated Use accessToken instead. Alias kept for backward compatibility. */
  token: string | null;
  accessToken: string | null;
  isAuthenticated: boolean;
  tenantId: string | null;
  expiresAt: number | null;

  // Actions
  setAuth: (data: {
    token?: string;
    access_token?: string;
    user: User;
    tenant_id?: string;
    expires_in?: number;
  }) => void;
  logout: () => Promise<void>;
  initializeAuth: () => Promise<boolean>;
  getValidToken: () => Promise<string | null>;
  refreshAccessToken: () => Promise<boolean>;
  clearAuth: () => void;
}

// Track ongoing refresh to prevent concurrent refreshes
let refreshPromise: Promise<boolean> | null = null;

function clearLegacyStorage(): void {
  // Remove old keys from both storages (migration cleanup)
  localStorage.removeItem("authToken");
  localStorage.removeItem("userRole");
  localStorage.removeItem("userName");
  // Clean up sessionStorage from old implementation
  sessionStorage.removeItem("authUser");
  sessionStorage.removeItem("tenantId");
  sessionStorage.removeItem("autoLoginFailed");
}

export const useAuthStore = create<AuthState>((set, get) => ({
  user: null,
  token: null,
  accessToken: null,
  isAuthenticated: false,
  tenantId: localStorage.getItem(STORAGE_KEYS.TENANT_ID),
  expiresAt: null,

  setAuth: ({ token, access_token, user, tenant_id, expires_in }) => {
    const finalToken = access_token || token || null;
    const expiresInMs = (expires_in || 900) * 1000;
    const expiresAt = Date.now() + expiresInMs;

    clearLegacyStorage();

    // Persist user metadata in localStorage (survives tab close)
    // Access token stays in-memory only (security best practice)
    localStorage.setItem(STORAGE_KEYS.AUTH_USER, JSON.stringify(user));

    if (tenant_id) {
      localStorage.setItem(STORAGE_KEYS.TENANT_ID, tenant_id);
    }

    set({
      token: finalToken,
      accessToken: finalToken,
      user,
      tenantId: tenant_id || get().tenantId,
      isAuthenticated: true,
      expiresAt,
    });
  },

  clearAuth: () => {
    clearLegacyStorage();
    localStorage.removeItem(STORAGE_KEYS.AUTH_USER);
    localStorage.removeItem(STORAGE_KEYS.TENANT_ID);

    set({
      user: null,
      token: null,
      accessToken: null,
      tenantId: null,
      isAuthenticated: false,
      expiresAt: null,
    });
  },

  refreshAccessToken: async (): Promise<boolean> => {
    // Prevent concurrent refresh attempts
    if (refreshPromise) {
      return refreshPromise;
    }

    refreshPromise = (async () => {
      try {
        const tenantId = get().tenantId;
        const response = await fetch(`${API_BASE_URL}/auth/refresh`, {
          method: "POST",
          credentials: "include", // IMPORTANT: Send HttpOnly cookies!
          headers: {
            "Content-Type": "application/json",
            ...(tenantId ? { "x-tenant-id": tenantId } : {}),
          },
        });

        if (response.ok) {
          const data = await response.json();
          // Backend wraps response as {success, data: {access_token, expires_in}}
          // Handle both wrapped and flat formats for robustness
          const payload = data.data || data;
          const accessToken = payload.access_token;
          const expiresIn = payload.expires_in;
          if (data.success && accessToken) {
            set(() => ({
              accessToken,
              token: accessToken,
              expiresAt: Date.now() + (expiresIn || 900) * 1000,
              isAuthenticated: true,
            }));
            return true;
          }
        }

        // Refresh failed - could be expired or revoked
        if (response.status === 401) {
          const data = await response.json().catch(() => ({}));
          if (data.code === "TOKEN_REUSE") {
            logger.error("Security alert: token reuse detected");
          }
        }
      } catch (error) {
        logger.error("Token refresh failed", { error });
      }

      // Clear auth on failure
      get().clearAuth();
      return false;
    })();

    try {
      return await refreshPromise;
    } finally {
      refreshPromise = null;
    }
  },

  getValidToken: async (): Promise<string | null> => {
    const { accessToken, expiresAt, user, refreshAccessToken } = get();

    // If we have a valid token, return it
    if (accessToken && expiresAt) {
      // Add 10 second buffer for network latency
      if (Date.now() < expiresAt - 10000) {
        return accessToken;
      }
    }

    // If we have user info (session), try to refresh
    if (user || localStorage.getItem(STORAGE_KEYS.AUTH_USER)) {
      const refreshed = await refreshAccessToken();
      if (refreshed) {
        return get().accessToken;
      }
    }

    return null;
  },

  initializeAuth: async (): Promise<boolean> => {
    const { accessToken, refreshAccessToken, clearAuth } = get();

    // Check for stored user in sessionStorage
    const storedUser = localStorage.getItem(STORAGE_KEYS.AUTH_USER);
    const storedTenantId = localStorage.getItem(STORAGE_KEYS.TENANT_ID);

    if (storedUser) {
      try {
        const user = JSON.parse(storedUser);
        // Restore user state immediately (optimistic)
        set({ user, tenantId: storedTenantId || null });

        // If no token, try to refresh
        if (!accessToken) {
          const success = await refreshAccessToken();
          if (!success) {
            // Refresh failed, clear everything
            clearAuth();
            return false;
          }
          return true;
        }
      } catch (err) {
        logger.warn("Failed to parse stored auth user", { error: err });
        localStorage.removeItem(STORAGE_KEYS.AUTH_USER);
      }
    }

    return !!get().accessToken;
  },

  logout: async (): Promise<void> => {
    const { accessToken, tenantId, clearAuth } = get();
    try {
      // Use apiClient for consistent auth header and error handling
      await apiClient.post("/auth/logout", undefined, {
        headers: {
          ...(accessToken ? { Authorization: `Bearer ${accessToken}` } : {}),
          ...(tenantId ? { "x-tenant-id": tenantId } : {}),
        },
      });
    } catch (error) {
      logger.error("Logout request failed", { error });
    } finally {
      clearAuth();
    }
  },
}));
