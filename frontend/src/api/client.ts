import axios, { AxiosInstance, AxiosRequestConfig, AxiosError } from "axios";
import { message } from "antd";
import { API_BASE_URL, API_TIMEOUT } from "@/lib/constants";
import { useAuthStore } from "@/stores/authStore";

/**
 * API Response Type - matches backend response format
 */
export interface ApiResponse<T = any> {
  success: boolean;
  data?: T;
  error?: string;
  message?: string;
}

/**
 * API Client Service - Axios instance with auth interceptors
 * Mirrors Vue frontend patterns for compatibility
 */
class ApiClient {
  client: AxiosInstance;

  constructor() {
    this.client = axios.create({
      baseURL: API_BASE_URL,
      timeout: API_TIMEOUT.DEFAULT,
      headers: { "Content-Type": "application/json" },
      withCredentials: true,
    });

    this.setupInterceptors();
  }

  /**
   * Setup request and response interceptors
   */
  private setupInterceptors(): void {
    // Request interceptor - add auth token
    this.client.interceptors.request.use(
      async (config) => {
        // Dev-mode logging
        if (import.meta.env.DEV) {
          console.debug(`[API] ${config.method?.toUpperCase()} ${config.url}`);
        }

        // Skip auth for login/refresh/dev-login endpoints to avoid loops
        // Auth endpoints (login, refresh, dev-login) do not need the bearer token
        // and attempting to get it might trigger a refresh loop
        if (
          config.url?.includes("/auth/login") ||
          config.url?.includes("/auth/refresh") ||
          config.url?.includes("/auth/dev-login")
        ) {
          return config;
        }

        // Get valid access token (handles auto-refresh)
        // This is async because it might need to refresh the token
        const token = await useAuthStore.getState().getValidToken();

        if (token) {
          config.headers.Authorization = `Bearer ${token}`;
        }

        // Add tenant ID from store
        const tenantId = useAuthStore.getState().tenantId;
        if (tenantId) {
          config.headers["x-tenant-id"] = tenantId;
        }

        return config;
      },
      (error) => {
        console.error("[API] Request failed:", error.message);
        return Promise.reject(error);
      },
    );

    // Response interceptor - handle errors
    this.client.interceptors.response.use(
      (response) => {
        // Dev-mode logging
        if (import.meta.env.DEV) {
          console.debug(`[API] ✓ ${response.status} ${response.config.url}`);
        }
        return response;
      },
      (error: AxiosError) => this.handleResponseError(error),
    );
  }

  /**
   * Handle response errors
   * - Timeout errors: log and reject
   * - Network errors: suggest backend URL issue
   * - 401 errors: clear auth and redirect to login
   * - 403 errors: show permission denied toast
   * - 500 errors: show error toast with backend message
   * - Other errors: reject with error message
   */
  private handleResponseError(error: AxiosError): Promise<never> {
    // Handle timeout
    if (error.code === "ECONNABORTED") {
      const timeoutMsg =
        "Request timeout - server is taking too long to respond";
      console.error("[API]", timeoutMsg);
      return Promise.reject(new Error(timeoutMsg));
    }

    // Handle network errors (no response)
    if (!error.response) {
      const host = window.location.hostname;
      const isLocalhost = host === "localhost" || host === "127.0.0.1";
      const backendUrl = isLocalhost
        ? window.location.origin.replace(/:\d+$/, "") + ":3000"
        : window.location.origin;
      const networkMsg = `Network error - cannot connect to ${backendUrl}`;
      console.error("[API]", networkMsg);
      return Promise.reject(new Error(networkMsg));
    }

    // Handle 401 Unauthorized - JWT expired or invalid
    if (error.response?.status === 401) {
      const currentPath = window.location.pathname;
      // Don't redirect if already on login page or auth endpoints
      if (currentPath !== "/login" && !error.config?.url?.includes("/auth/")) {
        console.info(
          "[API] JWT token expired or invalid - redirecting to login",
        );
        this.handleAuthExpired();
        return Promise.reject(
          new Error("Session expired - please login again"),
        );
      }
    }

    // Handle 403 Forbidden - permission denied
    if (error.response?.status === 403) {
      const errorMsg = "Permission denied";
      message.error(errorMsg);
      console.error(`[API] 403 Forbidden:`, errorMsg);
      return Promise.reject(new Error(errorMsg));
    }

    // Handle 500 Server Error
    if (error.response?.status === 500) {
      const backendError =
        (error.response?.data as { error?: string })?.error || "Server error";
      const errorMsg = `Server error - ${backendError}`;
      message.error("Server error - please try again");
      console.error(`[API] 500 Server Error:`, errorMsg);
      return Promise.reject(new Error(errorMsg));
    }

    // Generic error handling
    const errorMsg =
      (error.response?.data as { error?: string })?.error ||
      error.message ||
      "Unknown error";
    console.error(`[API] Error [${error.response?.status}]:`, errorMsg);
    return Promise.reject(error);
  }

  /**
   * Handle expired/invalid JWT token
   * Clears all auth-related state and redirects to login
   */
  private handleAuthExpired(): void {
    // Clear all auth-related state
    useAuthStore.getState().clearAuth();

    // Redirect to login page with return URL
    const currentPath = window.location.pathname;
    const returnUrl =
      currentPath !== "/"
        ? `?returnUrl=${encodeURIComponent(currentPath)}`
        : "";
    window.location.href = `/login${returnUrl}`;
  }

  /**
   * GET request
   */
  async get<T = any>(
    url: string,
    config: AxiosRequestConfig = {},
  ): Promise<ApiResponse<T>> {
    const response = await this.client.get<ApiResponse<T>>(url, {
      ...config,
      timeout: config.timeout ?? API_TIMEOUT.SHORT,
    });
    return response.data;
  }

  /**
   * POST request
   */
  async post<T = any>(
    url: string,
    data?: unknown,
    config: AxiosRequestConfig = {},
  ): Promise<ApiResponse<T>> {
    const response = await this.client.post<ApiResponse<T>>(url, data, {
      ...config,
      timeout: config.timeout ?? API_TIMEOUT.DEFAULT,
    });
    return response.data;
  }

  /**
   * PUT request
   */
  async put<T = any>(
    url: string,
    data?: unknown,
    config: AxiosRequestConfig = {},
  ): Promise<ApiResponse<T>> {
    const response = await this.client.put<ApiResponse<T>>(url, data, {
      ...config,
      timeout: config.timeout ?? API_TIMEOUT.DEFAULT,
    });
    return response.data;
  }

  /**
   * PATCH request
   */
  async patch<T = any>(
    url: string,
    data?: unknown,
    config: AxiosRequestConfig = {},
  ): Promise<ApiResponse<T>> {
    const response = await this.client.patch<ApiResponse<T>>(url, data, {
      ...config,
      timeout: config.timeout ?? API_TIMEOUT.SHORT,
    });
    return response.data;
  }

  /**
   * DELETE request
   */
  async delete<T = any>(
    url: string,
    config: AxiosRequestConfig = {},
  ): Promise<ApiResponse<T>> {
    const response = await this.client.delete<ApiResponse<T>>(url, {
      ...config,
      timeout: config.timeout ?? API_TIMEOUT.SHORT,
    });
    return response.data;
  }

  /**
   * Health check endpoint
   */
  async healthCheck(): Promise<ApiResponse<any>> {
    return this.get("/health", { timeout: API_TIMEOUT.HEALTH });
  }
}

// Export singleton instance
export const apiClient = new ApiClient();

// Export for convenience in components
export default apiClient;
