import type {
  MasterProduct,
  MasterProductPlatformLink,
  MasterProductSku,
} from "@/types/product";
import type { ProductData, ProductPlatform, ProductSku } from "../types";

const PLATFORM_ORDER = ["shopee", "tiktok", "lazada"] as const;
type PlatformName = (typeof PLATFORM_ORDER)[number];

interface PlatformMetrics {
  has_presence: boolean;
  has_pending: boolean;
  has_failed: boolean;
  last_sync?: string;
}

function isPlatformName(platform: string): platform is PlatformName {
  return PLATFORM_ORDER.includes(platform as PlatformName);
}

function getLastSync(
  current: string | undefined,
  incoming: string | undefined,
) {
  if (!incoming) return current;
  if (!current) return incoming;

  const currentTime = new Date(current).getTime();
  const incomingTime = new Date(incoming).getTime();
  return incomingTime > currentTime ? incoming : current;
}

function mapSku(sku: MasterProductSku): ProductSku {
  return {
    key: sku.id ? String(sku.id) : sku.seller_sku,
    id: sku.id,
    seller_sku: sku.seller_sku,
    variant_name: sku.variant_name || "",
    variant_data: sku.variant_data,
    stock: sku.stock,
    price: sku.price,
  };
}

function buildPlatformMetrics(
  product: MasterProduct,
): Record<PlatformName, PlatformMetrics> {
  const metrics: Record<PlatformName, PlatformMetrics> = {
    shopee: { has_presence: false, has_pending: false, has_failed: false },
    tiktok: { has_presence: false, has_pending: false, has_failed: false },
    lazada: { has_presence: false, has_pending: false, has_failed: false },
  };

  for (const sku of product.skus || []) {
    for (const link of sku.platform_links || []) {
      if (!isPlatformName(link.platform)) {
        continue;
      }

      applyLinkMetrics(metrics[link.platform], link);
    }
  }

  return metrics;
}

function applyLinkMetrics(
  metrics: PlatformMetrics,
  link: MasterProductPlatformLink,
) {
  const hasPresence = Boolean(
    link.platform_item_id || link.platform_product_id || link.platform_sku_id,
  );
  if (hasPresence) {
    metrics.has_presence = true;
  }

  if (
    link.sync_status === "pending" ||
    link.sync_status === "not_synced" ||
    link.sync_status === "outdated"
  ) {
    metrics.has_pending = true;
  }

  if (link.sync_status === "failed" || link.sync_status === "error") {
    metrics.has_failed = true;
  }

  metrics.last_sync = getLastSync(metrics.last_sync, link.last_synced_at);
}

export function buildPlatformRows(product: MasterProduct): ProductPlatform[] {
  const metrics = buildPlatformMetrics(product);

  return PLATFORM_ORDER.map((platform) => {
    const info = metrics[platform];
    let status = "not_synced";

    if (info.has_failed) {
      status = "failed";
    } else if (info.has_pending) {
      status = "pending";
    } else if (info.has_presence) {
      status = "synced";
    }

    return {
      platform,
      status,
      last_sync: info.last_sync || "-",
    };
  });
}

export function mapMasterProductToProductData(
  product: MasterProduct,
): ProductData {
  return {
    id: product.id,
    title: product.title,
    description: product.description,
    images: product.images || [],
    status: product.status,
    skus: (product.skus || []).map(mapSku),
    platforms: buildPlatformRows(product),
  };
}
