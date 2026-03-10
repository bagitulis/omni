import { AxiosError } from "axios";
import { message } from "@/components/AntStaticHolder";
import { useAuthStore } from "@/stores/authStore";
import { logger } from "@/lib/logger";
import { sanitizeForUser } from "@/lib/notificationSecurity";

/**
 * Handle expired/invalid JWT token — clear auth and redirect to login.
 */
export function handleAuthExpired(): void {
  useAuthStore.getState().clearAuth();
  const currentPath = window.location.pathname;
  const returnUrl =
    currentPath !== "/"
      ? `?returnUrl=${encodeURIComponent(currentPath)}`
      : "";
  window.location.href = `/login${returnUrl}`;
}

/**
 * Centralized HTTP error handler for Axios responses.
 * Returns a rejected promise with a descriptive Error.
 */
export function handleResponseError(error: AxiosError): Promise<never> {
  // Timeout
  if (error.code === "ECONNABORTED") {
    const timeoutMsg =
      "Request timeout - server is taking too long to respond";
    logger.error("[API]", { error: timeoutMsg });
    return Promise.reject(new Error(timeoutMsg));
  }

  // Network error (no response at all)
  if (!error.response) {
    const host = window.location.hostname;
    const isLocalhost = host === "localhost" || host === "127.0.0.1";
    const backendUrl = isLocalhost
      ? window.location.origin.replace(/:\d+$/, "") + ":3000"
      : window.location.origin;
    logger.error("[API] Network error", { backendUrl });
    const networkMsg = "Network error — cannot connect to server. Please check your connection.";
    return Promise.reject(new Error(networkMsg));
  }

  const { status, data } = error.response;
  const backendMsg = (data as { error?: string })?.error;

  // 401 Unauthorized
  if (status === 401) {
    const currentPath = window.location.pathname;
    if (currentPath !== "/login" && !error.config?.url?.includes("/auth/")) {
      logger.info(
        "[API] JWT token expired or invalid - redirecting to login",
      );
      handleAuthExpired();
      return Promise.reject(
        new Error("Session expired - please login again"),
      );
    }
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
    const errorMsg = backendMsg || "Server error — please try again";
    logger.error(`[API] 500 Server Error:`, { error: backendMsg });
    message.error("Server error — please try again");
    return Promise.reject(new Error(errorMsg));
  }

  // Generic error — sanitize before surfacing to the user
  const rawMsg = backendMsg || error.message || "Unknown error";
  logger.error(`[API] Error [${status}]:`, { error: rawMsg });
  const safeError = new Error(sanitizeForUser(rawMsg));
  return Promise.reject(safeError);
}
