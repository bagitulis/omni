import axios, { AxiosInstance } from "axios";
import { IConfigManager } from "../../config/managers/baseConfigManager";
import { backendLogger } from "../../utils/backendLogger";

/**
 * Base API Client
 * Mirrors Python backend's BaseAPIClient
 * Handles common HTTP operations and token refresh logic
 */
export abstract class BaseAPIClient {
  protected config: IConfigManager;
  protected httpClient: AxiosInstance;
  protected baseUrl: string = "";

  constructor(config: IConfigManager) {
    this.config = config;

    this.httpClient = axios.create({
      timeout: 30000,
      headers: {
        "Content-Type": "application/json",
      },
    });
  }

  protected log = {
    info: (msg: string) =>
      backendLogger.info(this.config.platform.toUpperCase(), msg),
    error: (msg: string) =>
      backendLogger.error(this.config.platform.toUpperCase(), msg),
    warn: (msg: string) =>
      backendLogger.warning(this.config.platform.toUpperCase(), msg),
  };

  /**
   * Mask sensitive parameters in URL for logging
   * Masks: access_token, refresh_token, app_key, app_secret, sign, timestamp, shop_cipher, partner_id, sign_method
   * Also masks identifiable data: shop_id, order_sn, etc
   */
  protected maskUrlForLogging(url: string): string {
    return url
      .replace(/access_token=[^&]*/g, "access_token=[MASKED]")
      .replace(/refresh_token=[^&]*/g, "refresh_token=[MASKED]")
      .replace(/app_key=[^&]*/g, "app_key=[MASKED]")
      .replace(/app_secret=[^&]*/g, "app_secret=[MASKED]")
      .replace(/shop_cipher=[^&]*/g, "shop_cipher=[MASKED]")
      .replace(/partner_id=[^&]*/g, "partner_id=[MASKED]")
      .replace(/partner_key=[^&]*/g, "partner_key=[MASKED]")
      .replace(/shop_id=[^&]*/g, "shop_id=[MASKED]")
      .replace(/sign=[^&]*/g, "sign=[MASKED]")
      .replace(/timestamp=[^&]*/g, "timestamp=[MASKED]")
      .replace(/sign_method=[^&]*/g, "sign_method=[MASKED]")
      .replace(/order_sn_list=[^&]*/g, "order_sn_list=[MASKED]")
      .replace(/\?.*&page_size/, "?[PARAMS_MASKED]&page_size"); // Mask complex query strings
  }

  /**
   * Make raw API request
   */
  protected async makeRequest(
    endpoint: string,
    method: "GET" | "POST" | "PUT" | "DELETE" = "GET",
    params: Record<string, any> = {},
    data: Record<string, any> = {}
  ): Promise<any> {
    let url = `${this.baseUrl}${endpoint}`;

    // For Lazada compatibility: manually append query params for POST requests
    // instead of relying on axios which sometimes doesn't send them correctly
    if (Object.keys(params).length > 0) {
      const queryString = new URLSearchParams();
      for (const [key, value] of Object.entries(params)) {
        queryString.append(key, String(value));
      }
      url = `${url}?${queryString.toString()}`;
    }

    try {
      // Log masked URL
      const maskedUrl = this.maskUrlForLogging(url);
      this.log.info(`${method} ${maskedUrl}`);

      const response = await this.httpClient({
        method,
        url, // Use the URL with query string already appended
        params: {}, // Don't use axios params - we built it manually
        data: method === "GET" ? undefined : data,
      });

      return response.data;
    } catch (error: any) {
      const status = error.response?.status;
      const errorMsg = error.response?.data?.message || error.message;

      this.log.error(`HTTP ${status}: ${errorMsg}`);
      throw error;
    }
  }

  /**
   * Public API for making requests (for use in order managers, etc)
   */
  async request(
    endpoint: string,
    method: "GET" | "POST" | "PUT" | "DELETE" = "GET",
    params: Record<string, any> = {},
    data: Record<string, any> = {}
  ): Promise<any> {
    return this.makeRequest(endpoint, method, params, data);
  }
  /**
   * Abstract method - implement in subclasses
   */
  abstract refreshAccessToken(): Promise<boolean>;
}
