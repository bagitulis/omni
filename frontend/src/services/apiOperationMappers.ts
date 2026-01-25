/**
 * API Operation Mappers
 * Maps platform operations to their respective endpoints
 */

export interface EndpointMapping {
  endpoint: string;
  method: string;
}

/**
 * Map Shopee operation names to endpoints
 * Updated to match Backend Go routes (unified inventory endpoints)
 */
export function mapShopeeOperation(operationName: string): EndpointMapping {
  const operationMap: Record<string, string> = {
    update_stock: "/inventory/update-stock",
    update_price: "/inventory/update-price",
    export_orders: "/shopee/orders/export",
    get_token: "/platform-auth/initiate/shopee",
    refresh_token: "/tokens/refresh/shopee",
    update_code: "/shopee/operations/update-code",
  };
  return {
    endpoint: operationMap[operationName] || "/execute",
    method: "POST",
  };
}

/**
 * Map Lazada operation names to endpoints
 * Updated to match Backend Go routes (unified inventory endpoints)
 */
export function mapLazadaOperation(operationName: string): EndpointMapping {
  const operationMap: Record<string, string> = {
    update_stock: "/inventory/update-stock",
    update_price: "/inventory/update-price",
    export_orders: "/lazada/orders/export",
    get_token: "/platform-auth/initiate/lazada",
    refresh_token: "/tokens/refresh/lazada",
    update_code: "/lazada/operations/update-code",
  };
  return {
    endpoint: operationMap[operationName] || "/execute",
    method: "POST",
  };
}

/**
 * Map TikTok operation names to endpoints
 * Updated to match Backend Go routes (unified inventory endpoints)
 */
export function mapTikTokOperation(operationName: string): EndpointMapping {
  const operationMap: Record<string, string> = {
    update_stock: "/inventory/update-stock",
    update_price: "/inventory/update-price",
    export_orders: "/tiktok/orders/export",
    get_token: "/platform-auth/initiate/tiktok",
    refresh_token: "/tokens/refresh/tiktok",
    update_code: "/tiktok/operations/update-code",
  };
  return {
    endpoint: operationMap[operationName] || "/execute",
    method: "POST",
  };
}

/**
 * Map sheets operation names to endpoints
 * Updated to match Backend Go routes (specific wallet/shipping endpoints)
 */
export function mapSheetsOperation(operation: string): string {
  const sheetOperationMap: Record<string, string> = {
    wallet_to_sheets: "/shopee/wallet/export-to-sheets",
    shipping_fee_to_sheets: "/shopee/shipping/export-to-sheets",
  };
  return sheetOperationMap[operation] || "/execute-sheets";
}

/**
 * Map order export to endpoints
 */
export function mapOrderExport(platform: string, orderType: string): string {
  const orderExportMap: Record<string, string> = {
    shopee_unpaid: "/orders/export/shopee/unpaid",
    shopee_cancelled: "/orders/export/shopee/cancelled",
    lazada_unpaid: "/orders/export/lazada/unpaid",
    lazada_cancelled: "/orders/export/lazada/cancelled",
    tiktok_unpaid: "/orders/export/tiktok/unpaid",
    tiktok_cancelled: "/orders/export/tiktok/cancelled",
  };

  const key = `${platform}_${orderType.toLowerCase()}`;
  return orderExportMap[key] || "/orders/export/by-platform";
}

/**
 * Check if order export key is direct mapping
 */
export function isDirectOrderExport(
  platform: string,
  orderType: string,
): boolean {
  const key = `${platform}_${orderType.toLowerCase()}`;
  const directKeys = [
    "shopee_unpaid",
    "shopee_cancelled",
    "lazada_unpaid",
    "lazada_cancelled",
    "tiktok_unpaid",
    "tiktok_cancelled",
  ];
  return directKeys.includes(key);
}

/**
 * Check if sheets operation is direct mapping
 */
export function isDirectSheetsOperation(operation: string): boolean {
  return ["wallet_to_sheets", "shipping_fee_to_sheets"].includes(operation);
}

/**
 * Get platform operation mapping
 */
export function getPlatformOperationMapping(
  operation: string,
  params: Record<string, any>,
): EndpointMapping | null {
  const operationMapping: Record<string, () => EndpointMapping> = {
    shopee_operation: () => mapShopeeOperation(params.operation),
    lazada_operation: () => mapLazadaOperation(params.operation),
    tiktok_operation: () => mapTikTokOperation(params.operation),
  };

  const mapper = operationMapping[operation];
  return mapper ? mapper() : null;
}
