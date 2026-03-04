import crypto from "crypto";
import { BaseAPIClient } from "./baseAPIClient";
import { LazadaConfigManager } from "../../config/managers/lazadaConfigManager";

/**
 * Lazada API Client
 * Handles Lazada-specific API operations and token refresh
 */
export class LazadaAPIClient extends BaseAPIClient {
  protected baseUrl: string = "";
  protected config: LazadaConfigManager;
  protected authUrl = "https://auth.lazada.com/rest"; // Token refresh endpoint ONLY

  // Country-specific order/product API gateways per official Lazada SDK
  private countryGateways: Record<string, string> = {
    sg: "https://api.lazada.sg/rest",
    th: "https://api.lazada.co.th/rest",
    my: "https://api.lazada.com.my/rest",
    vn: "https://api.lazada.vn/rest",
    ph: "https://api.lazada.com.ph/rest",
    id: "https://api.lazada.co.id/rest",
  };

  constructor(config: LazadaConfigManager) {
    super(config);
    this.config = config;
    // Set baseUrl based on country for order/product API calls
    const country = (config.country || "id").toLowerCase();
    this.baseUrl = this.countryGateways[country] || this.countryGateways.id;
    // Suppress debug log
  }

  /**
   * Generate Lazada signature using HMAC-SHA256
   * Per official Lazada SDK (lazop-sdk-python):
   * 1. Sort parameters alphabetically (excluding 'sign')
   * 2. Concatenate params as: key1value1key2value2...
   * 3. Prepend API path: /api/path + paramString
   * 4. HMAC-SHA256(appSecret, full_string)
   * 5. Convert to UPPERCASE hex
   */
  private generateSignature(
    params: Record<string, string | number>,
    apiPath: string,
  ): string {
    // Step 1: Sort all parameters alphabetically, excluding 'sign'
    const sortedKeys = Object.keys(params)
      .filter((key) => key !== "sign")
      .sort();

    // Step 2: Concatenate as key+value+key2+value2...
    const paramString = sortedKeys
      .map((key) => `${key}${String(params[key])}`)
      .join("");

    // Step 3: Prepend API path
    const preSignString = apiPath + paramString;

    // Signature source suppressed

    // Step 4: HMAC-SHA256
    const signature = crypto
      .createHmac("sha256", this.config.appSecret)
      .update(preSignString)
      .digest("hex")
      .toUpperCase(); // Step 5: UPPERCASE

    // Signature suppressed
    return signature;
  }

  /**
   * Override request method to add Lazada-specific signature for order endpoints
   */
  async request(
    endpoint: string,
    method: "GET" | "POST" | "PUT" | "DELETE" = "GET",
    params: Record<string, any> = {},
    data: Record<string, any> = {},
  ): Promise<any> {
    // For non-auth endpoints that need signature (order endpoints, etc)
    if (!endpoint.includes("/auth/")) {
      const timestamp = String(Math.floor(Date.now()));

      // Add required parameters
      params = {
        ...params,
        app_key: this.config.appKey,
        sign_method: "sha256",
        timestamp,
        access_token: this.config.accessToken,
      };

      // Generate signature with the actual API path
      const signature = this.generateSignature(params, endpoint);
      params.sign = signature;

      // Auth params added
    }

    return super.request(endpoint, method, params, data);
  }

  /**
   * Refresh Lazada access token
   * Endpoint: POST /auth/token/refresh
   * Per official Lazada SDK: https://github.com/branch8/lazada-open-platform-sdk
   * Required parameter: refresh_token
   *
   * IMPORTANT: Token refresh MUST use auth.lazada.com, not country-specific gateway
   */
  async refreshAccessToken(): Promise<boolean> {
    if (!this.config.refreshToken) {
      this.log.error("No refresh token available for Lazada");
      return false;
    }

    try {
      const timestamp = String(Math.floor(Date.now())); // milliseconds as STRING

      // Per official Lazada docs, only need these parameters
      const params: Record<string, string | number> = {
        app_key: this.config.appKey,
        sign_method: "sha256",
        timestamp,
        refresh_token: this.config.refreshToken,
      };

      // Generate signature with API path included
      const apiPath = "/auth/token/refresh";
      const signature = this.generateSignature(params, apiPath);
      params.sign = signature;

      // Build URL manually using authUrl (NOT baseUrl which is country-specific)
      const queryString = new URLSearchParams();
      for (const [key, value] of Object.entries(params)) {
        queryString.append(key, String(value));
      }
      const url = `${this.authUrl}${apiPath}?${queryString.toString()}`;

      // Log masked URL
      const maskedUrl = this.maskUrlForLogging(url);
      this.log.info(`POST ${maskedUrl}`);

      // Make direct HTTP request to auth endpoint
      const httpResponse = await this.httpClient({
        method: "POST",
        url,
        params: {},
        data: {},
      });

      const response = httpResponse.data;

      // Check response format and log errors
      if (!response) {
        this.log.error("Lazada token refresh failed: Empty response");
        return false;
      }

      if (response.code !== "0") {
        // Log detailed error from Lazada API
        this.log.error(
          `Lazada token refresh failed: code=${response.code}, message=${response.message || "Unknown error"}, request_id=${response.request_id || "N/A"}`,
        );
        return false;
      }

      if (!response.access_token) {
        this.log.error(
          "Lazada token refresh failed: No access_token in response",
        );
        return false;
      }

      const expiresIn = response.expires_in || 2592000; // 30 days default
      // Lazada refresh token is valid for same duration as access token
      const refreshExpiresIn = expiresIn; // 30 days

      // Update config and save to database
      await this.config.updateTokenInfo(
        response.access_token,
        response.refresh_token || this.config.refreshToken,
        expiresIn,
        refreshExpiresIn,
      );

      this.log.info(`Token refreshed successfully (expires in ${expiresIn}s)`);
      return true;
    } catch (error: any) {
      // Log detailed error including response data if available
      const status = error.response?.status;
      const responseData = error.response?.data;
      if (responseData) {
        this.log.error(
          `Lazada token refresh failed: HTTP ${status}, code=${responseData.code}, message=${responseData.message || error.message}, request_id=${responseData.request_id || "N/A"}`,
        );
      } else {
        this.log.error(`Lazada token refresh failed: ${error.message}`);
      }
      return false;
    }
  }
}
