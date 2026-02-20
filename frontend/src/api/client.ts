import axios, { AxiosInstance, AxiosRequestConfig } from "axios";
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
import { handleResponseError } from "./clientErrorHandler";

/**
 * API Response Type - matches backend response format
 */
export interface ApiResponse<T = unknown> {
  success: boolean;
  data?: T;
  error?: string;
  message?: string;
}

export interface ExecutionResponse {
  success: boolean;
  data?: unknown;
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

  private getCSRFToken(): string | null {
    const match = document.cookie.match(/(?:^|;\s*)csrf_token=([^;]*)/);
    return match ? decodeURIComponent(match[1]) : null;
  }

  private setupInterceptors(): void {
    this.client.interceptors.request.use(
      async (config) => {
        if (import.meta.env.DEV) {
          logger.debug(`[API] ${config.method?.toUpperCase()} ${config.url}`);
        }

        if (
          config.url?.includes("/auth/login") ||
          config.url?.includes("/auth/refresh") ||
          config.url?.includes("/auth/dev-login")
        ) {
          return config;
        }

        const token = await useAuthStore.getState().getValidToken();
        if (token) {
          config.headers.Authorization = `Bearer ${token}`;
        }

        const tenantId = useAuthStore.getState().tenantId;
        if (tenantId) {
          config.headers["x-tenant-id"] = tenantId;
        }

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

    this.client.interceptors.response.use(
      (response) => {
        if (import.meta.env.DEV) {
          logger.debug(`[API] ✓ ${response.status} ${response.config.url}`);
        }
        return response;
      },
      (error) => handleResponseError(error),
    );
  }

  // --- HTTP Methods ---

  async get<T = unknown>(
    url: string,
    config: AxiosRequestConfig = {},
  ): Promise<ApiResponse<T>> {
    const response = await this.client.get<ApiResponse<T>>(url, {
      ...config,
      timeout: config.timeout ?? API_TIMEOUT.SHORT,
    });
    return response.data;
  }

  async post<T = unknown>(
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

  async put<T = unknown>(
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

  async patch<T = unknown>(
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

  async delete<T = unknown>(
    url: string,
    config: AxiosRequestConfig = {},
  ): Promise<ApiResponse<T>> {
    const response = await this.client.delete<ApiResponse<T>>(url, {
      ...config,
      timeout: config.timeout ?? API_TIMEOUT.SHORT,
    });
    return response.data;
  }

  async healthCheck(): Promise<ApiResponse<unknown>> {
    return this.get("/health", { timeout: API_TIMEOUT.HEALTH });
  }

  // --- Domain Operations ---

  async executeOperation(
    operation: string,
    params: Record<string, unknown> = {},
  ): Promise<ExecutionResponse> {
    try {
      const mapping = getPlatformOperationMapping(operation, params);
      const endpoint = mapping ? mapping.endpoint : "/execute";
      const body = mapping ? params : { operation, params };

      const response = await this.client.post<ExecutionResponse>(
        endpoint,
        body,
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

      const response = await this.client.post<unknown>(endpoint, body, {
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

      const response = await this.client.post<unknown>(endpoint, body, {
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
