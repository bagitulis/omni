/**
 * Order Sync Controller
 * Handles API endpoints for order synchronization
 * Single Responsibility: Route definitions only
 */

import { Router } from "express";
import { OrderSyncRouteHandlers } from "./orderSyncRouteHandlers";

const router = Router();

// POST /api/orders/sync/:category - Sync orders by category for all platforms
router.post("/sync/:category", OrderSyncRouteHandlers.syncByCategory);

// GET /api/orders/sync - Query-based sync (DEPRECATED)
router.get("/sync", OrderSyncRouteHandlers.syncByCategoryQuery);

// GET /api/orders/locked-today - Retrieve saved locked items
router.get("/locked-today", OrderSyncRouteHandlers.getLockedOrders);

// POST /api/orders/locked-today - Fetch, aggregate, and save locked items
router.post("/locked-today", OrderSyncRouteHandlers.saveLockedOrders);

// GET /api/orders/today - Get saved Order Today items
router.get("/today", OrderSyncRouteHandlers.getOrdersToday);

// POST /api/orders/today - Fetch and save Order Today items from all platforms
router.post("/today", OrderSyncRouteHandlers.fetchOrdersToday);

// GET /api/orders/:category - Get all orders by category from database
router.get("/:category", OrderSyncRouteHandlers.getOrdersByCategory);

// GET /api/orders/:platform/:category - Get platform-specific orders
router.get("/:platform/:category", OrderSyncRouteHandlers.getPlatformOrders);

// POST /api/orders/:platform/sync/:category - Sync specific platform orders
router.post(
  "/:platform/sync/:category",
  OrderSyncRouteHandlers.syncPlatformOrders
);

// POST /api/orders/:platform/details - Get order details for specific order IDs
router.post("/:platform/details", OrderSyncRouteHandlers.getOrderDetails);

export default router;
