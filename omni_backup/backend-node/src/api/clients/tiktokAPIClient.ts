import crypto from "crypto";
import { BaseAPIClient } from "./baseAPIClient";
import { TiktokConfigManager } from "../../config/managers/tiktokConfigManager";

/**
 * TikTok API Client
 * Mirrors TikTok SDK signature generation from utils/generate-sign.ts
 *
 * Signature Algorithm (from SDK):
 * 1. Extract all query params excluding 'sign' and 'access_token', sort alphabetically
 * 2. Concatenate params as: {key}{value}
 * 3. Append request body if not multipart/form-data
 * 4. Prepend API pathname: /api/path + paramString + bodyString
 * 5. Wrap with app_secret: app_secret + fullString + app_secret
 * 6. Generate HMAC-SHA256(app_secret, wrappedString)
 */
export class TiktokAPIClient extends BaseAPIClient {
  protected baseUrl = "https://open-api.tiktokglobalshop.com";
  protected config: TiktokConfigManager;

  constructor(config: TiktokConfigManager) {
    super(config);
    this.config = config;
  }

  /**
   * Generate TikTok signature using official SDK algorithm
   * Mirrors: tiktok_sdk/utils/generate-sign.ts
   */
  protected generateSignature(
    pathname: string,
    queryParams: Record<string, any>,
    requestBody: Record<string, any> = {},
  ): string {
    // Step 1: Extract query params excluding 'sign' and 'access_token'
    const excludeKeys = ["access_token", "sign"];
    const sortedParams = Object.keys(queryParams)
      .filter((key) => !excludeKeys.includes(key))
      .sort()
      .map((key) => ({ key, value: queryParams[key] }));

    // Step 2: Concatenate params as {key}{value}
    const paramString = sortedParams
      .map(({ key, value }) => `${key}${value}`)
      .join("");

    // Step 3: Build sign string
    let signString = paramString;

    // Step 4: Append request body if not empty
    if (Object.keys(requestBody).length > 0) {
      const bodyString = JSON.stringify(requestBody);
      signString += bodyString;
    }

    // Step 5: Prepend pathname
    signString = `${pathname}${paramString}`;
    if (Object.keys(requestBody).length > 0) {
      signString += JSON.stringify(requestBody);
    }

    // Step 6: Wrap with app_secret and generate HMAC-SHA256
    const wrappedString = `${this.config.appSecret}${signString}${this.config.appSecret}`;
    const hmac = crypto.createHmac("sha256", this.config.appSecret);
    hmac.update(wrappedString);
    const signature = hmac.digest("hex");

    return signature;
  }

  /**
   * Override makeRequest to add TikTok signature to headers
   * TikTok requires signature and access_token in different places
   */
  protected async makeRequest(
    endpoint: string,
    method: "GET" | "POST" | "PUT" | "DELETE" = "GET",
    params: Record<string, any> = {},
    data: Record<string, any> = {},
  ): Promise<any> {
    let url = `${this.baseUrl}${endpoint}`;

    // Generate signature if this is not a token endpoint
    if (!endpoint.includes("/token/")) {
      // Add required TikTok params BEFORE generating signature
      // TikTok requires timestamp in SECONDS (not milliseconds)
      const timestamp = Math.floor(Date.now() / 1000);
      params.app_key = this.config.appKey;
      params.timestamp = timestamp;

      // Some endpoints don't require shop_cipher (e.g., images/upload, files/upload)
      // Per TikTok SDK: image upload is seller-level, not shop-level
      const noShopCipherEndpoints = ["/images/upload", "/files/upload"];
      const requiresShopCipher = !noShopCipherEndpoints.some((e) =>
        endpoint.includes(e),
      );

      if (requiresShopCipher && this.config.shopCipher) {
        params.shop_cipher = this.config.shopCipher;
      }
      // Add shop_id if available and endpoint requires it
      if (requiresShopCipher && this.config.shopId) {
        params.shop_id = this.config.shopId;
      }
      // Add access_token to query params (TikTok requires it in both header and query)
      if (this.config.accessToken) {
        params.access_token = this.config.accessToken;
      }
      // NOTE: TikTok API version is embedded in the URL path (e.g., /product/202309/...)
      // Do NOT add 'version' as a query parameter - it causes "Invalid API version" error

      // Generate signature with all params that will be sent
      const signature = this.generateSignature(endpoint, params, data);
      params.sign = signature;

      // Signature added
    }

    // Build URL with query params
    if (Object.keys(params).length > 0) {
      const queryString = new URLSearchParams();
      for (const [key, value] of Object.entries(params)) {
        queryString.append(key, String(value));
      }
      url = `${url}?${queryString.toString()}`;
    }

    try {
      // Log URL but mask sensitive parameters
      const maskedUrl = this.maskUrlForLogging(url);
      this.log.info(`${method} ${maskedUrl}`);

      // Prepare axios config with custom headers for TikTok
      const axiosConfig: any = {
        method,
        url,
        params: {}, // Don't use axios params - we built URL manually
        headers: {
          "Content-Type": "application/json",
          "x-tts-access-token": this.config.accessToken || "", // Add access token as header
        },
      };

      if (method !== "GET" && Object.keys(data).length > 0) {
        axiosConfig.data = data;
        // Only log request method, not the full body
        this.log.info(`Request sent`);
      }

      const response = await this.httpClient(axiosConfig);

      this.log.info(`✅ Response status: ${response.status}`);
      // Only log success code, not response data
      return response.data;
    } catch (error: any) {
      const status = error.response?.status;
      const errorMsg = error.response?.data?.message || error.message;

      this.log.error(`HTTP ${status}: ${errorMsg}`);

      throw error;
    }
  }

  /**
   * Refresh TikTok access token
   * Endpoint: https://auth.tiktok-shops.com/api/v2/token/refresh
   * Method: GET with query params (app_key, app_secret, grant_type, refresh_token)
   */
  async refreshAccessToken(): Promise<boolean> {
    if (!this.config.refreshToken) {
      return false;
    }

    try {
      // TikTok auth endpoint - simple GET request
      const authUrl = "https://auth.tiktok-shops.com/api/v2/token/refresh";
      const params = {
        app_key: this.config.appKey,
        app_secret: this.config.appSecret,
        grant_type: "refresh_token",
        refresh_token: this.config.refreshToken,
      };

      // Use axios directly for token refresh (no signature needed for this endpoint)
      const response = await this.httpClient.get(authUrl, { params });
      const responseData = response?.data;

      // Check TikTok response format: code=0 means success
      if (!responseData || responseData.code !== 0) {
        return false;
      }

      if (!responseData?.data?.access_token) {
        return false;
      }

      const tokenData = responseData.data;
      const accessTokenExpireTimestamp =
        tokenData.access_token_expire_in || 86400; // seconds

      // TikTok returns Unix timestamp (seconds) not "expires in X seconds"
      const nowSeconds = Math.floor(Date.now() / 1000);
      let accessTokenExpiryMs: number;

      if (accessTokenExpireTimestamp > nowSeconds) {
        // It's a Unix timestamp (seconds) - convert to milliseconds
        accessTokenExpiryMs = accessTokenExpireTimestamp * 1000;
      } else {
        // It's relative seconds - add to current time
        accessTokenExpiryMs = Date.now() + accessTokenExpireTimestamp * 1000;
      }

      // For refresh token: DO NOT use refresh_token_expire_in from API (unreliable)
      // Instead, use existing value from config or set reasonable default (90 days)
      let refreshTokenExpiryMs: number;

      // Check if existing value is reasonable (not more than 10 years in future)
      const maxReasonableExpiry = Date.now() + 10 * 365 * 24 * 60 * 60 * 1000; // 10 years

      if (
        this.config.refreshTokenExpiry &&
        this.config.refreshTokenExpiry > Date.now() &&
        this.config.refreshTokenExpiry < maxReasonableExpiry
      ) {
        // Use existing refresh token expiry if it's reasonable and still valid
        refreshTokenExpiryMs = this.config.refreshTokenExpiry;
      } else {
        // Use default: 90 days from now (reasonable for refresh tokens)
        refreshTokenExpiryMs = Date.now() + 90 * 24 * 60 * 60 * 1000;
      }

      // Update all token info in memory FIRST
      this.config.accessToken = tokenData.access_token;
      this.config.refreshToken =
        tokenData.refresh_token || this.config.refreshToken;
      this.config.tokenExpiry = accessTokenExpiryMs;
      this.config.refreshTokenExpiry = refreshTokenExpiryMs;

      // Update config map with all values
      await this.config.setConfig("accessToken", tokenData.access_token);
      await this.config.setConfig("refreshToken", this.config.refreshToken);
      await this.config.setConfig("tokenExpiry", String(accessTokenExpiryMs));
      await this.config.setConfig(
        "refreshTokenExpiry",
        String(refreshTokenExpiryMs),
      );

      // Save to database ONCE with all values
      await this.config.saveConfig();

      return true;
    } catch (error: any) {
      this.log.error(`TikTok token refresh failed: ${error.message}`);
      return false;
    }
  }

  /**
   * Search orders by status and date range
   * API: POST /order/202309/orders/search
   */
  async searchOrders(
    body: Record<string, unknown>,
    params: Record<string, unknown> = {},
  ): Promise<any> {
    return this.makeRequest(
      "/order/202309/orders/search",
      "POST",
      params,
      body,
    );
  }

  /**
   * Get transaction details for an order
   * API: GET /finance/202501/orders/{order_id}/statement_transactions
   */
  async getOrderTransactions(orderId: string): Promise<any> {
    return this.makeRequest(
      `/finance/202501/orders/${orderId}/statement_transactions`,
      "GET",
    );
  }

  /**
   * Upload image to TikTok using multipart/form-data
   * API: POST /product/202309/images/upload
   * This endpoint doesn't require shop_cipher - handled in makeRequest
   */
  async uploadImage(formData: any): Promise<any> {
    const endpoint = "/product/202309/images/upload";
    const timestamp = Math.floor(Date.now() / 1000);

    // Build params for signature (without shop_cipher for image upload)
    const params: Record<string, any> = {
      app_key: this.config.appKey,
      timestamp,
    };

    if (this.config.accessToken) {
      params.access_token = this.config.accessToken;
    }

    // Generate signature
    const signature = this.generateSignature(endpoint, params, {});
    params.sign = signature;

    // Build URL with params
    const queryString = new URLSearchParams();
    for (const [key, value] of Object.entries(params)) {
      queryString.append(key, String(value));
    }
    const url = `${this.baseUrl}${endpoint}?${queryString.toString()}`;

    try {
      const maskedUrl = this.maskUrlForLogging(url);
      this.log.info(`POST (multipart) ${maskedUrl}`);

      const response = await this.httpClient({
        method: "POST",
        url,
        data: formData,
        headers: {
          ...formData.getHeaders(),
          "x-tts-access-token": this.config.accessToken || "",
        },
      });

      this.log.info(`✅ Response status: ${response.status}`);
      return response.data;
    } catch (error: any) {
      const status = error.response?.status;
      const errorMsg = error.response?.data?.message || error.message;
      this.log.error(`HTTP ${status}: ${errorMsg}`);
      throw error;
    }
  }
}
