import axios, { AxiosInstance, AxiosRequestConfig, AxiosError } from "axios";
import { API_BASE_URL, API_TIMEOUT, STORAGE_KEYS } from "@/lib/constants";

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
 * API Client Service - Axios instance with auth and CSRF interceptors
 * Mirrors Vue frontend patterns for compatibility
 */
class ApiClient {
  client: AxiosInstance;

  constructor() {
    this.client = axios.create({
      baseURL: API_BASE_URL,
      timeout: API_TIMEOUT.DEFAULT,
      headers: { "Content-Type": "application/json" },
    });

    this.setupInterceptors();
  }

  /**
   * Setup request and response interceptors
   */
  private setupInterceptors(): void {
    // Request interceptor - add auth and CSRF tokens
    this.client.interceptors.request.use(
      (config) => {
        // Add Bearer token from localStorage
        const token = localStorage.getItem(STORAGE_KEYS.AUTH_TOKEN);
        if (token) {
          config.headers.Authorization = `Bearer ${token}`;
        }

        // Add tenant ID from localStorage
        const tenantId = localStorage.getItem(STORAGE_KEYS.TENANT_ID);
        if (tenantId) {
          config.headers["x-tenant-id"] = tenantId;
        }

        // Add CSRF token for mutating requests (POST, PUT, PATCH, DELETE)
        if (
          config.method &&
          ["post", "put", "patch", "delete"].includes(
            config.method.toLowerCase(),
          )
        ) {
          const csrfToken = this.getCSRFToken();
          if (csrfToken) {
            config.headers["x-csrf-token"] = csrfToken;
          }
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
      (response) => response,
      (error: AxiosError) => this.handleResponseError(error),
    );
  }

  /**
   * Extract CSRF token from cookies
   */
  private getCSRFToken(): string | null {
    const match = document.cookie.match(/(?:^|;\s*)csrf_token=([^;]*)/);
    return match ? decodeURIComponent(match[1]) : null;
  }

  /**
   * Handle response errors
   * - Timeout errors: log and reject
   * - Network errors: suggest backend URL issue
   * - 401 errors: clear auth and redirect to login
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

    // Generic error handling
    const errorMsg =
      (error.response?.data as any)?.error || error.message || "Unknown error";
    console.error(`[API] Error [${error.response?.status}]:`, errorMsg);
    return Promise.reject(error);
  }

  /**
   * Handle expired/invalid JWT token
   * Clears all auth-related localStorage and redirects to login
   */
  private handleAuthExpired(): void {
    // Clear all auth-related localStorage items
    localStorage.removeItem(STORAGE_KEYS.AUTH_TOKEN);
    localStorage.removeItem(STORAGE_KEYS.AUTH_USER);
    localStorage.removeItem(STORAGE_KEYS.TENANT_ID);
    localStorage.removeItem(STORAGE_KEYS.USER_ROLE);
    localStorage.removeItem(STORAGE_KEYS.USER_NAME);

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
    data?: any,
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
    data?: any,
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
    data?: any,
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
