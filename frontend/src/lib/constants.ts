/**
 * API Configuration and Constants
 */

export function getBackendUrl(): string {
  if (import.meta.env.DEV) {
    const currentUrl = window.location.origin;
    // Only add port 3000 for true localhost development
    const isLocalhost =
      window.location.hostname === "localhost" ||
      window.location.hostname === "127.0.0.1";
    return isLocalhost ? `${currentUrl.replace(/:\d+$/, "")}:3000/api` : "/api";
  }
  // Production: Use relative URL (nginx proxy)
  return import.meta.env.VITE_API_URL || "/api";
}

export const API_BASE_URL = getBackendUrl();

// Request/Response timeout values (in milliseconds)
export const API_TIMEOUT = {
  HEALTH: 10000,
  SHORT: 15000,
  DEFAULT: 30000,
  LONG: 60000,
  EXTRA_LONG: 180_000,
};

// localStorage keys - must match Vue frontend for compatibility
export const STORAGE_KEYS = {
  AUTH_USER: "authUser",
  TENANT_ID: "tenantId",
  USER_ROLE: "userRole",
  USER_NAME: "userName",
};

// Cookie names
export const COOKIE_NAMES = {};
