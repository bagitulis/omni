/**
 * Order Sync Platform Handlers
 * SRP: Handle platform-specific order operations
 */

import { Request, Response } from "express";
import { OrderFormatterService } from "../services/orders/orderFormatterService";
import { OrderSyncValidator } from "./orderSyncValidator";
import { OrderSyncResponseFormatter } from "./orderSyncResponseFormatter";
import { getOrderSyncService } from "../services/orderSyncService";

export class OrderSyncPlatformHandlers {
  /**
   * GET /api/orders/:platform/:category
   * Get platform-specific orders by category
   */
  static async getPlatformOrders(req: Request, res: Response): Promise<void> {
    try {
      const { platform, category } = req.params;
      const tenantId = OrderSyncValidator.extractTenantId(req);

      const platformValidation = OrderSyncValidator.validatePlatform(platform);
      if (!platformValidation.isValid) {
        res
          .status(400)
          .json(
            OrderSyncResponseFormatter.formatErrorResponse(
              platformValidation.error!
            )
          );
        return;
      }

      const categoryValidation = OrderSyncValidator.validateCategory(category);
      if (!categoryValidation.isValid) {
        res
          .status(400)
          .json(
            OrderSyncResponseFormatter.formatErrorResponse(
              categoryValidation.error!
            )
          );
        return;
      }

      const orderSyncService = await getOrderSyncService(tenantId);
      const orders = await orderSyncService.getPlatformOrders(
        platform as "shopee" | "lazada" | "tiktok",
        category as "unpaid" | "unprocess" | "processed"
      );

      const formattedItems = OrderFormatterService.formatOrdersForFrontend(
        platform,
        orders
      );

      res.json(
        OrderSyncResponseFormatter.formatPlatformOrdersResponse(
          platform,
          category,
          formattedItems
        )
      );
    } catch (error: any) {
      res
        .status(500)
        .json(OrderSyncResponseFormatter.formatErrorResponse(error.message));
    }
  }

  /**
   * POST /api/orders/:platform/sync/:category
   * Sync specific platform orders by category
   */
  static async syncPlatformOrders(req: Request, res: Response): Promise<void> {
    try {
      const { platform } = req.params;
      const { category = "unpaid", days = 7 } = req.body;

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

      const platformValidation = OrderSyncValidator.validatePlatform(platform);
      if (!platformValidation.isValid) {
        res
          .status(400)
          .json(
            OrderSyncResponseFormatter.formatErrorResponse(
              platformValidation.error!
            )
          );
        return;
      }

      const categoryValidation = OrderSyncValidator.validateCategory(category);
      if (!categoryValidation.isValid) {
        res
          .status(400)
          .json(
            OrderSyncResponseFormatter.formatErrorResponse(
              categoryValidation.error!
            )
          );
        return;
      }

      const orderSyncService = await getOrderSyncService(tenantId);
      const orders = await orderSyncService.syncPlatformOrders(
        platform as "shopee" | "lazada" | "tiktok",
        category as "unpaid" | "unprocess" | "processed",
        days
      );

      res.json(
        OrderSyncResponseFormatter.formatSyncPlatformResponse(
          platform,
          category,
          days,
          orders
        )
      );
    } catch (error: any) {
      res
        .status(500)
        .json(OrderSyncResponseFormatter.formatErrorResponse(error.message));
    }
  }

  /**
   * POST /api/orders/:platform/details
   * Get order details for specific order IDs
   */
  static async getOrderDetails(req: Request, res: Response): Promise<void> {
    try {
      const { platform } = req.params;
      const { orderIds = [] } = req.body;
      const tenantId = OrderSyncValidator.extractTenantId(req);

      const platformValidation = OrderSyncValidator.validatePlatform(platform);
      if (!platformValidation.isValid) {
        res
          .status(400)
          .json(
            OrderSyncResponseFormatter.formatErrorResponse(
              platformValidation.error!
            )
          );
        return;
      }

      const orderIdsValidation = OrderSyncValidator.validateOrderIds(orderIds);
      if (!orderIdsValidation.isValid) {
        res
          .status(400)
          .json(
            OrderSyncResponseFormatter.formatErrorResponse(
              orderIdsValidation.error!
            )
          );
        return;
      }

      const orderSyncService = await getOrderSyncService(tenantId);
      const details = await orderSyncService.getOrderDetails(
        platform as "shopee" | "lazada" | "tiktok",
        orderIds
      );

      res.json(
        OrderSyncResponseFormatter.formatOrderDetailsResponse(platform, details)
      );
    } catch (error: any) {
      res
        .status(500)
        .json(OrderSyncResponseFormatter.formatErrorResponse(error.message));
    }
  }
}
