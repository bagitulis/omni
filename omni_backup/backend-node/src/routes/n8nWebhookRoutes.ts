import { Router, Request, Response } from "express";
import { getPrisma } from "../services/prismaClient";
import { getOrderSyncService } from "../services/orderSyncService";

const router = Router();

/**
 * Route: POST /n8n/export-orders
 * Frontend Export button triggers this endpoint to send order data to n8n webhook
 *
 * Flow:
 * 1. Receive order data from frontend
 * 2. Send to n8n webhook for Google Sheets export
 *
 * Access: No auth required (data comes from frontend, already filtered)
 */
router.post(
  "/n8n/export-orders",
  async (req: Request, res: Response): Promise<void> => {
    try {
      // Get tenantId from header or body (frontend sends it)
      const tenantId =
        (req.headers["x-tenant-id"] as string) ||
        req.body.tenantId ||
        "yumna_bertigamart";

      const { orders, exportType = "today", metadata = {} } = req.body;

      if (!orders || !Array.isArray(orders) || orders.length === 0) {
        res.status(400).json({ error: "No orders data provided" });
        return;
      }

      console.log(
        `[n8n Export] Exporting ${orders.length} orders for tenant: ${tenantId}`
      );

      // Prepare webhook data
      const webhookData = {
        orders,
        timestamp: new Date().toISOString(),
        type: exportType,
        tenantId,
        metadata: {
          ...metadata,
          exportedBy: req.username || "unknown",
          totalOrders: orders.length,
        },
      };

      // Get n8n webhook URL from environment
      const n8nWebhookUrl =
        process.env.N8N_WEBHOOK_URL || "http://n8n:5678/webhook/export-orders";

      console.log(`[n8n Export] Sending to webhook: ${n8nWebhookUrl}`);

      const webhookResponse = await fetch(n8nWebhookUrl, {
        method: "POST",
        headers: {
          "Content-Type": "application/json",
        },
        body: JSON.stringify(webhookData),
      });

      if (!webhookResponse.ok) {
        const errorText = await webhookResponse.text();
        console.error(
          `[n8n Export] Webhook failed: ${webhookResponse.status} - ${errorText}`
        );
        throw new Error(
          `n8n webhook failed: ${webhookResponse.status} ${webhookResponse.statusText}`
        );
      }

      let webhookResult;
      try {
        webhookResult = await webhookResponse.json();
      } catch {
        webhookResult = { message: "Webhook received successfully" };
      }

      console.log(`[n8n Export] Success! Exported ${orders.length} orders`);

      res.json({
        success: true,
        message: `${orders.length} orders exported to n8n successfully`,
        n8nResponse: webhookResult,
      });
    } catch (error: any) {
      console.error("[n8n Export] Error:", error);
      res.status(500).json({
        success: false,
        error: "Failed to export orders to n8n",
        details: error.message,
      });
    }
  }
);

/**
 * Route: GET /order-manager/processed
 * Auto-trigger order sync for all platforms, then send to n8n webhook
 *
 * Flow:
 * 1. Sync processed orders from Shopee, TikTok, Lazada APIs → Database
 * 2. Send all synced orders to n8n webhook
 * 3. n8n workflow → Google Sheets
 *
 * Access: ?secret=<N8N_WEBHOOK_SECRET>&tenantId=<tenant>
 */
router.get(
  "/order-manager/processed",
  async (req: Request, res: Response): Promise<void> => {
    try {
      const { secret, tenantId: queryTenantId } = req.query;

      // Check authentication - Secret Key required
      const webhookSecret = process.env.N8N_WEBHOOK_SECRET;

      if (!secret || secret !== webhookSecret) {
        res.status(401).json({ error: "Invalid or missing secret key" });
        return;
      }

      if (!queryTenantId || typeof queryTenantId !== "string") {
        res.status(400).json({ error: "tenantId required" });
        return;
      }

      const tenantId = queryTenantId;

      console.log(`[n8n Webhook] Starting order sync for tenant: ${tenantId}`);

      // Step 1: Trigger order sync for all platforms
      const orderSyncService = await getOrderSyncService(tenantId);

      // Sync each platform separately (processed orders)
      const [shopeeResult, tiktokResult, lazadaResult] = await Promise.all([
        orderSyncService.syncPlatformOrders("shopee", "processed", 7),
        orderSyncService.syncPlatformOrders("tiktok", "processed", 7),
        orderSyncService.syncPlatformOrders("lazada", "processed", 7),
      ]);

      const syncResult = {
        shopee: shopeeResult.length,
        tiktok: tiktokResult.length,
        lazada: lazadaResult.length,
        total: shopeeResult.length + tiktokResult.length + lazadaResult.length,
      };

      console.log(`[n8n Webhook] Order sync completed:`, syncResult);

      // Step 2: Get all processed orders from database
      const prisma = getPrisma(tenantId);

      const [shopeeOrders, tiktokOrders, lazadaOrders] = await Promise.all([
        prisma.shopeeOrder.findMany({
          where: { orderStatus: "COMPLETED" },
          orderBy: { createdAt: "desc" },
          take: 100,
        }),
        prisma.tiktokOrder.findMany({
          where: { orderStatus: "COMPLETED" },
          orderBy: { updatedAt: "desc" },
          take: 100,
        }),
        prisma.lazadaOrder.findMany({
          where: { orderStatus: "delivered" },
          orderBy: { updatedAt: "desc" },
          take: 100,
        }),
      ]);

      // Combine all orders
      const allOrders = [
        ...shopeeOrders.map((order: any) => ({
          orderId: order.orderSn,
          buyerUsername: order.buyerUsername || "N/A",
          totalAmount: order.totalAmount || 0,
          status: order.orderStatus,
          processedAt: order.createdAt,
          platform: "shopee",
          tenantId,
        })),
        ...tiktokOrders.map((order: any) => ({
          orderId: order.orderSn,
          buyerUsername: order.buyerUsername || "N/A",
          totalAmount: order.totalAmount || 0,
          status: order.orderStatus,
          processedAt: order.updatedAt,
          platform: "tiktok",
          tenantId,
        })),
        ...lazadaOrders.map((order: any) => ({
          orderId: order.orderSn,
          buyerUsername: order.buyerUsername || "N/A",
          totalAmount: order.totalAmount || 0,
          status: order.orderStatus,
          processedAt: order.updatedAt,
          platform: "lazada",
          tenantId,
        })),
      ];

      if (allOrders.length === 0) {
        res.json({
          success: true,
          message: "No processed orders found after sync",
          syncResult,
          count: 0,
        });
        return;
      }

      // Step 3: Send to n8n webhook
      const webhookData = {
        orders: allOrders,
        timestamp: new Date().toISOString(),
        platforms: ["shopee", "tiktok", "lazada"],
        type: "processed",
        tenantId,
        syncResult, // Include sync statistics
      };

      console.log(`[n8n Webhook] Sending ${allOrders.length} orders to n8n...`);

      // Call n8n webhook
      const n8nWebhookUrl =
        process.env.N8N_WEBHOOK_URL ||
        "http://localhost:5678/webhook/processed-orders";

      const webhookResponse = await fetch(n8nWebhookUrl, {
        method: "POST",
        headers: {
          "Content-Type": "application/json",
        },
        body: JSON.stringify(webhookData),
      });

      if (!webhookResponse.ok) {
        throw new Error(
          `n8n webhook failed: ${webhookResponse.status} ${webhookResponse.statusText}`
        );
      }

      const webhookResult = await webhookResponse.json();

      console.log(`[n8n Webhook] Success! n8n response:`, webhookResult);

      res.json({
        success: true,
        message: "Orders synced from all platforms and sent to Google Sheets",
        summary: {
          totalOrders: allOrders.length,
          shopee: shopeeOrders.length,
          tiktok: tiktokOrders.length,
          lazada: lazadaOrders.length,
        },
        syncResult,
        n8nResponse: webhookResult,
      });
    } catch (error: any) {
      console.error("[n8n Webhook] Error:", error);
      res.status(500).json({
        error: "Failed to sync orders",
        details: error.message,
      });
    }
  }
);

export default router;
