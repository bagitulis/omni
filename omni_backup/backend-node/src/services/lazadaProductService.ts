import { PrismaClient } from "@prisma/client";
import { LazadaAPIClient } from "../api/clients/lazadaAPIClient";
import { LazadaConfigManager } from "../config/managers/lazadaConfigManager";
import { ProductService } from "./productService";
import { LazadaProductDBService } from "./lazada/lazadaProductDBService";
import { LazadaProductValidator } from "./lazada/lazadaProductValidator";

/**
 * Lazada Product Service
 * Coordinates product operations between API and database
 * Single Responsibility: Orchestrate Lazada product operations
 */
export class LazadaProductService extends ProductService {
  private apiClient: LazadaAPIClient;
  private dbService: LazadaProductDBService;

  constructor(
    prisma: PrismaClient,
    apiClient: LazadaAPIClient,
    _configManager: LazadaConfigManager,
    tenantId: string = "default"
  ) {
    super(prisma);
    this.apiClient = apiClient;
    this.dbService = new LazadaProductDBService(prisma, tenantId);
  }

  /**
   * Get products from Lazada API
   * GET /api/lazada/products
   */
  async getProducts(params: {
    filter?: string;
    offset?: number;
    limit?: number;
    syncToDb?: boolean;
  }): Promise<any> {
    try {
      const {
        offset = 0,
        limit = 50,
        filter = "live",
        syncToDb = true,
      } = params;

      const { offset: validOffset, limit: validLimit } =
        LazadaProductValidator.validatePagination(offset, limit);
      const validFilter = LazadaProductValidator.validateFilter(filter);

      this.logger.info(
        `🔄 Fetching Lazada products | Filter: ${validFilter}, Offset: ${validOffset}, Limit: ${validLimit}`
      );

      // Call Lazada API
      const endpoint = "/products/get";
      const apiParams: any = {
        offset: validOffset,
        limit: validLimit,
      };

      if (validFilter && validFilter !== "all") {
        apiParams.filter = validFilter;
      }

      const result = await this.apiClient.request(endpoint, "GET", apiParams);

      // Check if API returned error response
      if (result?.code && result?.code !== "0") {
        this.logger.error(
          `Lazada API Error (${result.code}): ${result.message}`
        );
        return {
          success: false,
          error: result.message || "Lazada API error",
          retrieved: 0,
          total_products: 0,
        };
      }

      // Check response structure
      if (!result?.data?.products) {
        this.logger.error(
          `Response structure error. Result: ${JSON.stringify(result).substring(
            0,
            200
          )}`
        );
        return {
          success: false,
          error: "No products found in response",
          retrieved: 0,
          total_products: 0,
        };
      }

      this.logger.info(
        `✅ Retrieved ${result.data.products.length} products | Total: ${result.data.total}`
      );

      // Save to database if requested
      if (syncToDb && result.data.products.length > 0) {
        await this.dbService.saveProducts(result.data.products, true);
      }

      return {
        success: true,
        products: result.data.products,
        detail_saved: syncToDb ? result.data.products.length : 0,
        total: result.data.total,
        message: `Fetched ${result.data.products.length} products${
          syncToDb ? " and saved to database" : ""
        }`,
      };
    } catch (error: any) {
      this.logger.error(`❌ Error fetching Lazada products: ${error.message}`);
      return {
        success: false,
        error: error.message,
        products: [],
        detail_saved: 0,
        total: 0,
      };
    }
  }

  /**
   * Get master products from database
   * GET /api/lazada/products/db
   */
  async getMasterProductsFromDb(params: {
    offset?: number;
    limit?: number;
  }): Promise<any> {
    const { offset = 0, limit = 100 } = params;
    const { offset: validOffset, limit: validLimit } =
      LazadaProductValidator.validatePagination(offset, limit);

    this.logger.info(
      `🔄 Retrieving Lazada products from DB | Limit: ${validLimit}, Offset: ${validOffset}`
    );

    return this.dbService.getMasterProducts(validOffset, validLimit);
  }

  /**
   * Get product list from database (quick lookup)
   */
  async getProductListFromDb(params: {
    offset?: number;
    limit?: number;
  }): Promise<any> {
    const { offset = 0, limit = 100 } = params;
    const { offset: validOffset, limit: validLimit } =
      LazadaProductValidator.validatePagination(offset, limit);

    this.logger.info(
      `🔄 Retrieving Lazada product list from DB | Limit: ${validLimit}, Offset: ${validOffset}`
    );

    return this.dbService.getProductList(validOffset, validLimit);
  }

  /**
   * Get product detail by Item ID
   */
  async getProductDetail(params: { itemId: string }): Promise<any> {
    const { itemId } = params;

    this.logger.info(
      `🔄 Retrieving Lazada product detail for Item ID: ${itemId}`
    );

    return this.dbService.getProductDetail(itemId);
  }
}
