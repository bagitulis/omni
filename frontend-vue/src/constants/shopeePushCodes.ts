/**
 * Shopee Webhook Push Codes Mapping
 * Reference: https://open.shopee.com/push-mechanism/
 *
 * This maps Shopee push codes to human-readable event names
 * Used for displaying webhook logs in the frontend
 */

export interface PushCodeInfo {
  name: string;
  category: string;
  description: string;
  icon: string;
}

/**
 * Complete mapping of Shopee Push Codes
 */
export const SHOPEE_PUSH_CODES: Record<number, PushCodeInfo> = {
  // ============ Shopee Push ============
  1: {
    name: "Shop Authorization",
    category: "Shopee",
    description: "Shop authorization notification",
    icon: "🔐",
  },
  2: {
    name: "Auth Canceled",
    category: "Shopee",
    description: "Shop authorization canceled",
    icon: "🚫",
  },
  5: {
    name: "Shopee Updates",
    category: "Shopee",
    description: "Shopee platform updates",
    icon: "📢",
  },
  12: {
    name: "Auth Expiry",
    category: "Shopee",
    description: "API authorization expiry warning",
    icon: "⏰",
  },
  28: {
    name: "Penalty Update",
    category: "Shopee",
    description: "Shop penalty update notification",
    icon: "⚠️",
  },
  38: {
    name: "Video Result",
    category: "Shopee",
    description: "Video upload result notification",
    icon: "🎬",
  },

  // ============ Order Push ============
  3: {
    name: "Order Status",
    category: "Order",
    description: "Order status update",
    icon: "📦",
  },
  4: {
    name: "Tracking No",
    category: "Order",
    description: "Order tracking number update",
    icon: "🚚",
  },
  15: {
    name: "Shipping Doc",
    category: "Order",
    description: "Shipping document status update",
    icon: "📄",
  },
  23: {
    name: "Booking Status",
    category: "Order",
    description: "Booking status update",
    icon: "📅",
  },
  24: {
    name: "Booking Track",
    category: "Order",
    description: "Booking tracking number update",
    icon: "📍",
  },
  25: {
    name: "Booking Doc",
    category: "Order",
    description: "Booking shipping document status",
    icon: "📋",
  },
  30: {
    name: "Fulfillment",
    category: "Order",
    description: "Package fulfillment status update",
    icon: "✅",
  },
  37: {
    name: "Courier Binding",
    category: "Order",
    description: "Courier delivery binding status",
    icon: "🔗",
  },
  47: {
    name: "Package Info",
    category: "Order",
    description: "Package information update",
    icon: "📬",
  },

  // ============ Product Push ============
  6: {
    name: "Item Banned",
    category: "Product",
    description: "Item banned notification",
    icon: "🚫",
  },
  8: {
    name: "Stock Change",
    category: "Product",
    description: "Reserved stock change",
    icon: "📊",
  },
  11: {
    name: "Video Upload",
    category: "Product",
    description: "Video upload notification",
    icon: "🎥",
  },
  13: {
    name: "Brand Result",
    category: "Product",
    description: "Brand registration result",
    icon: "🏷️",
  },
  16: {
    name: "Violation",
    category: "Product",
    description: "Item violation notification",
    icon: "⚠️",
  },
  22: {
    name: "Price Update",
    category: "Product",
    description: "Item price update",
    icon: "💰",
  },
  27: {
    name: "Publish Failed",
    category: "Product",
    description: "Scheduled publish failed",
    icon: "❌",
  },

  // ============ Marketing Push ============
  7: {
    name: "Promotion",
    category: "Marketing",
    description: "Item promotion notification",
    icon: "🎯",
  },
  9: {
    name: "Promo Update",
    category: "Marketing",
    description: "Promotion update notification",
    icon: "📣",
  },

  // ============ Return Push ============
  29: {
    name: "Return Update",
    category: "Return",
    description: "Return/refund update",
    icon: "↩️",
  },

  // ============ Webchat Push ============
  10: {
    name: "Chat",
    category: "Webchat",
    description: "Chat message notification",
    icon: "💬",
  },

  // ============ FBS Push ============
  31: {
    name: "FBS Invoice",
    category: "FBS",
    description: "FBS Brazil invoice issued",
    icon: "🧾",
  },
  33: {
    name: "FBS Error",
    category: "FBS",
    description: "FBS Brazil invoice error",
    icon: "❗",
  },
  34: {
    name: "FBS Shop Block",
    category: "FBS",
    description: "FBS Brazil shop blocked",
    icon: "🔒",
  },
  35: {
    name: "FBS SKU Block",
    category: "FBS",
    description: "FBS Brazil SKU blocked",
    icon: "🔒",
  },
  36: {
    name: "FBS Stock",
    category: "FBS",
    description: "FBS sellable stock update",
    icon: "📦",
  },
};

/**
 * Get push code info by code number or name
 */
export function getPushCodeInfo(codeOrName: number | string): PushCodeInfo {
  // If it's a number, lookup directly
  if (typeof codeOrName === "number") {
    return (
      SHOPEE_PUSH_CODES[codeOrName] || {
        name: `Unknown (${codeOrName})`,
        category: "Unknown",
        description: `Unknown push code: ${codeOrName}`,
        icon: "❓",
      }
    );
  }

  // If it's a string, check if it's a numeric string first
  const numCode = parseInt(codeOrName, 10);
  if (!isNaN(numCode) && SHOPEE_PUSH_CODES[numCode]) {
    return SHOPEE_PUSH_CODES[numCode];
  }

  // Otherwise, search by name pattern (for new format like "package_fulfillment_status_push")
  for (const [, info] of Object.entries(SHOPEE_PUSH_CODES)) {
    const snakeCaseName = info.name.toLowerCase().replace(/\s+/g, "_");
    if (
      codeOrName.toLowerCase().includes(snakeCaseName) ||
      snakeCaseName.includes(codeOrName.toLowerCase().replace(/_push$/, ""))
    ) {
      return info;
    }
  }

  // Return the original string as name if not found
  return {
    name: codeOrName,
    category: "Unknown",
    description: codeOrName,
    icon: "📨",
  };
}

/**
 * Format event type for display
 * Handles both numeric codes (legacy) and string names (new format)
 */
export function formatEventType(eventType: string | null): string {
  if (!eventType || eventType === "unknown") return "Unknown";

  const info = getPushCodeInfo(eventType);
  return `${info.icon} ${info.name}`;
}

/**
 * Get category badge color
 */
export function getCategoryColor(category: string): string {
  const colors: Record<string, string> = {
    Order: "#4CAF50",
    Product: "#2196F3",
    Marketing: "#FF9800",
    Shopee: "#EE4D2D",
    Return: "#9C27B0",
    Webchat: "#00BCD4",
    FBS: "#795548",
    Unknown: "#9E9E9E",
  };
  return colors[category] || colors.Unknown;
}
