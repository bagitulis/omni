/**
 * Order Sync Category Handlers
 * SRP: Handle category-based order sync operations
 */

import { Request, Response } from "express";
import { OrderFormatterService } from "../services/orders/orderFormatterService";
import { OrderSyncValidator } from "./orderSyncValidator";
import { OrderSyncResponseFormatter } from "./orderSyncResponseFormatter";
import { getOrderSyncService } from "../services/orderSyncService";

export class OrderSyncCategoryHandlers {
  /**
   * POST /api/orders/sync/:category
   * Sync orders by category for all platforms
   */
  static async syncByCategory(req: Request, res: Response): Promise<void> {
    try {
      const { category } = req.params;
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
      const results = await orderSyncService.syncByCategory(
        category as "unpaid" | "unprocess" | "processed",
        Number(days)
      );

      res.json(
        OrderSyncResponseFormatter.formatSyncResponse(
          category,
          Number(days),
          results
        )
      );
    } catch (error: any) {
      res
        .status(500)
        .json(OrderSyncResponseFormatter.formatErrorResponse(error.message));
    }
  }

  /**
   * GET /api/orders/sync (Query-based - DEPRECATED)
   */
  static async syncByCategoryQuery(req: Request, res: Response): Promise<void> {
    try {
      const { category = "unpaid", days = "7" } = req.query;

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

      const categoryValidation = OrderSyncValidator.validateCategory(
        String(category)
      );
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
      const results = await orderSyncService.syncByCategory(
        String(category) as "unpaid" | "unprocess" | "processed",
        parseInt(String(days), 10)
      );

      res.json(
        OrderSyncResponseFormatter.formatSyncResponse(
          String(category),
          parseInt(String(days), 10),
          results
        )
      );
    } catch (error: any) {
      res
        .status(500)
        .json(OrderSyncResponseFormatter.formatErrorResponse(error.message));
    }
  }

  /**
   * GET /api/orders/:category
   * Get all orders by category from database
   */
  static async getOrdersByCategory(req: Request, res: Response): Promise<void> {
    try {
      const { category } = req.params;
      const tenantId = OrderSyncValidator.extractTenantId(req);

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

      const ordersMap: Record<string, any[]> = {
        shopee: [],
        lazada: [],
        tiktok: [],
      };

      const orderSyncService = await getOrderSyncService(tenantId);
      const shopeeOrders = await orderSyncService.getOrdersByCategory(
        category as any,
        "shopee"
      );
      const lazadaOrders = await orderSyncService.getOrdersByCategory(
        category as any,
        "lazada"
      );
      const tiktokOrders = await orderSyncService.getOrdersByCategory(
        category as any,
        "tiktok"
      );

      ordersMap.shopee = shopeeOrders.filter(
        (o: any) => o._platform === "shopee"
      );
      ordersMap.lazada = lazadaOrders.filter(
        (o: any) => o._platform === "lazada"
      );
      ordersMap.tiktok = tiktokOrders.filter(
        (o: any) => o._platform === "tiktok"
      );

      const items =
        OrderFormatterService.formatAllPlatformsForCategory(ordersMap);

      res.json(
        OrderSyncResponseFormatter.formatOrdersListResponse(category, items)
      );
    } catch (error: any) {
      res
        .status(500)
        .json(OrderSyncResponseFormatter.formatErrorResponse(error.message));
    }
  }
}
