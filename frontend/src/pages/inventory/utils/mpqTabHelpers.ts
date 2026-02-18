import { calculateTiersLocal, type WholesaleSettings } from "@/api/wholesale";

export type TierKey = "normal" | "tier1" | "tier2" | "tier3";

export interface PreviewRow {
  key: string;
  platform: string;
  sku: string;
  base_price: number;
  updated_price: number;
  mpq: number;
}

export const DEFAULT_SETTINGS: WholesaleSettings = {
  admin_fee: 1500,
  min_order_1: 2,
  max_order_1: 3,
  max_order_tier_3: 1000,
};

export function formatCurrency(value: number): string {
  return new Intl.NumberFormat("id-ID").format(value);
}

export function readCount(payload: unknown, field: string): number {
  if (typeof payload !== "object" || payload === null) {
    return 0;
  }

  const raw = (payload as Record<string, unknown>)[field];
  return typeof raw === "number" && Number.isFinite(raw) ? raw : 0;
}

export function getAdjustedPrice(
  price: number,
  settings: WholesaleSettings,
  selectedTier: TierKey,
): number {
  if (selectedTier === "normal") {
    return price;
  }

  const tiers = calculateTiersLocal(price, settings);
  const tierIndex =
    selectedTier === "tier1" ? 0 : selectedTier === "tier2" ? 1 : 2;
  return tiers[tierIndex]?.unit_price ?? price;
}
