/**
 * Route Constants
 * Centralized route definitions for DRY principle
 * Usage: Import these constants instead of hardcoding route strings
 */

// Route name constants - PascalCase
export const ROUTE_NAMES = {
  DASHBOARD: "Dashboard",
  OPERATION: "Operation",
  OPERATION_SHOPEE: "OperationShopee",
  OPERATION_LAZADA: "OperationLazada",
  OPERATION_TIKTOK: "OperationTiktok",
  PRODUCT_MANAGER: "ProductManager",
  PRODUCT_MANAGER_SHOPEE: "ProductManagerShopee",
  PRODUCT_MANAGER_LAZADA: "ProductManagerLazada",
  PRODUCT_MANAGER_TIKTOK: "ProductManagerTiktok",
  ORDER_MANAGER: "OrderManager",
  ORDER_MANAGER_SHOPEE: "OrderManagerShopee",
  ORDER_MANAGER_LAZADA: "OrderManagerLazada",
  ORDER_MANAGER_TIKTOK: "OrderManagerTiktok",
  INVENTORY: "Inventory",
  SETTINGS: "Settings",
  SETTINGS_GOOGLE_SHEETS: "SettingsGoogleSheets",
  SETTINGS_SCRIPT_MONITOR: "SettingsScriptMonitor",
  ROUTE_MAPPING: "RouteMapping",
} as const;

// Route path constants - kebab-case
export const ROUTE_PATHS = {
  DASHBOARD: "/",
  OPERATION: "/operation",
  OPERATION_SHOPEE: "/operation/shopee",
  OPERATION_LAZADA: "/operation/lazada",
  OPERATION_TIKTOK: "/operation/tiktok",
  PRODUCT_MANAGER: "/product-manager",
  PRODUCT_MANAGER_SHOPEE: "/product-manager/shopee",
  PRODUCT_MANAGER_LAZADA: "/product-manager/lazada",
  PRODUCT_MANAGER_TIKTOK: "/product-manager/tiktok",
  ORDER_MANAGER: "/order-manager",
  ORDER_MANAGER_SHOPEE: "/order-manager/shopee",
  ORDER_MANAGER_LAZADA: "/order-manager/lazada",
  ORDER_MANAGER_TIKTOK: "/order-manager/tiktok",
  INVENTORY: "/inventory",
  SETTINGS: "/settings",
  SETTINGS_GOOGLE_SHEETS: "/settings/google-sheets",
  SETTINGS_SCRIPT_MONITOR: "/settings/script-monitor",
  ROUTE_MAPPING: "/route-mapping",
} as const;

// Metadata constants
export const ROUTE_SECTIONS = {
  OPERATION: "operation",
  PRODUCT_MANAGER: "product-manager",
  ORDER_MANAGER: "order-manager",
  INVENTORY: "inventory",
  SETTINGS: "settings",
  ROUTE_MAPPING: "route-mapping",
} as const;

export const PLATFORMS = {
  SHOPEE: "shopee",
  LAZADA: "lazada",
  TIKTOK: "tiktok",
} as const;

export const SETTINGS_SUBSECTIONS = {
  GOOGLE_SHEETS: "google-sheets",
  SCRIPT_MONITOR: "script-monitor",
} as const;

// Type exports for TypeScript safety
export type RouteName = (typeof ROUTE_NAMES)[keyof typeof ROUTE_NAMES];
export type RoutePath = (typeof ROUTE_PATHS)[keyof typeof ROUTE_PATHS];
export type RouteSection = (typeof ROUTE_SECTIONS)[keyof typeof ROUTE_SECTIONS];
export type Platform = (typeof PLATFORMS)[keyof typeof PLATFORMS];
export type SettingsSubsection =
  (typeof SETTINGS_SUBSECTIONS)[keyof typeof SETTINGS_SUBSECTIONS];

/**
 * Helper function to build platform-specific routes
 * @param baseSection - Base section (e.g., 'operation', 'product-manager')
 * @param platform - Platform name (e.g., 'shopee', 'lazada', 'tiktok')
 * @returns Full route path
 */
export function getPlatformRoutePath(
  baseSection: string,
  platform: Platform,
): string {
  return `/${baseSection}/${platform}`;
}

/**
 * Helper function to get route name for platform
 * @param baseName - Base name (e.g., 'Operation', 'ProductManager')
 * @param platform - Platform name (capitalized)
 * @returns Route name with platform
 */
export function getPlatformRouteName(
  baseName: string,
  platform: string,
): string {
  const capitalizedPlatform =
    platform.charAt(0).toUpperCase() + platform.slice(1);
  return `${baseName}${capitalizedPlatform}`;
}
