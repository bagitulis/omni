import { Request, Response } from "express";
import { PrismaClient } from "@prisma/client";
import { ProductControllerBase } from "./base/ProductControllerBase";
import { TiktokProductService } from "../services/tiktokProductService";
import { TiktokAPIClient } from "../api/clients/tiktokAPIClient";
import { TiktokConfigManager } from "../config/managers/tiktokConfigManager";

/**
 * TikTok Product Controller
 * Handles TikTok product search, sync, and database operations
 * Product creation moved to TiktokProductCreateController (SRP compliance)
 * Max 300 lines - AGENTS.MD compliant
 */
export class TiktokProductController extends ProductControllerBase {
  private productService: TiktokProductService;
  private configManager: TiktokConfigManager;

  constructor(
    prisma: PrismaClient,
    apiClient: TiktokAPIClient,
    configManager: TiktokConfigManager,
  ) {
    super(prisma);
    this.configManager = configManager;
    this.productService = new TiktokProductService(
      prisma,
      apiClient,
      configManager,
    );
  }

  /**
   * POST /api/tiktok/products/search
   * Search products from TikTok API
   * Supports limit parameter for testing (e.g., limit=5)
   */
  async searchProducts(req: Request, res: Response): Promise<any> {
    try {
      // Ensure config is loaded
      await this.configManager.loadConfig();

      const {
        status = "ACTIVATE",
        page_size = 100,
        page_token,
        limit,
        sync_to_db = true,
      } = req.body;

      const result = await this.productService.searchProducts({
        status,
        pageSize: page_size,
        limit,
        pageToken: page_token,
        syncToDb: sync_to_db,
      });

      return this.sendSuccess(res, result, result.success ? 200 : 400);
    } catch (error: any) {
      return this.sendError(res, error);
    }
  }

  /**
   * POST /api/tiktok/products/get-all
   * Fetch all products with pagination
   */
  async getAllProducts(req: Request, res: Response): Promise<any> {
    try {
      const {
        status = "ACTIVATE",
        page_size = 100,
        max_pages,
        sync_to_db = true,
      } = req.body;

      const result = await this.productService.getAllProducts({
        status,
        pageSize: page_size,
        maxPages: max_pages,
        syncToDb: sync_to_db,
      });

      return this.sendSuccess(res, result, result.success ? 200 : 400);
    } catch (error: any) {
      return this.sendError(res, error);
    }
  }

  /**
   * GET /api/tiktok/products/:productId
   * Get product detail by ID
   */
  async getProductDetail(req: Request, res: Response): Promise<any> {
    try {
      const { productId } = req.params;
      const sync_to_db = req.query.sync_to_db !== "false";

      const result = await this.productService.getProductDetail({
        productId,
        syncToDb: sync_to_db,
      });

      return this.sendSuccess(res, result, result.success ? 200 : 400);
    } catch (error: any) {
      return this.sendError(res, error);
    }
  }

  /**
   * GET /api/tiktok/db/products/list
   * Get products from database
   */
  async getProductsFromDb(req: Request, res: Response): Promise<any> {
    try {
      const offset = parseInt(req.query.offset as string) || 0;
      const limit = parseInt(req.query.limit as string) || 100;

      const result = await this.productService.getProductListFromDb({
        offset,
        limit,
      });

      return this.sendSuccess(res, result, result.success ? 200 : 400);
    } catch (error: any) {
      return this.sendError(res, error);
    }
  }

  /**
   * GET /api/tiktok/db/products/master
   * Get master products from database
   */
  async getMasterProductsFromDb(req: Request, res: Response): Promise<any> {
    try {
      const offset = parseInt(req.query.offset as string) || 0;
      const limit = parseInt(req.query.limit as string) || 100000;

      const result = await this.productService.getMasterProductsFromDb({
        offset,
        limit,
      });

      return this.sendSuccess(res, result, result.success ? 200 : 400);
    } catch (error: any) {
      return this.sendError(res, error);
    }
  }

  /**
   * GET /api/tiktok/db/products/:productId
   * Get product by ID from database
   */
  async getProductByIdFromDb(req: Request, res: Response): Promise<any> {
    try {
      const { productId } = req.params;

      const result = await this.productService.getProductByIdFromDb(productId);

      return this.sendSuccess(res, result, result.success ? 200 : 404);
    } catch (error: any) {
      return this.sendError(res, error);
    }
  }

  /**
   * GET /api/tiktok/db/search?q=...&field=...
   * Search products in database
   */
  async searchProductsInDb(req: Request, res: Response): Promise<any> {
    try {
      const query = req.query.q as string;
      const field = (req.query.field as string) || "name";
      const offset = parseInt(req.query.offset as string) || 0;
      const limit = parseInt(req.query.limit as string) || 100;

      if (!query) {
        return res.status(400).json({
          success: false,
          error: "Search query required",
        });
      }

      const result = await this.productService.searchProductsInDb({
        query,
        field,
        offset,
        limit,
      });

      return this.sendSuccess(res, result, result.success ? 200 : 400);
    } catch (error: any) {
      return this.sendError(res, error);
    }
  }

  /**
   * GET /api/tiktok/db/products/status/:status
   * Get products by status
   */
  async getProductsByStatus(req: Request, res: Response): Promise<any> {
    try {
      const { status } = req.params;
      const offset = parseInt(req.query.offset as string) || 0;
      const limit = parseInt(req.query.limit as string) || 100;

      const result = await this.productService.getProductsByStatusFromDb({
        status,
        offset,
        limit,
      });

      return this.sendSuccess(res, result, result.success ? 200 : 400);
    } catch (error: any) {
      return this.sendError(res, error);
    }
  }

  /**
   * GET /api/tiktok/db/statistics
   * Get database statistics
   */
  async getStatistics(_req: Request, res: Response): Promise<any> {
    try {
      const result = await this.productService.getDatabaseStatistics();

      return this.sendSuccess(res, result, result.success ? 200 : 400);
    } catch (error: any) {
      return this.sendError(res, error);
    }
  }

  /**
   * DELETE /api/tiktok/db/products/:productId
   * Delete product from database
   */
  async deleteProductFromDb(req: Request, res: Response): Promise<any> {
    try {
      const { productId } = req.params;

      const result = await this.productService.deleteProductFromDb(productId);

      return this.sendSuccess(res, result, result.success ? 200 : 404);
    } catch (error: any) {
      return this.sendError(res, error);
    }
  }
}
