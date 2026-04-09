/**
 * API Configuration and Constants
 */

export function getBackendUrl(): string {
  // Always use relative /api path:
  // - Dev: Vite proxy forwards /api → localhost:3000 (avoids CORS)
  // - Production: Nginx proxy forwards /api → backend:3000
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
