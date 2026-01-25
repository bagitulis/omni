/**
 * Shopee Webhook Push Codes Mapping
 * Reference: https://open.shopee.com/push-mechanism/
 *
 * This maps Shopee push codes to human-readable event names
 */

export interface PushCodeInfo {
  name: string;
  category: string;
  description: string;
}

/**
 * Complete mapping of Shopee Push Codes
 * Based on official Shopee documentation
 */
export const SHOPEE_PUSH_CODES: Record<number, PushCodeInfo> = {
  // ============ Shopee Push ============
  1: {
    name: "shop_authorization_push",
    category: "Shopee",
    description: "Shop authorization notification",
  },
  2: {
    name: "shop_authorization_canceled_push",
    category: "Shopee",
    description: "Shop authorization canceled",
  },
  5: {
    name: "shopee_updates",
    category: "Shopee",
    description: "Shopee platform updates",
  },
  12: {
    name: "open_api_authorization_expiry",
    category: "Shopee",
    description: "API authorization expiry warning",
  },
  28: {
    name: "shop_penalty_update_push",
    category: "Shopee",
    description: "Shop penalty update notification",
  },
  38: {
    name: "video_upload_result_push",
    category: "Shopee",
    description: "Video upload result notification",
  },

  // ============ Order Push ============
  3: {
    name: "order_status_push",
    category: "Order",
    description: "Order status update",
  },
  4: {
    name: "order_trackingno_push",
    category: "Order",
    description: "Order tracking number update",
  },
  15: {
    name: "shipping_document_status_push",
    category: "Order",
    description: "Shipping document status update",
  },
  23: {
    name: "booking_status_push",
    category: "Order",
    description: "Booking status update",
  },
  24: {
    name: "booking_trackingno_push",
    category: "Order",
    description: "Booking tracking number update",
  },
  25: {
    name: "booking_shipping_document_status_push",
    category: "Order",
    description: "Booking shipping document status",
  },
  30: {
    name: "package_fulfillment_status_push",
    category: "Order",
    description: "Package fulfillment status update",
  },
  37: {
    name: "courier_delivery_binding_status_push",
    category: "Order",
    description: "Courier delivery binding status",
  },
  47: {
    name: "package_info_push",
    category: "Order",
    description: "Package information update",
  },

  // ============ Product Push ============
  6: {
    name: "banned_item_push",
    category: "Product",
    description: "Item banned notification",
  },
  8: {
    name: "reserved_stock_change_push",
    category: "Product",
    description: "Reserved stock change",
  },
  11: {
    name: "video_upload_push",
    category: "Product",
    description: "Video upload notification",
  },
  13: {
    name: "brand_register_result",
    category: "Product",
    description: "Brand registration result",
  },
  16: {
    name: "violation_item_push",
    category: "Product",
    description: "Item violation notification",
  },
  22: {
    name: "item_price_update_push",
    category: "Product",
    description: "Item price update",
  },
  27: {
    name: "item_scheduled_publish_failed_push",
    category: "Product",
    description: "Scheduled publish failed",
  },

  // ============ Marketing Push ============
  7: {
    name: "item_promotion_push",
    category: "Marketing",
    description: "Item promotion notification",
  },
  9: {
    name: "promotion_update_push",
    category: "Marketing",
    description: "Promotion update notification",
  },

  // ============ Return Push ============
  29: {
    name: "return_updates_push",
    category: "Return",
    description: "Return/refund update",
  },

  // ============ Webchat Push ============
  10: {
    name: "webchat_push",
    category: "Webchat",
    description: "Chat message notification",
  },

  // ============ Fulfillment by Shopee (FBS) Push ============
  31: {
    name: "fbs_br_invoice_issued_push",
    category: "FBS",
    description: "FBS Brazil invoice issued",
  },
  33: {
    name: "fbs_br_invoice_error_push",
    category: "FBS",
    description: "FBS Brazil invoice error",
  },
  34: {
    name: "fbs_br_block_shop_push",
    category: "FBS",
    description: "FBS Brazil shop blocked",
  },
  35: {
    name: "fbs_br_block_sku_push",
    category: "FBS",
    description: "FBS Brazil SKU blocked",
  },
  36: {
    name: "fbs_sellable_stock",
    category: "FBS",
    description: "FBS sellable stock update",
  },
};

/**
 * Get push code info by code number
 */
export function getPushCodeInfo(code: number): PushCodeInfo {
  return (
    SHOPEE_PUSH_CODES[code] || {
      name: `unknown_push_${code}`,
      category: "Unknown",
      description: `Unknown push code: ${code}`,
    }
  );
}

/**
 * Get human-readable event name from push code
 */
export function getPushCodeName(code: number): string {
  const info = getPushCodeInfo(code);
  return info.name;
}

/**
 * Get category from push code
 */
export function getPushCodeCategory(code: number): string {
  const info = getPushCodeInfo(code);
  return info.category;
}

/**
 * Format push code for display (e.g., "30 - Package Fulfillment")
 */
export function formatPushCodeDisplay(code: number): string {
  const info = getPushCodeInfo(code);
  return `${info.name} (${code})`;
}
