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
 */
export function mapShopeeOperation(operationName: string): EndpointMapping {
  const operationMap: Record<string, string> = {
    update_stock: "/shopee/operations/update-stock",
    update_price: "/shopee/operations/update-price",
    export_orders: "/shopee/operations/export-orders",
    get_token: "/shopee/operations/get-token",
    refresh_token: "/shopee/operations/refresh-token",
    update_code: "/shopee/operations/update-code",
  };
  return {
    endpoint: operationMap[operationName] || "/execute",
    method: "POST",
  };
}

/**
 * Map Lazada operation names to endpoints
 */
export function mapLazadaOperation(operationName: string): EndpointMapping {
  const operationMap: Record<string, string> = {
    update_stock: "/lazada/operations/update-stock",
    update_price: "/lazada/operations/update-price",
    export_orders: "/lazada/operations/export-orders",
    get_token: "/lazada/operations/get-token",
    refresh_token: "/lazada/operations/refresh-token",
    update_code: "/lazada/operations/update-code",
  };
  return {
    endpoint: operationMap[operationName] || "/execute",
    method: "POST",
  };
}

/**
 * Map TikTok operation names to endpoints
 */
export function mapTikTokOperation(operationName: string): EndpointMapping {
  const operationMap: Record<string, string> = {
    update_stock: "/tiktok/operations/update-stock",
    update_price: "/tiktok/operations/update-price",
    export_orders: "/tiktok/operations/export-orders",
    get_token: "/tiktok/operations/get-token",
    refresh_token: "/tiktok/operations/refresh-token",
    update_code: "/tiktok/operations/update-code",
  };
  return {
    endpoint: operationMap[operationName] || "/execute",
    method: "POST",
  };
}

/**
 * Map sheets operation names to endpoints
 */
export function mapSheetsOperation(operation: string): string {
  const sheetOperationMap: Record<string, string> = {
    wallet_to_sheets: "/sheets/operations/wallet-to-sheets",
    shipping_fee_to_sheets: "/sheets/operations/shipping-fee-to-sheets",
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
  orderType: string
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
  params: Record<string, any>
): EndpointMapping | null {
  const operationMapping: Record<string, () => EndpointMapping> = {
    shopee_operation: () => mapShopeeOperation(params.operation),
    lazada_operation: () => mapLazadaOperation(params.operation),
    tiktok_operation: () => mapTikTokOperation(params.operation),
  };

  const mapper = operationMapping[operation];
  return mapper ? mapper() : null;
}
