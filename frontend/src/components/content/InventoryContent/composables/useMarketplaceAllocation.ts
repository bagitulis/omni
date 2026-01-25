import { useMarketplaceSettings } from "./useMarketplaceSettings";
import { parseItemData } from "./useInventoryItemHelpers";

/**
 * useMarketplaceAllocation Composable
 * RESPONSIBILITY: Calculate marketplace stock allocation
 * Based on Excel formula:
 * - SHOPEE: MIN(CEILING(0.6 * Total, 1), Total)
 * - TIKTOK: MIN(CEILING(0.3 * Total, 1), MAX(0, Total - Shopee))
 * - LAZADA: Total - Shopee - Tiktok
 *
 * If Auto column is TRUE, all platforms get Total value
 */

export interface MarketplaceAllocation {
  shopee: number;
  tiktok: number;
  lazada: number;
}

export interface AllocationInput {
  total: number | null;
  isAuto: boolean;
  shopeeOverride?: number;
  tiktokOverride?: number;
}

/**
 * Math.ceil equivalent - rounds up to nearest integer
 */
function ceiling(value: number, significance: number = 1): number {
  return Math.ceil(value / significance) * significance;
}

/**
 * Parse auto column value to boolean
 */
function parseAutoValue(value: any): boolean {
  if (value === null || value === undefined) return false;
  if (typeof value === "boolean") return value;
  const strVal = String(value).toLowerCase().trim();
  return strVal === "true" || strVal === "1" || strVal === "yes";
}

/**
 * Calculate marketplace allocation (standalone function)
 * Can be used without Vue composable context
 * Supports platform availability check - if platform is not available (red), set to 0 and redistribute
 */
function calculateAllocationStandalone(
  input: AllocationInput,
  platformAvailability?: {
    shopee?: boolean;
    tiktok?: boolean;
    lazada?: boolean;
  },
): MarketplaceAllocation {
  const { total, isAuto } = input;

  // Handle null/undefined total
  if (total === null || total === undefined || isNaN(total)) {
    return { shopee: 0, tiktok: 0, lazada: 0 };
  }

  // Default: all platforms available if not specified
  const available = {
    shopee: platformAvailability?.shopee !== false,
    tiktok: platformAvailability?.tiktok !== false,
    lazada: platformAvailability?.lazada !== false,
  };

  // Count available platforms
  const availableCount = [
    available.shopee,
    available.tiktok,
    available.lazada,
  ].filter(Boolean).length;

  // If no platforms available, return 0 for all
  if (availableCount === 0) {
    return { shopee: 0, tiktok: 0, lazada: 0 };
  }

  // If Auto is TRUE, all AVAILABLE platforms get total value
  if (isAuto) {
    return {
      shopee: available.shopee ? total : 0,
      tiktok: available.tiktok ? total : 0,
      lazada: available.lazada ? total : 0,
    };
  }

  // Calculate allocations based on ratios, considering availability
  const shopeeRatio = 0.6;
  const tiktokRatio = 0.3;

  let shopee = 0;
  let tiktok = 0;
  let lazada = 0;

  // Case 1: All platforms available - normal allocation
  if (availableCount === 3) {
    shopee = Math.min(ceiling(shopeeRatio * total, 1), total);
    const remaining = Math.max(0, total - shopee);
    tiktok = Math.min(ceiling(tiktokRatio * total, 1), remaining);
    lazada = Math.max(0, total - shopee - tiktok);
  }
  // Case 2: Only 2 platforms available
  else if (availableCount === 2) {
    if (available.shopee && available.tiktok) {
      // Shopee + Tiktok: split 60/40
      shopee = Math.min(ceiling(0.6 * total, 1), total);
      tiktok = Math.max(0, total - shopee);
      lazada = 0;
    } else if (available.shopee && available.lazada) {
      // Shopee + Lazada: split 60/40
      shopee = Math.min(ceiling(0.6 * total, 1), total);
      lazada = Math.max(0, total - shopee);
      tiktok = 0;
    } else if (available.tiktok && available.lazada) {
      // Tiktok + Lazada: split 50/50
      tiktok = Math.min(ceiling(0.5 * total, 1), total);
      lazada = Math.max(0, total - tiktok);
      shopee = 0;
    }
  }
  // Case 3: Only 1 platform available
  else if (availableCount === 1) {
    if (available.shopee) {
      shopee = total;
    } else if (available.tiktok) {
      tiktok = total;
    } else if (available.lazada) {
      lazada = total;
    }
  }

  return { shopee, tiktok, lazada };
}

/**
 * Get marketplace allocation for an inventory item (standalone function)
 * @param item - Inventory item object
 * @param totalColumn - Column name for Total value
 * @param autoColumn - Column name for Auto flag
 * @param platformAvailability - Optional platform availability check from Sync Status
 */
export function getAllocationForItem(
  item: any,
  totalColumn: string,
  autoColumn: string,
  platformAvailability?: {
    shopee?: boolean;
    tiktok?: boolean;
    lazada?: boolean;
  },
): MarketplaceAllocation {
  const itemData = parseItemData(item);

  // Get total value
  const totalRaw = itemData[totalColumn];
  const total =
    totalRaw !== null && totalRaw !== undefined
      ? parseInt(String(totalRaw), 10)
      : 0;

  // Get auto flag (TRUE/FALSE/1/0/true/false)
  const autoRaw = itemData[autoColumn];
  const isAuto = parseAutoValue(autoRaw);

  return calculateAllocationStandalone({ total, isAuto }, platformAvailability);
}

export function useMarketplaceAllocation() {
  const { shopeeRatio, tiktokRatio } = useMarketplaceSettings();

  /**
   * Calculate Shopee allocation
   * Formula: MIN(CEILING(0.6 * Total, 1), Total)
   */
  function calculateShopee(total: number): number {
    if (total <= 0) return 0;
    const ratio = shopeeRatio.value;
    const calculated = ceiling(ratio * total, 1);
    return Math.min(calculated, total);
  }

  /**
   * Calculate Tiktok allocation
   * Formula: MIN(CEILING(0.3 * Total, 1), MAX(0, Total - Shopee))
   */
  function calculateTiktok(total: number, shopeeAlloc: number): number {
    if (total <= 0) return 0;
    const ratio = tiktokRatio.value;
    const remaining = Math.max(0, total - shopeeAlloc);
    const calculated = ceiling(ratio * total, 1);
    return Math.min(calculated, remaining);
  }

  /**
   * Calculate Lazada allocation
   * Formula: Total - Shopee - Tiktok
   */
  function calculateLazada(
    total: number,
    shopeeAlloc: number,
    tiktokAlloc: number,
  ): number {
    if (total <= 0) return 0;
    return Math.max(0, total - shopeeAlloc - tiktokAlloc);
  }

  /**
   * Calculate all marketplace allocations
   * @param input - AllocationInput with total and isAuto flag
   * @param platformAvailability - Optional platform availability check from Sync Status
   * @returns MarketplaceAllocation object
   */
  function calculateAllocation(
    input: AllocationInput,
    platformAvailability?: {
      shopee?: boolean;
      tiktok?: boolean;
      lazada?: boolean;
    },
  ): MarketplaceAllocation {
    const { total, isAuto } = input;

    // Handle null/undefined total
    if (total === null || total === undefined || isNaN(total)) {
      return { shopee: 0, tiktok: 0, lazada: 0 };
    }

    // Default: all platforms available if not specified
    const available = {
      shopee: platformAvailability?.shopee !== false,
      tiktok: platformAvailability?.tiktok !== false,
      lazada: platformAvailability?.lazada !== false,
    };

    // Count available platforms
    const availableCount = [
      available.shopee,
      available.tiktok,
      available.lazada,
    ].filter(Boolean).length;

    // If no platforms available, return 0 for all
    if (availableCount === 0) {
      return { shopee: 0, tiktok: 0, lazada: 0 };
    }

    // If Auto is TRUE, all AVAILABLE platforms get total value
    if (isAuto) {
      return {
        shopee: available.shopee ? total : 0,
        tiktok: available.tiktok ? total : 0,
        lazada: available.lazada ? total : 0,
      };
    }

    // Calculate allocations based on ratios, considering availability
    let shopee = 0;
    let tiktok = 0;
    let lazada = 0;

    // Case 1: All platforms available - normal allocation
    if (availableCount === 3) {
      shopee = calculateShopee(total);
      tiktok = calculateTiktok(total, shopee);
      lazada = calculateLazada(total, shopee, tiktok);
    }
    // Case 2: Only 2 platforms available
    else if (availableCount === 2) {
      if (available.shopee && available.tiktok) {
        // Shopee + Tiktok: split 60/40
        shopee = Math.min(Math.ceil(0.6 * total), total);
        tiktok = Math.max(0, total - shopee);
        lazada = 0;
      } else if (available.shopee && available.lazada) {
        // Shopee + Lazada: split 60/40
        shopee = Math.min(Math.ceil(0.6 * total), total);
        lazada = Math.max(0, total - shopee);
        tiktok = 0;
      } else if (available.tiktok && available.lazada) {
        // Tiktok + Lazada: split 50/50
        tiktok = Math.min(Math.ceil(0.5 * total), total);
        lazada = Math.max(0, total - tiktok);
        shopee = 0;
      }
    }
    // Case 3: Only 1 platform available
    else if (availableCount === 1) {
      if (available.shopee) {
        shopee = total;
      } else if (available.tiktok) {
        tiktok = total;
      } else if (available.lazada) {
        lazada = total;
      }
    }

    return { shopee, tiktok, lazada };
  }

  /**
   * Parse item data from inventory row
   * Supports both string (Node.js) and object (Go) formats
   */
  function parseItemData(item: any): any {
    if (item?.data !== undefined && item?.data !== null) {
      if (typeof item.data === "string") {
        try {
          return JSON.parse(item.data);
        } catch {
          return item;
        }
      } else if (typeof item.data === "object") {
        return item.data;
      }
    }
    return item;
  }

  /**
   * Get allocation for an inventory item
   * @param item - Inventory item row
   * @param totalColumn - Column name for Total value
   * @param autoColumn - Column name for Auto flag
   * @param platformAvailability - Optional platform availability check from Sync Status
   */
  function getAllocationForItem(
    item: any,
    totalColumn: string,
    autoColumn: string,
    platformAvailability?: {
      shopee?: boolean;
      tiktok?: boolean;
      lazada?: boolean;
    },
  ): MarketplaceAllocation {
    const itemData = parseItemData(item);

    // Get total value
    const totalRaw = itemData[totalColumn];
    const total =
      totalRaw !== null && totalRaw !== undefined
        ? parseInt(String(totalRaw), 10)
        : 0;

    // Get auto flag (TRUE/FALSE/1/0/true/false)
    const autoRaw = itemData[autoColumn];
    const isAuto = parseAutoValue(autoRaw);

    return calculateAllocation({ total, isAuto }, platformAvailability);
  }

  /**
   * Parse auto column value to boolean
   */
  function parseAutoValue(value: any): boolean {
    if (value === null || value === undefined) return false;
    if (typeof value === "boolean") return value;
    const strVal = String(value).toLowerCase().trim();
    return strVal === "true" || strVal === "1" || strVal === "yes";
  }

  return {
    calculateAllocation,
    calculateShopee,
    calculateTiktok,
    calculateLazada,
    getAllocationForItem,
    parseAutoValue,
  };
}
