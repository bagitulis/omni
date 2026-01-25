import { PrismaClient } from "@prisma/client";
import { ShopeeAPIClient } from "../api/clients/shopeeAPIClient";
import { ShopeeConfigManager } from "../config/managers/shopeeConfigManager";
import { ProductService } from "./productService";
import { ShopeeProductAPIService } from "./shopeeProductAPIService";
import { ShopeeProductDBService } from "./shopeeProductDBService";
import { ShopeeProductQueryService } from "./shopeeProductQueryService";

/**
 * Shopee Product Service - Orchestrator
 * Orchestrates API calls and database operations for Shopee products
 */
export class ShopeeProductService extends ProductService {
  private apiService: ShopeeProductAPIService;
  private dbService: ShopeeProductDBService;
  private queryService: ShopeeProductQueryService;

  constructor(
    prisma: PrismaClient,
    apiClient: ShopeeAPIClient,
    _configManager: ShopeeConfigManager
  ) {
    super(prisma);
    this.apiService = new ShopeeProductAPIService(apiClient);
    this.dbService = new ShopeeProductDBService(prisma, this.logger);
    this.queryService = new ShopeeProductQueryService(prisma, this.logger);
  }

  /**
   * Get products from Shopee API and optionally save to database
   * Now supports automatic pagination to get ALL products
   */
  async getProducts(params: {
    itemStatus?: string;
    offset?: number;
    limit?: number;
    syncToDb?: boolean;
    fetchAll?: boolean;
  }): Promise<any> {
    try {
      const {
        itemStatus = "NORMAL",
        offset = 0,
        limit = 50,
        syncToDb = true,
        fetchAll = true, // Default to fetch all products
      } = params;

      this.logger.info(
        `🔄 Fetching Shopee products | Status: ${itemStatus}, FetchAll: ${fetchAll}`
      );

      let mergedProducts: any[];
      let totalCount: number;
      let pageCount: number = 1;

      if (fetchAll) {
        // Use pagination to get ALL products
        const result = await this.apiService.fetchAllProducts(
          itemStatus,
          limit > 100 ? undefined : undefined
        );
        mergedProducts = result.products;
        totalCount = result.totalCount;
        pageCount = result.pageCount;
      } else {
        // Legacy: single page fetch
        mergedProducts = await this.apiService.fetchAndMergeProducts(
          itemStatus,
          offset,
          limit
        );
        totalCount = mergedProducts.length;
      }

      if (!mergedProducts || mergedProducts.length === 0) {
        return {
          success: true,
          products: [],
          total: 0,
          message: "No products found",
        };
      }

      // Save to database if requested
      if (syncToDb) {
        await this.dbService.saveProducts(mergedProducts, true);
      }

      return {
        success: true,
        products: mergedProducts,
        total: mergedProducts.length,
        totalCount,
        pageCount,
        message: `Fetched ${
          mergedProducts.length
        } products (${pageCount} pages)${
          syncToDb ? " and saved to database" : ""
        }`,
      };
    } catch (error: any) {
      this.logger.error(`❌ Error fetching Shopee products: ${error.message}`);
      return {
        success: false,
        error: error.message,
        products: [],
        total: 0,
      };
    }
  }

  /**
   * Save Shopee products to database
   */
  async saveProductsToDb(
    products: any[],
    cleanOld: boolean = false
  ): Promise<void> {
    return this.dbService.saveProducts(products, cleanOld);
  }

  /**
   * Get master products from database
   */
  async getMasterProductsFromDb(params: {
    offset?: number;
    limit?: number;
  }): Promise<any> {
    return this.queryService.getMasterProductsFromDb(params);
  }

  /**
   * Get master product list (flat structure)
   */
  async getMasterProductsList(params: {
    limit?: number;
    offset?: number;
  }): Promise<any> {
    return this.queryService.getMasterProductsList(params);
  }

  /**
   * Get master products in formatted table structure
   */
  async getMasterProductsFromDbFormatted(): Promise<any> {
    return this.queryService.getMasterProductsFormatted();
  }

  /**
   * Get master product statistics
   */
  async getMasterProductsStats(): Promise<any> {
    return this.queryService.getMasterProductsStats();
  }
}
