import type { ApiResponse } from "./client";

/**
 * Wholesale Helper Functions
 * Pure utility functions for wholesale data processing.
 * Separated from API calls per SRP principle.
 */

// --- Internal Types ---

interface LegacyWholesaleSettings {
  min_qty_1?: number;
  min_qty_2?: number;
  min_qty_3?: number;
}

interface SettingsEnvelope {
  settings?: unknown;
}

interface TiktokMpqProduct {
  product_id: string;
  sku_id?: string;
  mpq: number;
}

// --- Wholesale Settings Shape (API-layer specific) ---

export interface WholesaleSettingsApi {
  admin_fee: number;
  min_order_1: number;
  max_order_1: number;
  max_order_tier_3: number;
}

export const DEFAULT_SETTINGS: WholesaleSettingsApi = {
  admin_fee: 0,
  min_order_1: 1,
  max_order_1: 1,
  max_order_tier_3: 3,
};

// --- Type Guards ---

export function isRecord(value: unknown): value is Record<string, unknown> {
  return typeof value === "object" && value !== null;
}

export function toPositiveInt(value: unknown): number | null {
  if (typeof value !== "number" || !Number.isFinite(value)) {
    return null;
  }

  const parsed = Math.trunc(value);
  return parsed > 0 ? parsed : null;
}

export function getMessage(error: unknown, fallback: string): string {
  return error instanceof Error && error.message ? error.message : fallback;
}

// --- Settings Normalization ---

export function normalizeWholesaleSettings(
  payload: unknown,
): WholesaleSettingsApi | null {
  if (!isRecord(payload)) {
    return null;
  }

  const adminFee = payload.admin_fee;
  const minOrder1 = payload.min_order_1;
  const maxOrder1 = payload.max_order_1;
  const maxOrderTier3 = payload.max_order_tier_3;

  if (
    typeof adminFee === "number" &&
    typeof minOrder1 === "number" &&
    typeof maxOrder1 === "number" &&
    typeof maxOrderTier3 === "number"
  ) {
    return {
      admin_fee: adminFee,
      min_order_1: minOrder1,
      max_order_1: maxOrder1,
      max_order_tier_3: maxOrderTier3,
    };
  }

  const legacy = payload as LegacyWholesaleSettings;
  const minQty1 = toPositiveInt(legacy.min_qty_1);
  const minQty2 = toPositiveInt(legacy.min_qty_2);
  const minQty3 = toPositiveInt(legacy.min_qty_3);

  if (minQty1 === null || minQty2 === null || minQty3 === null) {
    return null;
  }

  const maxOrder1Value = Math.max(minQty1, minQty2 - 1);
  const maxOrderTier3Value = Math.max(maxOrder1Value + 2, minQty3);

  return {
    admin_fee: DEFAULT_SETTINGS.admin_fee,
    min_order_1: minQty1,
    max_order_1: maxOrder1Value,
    max_order_tier_3: maxOrderTier3Value,
  };
}

export function extractSettingsPayload(
  response: ApiResponse<WholesaleSettingsApi>,
): unknown {
  if (response.data !== undefined) {
    return response.data;
  }

  if (isRecord(response)) {
    return (response as SettingsEnvelope).settings;
  }

  return null;
}

// --- TikTok MPQ Builder ---

export interface BatchUpdateItemInput {
  sku: string;
  item_id?: number;
  price?: number;
}

export function buildTiktokMpqProducts(
  items: BatchUpdateItemInput[],
  mpq: number,
): TiktokMpqProduct[] {
  const products: TiktokMpqProduct[] = [];

  for (const item of items) {
    const sku = item.sku.trim();
    const productId = item.item_id ? String(item.item_id) : sku;

    if (!productId) {
      continue;
    }

    const product: TiktokMpqProduct = {
      product_id: productId,
      mpq,
    };

    if (sku) {
      product.sku_id = sku;
    }

    products.push(product);
  }

  return products;
}
