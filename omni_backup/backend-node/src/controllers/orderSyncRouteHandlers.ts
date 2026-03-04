/**
 * Order Sync Route Handlers
 * Facade that delegates to specialized handler classes
 * SRP: Route delegation only
 */

import { OrderSyncCategoryHandlers } from "./orderSyncCategoryHandlers";
import { OrderSyncPlatformHandlers } from "./orderSyncPlatformHandlers";
import { OrderSyncLockedHandlers } from "./orderSyncLockedHandlers";
import { OrderSyncTodayHandlers } from "./orderSyncTodayHandlers";

/**
 * Order Sync Route Handlers - Facade Pattern
 * Delegates to specialized handlers:
 * - CategoryHandlers: syncByCategory, syncByCategoryQuery, getOrdersByCategory
 * - PlatformHandlers: getPlatformOrders, syncPlatformOrders, getOrderDetails
 * - LockedHandlers: getLockedOrders, saveLockedOrders
 * - TodayHandlers: fetchOrdersToday, getOrdersToday
 */
export class OrderSyncRouteHandlers {
  // Category handlers
  static syncByCategory = OrderSyncCategoryHandlers.syncByCategory;
  static syncByCategoryQuery = OrderSyncCategoryHandlers.syncByCategoryQuery;
  static getOrdersByCategory = OrderSyncCategoryHandlers.getOrdersByCategory;

  // Platform handlers
  static getPlatformOrders = OrderSyncPlatformHandlers.getPlatformOrders;
  static syncPlatformOrders = OrderSyncPlatformHandlers.syncPlatformOrders;
  static getOrderDetails = OrderSyncPlatformHandlers.getOrderDetails;

  // Locked order handlers
  static getLockedOrders = OrderSyncLockedHandlers.getLockedOrders;
  static saveLockedOrders = OrderSyncLockedHandlers.saveLockedOrders;

  // Order today handlers
  static fetchOrdersToday = OrderSyncTodayHandlers.fetchOrdersToday;
  static getOrdersToday = OrderSyncTodayHandlers.getOrdersToday;
}

// Re-export individual handlers for direct usage if needed
export {
  OrderSyncCategoryHandlers,
  OrderSyncPlatformHandlers,
  OrderSyncLockedHandlers,
  OrderSyncTodayHandlers,
};
