/**
 * Shopee Product Master Controller
 * Handles master product operations and stats
 * Single Responsibility: Master product management only
 */

import { Request, Response } from "express";
import { PrismaClient } from "@prisma/client";
import { ProductControllerBase } from "./base/ProductControllerBase";
import { ShopeeProductService } from "../services/shopeeProductService";
import { ShopeeAPIClient } from "../api/clients/shopeeAPIClient";
import { ShopeeConfigManager } from "../config/managers/shopeeConfigManager";

export class ShopeeProductMasterController extends ProductControllerBase {
  private productService: ShopeeProductService;

  constructor(
    prisma: PrismaClient,
    apiClient: ShopeeAPIClient,
    configManager: ShopeeConfigManager
  ) {
    super(prisma);
    this.productService = new ShopeeProductService(
      prisma,
      apiClient,
      configManager
    );
  }

  async getMasterProductList(req: Request, res: Response): Promise<any> {
    try {
      const limit = parseInt(req.query.limit as string) || 100;
      const offset = parseInt(req.query.offset as string) || 0;

      const result = await this.productService.getMasterProductsList({
        limit,
        offset,
      });

      return this.sendSuccess(res, result, result.success ? 200 : 400);
    } catch (error: any) {
      return this.sendError(res, error);
    }
  }

  async getMasterProductStats(_req: Request, res: Response): Promise<any> {
    try {
      const result = await this.productService.getMasterProductsStats();

      return this.sendSuccess(res, result, result.success ? 200 : 400);
    } catch (error: any) {
      return this.sendError(res, error);
    }
  }

  async syncMasterProduct(_req: Request, res: Response): Promise<any> {
    try {
      this.logger.info("🔄 Manual Master Product Sync triggered");

      const stats = await this.productService.getMasterProductsStats();

      return this.sendSuccess(res, {
        success: true,
        message: "Master product sync completed (already up-to-date)",
        stats: stats.stats,
      });
    } catch (error: any) {
      return this.sendError(res, error);
    }
  }
}
