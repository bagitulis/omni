/**
 * TikTok Product Service
 * SRP: Coordinate TikTok product operations between API and DB
 */

import { PrismaClient } from "@prisma/client";
import { TiktokAPIClient } from "../api/clients/tiktokAPIClient";
import { TiktokConfigManager } from "../config/managers/tiktokConfigManager";
import { ProductService } from "./productService";
import { TiktokProductAPIService } from "./tiktokProductAPIService";
import { TiktokProductDBService } from "./tiktokProductDBService";

export class TiktokProductService extends ProductService {
  private apiService: TiktokProductAPIService;
  private dbService: TiktokProductDBService;

  constructor(prisma: PrismaClient, apiClient: TiktokAPIClient, _configManager: TiktokConfigManager) {
    super(prisma);
    this.apiService = new TiktokProductAPIService(apiClient);
    this.dbService = new TiktokProductDBService(prisma);
  }

  async getProducts(params: any): Promise<any> {
    return this.searchProducts(params);
  }

  async searchProducts(params: { status?: string; pageSize?: number; limit?: number; pageToken?: string; syncToDb?: boolean }): Promise<any> {
    try {
      const { status = "ACTIVATE", pageSize = 100, limit = undefined, syncToDb = true } = params;
      const result = await this.apiService.searchProducts({ status, pageSize, limit });

      if (syncToDb && result.products.length > 0) {
        await this.dbService.saveProducts(result.products, true);
      }

      return { success: true, data: { products: result.products, next_page_token: undefined }, pageCount: result.pageCount, totalProducts: result.totalProducts, sync_result: syncToDb ? "saved" : "skipped" };
    } catch (error: any) {
      this.logger.error(`❌ Error searching TikTok products: ${error.message}`);
      return { success: false, error: error.message, data: { products: [], next_page_token: undefined } };
    }
  }

  async getAllProducts(params: { status?: string; pageSize?: number; maxPages?: number; syncToDb?: boolean }): Promise<any> {
    try {
      const { status = "ACTIVATE", pageSize = 100, maxPages = undefined, syncToDb = true } = params;
      const result = await this.apiService.getAllProducts({ status, pageSize, maxPages });

      if (syncToDb && result.products.length > 0) {
        await this.dbService.saveProducts(result.products, true);
      }

      return { success: true, total_products: result.totalProducts, total_pages: result.pageCount, data: result.products, sync_result: syncToDb ? "saved" : "skipped" };
    } catch (error: any) {
      this.logger.error(`❌ Error fetching all TikTok products: ${error.message}`);
      return { success: false, error: error.message, total_products: 0, total_pages: 0, data: [] };
    }
  }

  async getProductDetail(params: { productId: string; syncToDb?: boolean }): Promise<any> {
    try {
      const { productId, syncToDb = true } = params;
      const product = await this.apiService.getProductDetail(productId);

      if (syncToDb) await this.dbService.saveProducts([product]);

      return { success: true, data: product, sync_result: syncToDb ? "saved" : "skipped" };
    } catch (error: any) {
      this.logger.error(`❌ Error fetching product detail: ${error.message}`);
      return { success: false, error: error.message };
    }
  }

  async getProductListFromDb(params: { offset?: number; limit?: number }): Promise<any> {
    try {
      const { offset = 0, limit = 100 } = params;
      const result = await this.dbService.getProductList(offset, limit);
      return { success: true, count: result.products.length, total: result.total, offset: result.offset, limit: result.limit, data: result.products };
    } catch (error: any) {
      this.logger.error(`❌ Error retrieving products from DB: ${error.message}`);
      return { success: false, error: error.message, count: 0, total: 0, data: [] };
    }
  }

  async getMasterProductsFromDb(params: { offset?: number; limit?: number }): Promise<any> {
    try {
      const { offset = 0, limit = 100000 } = params;
      const result = await this.dbService.getMasterProducts(offset, limit);
      return { success: true, products: result.flattened, total: result.total, offset: result.offset, limit: result.limit, count: result.flattened.length };
    } catch (error: any) {
      this.logger.error(`❌ Error getting master products from DB: ${error.message}`);
      return { success: false, error: error.message, products: [], total: 0 };
    }
  }

  async getProductByIdFromDb(productId: string): Promise<any> {
    try {
      const product = await this.dbService.getProductById(productId);
      return { success: true, data: product };
    } catch (error: any) {
      this.logger.error(`❌ Error retrieving product by ID from DB: ${error.message}`);
      return { success: false, error: error.message };
    }
  }

  async searchProductsInDb(params: { query: string; field?: string; offset?: number; limit?: number }): Promise<any> {
    try {
      const { query, offset = 0, limit = 100 } = params;
      if (!query) return { success: false, error: "Search query required" };

      const result = await this.dbService.searchProducts(query, offset, limit);
      return { success: true, count: result.products.length, total: result.total, offset: result.offset, limit: result.limit, data: result.products, search_query: query };
    } catch (error: any) {
      this.logger.error(`❌ Error searching products in DB: ${error.message}`);
      return { success: false, error: error.message, count: 0, total: 0, data: [] };
    }
  }

  async getProductsByStatusFromDb(params: { status: string; offset?: number; limit?: number }): Promise<any> {
    try {
      const { status, offset = 0, limit = 100 } = params;
      const result = await this.dbService.getProductsByStatus(status, offset, limit);
      return { success: true, count: result.products.length, total: result.total, offset: result.offset, limit: result.limit, status, data: result.products };
    } catch (error: any) {
      this.logger.error(`❌ Error retrieving products by status: ${error.message}`);
      return { success: false, error: error.message, count: 0, total: 0, data: [] };
    }
  }

  async getDatabaseStatistics(): Promise<any> {
    try {
      const statistics = await this.dbService.getStatistics();
      return { success: true, statistics };
    } catch (error: any) {
      this.logger.error(`❌ Error retrieving statistics: ${error.message}`);
      return { success: false, error: error.message, statistics: {} };
    }
  }

  async deleteProductFromDb(productId: string): Promise<any> {
    try {
      const deleted = await this.dbService.deleteProduct(productId);
      return { success: true, data: deleted };
    } catch (error: any) {
      if (error.code === "P2025") return { success: false, error: "Product not found" };
      this.logger.error(`❌ Error deleting product: ${error.message}`);
      return { success: false, error: error.message };
    }
  }
}
