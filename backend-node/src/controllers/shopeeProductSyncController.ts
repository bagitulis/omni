/**
 * Shopee Product Sync Controller
 * Handles product synchronization operations
 * Single Responsibility: Product sync only
 */

import { Request, Response } from "express";
import { PrismaClient } from "@prisma/client";
import { ProductControllerBase } from "./base/ProductControllerBase";
import { ShopeeProductService } from "../services/shopeeProductService";
import { ShopeeAPIClient } from "../api/clients/shopeeAPIClient";
import { ShopeeConfigManager } from "../config/managers/shopeeConfigManager";

export class ShopeeProductSyncController extends ProductControllerBase {
  private productService: ShopeeProductService;
  private configManager: ShopeeConfigManager;

  constructor(
    prisma: PrismaClient,
    apiClient: ShopeeAPIClient,
    configManager: ShopeeConfigManager
  ) {
    super(prisma);
    this.configManager = configManager;
    this.productService = new ShopeeProductService(
      prisma,
      apiClient,
      configManager
    );
  }

  async syncProducts(req: Request, res: Response): Promise<any> {
    try {
      await this.configManager.loadConfig();

      const itemStatus =
        (req.body.item_status as string) ||
        (req.query.itemStatus as string) ||
        "NORMAL";
      const limit =
        parseInt(req.body.page_size as string) ||
        parseInt(req.query.limit as string) ||
        50;

      const result = await this.productService.getProducts({
        itemStatus,
        offset: 0,
        limit,
      });

      if (result.success) {
        return this.sendSuccess(res, {
          success: true,
          message: "Successfully synced products from Shopee API",
          total_fetched: result.total,
          total_saved: result.products?.length || 0,
        });
      }

      return this.sendSuccess(res, result, 400);
    } catch (error: any) {
      return this.sendError(res, error);
    }
  }

  async batchSync(req: Request, res: Response): Promise<any> {
    return this.syncProducts(req, res);
  }

  async getUnprocessedItems(_req: Request, res: Response): Promise<any> {
    return this.sendSuccess(res, { success: true, items: [] });
  }

  async getItemsWithoutModels(_req: Request, res: Response): Promise<any> {
    return this.sendSuccess(res, { success: true, items: [] });
  }

  async getSyncLogs(_req: Request, res: Response): Promise<any> {
    return this.sendSuccess(res, { success: true, logs: [] });
  }
}
