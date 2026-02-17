import type { InventoryRecord } from "@/types/inventory";

export type SupportedPlatform = "shopee" | "tiktok" | "lazada";

export interface BulkPricingItem {
  sku: string;
  price: number;
  platform: SupportedPlatform;
}

export interface BulkPricingExtractionResult {
  items: BulkPricingItem[];
  skipped_skus: string[];
}

const PRICE_CANDIDATE_KEYS = [
  "price",
  "Price",
  "PRICE",
  "harga",
  "Harga",
  "HARGA",
  "harga jual",
  "Harga Jual",
  "selling price",
  "Selling Price",
  "unit price",
  "Unit Price",
];

const PLATFORM_CANDIDATE_KEYS = [
  "platform",
  "Platform",
  "marketplace",
  "Marketplace",
];

function parseNumericValue(value: unknown): number | null {
  if (typeof value === "number") {
    return Number.isFinite(value) ? value : null;
  }

  if (typeof value !== "string") {
    return null;
  }

  const trimmed = value.trim();
  if (!trimmed) {
    return null;
  }

  const sanitized = trimmed.replace(/[^\d,.-]/g, "");
  if (!sanitized) {
    return null;
  }

  const commaCount = (sanitized.match(/,/g) ?? []).length;
  const dotCount = (sanitized.match(/\./g) ?? []).length;

  let normalized = sanitized;

  if (commaCount > 0 && dotCount > 0) {
    const lastComma = sanitized.lastIndexOf(",");
    const lastDot = sanitized.lastIndexOf(".");

    if (lastComma > lastDot) {
      normalized = sanitized.replace(/\./g, "").replace(",", ".");
    } else {
      normalized = sanitized.replace(/,/g, "");
    }
  } else if (commaCount > 0) {
    if (commaCount > 1) {
      normalized = sanitized.replace(/,/g, "");
    } else {
      const commaIndex = sanitized.indexOf(",");
      const fractionLength = sanitized.length - commaIndex - 1;
      normalized =
        fractionLength === 3
          ? sanitized.replace(/,/g, "")
          : sanitized.replace(",", ".");
    }
  } else if (dotCount > 0) {
    if (dotCount > 1) {
      normalized = sanitized.replace(/\./g, "");
    } else {
      const dotIndex = sanitized.indexOf(".");
      const fractionLength = sanitized.length - dotIndex - 1;
      normalized =
        fractionLength === 3 ? sanitized.replace(/\./g, "") : sanitized;
    }
  }

  const parsed = Number.parseFloat(normalized);
  return Number.isFinite(parsed) ? parsed : null;
}

function extractPrice(record: InventoryRecord): number | null {
  for (const key of PRICE_CANDIDATE_KEYS) {
    if (Object.prototype.hasOwnProperty.call(record.data, key)) {
      const parsed = parseNumericValue(record.data[key]);
      if (parsed !== null && parsed > 0) {
        return parsed;
      }
    }
  }

  return null;
}

function normalizePlatform(value: string): SupportedPlatform | null {
  const normalized = value.toLowerCase();

  if (normalized.includes("shopee")) {
    return "shopee";
  }

  if (normalized.includes("tiktok")) {
    return "tiktok";
  }

  if (normalized.includes("lazada")) {
    return "lazada";
  }

  return null;
}

function extractPlatforms(record: InventoryRecord): SupportedPlatform[] {
  const fromStatus = (record.platform_status ?? [])
    .map((status) => normalizePlatform(status.platform))
    .filter((platform): platform is SupportedPlatform => platform !== null);

  if (fromStatus.length > 0) {
    return [...new Set(fromStatus)];
  }

  for (const key of PLATFORM_CANDIDATE_KEYS) {
    const candidate = record.data[key];
    if (typeof candidate !== "string") {
      continue;
    }

    const fromCell = candidate
      .split(",")
      .map((value) => normalizePlatform(value.trim()))
      .filter((platform): platform is SupportedPlatform => platform !== null);

    if (fromCell.length > 0) {
      return [...new Set(fromCell)];
    }
  }

  return ["shopee"];
}

export function extractBulkPricingItems(
  records: InventoryRecord[],
): BulkPricingExtractionResult {
  const items: BulkPricingItem[] = [];
  const skippedSkus: string[] = [];
  const dedupe = new Set<string>();

  for (const record of records) {
    const sku = record.key_value?.trim();
    if (!sku) {
      continue;
    }

    const price = extractPrice(record);
    if (price === null) {
      skippedSkus.push(sku);
      continue;
    }

    for (const platform of extractPlatforms(record)) {
      const dedupeKey = `${sku}:${platform}`;
      if (dedupe.has(dedupeKey)) {
        continue;
      }
      dedupe.add(dedupeKey);
      items.push({ sku, price, platform });
    }
  }

  return { items, skipped_skus: skippedSkus };
}
