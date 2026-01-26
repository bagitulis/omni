import axios, { AxiosInstance, AxiosRequestConfig, AxiosError } from "axios";
import { logger } from "./logger";
import {
  getPlatformOperationMapping,
  mapSheetsOperation,
  mapOrderExport,
  isDirectOrderExport,
  isDirectSheetsOperation,
} from "./apiOperationMappers";
import type {
  HealthCheckResponse,
  StatusResponse,
  ExecutionResponse,
  ShippingFileResponse,
  ShippingProcessResponse,
} from "../types/api";

function getBackendUrl(): string {
  if (import.meta.env.DEV) {
    const currentUrl = window.location.origin;
    return `${currentUrl.replace(/:\d+$/, "")}:3000/api`;
  }
  return import.meta.env.VITE_API_URL || "/api";
}

const API_BASE_URL = getBackendUrl();

class ApiService {
  client: AxiosInstance;

  constructor() {
    this.client = axios.create({
      baseURL: API_BASE_URL,
      timeout: 180000,
      headers: { "Content-Type": "application/json" },
    });

    logger.apiUrl(API_BASE_URL);
    this.setupInterceptors();
  }

  private setupInterceptors(): void {
    this.client.interceptors.request.use(
      (config) => {
        const token = localStorage.getItem("authToken");
        if (token) config.headers.Authorization = `Bearer ${token}`;

        const tenantId = localStorage.getItem("tenantId");
        if (tenantId) config.headers["x-tenant-id"] = tenantId;

        // Add CSRF token for mutating requests
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
        logger.error("Request failed", error.message);
        return Promise.reject(error);
      },
    );

    this.client.interceptors.response.use(
      (response) => response,
      (error: AxiosError) => this.handleResponseError(error),
    );
  }

  /**
   * Get CSRF token from cookies
   */
  private getCSRFToken(): string | null {
    const match = document.cookie.match(/(?:^|;\s*)csrf_token=([^;]*)/);
    return match ? decodeURIComponent(match[1]) : null;
  }

  private handleResponseError(error: AxiosError): Promise<never> {
    if (error.code === "ECONNABORTED") {
      logger.error("Request timeout - server is taking too long to respond");
      return Promise.reject(new Error("Request timeout"));
    }

    if (!error.response) {
      const host = window.location.hostname;
      const isLocalhost = host === "localhost" || host === "127.0.0.1";
      const backendUrl = isLocalhost
        ? window.location.origin.replace(/:\d+$/, "") + ":3000"
        : window.location.origin;
      logger.error("Network error - Backend not running?", backendUrl);
      return Promise.reject(
        new Error(`Network error - cannot connect to ${backendUrl}`),
      );
    }

    // Handle 401 Unauthorized - JWT expired or invalid
    if (error.response?.status === 401) {
      const currentPath = window.location.pathname;
      // Don't redirect if already on login page or auth endpoints
      if (currentPath !== "/login" && !error.config?.url?.includes("/auth/")) {
        logger.info("JWT token expired or invalid - redirecting to login");
        this.handleAuthExpired();
        return Promise.reject(
          new Error("Session expired - please login again"),
        );
      }
    }

    const errorMsg = (error.response?.data as any)?.error || error.message;
    logger.error(`API Error [${error.response?.status}]`, errorMsg);
    return Promise.reject(error);
  }

  /**
   * Handle expired/invalid JWT token
   * Clear auth data and redirect to login page
   */
  private handleAuthExpired(): void {
    // Clear all auth-related localStorage items
    localStorage.removeItem("authToken");
    localStorage.removeItem("authUser");
    localStorage.removeItem("tenantId");
    localStorage.removeItem("userRole");
    localStorage.removeItem("userName");

    // Redirect to login page with return URL
    const currentPath = window.location.pathname;
    const returnUrl =
      currentPath !== "/"
        ? `?returnUrl=${encodeURIComponent(currentPath)}`
        : "";
    window.location.href = `/login${returnUrl}`;
  }

  async get<T = any>(url: string, config: AxiosRequestConfig = {}): Promise<T> {
    try {
      const response = await this.client.get<T>(url, {
        ...config,
        timeout: 15000,
      });
      return response.data;
    } catch (error) {
      throw this.handleError(error);
    }
  }

  async post<T = any>(
    url: string,
    data?: any,
    config: AxiosRequestConfig = {},
  ): Promise<T> {
    try {
      const response = await this.client.post<T>(url, data, {
        ...config,
        timeout: 30000,
      });
      return response.data;
    } catch (error) {
      throw this.handleError(error);
    }
  }

  async patch<T = any>(
    url: string,
    data?: any,
    config: AxiosRequestConfig = {},
  ): Promise<T> {
    try {
      const response = await this.client.patch<T>(url, data, {
        ...config,
        timeout: 15000,
      });
      return response.data;
    } catch (error) {
      throw this.handleError(error);
    }
  }

  async delete<T = any>(
    url: string,
    config: AxiosRequestConfig = {},
  ): Promise<T> {
    try {
      const response = await this.client.delete<T>(url, {
        ...config,
        timeout: 15000,
      });
      return response.data;
    } catch (error) {
      throw this.handleError(error);
    }
  }

  async healthCheck(): Promise<HealthCheckResponse> {
    try {
      const response = await this.client.get<HealthCheckResponse>("/health", {
        timeout: 10000,
      });
      return response.data;
    } catch (error) {
      throw this.handleError(error);
    }
  }

  async getStatus(): Promise<StatusResponse> {
    try {
      const response = await this.client.get<StatusResponse>("/status", {
        timeout: 15000,
      });
      return response.data;
    } catch (error) {
      throw this.handleError(error);
    }
  }

  async getTokenStatus(): Promise<any> {
    try {
      const response = await this.client.get("/token-status", {
        timeout: 15000,
      });
      return response.data;
    } catch (error) {
      throw this.handleError(error);
    }
  }

  async refreshAllTokens(force: boolean = true): Promise<any> {
    try {
      const response = await this.client.post(
        `/tokens/refresh-all?force=${force}`,
        {},
        {
          timeout: 30000,
        },
      );
      return response.data;
    } catch (error) {
      throw this.handleError(error);
    }
  }

  async executeOperation(
    operation: string,
    params: Record<string, any> = {},
  ): Promise<ExecutionResponse> {
    try {
      const mapping = getPlatformOperationMapping(operation, params);

      if (mapping) {
        const response = await this.client.post<ExecutionResponse>(
          mapping.endpoint,
          params,
          { timeout: 60000 },
        );
        return response.data;
      }

      const response = await this.client.post<ExecutionResponse>(
        "/execute",
        { operation, params },
        { timeout: 60000 },
      );
      return response.data;
    } catch (error) {
      console.error(`[API] Operation ${operation} failed:`, error);
      throw this.handleError(error);
    }
  }

  async executeSheetsOperation(
    operation: string,
    params: Record<string, any> = {},
  ): Promise<any> {
    try {
      const endpoint = mapSheetsOperation(operation);
      const body = isDirectSheetsOperation(operation)
        ? params
        : { operation, params };

      const response = await this.client.post<any>(endpoint, body, {
        timeout: 60000,
      });
      return response.data;
    } catch (error) {
      throw this.handleError(error);
    }
  }

  async exportOrders(
    platform: string,
    orderType: string,
    days: number = 7,
  ): Promise<any> {
    try {
      const endpoint = mapOrderExport(platform, orderType);
      const body = isDirectOrderExport(platform, orderType)
        ? { days }
        : { platform, order_type: orderType, days };

      const response = await this.client.post<any>(endpoint, body, {
        timeout: 60000,
      });
      return response.data;
    } catch (error) {
      throw this.handleError(error);
    }
  }

  async getShippingFiles(): Promise<ShippingFileResponse> {
    try {
      const response = await this.client.get<ShippingFileResponse>(
        "/shipping/files",
        { timeout: 15000 },
      );
      return response.data;
    } catch (error) {
      throw this.handleError(error);
    }
  }

  async processShippingFile(
    filename: string,
  ): Promise<ShippingProcessResponse> {
    try {
      const response = await this.client.post<ShippingProcessResponse>(
        "/shipping/process-file",
        { filename },
        { timeout: 30000 },
      );
      return response.data;
    } catch (error) {
      throw this.handleError(error);
    }
  }

  async debugCheckFunctions(): Promise<any> {
    try {
      const response = await this.client.get<any>("/debug/check-functions", {
        timeout: 15000,
      });
      return response.data;
    } catch (error) {
      throw this.handleError(error);
    }
  }

  private handleError(error: any): Error {
    if (error.response) {
      const message =
        error.response.data?.error || `Server error: ${error.response.status}`;
      console.error("API Error Response:", error.response);
      return new Error(message);
    }

    if (error.request) {
      console.error("API No Response:", error.request);
      const host = window.location.hostname;
      const isLocalhost = host === "localhost" || host === "127.0.0.1";
      const backendUrl = isLocalhost
        ? window.location.origin.replace(/:\d+$/, "") + ":3000"
        : window.location.origin;
      return new Error(`Network error - cannot connect to ${backendUrl}`);
    }

    console.error("API Error:", error.message);
    return new Error(error.message);
  }
}

export default new ApiService();
