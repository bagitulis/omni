import { create } from "zustand";
import { STORAGE_KEYS } from "@/lib/constants";
import { User } from "@/types/auth";

interface AuthState {
  user: User | null;
  token: string | null;
  isAuthenticated: boolean;
  tenantId: string | null;

  // Actions
  setAuth: (data: { token: string; user: User; tenant_id?: string }) => void;
  logout: () => void;
  initialize: () => void;
}

export const useAuthStore = create<AuthState>((set) => ({
  user: null,
  token: null,
  isAuthenticated: false,
  tenantId: null,

  setAuth: ({ token, user, tenant_id }) => {
    // Update LocalStorage (Critical for API Client)
    localStorage.setItem(STORAGE_KEYS.AUTH_TOKEN, token);
    localStorage.setItem(STORAGE_KEYS.AUTH_USER, JSON.stringify(user));

    if (tenant_id) {
      localStorage.setItem(STORAGE_KEYS.TENANT_ID, tenant_id);
    }

    // Legacy keys for compatibility if needed (based on Vue login view)
    localStorage.setItem(STORAGE_KEYS.USER_ROLE, user.role);
    localStorage.setItem(STORAGE_KEYS.USER_NAME, user.username);

    set({
      token,
      user,
      tenantId: tenant_id || null,
      isAuthenticated: true,
    });
  },

  logout: () => {
    // LocalStorage clearing is also handled in api/auth.ts logout(),
    // but good to have here for state consistency
    localStorage.removeItem(STORAGE_KEYS.AUTH_TOKEN);
    localStorage.removeItem(STORAGE_KEYS.AUTH_USER);
    localStorage.removeItem(STORAGE_KEYS.TENANT_ID);
    localStorage.removeItem(STORAGE_KEYS.USER_ROLE);
    localStorage.removeItem(STORAGE_KEYS.USER_NAME);

    set({
      user: null,
      token: null,
      tenantId: null,
      isAuthenticated: false,
    });
  },

  initialize: () => {
    const token = localStorage.getItem(STORAGE_KEYS.AUTH_TOKEN);
    const userStr = localStorage.getItem(STORAGE_KEYS.AUTH_USER);
    const tenantId = localStorage.getItem(STORAGE_KEYS.TENANT_ID);

    if (token && userStr) {
      try {
        const user = JSON.parse(userStr);
        set({
          token,
          user,
          tenantId,
          isAuthenticated: true,
        });
      } catch (e) {
        console.error("Failed to parse user from localStorage", e);
        // Invalid state, clear it
        localStorage.removeItem(STORAGE_KEYS.AUTH_TOKEN);
        localStorage.removeItem(STORAGE_KEYS.AUTH_USER);
        set({ user: null, token: null, isAuthenticated: false });
      }
    }
  },
}));
