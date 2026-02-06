import { defineStore } from "pinia";
import { ref, computed } from "vue";

interface User {
  id: string;
  username: string;
  email: string;
  role: string;
}

// API base URL for auth endpoints
const API_BASE = import.meta.env.VITE_API_URL || "/api";

export const useAuthStore = defineStore("auth", () => {
  // SECURITY: Access token stored in MEMORY only (not localStorage!)
  // This prevents XSS attacks from stealing the token
  const accessToken = ref<string | null>(null);
  const expiresAt = ref<number | null>(null);

  // User info can be in sessionStorage (less sensitive, cleared on tab close)
  const user = ref<User | null>(null);
  const storedUser = sessionStorage.getItem("authUser");
  if (storedUser) {
    try {
      user.value = JSON.parse(storedUser);
    } catch {
      sessionStorage.removeItem("authUser");
    }
  }

  // Tenant context
  const tenantId = ref<string | null>(sessionStorage.getItem("tenantId"));

  // Computed properties
  const isAuthenticated = computed(() => !!accessToken.value && !!user.value);

  const isTokenExpired = computed(() => {
    if (!expiresAt.value) return true;
    // Add 10 second buffer for network latency
    return Date.now() >= expiresAt.value - 10000;
  });

  // For backward compatibility with existing code that uses 'token'
  const token = computed(() => accessToken.value);

  // Track ongoing refresh to prevent concurrent refreshes
  let refreshPromise: Promise<boolean> | null = null;

  /**
   * Set authentication data after login
   */
  const setAuth = (auth: {
    token?: string;
    access_token?: string;
    user: User;
    tenant_id?: string;
    expires_in?: number;
  }) => {
    // Support both 'token' and 'access_token' field names
    accessToken.value = auth.access_token || auth.token || null;
    user.value = auth.user;

    // Calculate expiration time (default 15 minutes if not provided)
    const expiresInMs = (auth.expires_in || 900) * 1000;
    expiresAt.value = Date.now() + expiresInMs;

    if (auth.tenant_id) {
      tenantId.value = auth.tenant_id;
      sessionStorage.setItem("tenantId", auth.tenant_id);
    }

    // Store user info only (not token!) in sessionStorage
    sessionStorage.setItem("authUser", JSON.stringify(auth.user));
  };

  /**
   * Clear all authentication data
   */
  const clearAuth = () => {
    accessToken.value = null;
    user.value = null;
    expiresAt.value = null;
    tenantId.value = null;

    sessionStorage.removeItem("authUser");
    sessionStorage.removeItem("tenantId");

    // Also clear any legacy localStorage items
    localStorage.removeItem("authToken");
    localStorage.removeItem("authUser");
  };

  /**
   * Refresh access token using HttpOnly cookie (sent automatically)
   * The refresh token is stored in a secure HttpOnly cookie by the server
   */
  const refreshAccessToken = async (): Promise<boolean> => {
    // Prevent concurrent refresh attempts
    if (refreshPromise) {
      return refreshPromise;
    }

    refreshPromise = (async () => {
      try {
        const response = await fetch(`${API_BASE}/auth/refresh`, {
          method: "POST",
          credentials: "include", // IMPORTANT: Send HttpOnly cookies!
          headers: {
            "Content-Type": "application/json",
            ...(tenantId.value ? { "x-tenant-id": tenantId.value } : {}),
          },
        });

        if (response.ok) {
          const data = await response.json();
          if (data.success && data.access_token) {
            accessToken.value = data.access_token;
            expiresAt.value = Date.now() + (data.expires_in || 900) * 1000;
            return true;
          }
        }

        // Refresh failed - could be expired or revoked
        // Check if it's a token reuse attack
        if (response.status === 401) {
          const data = await response.json().catch(() => ({}));
          if (data.code === "TOKEN_REUSE") {
            console.error("Security alert: Token reuse detected");
          }
        }
      } catch (error) {
        console.error("Token refresh failed:", error);
      }

      // Clear auth on failure
      clearAuth();
      return false;
    })();

    try {
      return await refreshPromise;
    } finally {
      refreshPromise = null;
    }
  };

  /**
   * Get a valid access token, refreshing if needed
   * Use this method before making authenticated API calls
   */
  const getValidToken = async (): Promise<string | null> => {
    // If we have a valid token, return it
    if (accessToken.value && !isTokenExpired.value) {
      return accessToken.value;
    }

    // If we have user info (session), try to refresh
    if (user.value) {
      const refreshed = await refreshAccessToken();
      if (refreshed) {
        return accessToken.value;
      }
    }

    return null;
  };

  /**
   * Initialize auth state on page load
   * Attempts to restore session using refresh token cookie
   */
  const initializeAuth = async (): Promise<boolean> => {
    // If we have stored user info but no token, try to refresh
    if (user.value && !accessToken.value) {
      return await refreshAccessToken();
    }
    return !!accessToken.value;
  };

  /**
   * Logout - clears local state and calls server logout
   */
  const logout = async (): Promise<void> => {
    try {
      await fetch(`${API_BASE}/auth/logout`, {
        method: "POST",
        credentials: "include",
        headers: {
          ...(accessToken.value
            ? { Authorization: `Bearer ${accessToken.value}` }
            : {}),
          ...(tenantId.value ? { "x-tenant-id": tenantId.value } : {}),
        },
      });
    } catch (error) {
      console.error("Logout request failed:", error);
    } finally {
      clearAuth();
    }
  };

  /**
   * Update tenant context
   */
  const setTenant = (newTenantId: string, newToken?: string) => {
    tenantId.value = newTenantId;
    sessionStorage.setItem("tenantId", newTenantId);
    if (newToken) {
      accessToken.value = newToken;
      // Reset expiration for new token
      expiresAt.value = Date.now() + 900 * 1000; // Default 15 min
    }
  };

  return {
    // State (reactive)
    accessToken,
    token, // Alias for backward compatibility
    user,
    tenantId,
    expiresAt,

    // Computed
    isAuthenticated,
    isTokenExpired,

    // Actions
    setAuth,
    clearAuth,
    refreshAccessToken,
    getValidToken,
    initializeAuth,
    logout,
    setTenant,
  };
});
