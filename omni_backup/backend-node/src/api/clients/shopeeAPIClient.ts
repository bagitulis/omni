import crypto from "crypto";
import { BaseAPIClient } from "./baseAPIClient";
import { ShopeeConfigManager } from "../../config/managers/shopeeConfigManager";

/**
 * Shopee API Client
 * Handles Shopee-specific API operations and token refresh
 */
export class ShopeeAPIClient extends BaseAPIClient {
  protected baseUrl = "https://partner.shopeemobile.com"; // ✅ CORRECT endpoint
  protected config: ShopeeConfigManager;

  constructor(config: ShopeeConfigManager) {
    super(config);
    this.config = config;
  }

  /**
   * Generate signature for Shopee API
   * Format: SHA256(partner_id + endpoint + timestamp + extra_string)
   * extra_string = access_token + shop_id (for regular requests)
   * extra_string = "" (for auth endpoints like token refresh)
   * IMPORTANT: Use partnerKey.encode() for HMAC (matches Python backend)
   */
  private generateSignature(
    endpoint: string,
    extraString: string = "",
  ): string {
    const timestamp = Math.floor(Date.now() / 1000);

    // Signature format: partner_id + endpoint + timestamp + extra_string
    const baseString = `${this.config.partnerId}${endpoint}${timestamp}${extraString}`;

    // Signature generation suppressed

    // Use partnerKey as-is for HMAC (just like Python does .encode())
    const signature = crypto
      .createHmac("sha256", this.config.partnerKey)
      .update(baseString)
      .digest("hex");

    // Signature suppressed
    return signature;
  }

  /**
   * Override request method to add Shopee-specific signature
   * For regular API calls (not auth), we need to add partner_id, timestamp, and sign to query params
   */
  async request(
    endpoint: string,
    method: "GET" | "POST" | "PUT" | "DELETE" = "GET",
    params: Record<string, any> = {},
    data: Record<string, any> = {},
  ): Promise<any> {
    // For non-auth endpoints, add signature
    if (!endpoint.includes("/auth/")) {
      const timestamp = Math.floor(Date.now() / 1000);
      const extraString = `${this.config.accessToken}${this.config.shopId}`;
      const signature = this.generateSignature(endpoint, extraString);

      // Add signature parameters to params
      params = {
        ...params,
        partner_id: this.config.partnerId,
        shop_id: this.config.shopId,
        timestamp,
        sign: signature,
        access_token: this.config.accessToken,
      };

      // Signature params added
    }

    return super.request(endpoint, method, params, data);
  }

  /**
   * Refresh Shopee access token
   * Endpoint: POST /api/v2/auth/access_token/get
   * Payload: { refresh_token, partner_id, shop_id }
   */
  async refreshAccessToken(): Promise<boolean> {
    if (!this.config.refreshToken) {
      return false;
    }

    try {
      const endpoint = "/api/v2/auth/access_token/get";
      const timestamp = Math.floor(Date.now() / 1000);

      // For auth endpoints, extra_string is empty
      const signature = this.generateSignature(endpoint, "");

      // Query params
      const params = {
        partner_id: this.config.partnerId,
        timestamp,
        sign: signature,
      };

      // Request body
      const payload = {
        refresh_token: this.config.refreshToken,
        partner_id: this.config.partnerId,
        shop_id: this.config.shopId,
      };

      const response = await this.request(endpoint, "POST", params, payload);

      // Check for errors
      if (!response || response.error) {
        return false;
      }

      if (!response?.access_token) {
        return false;
      }

      const expiresIn = response.expire_in || 14400; // 4 hours default
      // Shopee refresh token is valid for 30 days from refresh
      const refreshExpiresIn = 30 * 24 * 60 * 60; // 30 days in seconds

      // Update config and save to database
      await this.config.updateTokenInfo(
        response.access_token,
        response.refresh_token || this.config.refreshToken,
        expiresIn,
        refreshExpiresIn,
      );

      return true;
    } catch (error: any) {
      this.log.error(`Shopee token refresh failed: ${error.message}`);
      return false;
    }
  }

  /**
   * Search package list with filters (new API - recommended)
   * POST /api/v2/order/search_package_list
   * @param packageStatus 0=All, 1=Pending, 2=ToProcess, 3=Processed
   */
  async searchPackageList(
    packageStatus: number = 3,
    cursor: string = "",
    pageSize: number = 100,
  ): Promise<{
    packages: Array<{
      order_sn: string;
      package_number: string;
      logistics_channel_id: number;
      is_shipment_arranged: boolean;
    }>;
    nextCursor: string;
    more: boolean;
    totalCount: number;
  }> {
    const requestBody = {
      filter: {
        package_status: packageStatus,
        fulfillment_type: 2, // Seller fulfilled
      },
      pagination: {
        page_size: pageSize,
        cursor: cursor,
      },
      sort: {
        sort_type: 1, // ShipByDate
        ascending: true,
      },
    };

    const response = await this.request(
      "/api/v2/order/search_package_list",
      "POST",
      {},
      requestBody,
    );

    if (!response?.response) {
      return { packages: [], nextCursor: "", more: false, totalCount: 0 };
    }

    const packagesList = response.response.packages_list || [];
    return {
      packages: packagesList.map((p: any) => ({
        order_sn: p.order_sn,
        package_number: p.package_number || "",
        logistics_channel_id: p.logistics_channel_id || 0,
        is_shipment_arranged: p.is_shipment_arranged || false,
      })),
      nextCursor: response.response.pagination?.next_cursor || "",
      more: response.response.pagination?.more || false,
      totalCount: response.response.pagination?.total_count || 0,
    };
  }

  /**
   * @deprecated Use searchPackageList instead - get_shipment_list only returns READY_TO_SHIP orders
   * Get shipment list for processed orders
   * Returns order_sn and package_number for tracking lookup
   */
  async getShipmentList(
    cursor: string = "",
    pageSize: number = 100,
  ): Promise<{
    orders: Array<{ order_sn: string; package_number: string }>;
    nextCursor: string;
    more: boolean;
  }> {
    const response = await this.request(
      "/api/v2/order/get_shipment_list",
      "GET",
      {
        page_size: pageSize,
        cursor: cursor,
      },
    );

    if (!response?.response) {
      return { orders: [], nextCursor: "", more: false };
    }

    const orderList = response.response.order_list || [];
    return {
      orders: orderList.map((o: any) => ({
        order_sn: o.order_sn,
        package_number: o.package_number || "",
      })),
      nextCursor: response.response.next_cursor || "",
      more: response.response.more || false,
    };
  }

  /**
   * Get package details including tracking number, carrier, and item details
   * Accepts up to 50 package numbers at once
   */
  async getPackageDetail(packageNumbers: string[]): Promise<
    Array<{
      order_sn: string;
      package_number: string;
      shipping_carrier: string;
      tracking_number: string;
      items: Array<{
        item_id: number;
        model_id: number;
        item_sku: string;
        model_sku: string;
        model_quantity: number;
      }>;
    }>
  > {
    if (packageNumbers.length === 0) return [];

    const response = await this.request(
      "/api/v2/order/get_package_detail",
      "GET",
      {
        package_number_list: packageNumbers.slice(0, 50).join(","),
      },
    );

    if (!response?.response?.package_list) {
      return [];
    }

    return response.response.package_list.map((pkg: any) => ({
      order_sn: pkg.order_sn || "",
      package_number: pkg.package_number || "",
      shipping_carrier: pkg.shipping_carrier || "",
      tracking_number: pkg.tracking_number || "",
      items: (pkg.item_list || []).map((item: any) => ({
        item_id: item.item_id || 0,
        model_id: item.model_id || 0,
        item_sku: item.item_sku || "",
        model_sku: item.model_sku || "",
        model_quantity: item.model_quantity || 1,
      })),
    }));
  }
}
