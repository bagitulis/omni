import { TiktokAPIClient } from "../api/clients/tiktokAPIClient";
import { Logger } from "winston";
import { getLogger } from "../utils/logger";

interface SearchParams {
  status?: string;
  pageSize?: number;
  limit?: number;
}

export class TiktokProductAPIService {
  private apiClient: TiktokAPIClient;
  private logger: Logger;

  constructor(apiClient: TiktokAPIClient) {
    this.apiClient = apiClient;
    this.logger = getLogger("TiktokProductAPIService");
  }

  async searchProducts(params: SearchParams): Promise<any> {
    const { status = "ACTIVATE", pageSize = 100, limit = undefined } = params;

    this.logger.info(
      `🔄 Searching TikTok products | Status: ${status}, Page Size: ${pageSize}${limit ? `, Limit: ${limit}` : ""}`
    );

    const endpoint = "/product/202502/products/search";
    let pageToken: string | undefined;
    let pageCount = 0;
    let totalProducts = 0;
    const allProducts: any[] = [];

    // eslint-disable-next-line no-constant-condition
    while (true) {
      // TikTok v202502 requires status in POST body, pagination in query params
      const queryParams: any = { page_size: pageSize };
      if (pageToken) queryParams.page_token = pageToken;
      const bodyParams: any = { status };

      this.logger.info(`   📤 Fetching page ${pageCount + 1} with body: ${JSON.stringify(bodyParams)}`);

      const result = await this.apiClient.request(
        endpoint,
        "POST",
        queryParams,
        bodyParams
      );

      if (!result?.data?.products) break;

      const pageProducts = result.data.products;
      allProducts.push(...pageProducts);
      totalProducts += pageProducts.length;
      pageCount++;

      this.logger.debug(
        `   Page ${pageCount}: ${pageProducts.length} products (total: ${totalProducts})`
      );

      pageToken = result.data.next_page_token;
      if (!pageToken) {
        this.logger.info(`   ✅ Reached last page (page ${pageCount})`);
        break;
      }

      if (limit && totalProducts >= limit) {
        this.logger.info(`   ✅ Reached limit of ${limit} products`);
        allProducts.splice(limit);
        totalProducts = limit;
        break;
      }
    }

    this.logger.info(
      `✅ Search completed | Pages: ${pageCount}, Total: ${totalProducts}`
    );

    return { products: allProducts, pageCount, totalProducts };
  }

  async getAllProducts(params: {
    status?: string;
    pageSize?: number;
    maxPages?: number;
  }): Promise<any> {
    const {
      status = "ACTIVATE",
      pageSize = 100,
      maxPages = undefined,
    } = params;

    this.logger.info(
      `🔄 Fetching all TikTok products | Status: ${status}, Page Size: ${pageSize}`
    );

    const endpoint = "/product/202502/products/search";
    let pageToken: string | undefined;
    let pageCount = 0;
    let totalProducts = 0;
    const allProducts: any[] = [];

    while (
      pageCount === 0 ||
      (pageToken && (!maxPages || pageCount < maxPages))
    ) {
      // TikTok v202502 requires status in POST body, pagination in query params
      const queryParams: any = { page_size: pageSize };
      if (pageToken) queryParams.page_token = pageToken;
      const bodyParams: any = { status };

      const result = await this.apiClient.request(endpoint, "POST", queryParams, bodyParams);

      if (!result?.data?.products) break;

      allProducts.push(...result.data.products);
      totalProducts += result.data.products.length;
      pageToken = result.data.next_page_token;
      pageCount++;

      this.logger.debug(
        `   Page ${pageCount}: ${result.data.products.length} products`
      );

      if (!pageToken) break;
    }

    this.logger.info(
      `✅ Fetched all products | Pages: ${pageCount}, Total: ${totalProducts}`
    );

    return { products: allProducts, pageCount, totalProducts };
  }

  async getProductDetail(productId: string): Promise<any> {
    this.logger.info(
      `🔄 Fetching TikTok product detail | Product ID: ${productId}`
    );

    const endpoint = `/product/202309/products/${productId}`;
    const result = await this.apiClient.request(endpoint, "GET", {});

    if (!result?.data) {
      throw new Error("No product data in response");
    }

    this.logger.info(`✅ Retrieved product detail | Product ID: ${productId}`);
    return result.data;
  }
}
