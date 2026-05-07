import axios from "axios";
import type { AxiosError, InternalAxiosRequestConfig } from "axios";
import { message } from "@/components/AntStaticApi";
import { useAuthStore } from "@/stores/authStore";
import { logger } from "@/lib/logger";
import { sanitizeForUser } from "@/lib/notificationSecurity";

// Track if a 401 retry is already in progress to prevent infinite loops
let isRetrying = false;

/**
 * Handle expired/invalid JWT token — clear auth and redirect to login.
 */
export function handleAuthExpired(): void {
  useAuthStore.getState().clearAuth();
  const currentPath = window.location.pathname;
  const returnUrl =
    currentPath !== "/" ? `?returnUrl=${encodeURIComponent(currentPath)}` : "";
  window.location.href = `/login${returnUrl}`;
}

/**
 * Centralized HTTP error handler for Axios responses.
 * On 401 for non-auth endpoints: attempts token refresh before redirecting.
 * Returns a rejected promise with a descriptive Error.
 */
export function handleResponseError(error: AxiosError): Promise<never> {
  // Timeout
  if (error.code === "ECONNABORTED") {
    const timeoutMsg = "Request timeout - server is taking too long to respond";
    logger.error("[API]", { error: timeoutMsg });
    return Promise.reject(new Error(timeoutMsg));
  }

  // Network error (no response at all)
  if (!error.response) {
    logger.error("[API] Network error", {
      url: error.config?.url,
      origin: window.location.origin,
    });
    const networkMsg =
      "Network error \u2014 cannot connect to server. Please check your connection.";
    return Promise.reject(new Error(networkMsg));
  }

  const { status, data } = error.response;
  const backendMsg = (data as { error?: string })?.error;

  // 401 Unauthorized
  if (status === 401) {
    const currentPath = window.location.pathname;
    const isAuthEndpoint = error.config?.url?.includes("/auth/") ?? false;

    // Auth endpoints (login, register, refresh) — pass backend message through directly
    // These messages are controlled constants (e.g. "Invalid username or password")
    if (isAuthEndpoint) {
      const authMsg = backendMsg || "Authentication failed";
      logger.info(`[API] Auth endpoint 401:`, { error: authMsg });
      return Promise.reject(new Error(authMsg));
    }

    // Non-auth 401 — attempt token refresh before giving up
    // This handles the case where access token expired between getValidToken() and server receipt
    if (currentPath !== "/login" && !isRetrying) {
      isRetrying = true;
      return useAuthStore
        .getState()
        .refreshAccessToken()
        .then((success) => {
          isRetrying = false;
          if (success && error.config) {
            // Retry the original request with new token
            const config = error.config as InternalAxiosRequestConfig;
            const newToken = useAuthStore.getState().accessToken;
            if (newToken) {
              config.headers.Authorization = `Bearer ${newToken}`;
            }
            // Retry the failed request with the refreshed token
            return axios(config);
          }
          // Refresh failed definitively — redirect to login
          handleAuthExpired();
          return Promise.reject(
            new Error("Session expired - please login again"),
          );
        })
        .catch((retryError: unknown) => {
          isRetrying = false;
          handleAuthExpired();
          return Promise.reject(
            retryError instanceof Error
              ? retryError
              : new Error("Session expired - please login again"),
          );
        }) as Promise<never>;
    }

    // Already retrying or on login page — just reject
    if (currentPath !== "/login") {
      handleAuthExpired();
    }
    return Promise.reject(new Error("Session expired - please login again"));
  }

  // 403 Forbidden
  if (status === 403) {
    const errorMsg = backendMsg || "Permission denied";
    message.error("Permission denied");
    logger.error(`[API] 403 Forbidden:`, { error: errorMsg });
    return Promise.reject(new Error(errorMsg));
  }

  // 500 Server Error
  if (status === 500) {
    const errorMsg = "Server error \u2014 please try again";
    logger.error(`[API] 500 Server Error:`, { error: backendMsg });
    message.error("Server error \u2014 please try again");
    return Promise.reject(new Error(errorMsg));
  }

  // Generic error — sanitize before surfacing to the user
  const rawMsg = backendMsg || error.message || "Unknown error";
  logger.error(`[API] Error [${status}]:`, { error: rawMsg });
  const safeError = new Error(sanitizeForUser(rawMsg));
  return Promise.reject(safeError);
}
