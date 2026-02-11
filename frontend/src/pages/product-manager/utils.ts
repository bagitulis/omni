import type {
  DbProductRow,
  ProductManagerPlatform,
} from "@/types/product_manager";

export const PLATFORMS: Array<{ key: ProductManagerPlatform; label: string }> =
  [
    { key: "shopee", label: "Shopee" },
    { key: "lazada", label: "Lazada" },
    { key: "tiktok", label: "TikTok" },
  ];

export function coercePlatform(
  input: string | undefined,
): ProductManagerPlatform {
  if (input === "lazada" || input === "tiktok" || input === "shopee") {
    return input;
  }
  return "shopee";
}

export function getRowKey(
  platform: ProductManagerPlatform,
  row: DbProductRow,
): string {
  const itemId = typeof row.item_id === "string" ? row.item_id : undefined;
  const modelId = typeof row.model_id === "string" ? row.model_id : undefined;
  const skuId = typeof row.sku_id === "string" ? row.sku_id : undefined;
  const productId =
    typeof row.product_id === "string" ? row.product_id : undefined;

  if (platform === "shopee") {
    return modelId
      ? `${itemId ?? ""}-${modelId}`
      : (itemId ?? JSON.stringify(row));
  }

  if (platform === "lazada") {
    return skuId ? `${itemId ?? ""}-${skuId}` : (itemId ?? JSON.stringify(row));
  }

  return skuId
    ? `${productId ?? ""}-${skuId}`
    : (productId ?? JSON.stringify(row));
}

export function formatImageSrc(value: unknown): string | null {
  if (typeof value === "string" && value.trim() !== "") {
    return value;
  }

  if (Array.isArray(value) && value.length > 0) {
    const first = value[0];
    if (typeof first === "string" && first.trim() !== "") {
      return first;
    }
  }

  return null;
}

export function getProductImage(row: DbProductRow): string | null {
  const fromLocal = formatImageSrc(row.local_images);
  if (fromLocal) return fromLocal;

  const fromImage = formatImageSrc(row.image);
  if (fromImage) return fromImage;

  return null;
}

export function extractSku(row: DbProductRow): string {
  return typeof row.sku === "string"
    ? row.sku
    : typeof row.sku_name === "string"
      ? row.sku_name
      : typeof row.seller_sku === "string"
        ? row.seller_sku
        : typeof row.shop_sku === "string"
          ? row.shop_sku
          : "";
}
