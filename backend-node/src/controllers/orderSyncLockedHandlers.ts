/**
 * Order Sync Locked Order Handlers
 * SRP: Handle locked order operations
 */

import { Request, Response } from "express";
import { LockedOrderService } from "../services/lockedOrderService";
import { getPrisma } from "../services/prismaClient";
import { OrderSyncValidator } from "./orderSyncValidator";
import { OrderSyncResponseFormatter } from "./orderSyncResponseFormatter";
import { LockedOrderAggregationService } from "./lockedOrderAggregationService";
import { JobHistoryManager } from "../services/jobHistoryManager";
import { v4 as uuidv4 } from "uuid";
import { getLogger } from "../utils/logger";

const logger = getLogger("OrderSyncLockedHandlers");
import { getOrderSyncService } from "../services/orderSyncService";

const buildLockedOrderService = (tenantId: string) =>
  new LockedOrderService(getPrisma(tenantId));

export class OrderSyncLockedHandlers {
  /**
   * GET /api/orders/locked-today
   * Retrieve saved locked items from database
   */
  static async getLockedOrders(req: Request, res: Response): Promise<void> {
    try {
      const { days = 7 } = req.query;
      const tenantId = OrderSyncValidator.extractTenantId(req);
      const lockedOrderService = buildLockedOrderService(tenantId);

      const lockedOrders = await lockedOrderService.getLockedOrders(tenantId);
      const totalQty = await lockedOrderService.getTotalQty(tenantId);

      res.json(
        OrderSyncResponseFormatter.formatSavedLockedOrdersResponse(
          tenantId,
          Number(days),
          lockedOrders,
          totalQty
        )
      );
    } catch (error: any) {
      res
        .status(500)
        .json(OrderSyncResponseFormatter.formatErrorResponse(error.message));
    }
  }

  /**
   * POST /api/orders/locked-today
   * Fetch, aggregate, and save locked items
   *
   * Time-based logic:
   * - 00:00-14:00 (before 2pm): unprocess + processed orders
   * - 14:00-24:00 (after 2pm): unprocess orders only
   */
  static async saveLockedOrders(req: Request, res: Response): Promise<void> {
    const startTime = Date.now();
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

      const lockedOrderService = buildLockedOrderService(tenantId);
      const orderSyncService = await getOrderSyncService(tenantId);

      // Time-based logic: Check current hour (Jakarta timezone UTC+7)
      const now = new Date();
      const jakartaOffset = 7 * 60; // UTC+7 in minutes
      const utcMinutes = now.getUTCHours() * 60 + now.getUTCMinutes();
      const jakartaMinutes = utcMinutes + jakartaOffset;
      const jakartaHour = Math.floor((jakartaMinutes % (24 * 60)) / 60);

      const includeProcessed = jakartaHour < 14; // Before 2pm include processed

      logger.info(
        `[LockedOrders] Current Jakarta hour: ${jakartaHour}, includeProcessed: ${includeProcessed}`
      );

      // Always sync and fetch unprocess orders
      await orderSyncService.syncByCategory("unprocess", Number(days));
      const unprocessMap =
        await LockedOrderAggregationService.fetchOrdersByCategory(
          orderSyncService,
          "unprocess"
        );

      // Only sync and fetch processed orders before 2pm
      let processedMap: Record<string, any[]> = {
        shopee: [],
        lazada: [],
        tiktok: [],
      };
      if (includeProcessed) {
        await orderSyncService.syncByCategory("processed", Number(days));
        processedMap =
          await LockedOrderAggregationService.fetchOrdersByCategory(
            orderSyncService,
            "processed"
          );
      }

      const lockedItems = LockedOrderAggregationService.aggregateLockedOrders(
        unprocessMap,
        processedMap
      );

      const savedCount = await lockedOrderService.saveLockedOrders(
        tenantId,
        lockedItems
      );
      const totalQty = await lockedOrderService.getTotalQty(tenantId);

      // Log to history
      const durationMs = Date.now() - startTime;
      const mode = includeProcessed ? "unprocess+processed" : "unprocess-only";
      try {
        const historyManager = new JobHistoryManager();
        historyManager.addDirectHistory(
          `locked-orders-${uuidv4().substring(0, 8)}`,
          `locked_orders:${lockedItems.length} items (${mode})`,
          "completed",
          undefined,
          durationMs
        );
      } catch (historyError) {
        logger.warn("[OrderSync] Failed to log history:", historyError);
      }

      // Build response with mode info
      const response = OrderSyncResponseFormatter.formatLockedOrdersResponse(
        tenantId,
        Number(days),
        lockedItems,
        savedCount,
        totalQty
      );

      res.json({
        ...response,
        mode,
        jakartaHour,
        message: includeProcessed
          ? `Sebelum jam 14:00 - menghitung unprocess + processed orders`
          : `Setelah jam 14:00 - hanya menghitung unprocess orders`,
      });
    } catch (error: any) {
      const errorDurationMs = Date.now() - startTime;
      try {
        const historyManager = new JobHistoryManager();
        historyManager.addDirectHistory(
          `locked-orders-${uuidv4().substring(0, 8)}`,
          "locked_orders",
          "failed",
          error.message,
          errorDurationMs
        );
      } catch (historyError) {
        logger.warn("[OrderSync] Failed to log history:", historyError);
      }

      res
        .status(500)
        .json(OrderSyncResponseFormatter.formatErrorResponse(error.message));
    }
  }
}
