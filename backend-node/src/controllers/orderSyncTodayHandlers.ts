/**
 * Order Sync Today Handlers
 * SRP: Handle Order Today operations
 */

import { Request, Response } from "express";
import { OrderTodayService } from "../services/orders/orderTodayService";
import { getPrisma } from "../services/prismaClient";
import { OrderSyncValidator } from "./orderSyncValidator";
import { OrderSyncResponseFormatter } from "./orderSyncResponseFormatter";
import { getLogger } from "../utils/logger";

const logger = getLogger("OrderSyncTodayHandlers");

const buildOrderTodayService = (tenantId: string) =>
  new OrderTodayService(getPrisma(tenantId), tenantId);

export class OrderSyncTodayHandlers {
  /**
   * POST /api/orders/today
   * Fetch and save Order Today items from all platforms
   */
  static async fetchOrdersToday(req: Request, res: Response): Promise<void> {
    try {
      const { days = 7 } = req.body || {};
      const tenantId = OrderSyncValidator.extractTenantId(req);

      const tenantValidation =
        OrderSyncValidator.validateTenantHasPlatformAccess(tenantId);
      if (!tenantValidation.isValid) {
        res
          .status(403)
          .json(
            OrderSyncResponseFormatter.formatErrorResponse(
              tenantValidation.error!
            )
          );
        return;
      }

      const orderTodayService = buildOrderTodayService(tenantId);
      const items = await orderTodayService.fetchAndSaveOrdersToday(
        Number(days)
      );

      res.json({
        success: true,
        message: `Fetched ${items.length} items for Order Today`,
        count: items.length,
        items: items,
      });
    } catch (error: any) {
      logger.error(`Failed to fetch Order Today: ${error.message}`);
      res
        .status(500)
        .json(OrderSyncResponseFormatter.formatErrorResponse(error.message));
    }
  }

  /**
   * GET /api/orders/today
   * Get saved Order Today items from database
   */
  static async getOrdersToday(req: Request, res: Response): Promise<void> {
    try {
      const tenantId = OrderSyncValidator.extractTenantId(req);
      const orderTodayService = buildOrderTodayService(tenantId);
      const items = await orderTodayService.getOrderTodayItems();

      res.json({
        success: true,
        count: items.length,
        items: items,
      });
    } catch (error: any) {
      logger.error(`Failed to get Order Today: ${error.message}`);
      res
        .status(500)
        .json(OrderSyncResponseFormatter.formatErrorResponse(error.message));
    }
  }
}
