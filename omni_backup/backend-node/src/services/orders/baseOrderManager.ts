/**
 * Base Order Manager
 * Abstract base class for all platform-specific order managers
 * Mirrors Python backend's BaseOrderManager
 */

import { BaseAPIClient } from "../../api/clients/baseAPIClient";

export abstract class BaseOrderManager {
  protected api: BaseAPIClient;
  protected logger: any;
  protected MAX_DAYS = 15;
  protected BATCH_SIZE = 50;
  protected RATE_LIMIT_DELAY = 500; // milliseconds

  constructor(apiClient: BaseAPIClient) {
    this.api = apiClient;
    this.logger = {
      info: (msg: string) => console.log(`[OrderManager] ℹ️  ${msg}`),
      debug: (_msg: string) => {}, // Suppress debug logs
      error: (msg: string) => console.error(`[OrderManager] ❌ ${msg}`),
      warn: (msg: string) => console.warn(`[OrderManager] ⚠️  ${msg}`),
    };
  }

  /**
   * Get time range for API requests
   */
  protected getTimeRange(days: number): [number, number] {
    days = Math.min(days, this.MAX_DAYS);
    const timeFrom = Math.floor(Date.now() / 1000) - days * 86400;
    const timeTo = Math.floor(Date.now() / 1000);
    return [timeFrom, timeTo];
  }

  /**
   * Sleep for rate limiting
   */
  protected async sleep(ms: number): Promise<void> {
    return new Promise((resolve) => setTimeout(resolve, ms));
  }

  /**
   * Fetch paginated data from API
   */
  protected async fetchPaginatedData(
    endpoint: string,
    baseParams: Record<string, any>,
    dataKey: string
  ): Promise<any[]> {
    const allData: any[] = [];
    let nextCursor: string | null = null;
    let pageCount = 0;

    // eslint-disable-next-line no-constant-condition
    while (true) {
      pageCount++;
      const params = { ...baseParams };

      if (nextCursor) {
        params.cursor = nextCursor;
      }

      try {
        const response = await this.api.request(endpoint, "GET", params);

        if (!response || !response.response) {
          break;
        }

        const pageData = response.response[dataKey] || [];
        if (!pageData || pageData.length === 0) {
          break;
        }

        allData.push(...pageData);
        nextCursor = response.response.next_cursor || null;
        const hasMore = response.response.more || false;

        if (!hasMore || !nextCursor) {
          break;
        }

        await this.sleep(this.RATE_LIMIT_DELAY);
      } catch (error) {
        this.logger.error(`Page ${pageCount} failed: ${error}`);
        break;
      }
    }

    this.logger.info(
      `Fetched ${allData.length} items in ${pageCount} pages from ${endpoint}`
    );

    return allData;
  }

  /**
   * Abstract methods - implement in subclasses
   */
  abstract getOrderList(status: string, days?: number): Promise<any[]>;
  abstract getOrderDetails(orderIds: string[]): Promise<any[]>;
  abstract formatOrdersForExport(orders: any[]): Record<string, any>[];
}
