/**
 * Wholesale Controller
 * RESPONSIBILITY: Handle single-item wholesale operations
 *
 * Endpoints:
 * - DELETE /api/wholesale/shopee/:itemId - Delete wholesale tiers
 * - PUT /api/wholesale/shopee/:itemId - Update wholesale tiers
 * - GET /api/wholesale/shopee/:itemId - Get wholesale info
 * - GET /api/wholesale/shopee/lookup/:sku - Lookup itemId by SKU
 */

import { Request, Response } from "express";
import {
  ShopeeWholesaleService,
  WholesaleTier,
} from "../services/wholesale/shopeeWholesaleService";
import { ShopeeAPIClient } from "../api/clients/shopeeAPIClient";
import { ShopeeConfigManager } from "../config/managers/shopeeConfigManager";
import { getLogger } from "../utils/logger";
import { PrismaClient } from "@prisma/client";

const logger = getLogger("WholesaleController");

export class WholesaleController {
  private shopeeService: ShopeeWholesaleService;

  constructor(
    shopeeApiClient: ShopeeAPIClient,
    shopeeConfig: ShopeeConfigManager,
    prisma?: PrismaClient
  ) {
    this.shopeeService = new ShopeeWholesaleService(
      shopeeApiClient,
      shopeeConfig,
      prisma
    );
  }

  /**
   * DELETE /api/wholesale/shopee/:itemId
   * Delete all wholesale tiers for a Shopee product
   */
  async deleteShopeeWholesale(req: Request, res: Response): Promise<Response> {
    try {
      const { itemId } = req.params;

      if (!itemId || isNaN(Number(itemId))) {
        return res.status(400).json({
          success: false,
          error: "Valid itemId is required",
        });
      }

      logger.info(`🗑️ Deleting wholesale for Shopee item: ${itemId}`);

      const result = await this.shopeeService.deleteWholesaleTiers(
        Number(itemId)
      );

      if (!result.success) {
        return res.status(400).json({
          success: false,
          error: result.error,
        });
      }

      return res.json({
        success: true,
        message: result.message,
        data: { itemId: result.itemId },
      });
    } catch (error: any) {
      logger.error(`❌ Delete wholesale error: ${error.message}`);
      return res.status(500).json({
        success: false,
        error: error.message,
      });
    }
  }

  /**
   * PUT /api/wholesale/shopee/:itemId
   * Update wholesale tiers for a Shopee product
   *
   * Body: { tiers: [{ minCount, maxCount, unitPrice }] }
   */
  async updateShopeeWholesale(req: Request, res: Response): Promise<Response> {
    try {
      const { itemId } = req.params;
      const { tiers } = req.body;

      if (!itemId || isNaN(Number(itemId))) {
        return res.status(400).json({
          success: false,
          error: "Valid itemId is required",
        });
      }

      if (!Array.isArray(tiers)) {
        return res.status(400).json({
          success: false,
          error: "tiers must be an array",
        });
      }

      logger.info(`📦 Updating wholesale for Shopee item: ${itemId}`);

      const result = await this.shopeeService.updateWholesaleTiers(
        Number(itemId),
        tiers as WholesaleTier[]
      );

      if (!result.success) {
        return res.status(400).json({
          success: false,
          error: result.error,
        });
      }

      return res.json({
        success: true,
        message: result.message,
        data: { itemId: result.itemId },
      });
    } catch (error: any) {
      logger.error(`❌ Update wholesale error: ${error.message}`);
      return res.status(500).json({
        success: false,
        error: error.message,
      });
    }
  }

  /**
   * GET /api/wholesale/shopee/:itemId
   * Get current wholesale info for a Shopee product
   */
  async getShopeeWholesale(req: Request, res: Response): Promise<Response> {
    try {
      const { itemId } = req.params;

      if (!itemId || isNaN(Number(itemId))) {
        return res.status(400).json({
          success: false,
          error: "Valid itemId is required",
        });
      }

      logger.info(`📥 Getting wholesale for Shopee item: ${itemId}`);

      const result = await this.shopeeService.getWholesaleTiers(Number(itemId));

      if (!result) {
        return res.status(404).json({
          success: false,
          error: "Product not found",
        });
      }

      return res.json({
        success: true,
        data: result,
      });
    } catch (error: any) {
      logger.error(`❌ Get wholesale error: ${error.message}`);
      return res.status(500).json({
        success: false,
        error: error.message,
      });
    }
  }

  /**
   * GET /api/wholesale/shopee/lookup/:sku
   * Lookup item_id from SKU
   */
  async lookupItemBySku(req: Request, res: Response): Promise<Response> {
    try {
      const { sku } = req.params;

      if (!sku) {
        return res.status(400).json({
          success: false,
          error: "SKU is required",
        });
      }

      const result = await this.shopeeService.lookupItemIdBySku(sku);

      if (!result) {
        return res.status(404).json({
          success: false,
          error: `SKU not found: ${sku}`,
        });
      }

      return res.json({
        success: true,
        data: result,
      });
    } catch (error: any) {
      logger.error(`❌ Lookup error: ${error.message}`);
      return res.status(500).json({
        success: false,
        error: error.message,
      });
    }
  }
}
