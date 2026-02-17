import { useMemo } from "react";
import type {
  Platform,
  PlatformIndicatorData,
  PlatformLinkStatus,
  PlatformSyncState,
  UnifiedProductRow,
} from "@/types/shared";

/**
 * Map platform link status to sync state for indicator
 * linked → success, pending → syncing, error → error, not_linked → idle
 */
function mapLinkStatusToSyncState(
  status: PlatformLinkStatus,
): PlatformSyncState {
  switch (status) {
    case "linked":
      return "success";
    case "pending":
      return "syncing";
    case "error":
      return "error";
    case "not_linked":
      return "idle";
    default:
      return "idle";
  }
}

/**
 * Derive platform indicator data from a product row
 * Aggregates platform link status and metadata for display
 */
export function derivePlatformStatus(
  product: UnifiedProductRow | null | undefined,
  platform: Platform,
): PlatformIndicatorData {
  if (!product) {
    return {
      platform,
      linked: false,
      has_update: false,
      sync_state: "idle",
    };
  }

  const linkStatus = product.platform_summary[platform];
  const linked = linkStatus === "linked";
  const syncState = mapLinkStatusToSyncState(linkStatus);

  // Check if any SKU has platform link for this platform
  let platformProductId: string | undefined;
  let platformSkuId: string | undefined;
  let lastSyncedAt: string | undefined;
  let errorMessage: string | undefined;

  // Iterate SKUs to find platform link metadata
  for (const sku of product.skus) {
    const link = sku.platform_links?.find((l) => l.platform === platform);
    if (link) {
      // Use first found link's metadata
      if (!platformProductId) platformProductId = link.platform_product_id;
      if (!platformSkuId) platformSkuId = link.platform_sku_id;
      if (!lastSyncedAt) lastSyncedAt = link.last_synced_at;
      if (!errorMessage && link.error_message)
        errorMessage = link.error_message;

      // If we found error message, keep it (highest priority)
      if (link.error_message) {
        errorMessage = link.error_message;
        break;
      }
    }
  }

  // Determine has_update flag (sync_status is "outdated" or pending with existing link)
  const hasUpdate = product.skus.some((sku) => {
    const link = sku.platform_links?.find((l) => l.platform === platform);
    return (
      link?.sync_status === "outdated" ||
      (link?.sync_status === "pending" && link.platform_product_id)
    );
  });

  return {
    platform,
    linked,
    has_update: hasUpdate,
    sync_state: syncState,
    platform_product_id: platformProductId,
    platform_sku_id: platformSkuId,
    last_synced_at: lastSyncedAt,
    error_message: errorMessage,
  };
}

/**
 * Hook to derive platform status indicators from product data
 * Returns stable, memoized platform indicator data for all platforms
 */
export function usePlatformStatus(
  product: UnifiedProductRow | null | undefined,
): Record<Platform, PlatformIndicatorData> {
  return useMemo(() => {
    const platforms: Platform[] = ["shopee", "tiktok", "lazada"];
    const result: Partial<Record<Platform, PlatformIndicatorData>> = {};

    for (const platform of platforms) {
      result[platform] = derivePlatformStatus(product, platform);
    }

    return result as Record<Platform, PlatformIndicatorData>;
  }, [product]);
}
