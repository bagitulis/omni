/**
 * Wholesale Settings Controller
 * RESPONSIBILITY: Handle HTTP requests for wholesale settings
 *
 * Endpoints:
 * - GET /api/wholesale/settings - Get tenant settings
 * - PUT /api/wholesale/settings - Update tenant settings
 * - POST /api/wholesale/preview - Preview calculated tiers
 */

import { Request, Response } from "express";
import { WholesaleSettingsService } from "../services/wholesale/wholesaleSettingsService";
import { getLogger } from "../utils/logger";
import { PrismaClient } from "@prisma/client";

const logger = getLogger("WholesaleSettingsController");

export class WholesaleSettingsController {
  private settingsService: WholesaleSettingsService;

  constructor(prisma?: PrismaClient) {
    this.settingsService = new WholesaleSettingsService(prisma);
  }

  /**
   * GET /api/wholesale/settings
   * Get wholesale settings for current tenant
   */
  async getSettings(req: Request, res: Response): Promise<Response> {
    try {
      const tenantId = (req as any).tenantId;
      if (!tenantId) {
        return res
          .status(401)
          .json({
            success: false,
            error: "Missing tenantId - authentication required",
          });
      }
      logger.info(`📥 Getting wholesale settings for tenant: ${tenantId}`);

      const settings = await this.settingsService.getSettings(tenantId);

      return res.json({
        success: true,
        data: settings,
      });
    } catch (error: any) {
      logger.error(`❌ Get settings error: ${error.message}`);
      return res.status(500).json({
        success: false,
        error: error.message,
      });
    }
  }

  /**
   * PUT /api/wholesale/settings
   * Update wholesale settings for current tenant
   *
   * Body: { adminFee?, minOrder1?, maxOrder1?, maxOrderTier3? }
   */
  async updateSettings(req: Request, res: Response): Promise<Response> {
    try {
      const tenantId = (req as any).tenantId;
      if (!tenantId) {
        return res
          .status(401)
          .json({
            success: false,
            error: "Missing tenantId - authentication required",
          });
      }
      const { adminFee, minOrder1, maxOrder1, maxOrderTier3 } = req.body;

      logger.info(`📝 Updating wholesale settings for tenant: ${tenantId}`);

      const settings = await this.settingsService.updateSettings(tenantId, {
        adminFee,
        minOrder1,
        maxOrder1,
        maxOrderTier3,
      });

      return res.json({
        success: true,
        message: "Settings updated successfully",
        data: settings,
      });
    } catch (error: any) {
      logger.error(`❌ Update settings error: ${error.message}`);
      return res.status(500).json({
        success: false,
        error: error.message,
      });
    }
  }

  /**
   * POST /api/wholesale/preview
   * Preview calculated tiers for a given base price
   *
   * Body: { basePrice: number, settings?: {...} }
   */
  async previewTiers(req: Request, res: Response): Promise<Response> {
    try {
      const tenantId = (req as any).tenantId;
      if (!tenantId) {
        return res
          .status(401)
          .json({
            success: false,
            error: "Missing tenantId - authentication required",
          });
      }
      const { basePrice, settings: customSettings } = req.body;

      if (!basePrice || basePrice <= 0) {
        return res.status(400).json({
          success: false,
          error: "Valid basePrice is required",
        });
      }

      logger.info(`🔍 Preview tiers for price: ${basePrice}`);

      // Get current settings or use custom
      const settings =
        customSettings || (await this.settingsService.getSettings(tenantId));

      const preview = this.settingsService.preview(basePrice, settings);

      return res.json({
        success: true,
        data: preview,
      });
    } catch (error: any) {
      logger.error(`❌ Preview error: ${error.message}`);
      return res.status(500).json({
        success: false,
        error: error.message,
      });
    }
  }
}
