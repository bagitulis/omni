/**
 * Marketplace stock allocation logic.
 * Ported from Vue useMarketplaceAllocation.ts.
 *
 * Non-AUTO mode: Shopee 60%, TikTok 30%, Lazada remainder.
 * AUTO mode: all available platforms get the total value.
 * Handles platform unavailability (redistributes to remaining).
 */

// --- Types ---

export interface MarketplaceAllocation {
  shopee: number;
  tiktok: number;
  lazada: number;
}

export interface PlatformAvailability {
  shopee?: boolean;
  tiktok?: boolean;
  lazada?: boolean;
}

// --- Helpers ---

function ceiling(value: number, significance: number = 1): number {
  return Math.ceil(value / significance) * significance;
}

// --- Core Calculation ---

/**
 * Calculate marketplace stock allocation.
 * @param total       Total sellable stock (after locked deduction).
 * @param isAuto      If true, all platforms get the full total.
 * @param availability Which platforms are available (linked).
 */
export function calculateAllocation(
  total: number,
  isAuto: boolean,
  availability?: PlatformAvailability,
): MarketplaceAllocation {
  if (total <= 0 || isNaN(total)) {
    return { shopee: 0, tiktok: 0, lazada: 0 };
  }

  const available = {
    shopee: availability?.shopee !== false,
    tiktok: availability?.tiktok !== false,
    lazada: availability?.lazada !== false,
  };

  const count = [available.shopee, available.tiktok, available.lazada].filter(
    Boolean,
  ).length;

  if (count === 0) {
    return { shopee: 0, tiktok: 0, lazada: 0 };
  }

  // AUTO mode: all available platforms get total
  if (isAuto) {
    return {
      shopee: available.shopee ? total : 0,
      tiktok: available.tiktok ? total : 0,
      lazada: available.lazada ? total : 0,
    };
  }

  let shopee = 0;
  let tiktok = 0;
  let lazada = 0;

  if (count === 3) {
    // All platforms: 60/30/10
    shopee = Math.min(ceiling(0.6 * total), total);
    const remaining = Math.max(0, total - shopee);
    tiktok = Math.min(ceiling(0.3 * total), remaining);
    lazada = Math.max(0, total - shopee - tiktok);
  } else if (count === 2) {
    if (available.shopee && available.tiktok) {
      shopee = Math.min(ceiling(0.6 * total), total);
      tiktok = Math.max(0, total - shopee);
    } else if (available.shopee && available.lazada) {
      shopee = Math.min(ceiling(0.6 * total), total);
      lazada = Math.max(0, total - shopee);
    } else if (available.tiktok && available.lazada) {
      tiktok = Math.min(ceiling(0.5 * total), total);
      lazada = Math.max(0, total - tiktok);
    }
  } else {
    // Only 1 platform
    if (available.shopee) shopee = total;
    else if (available.tiktok) tiktok = total;
    else if (available.lazada) lazada = total;
  }

  return { shopee, tiktok, lazada };
}

/**
 * React hook for marketplace allocation (optional composable wrapper).
 */
export function useMarketplaceAllocation() {
  return { calculateAllocation };
}
