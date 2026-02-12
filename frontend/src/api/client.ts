import axios, { AxiosInstance, AxiosRequestConfig, AxiosError } from "axios";
import { message } from "antd";
import { API_BASE_URL, API_TIMEOUT } from "@/lib/constants";
import { useAuthStore } from "@/stores/authStore";
import { logger } from "@/lib/logger";
import {
  getPlatformOperationMapping,
  mapSheetsOperation,
  mapOrderExport,
  isDirectOrderExport,
  isDirectSheetsOperation,
} from "./operationMappers";

/**
 * API Response Type - matches backend response format
 */
export interface ApiResponse<T = any> {
  success: boolean;
  data?: T;
  error?: string;
  message?: string;
}

export interface ExecutionResponse {
  success: boolean;
  data?: any;
  error?: string;
  message?: string;
}

export interface ShippingFileResponse {
  success: boolean;
  data?: { files: string[] };
  error?: string;
}

export interface ShippingProcessResponse {
  success: boolean;
  data?: { processed: number; errors: string[] };
  error?: string;
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
   * Get CSRF token from cookies
   */
  private getCSRFToken(): string | null {
    const match = document.cookie.match(/(?:^|;\s*)csrf_token=([^;]*)/);
    return match ? decodeURIComponent(match[1]) : null;
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
          logger.debug(`[API] ${config.method?.toUpperCase()} ${config.url}`);
        }

        // Skip auth for login/refresh/dev-login endpoints to avoid loops
        if (
          config.url?.includes("/auth/login") ||
          config.url?.includes("/auth/refresh") ||
          config.url?.includes("/auth/dev-login")
        ) {
          return config;
        }

        // Get valid access token (handles auto-refresh)
        const token = await useAuthStore.getState().getValidToken();

        if (token) {
          config.headers.Authorization = `Bearer ${token}`;
        }

        // Add tenant ID from store
        const tenantId = useAuthStore.getState().tenantId;
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
        logger.error("[API] Request failed:", { message: error.message });
        return Promise.reject(error);
      },
    );

    // Response interceptor - handle errors
    this.client.interceptors.response.use(
      (response) => {
        // Dev-mode logging
        if (import.meta.env.DEV) {
          logger.debug(`[API] ✓ ${response.status} ${response.config.url}`);
        }
        return response;
      },
      (error: AxiosError) => this.handleResponseError(error),
    );
  }

  /**
   * Handle response errors
   */
  private handleResponseError(error: AxiosError): Promise<never> {
    if (error.code === "ECONNABORTED") {
      const timeoutMsg =
        "Request timeout - server is taking too long to respond";
      logger.error("[API]", { error: timeoutMsg });
      return Promise.reject(new Error(timeoutMsg));
    }

    if (!error.response) {
      const host = window.location.hostname;
      const isLocalhost = host === "localhost" || host === "127.0.0.1";
      const backendUrl = isLocalhost
        ? window.location.origin.replace(/:\d+$/, "") + ":3000"
        : window.location.origin;
      const networkMsg = `Network error - cannot connect to ${backendUrl}`;
      logger.error("[API]", { error: networkMsg });
      return Promise.reject(new Error(networkMsg));
    }

    // Handle 401 Unauthorized - JWT expired or invalid
    if (error.response?.status === 401) {
      const currentPath = window.location.pathname;
      // Don't redirect if already on login page or auth endpoints
      if (currentPath !== "/login" && !error.config?.url?.includes("/auth/")) {
        logger.info(
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
      logger.error(`[API] 403 Forbidden:`, { error: errorMsg });
      return Promise.reject(new Error(errorMsg));
    }

    // Handle 500 Server Error
    if (error.response?.status === 500) {
      const backendError =
        (error.response?.data as { error?: string })?.error || "Server error";
      const errorMsg = `Server error - ${backendError}`;
      message.error("Server error - please try again");
      logger.error(`[API] 500 Server Error:`, { error: errorMsg });
      return Promise.reject(new Error(errorMsg));
    }

    // Generic error handling
    const errorMsg =
      (error.response?.data as { error?: string })?.error ||
      error.message ||
      "Unknown error";
    logger.error(`[API] Error [${error.response?.status}]:`, {
      error: errorMsg,
    });
    return Promise.reject(error);
  }

  /**
   * Handle expired/invalid JWT token
   */
  private handleAuthExpired(): void {
    useAuthStore.getState().clearAuth();
    const currentPath = window.location.pathname;
    const returnUrl =
      currentPath !== "/"
        ? `?returnUrl=${encodeURIComponent(currentPath)}`
        : "";
    window.location.href = `/login${returnUrl}`;
  }

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

  async healthCheck(): Promise<ApiResponse<any>> {
    return this.get("/health", { timeout: API_TIMEOUT.HEALTH });
  }

  // --- NEW OPERATION METHODS ---

  async executeOperation(
    operation: string,
    // eslint-disable-next-line @typescript-eslint/no-explicit-any
    params: Record<string, any> = {},
  ): Promise<ExecutionResponse> {
    try {
      const mapping = getPlatformOperationMapping(operation, params);

      if (mapping) {
        const response = await this.client.post<ExecutionResponse>(
          mapping.endpoint,
          params,
          { timeout: API_TIMEOUT.LONG },
        );
        return response.data;
      }

      const response = await this.client.post<ExecutionResponse>(
        "/execute",
        { operation, params },
        { timeout: API_TIMEOUT.LONG },
      );
      return response.data;
    } catch (error) {
      logger.error(`[API] Operation ${operation} failed:`, { error });
      throw error;
    }
  }

  async executeSheetsOperation(
    operation: string,
    // eslint-disable-next-line @typescript-eslint/no-explicit-any
    params: Record<string, any> = {},
    // eslint-disable-next-line @typescript-eslint/no-explicit-any
  ): Promise<any> {
    try {
      const endpoint = mapSheetsOperation(operation);
      const body = isDirectSheetsOperation(operation)
        ? params
        : { operation, params };

      const response = await this.client.post<any>(endpoint, body, {
        timeout: API_TIMEOUT.LONG,
      });
      return response.data;
    } catch (error) {
      logger.error(`[API] Sheets operation ${operation} failed:`, { error });
      throw error;
    }
  }

  async exportOrders(
    platform: string,
    orderType: string,
    days: number = 7,
    // eslint-disable-next-line @typescript-eslint/no-explicit-any
  ): Promise<any> {
    try {
      const endpoint = mapOrderExport(platform, orderType);
      const body = isDirectOrderExport(platform, orderType)
        ? { days }
        : { platform, order_type: orderType, days };

      const response = await this.client.post<any>(endpoint, body, {
        timeout: API_TIMEOUT.LONG,
      });
      return response.data;
    } catch (error) {
      logger.error(`[API] Export orders ${platform} failed:`, { error });
      throw error;
    }
  }

  async getShippingFiles(): Promise<ShippingFileResponse> {
    try {
      const response = await this.client.get<ShippingFileResponse>(
        "/shipping/files",
        { timeout: API_TIMEOUT.SHORT },
      );
      return response.data;
    } catch (error) {
      logger.error("[API] Get shipping files failed:", { error });
      throw error;
    }
  }

  async processShippingFile(
    filename: string,
  ): Promise<ShippingProcessResponse> {
    try {
      const response = await this.client.post<ShippingProcessResponse>(
        "/shipping/process-file",
        { filename },
        { timeout: API_TIMEOUT.DEFAULT },
      );
      return response.data;
    } catch (error) {
      logger.error("[API] Process shipping file failed:", { error });
      throw error;
    }
  }
}

export const apiClient = new ApiClient();
export default apiClient;
