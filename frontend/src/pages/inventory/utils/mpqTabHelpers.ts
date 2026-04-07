import {
	batchShopeeMpq,
	batchTiktokMpq,
	calculateTiersLocal,
	type WholesaleSettings,
} from "@/api/wholesale";

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

export const MPQ_PREVIEW_COLUMNS = [
  { title: "Platform", dataIndex: "platform", key: "platform" },
  { title: "SKU", dataIndex: "sku", key: "sku" },
  {
    title: "Base Price",
    dataIndex: "base_price",
    key: "base_price",
    render: (value: number) => formatCurrency(value),
  },
  {
    title: "Updated Price",
    dataIndex: "updated_price",
    key: "updated_price",
    render: (value: number) => formatCurrency(value),
  },
  { title: "MPQ", dataIndex: "mpq", key: "mpq" },
];

interface SkuPriceItem {
  sku: string;
  price: number;
}

export interface MpqUpdateResult {
  message: string;
  type: "success" | "warning" | "error";
}

/**
 * Execute MPQ update for Shopee and/or TikTok items.
 * Returns per-platform result for display in Alert.
 */
export async function executeMpqUpdate(
  shopeeItems: SkuPriceItem[],
  tiktokItems: SkuPriceItem[],
  enableShopee: boolean,
  enableTiktok: boolean,
  settings: WholesaleSettings,
  selectedTier: TierKey,
  selectedMinQty: number,
): Promise<MpqUpdateResult> {
  const shopeeActive = enableShopee && shopeeItems.length > 0;
  const tiktokActive = enableTiktok && tiktokItems.length > 0;

  if (!shopeeActive && !tiktokActive) {
    return { message: "No platform selected or no items available", type: "warning" };
  }

  let totalProcessed = 0;
  let totalFailed = 0;
  const summaries: string[] = [];

  // Shopee MPQ
  if (shopeeActive) {
    const shopeePayload = shopeeItems.map((item) => ({
      sku: item.sku,
      price: getAdjustedPrice(item.price, settings, selectedTier),
    }));

    const shopeeResult = await batchShopeeMpq(shopeePayload, selectedMinQty);
    const processed = readCount(shopeeResult.data, "processed");
    const failed = readCount(shopeeResult.data, "failed");
    totalProcessed += processed;
    totalFailed += failed;
    summaries.push(`Shopee: ${processed} OK, ${failed} Failed`);
  }

  // TikTok MPQ
  if (tiktokActive) {
    const tiktokPayload = tiktokItems.map((item) => ({
      sku: item.sku,
      price: getAdjustedPrice(item.price, settings, selectedTier),
    }));

    try {
      const tiktokResult = await batchTiktokMpq(tiktokPayload, selectedMinQty);
      const processed = readCount(tiktokResult.data, "processed");
      const failed = readCount(tiktokResult.data, "failed");
      totalProcessed += processed;
      totalFailed += failed;
      summaries.push(`TikTok: ${processed} OK, ${failed} Failed`);
    } catch (err) {
      const errMsg = err instanceof Error ? err.message : "TikTok MPQ failed";
      summaries.push(`TikTok Error: ${errMsg}`);
      totalFailed++;
    }
  }

  const summary = `Wholesale Sync (Min Qty: ${selectedMinQty}): ${totalProcessed} Succeeded, ${totalFailed} Failed. (${summaries.join(" | ")})`;

  if (totalFailed > 0) {
    return { message: summary, type: "warning" };
  }
  return { message: summary, type: "success" };
}
