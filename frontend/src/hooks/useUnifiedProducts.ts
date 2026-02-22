import { useQuery } from "@tanstack/react-query";
import { getProducts } from "@/api/products";
import type { MasterProduct } from "@/types/product";
import type {
  Platform,
  PlatformLinkStatus,
  ProductFilterValues,
  UnifiedProductRow,
} from "@/types/shared";

/**
 * Transform MasterProduct[] from backend to UnifiedProductRow[] for table display
 * Aggregates all SKUs' platform links into a unified summary
 */
function transformToUnifiedRows(
  products: MasterProduct[],
): UnifiedProductRow[] {
  return products.map((product) => {
    const skus = product.skus || [];

    // Use first SKU as primary fallback
    const primarySku = skus[0] || null;

    // Aggregate platform link status across all SKUs
    const platformSummary = aggregatePlatformStatus(skus);

    // Keep backend status contract as-is: active | archived | draft
    const status = mapProductStatus(product.status);

    return {
      id: product.id,
      title: product.title,
      description: product.description,
      images: product.images,
      status,
      skus: skus.map((sku) => ({
        id: sku.id,
        seller_sku: sku.seller_sku,
        variant_name: sku.variant_name,
        price: sku.price,
        stock: sku.stock,
        platform_prices: sku.platform_prices,
        inventory_price: sku.inventory_price,
        inventory_stock: sku.inventory_stock,
        platform_links: (sku.platform_links || []).map((link) => ({
          platform: link.platform,
          platform_product_id:
            link.platform_product_id || link.platform_item_id,
          platform_item_id: link.platform_item_id,
          platform_sku_id: link.platform_sku_id,
          sync_status: mapSyncStatus(link.sync_status),
          last_synced_at: link.last_synced_at,
          error_message: link.error_message,
        })),
      })),
      primary_sku: primarySku?.seller_sku || `MP-${product.id}`,
      primary_price: primarySku?.price || 0,
      primary_stock: primarySku?.stock || 0,
      platform_summary: platformSummary,
      created_at: product.created_at,
      updated_at: product.updated_at,
    };
  });
}

/**
 * Map backend product status to output type
 * Backend: active | archived | draft
 * Output: active | archived | draft
 */
function mapProductStatus(
  status: "active" | "archived" | "draft",
): "active" | "archived" | "draft" {
  return status;
}

/**
 * Map backend sync_status to output sync_status
 * Backend: synced | pending | error | outdated
 * Legacy values still accepted: failed | not_synced
 * Output: pending | synced | error | outdated
 */
function mapSyncStatus(
  backendStatus:
    | "synced"
    | "pending"
    | "error"
    | "outdated"
    | "failed"
    | "not_synced"
    | "success"
    | "linked"
    | string
    | null
    | undefined,
): "pending" | "synced" | "error" | "outdated" {
  const normalizedStatus = String(backendStatus ?? "")
    .trim()
    .toLowerCase();

  switch (normalizedStatus) {
    case "synced":
    case "success":
    case "linked":
      return "synced";
    case "pending":
      return "pending";
    case "outdated":
      return "outdated";
    case "error":
    case "failed":
      return "error";
    case "not_synced":
      return "pending"; // Treat not_synced as pending
    default:
      if (normalizedStatus.length > 0) {
        return "pending";
      }
      return "pending";
  }
}

function toPlatformLinkStatus(
  syncStatus: "pending" | "synced" | "error" | "outdated",
): PlatformLinkStatus {
  if (syncStatus === "synced" || syncStatus === "outdated") {
    return "linked";
  }

  if (syncStatus === "pending") {
    return "pending";
  }

  return "error";
}

/**
 * Aggregate platform link status across all SKUs
 * Priority (worst-state-wins): error > pending > linked > not_linked
 */
function aggregatePlatformStatus(
  skus: Array<{
    platform_links?: Array<{
      platform: Platform;
      sync_status: string;
    }>;
  }>,
): {
  shopee: PlatformLinkStatus;
  tiktok: PlatformLinkStatus;
  lazada: PlatformLinkStatus;
} {
  const statusMap: Record<Platform, PlatformLinkStatus> = {
    shopee: "not_linked",
    tiktok: "not_linked",
    lazada: "not_linked",
  };

  // Iterate all SKUs' platform links
  skus.forEach((sku) => {
    (sku.platform_links || []).forEach((link) => {
      const platform = link.platform;
      const currentStatus = statusMap[platform];

      // Map sync_status to PlatformLinkStatus
      const linkStatus = toPlatformLinkStatus(mapSyncStatus(link.sync_status));

      // Priority (worst-state-wins): error > pending > linked > not_linked
      statusMap[platform] = getHigherPriorityStatus(currentStatus, linkStatus);
    });
  });

  return statusMap;
}

/**
 * Determine higher priority status using worst-state-wins logic
 * Priority: error > pending > linked > not_linked
 * "error" has lowest index (highest concern) and always wins.
 */
function getHigherPriorityStatus(
  current: PlatformLinkStatus,
  incoming: PlatformLinkStatus,
): PlatformLinkStatus {
  const priority: PlatformLinkStatus[] = [
    "error",
    "pending",
    "linked",
    "not_linked",
  ];
  const currentIndex = priority.indexOf(current);
  const incomingIndex = priority.indexOf(incoming);
  return incomingIndex < currentIndex ? incoming : current;
}

/**
 * Hook to fetch unified products with platform link aggregation
 * Returns transformed UnifiedProductRow[] + total count
 */
export function useUnifiedProducts(
  filters: ProductFilterValues,
  page: number,
  pageSize: number,
) {
  const { search, platform, status, mapping } = filters;

  return useQuery({
    queryKey: ["unified-products", filters, page, pageSize],
    queryFn: async () => {
      const response = await getProducts({
        page,
        limit: pageSize,
        search: search || undefined,
        status: status !== "all" ? status : undefined,
        platform: platform !== "all" ? platform : undefined,
        linked_only: mapping === "mapped" ? true : false,
        unmapped_only: mapping === "unmapped" ? true : false,
      });

      const products = transformToUnifiedRows(response.data || []);

      return {
        products,
        total: response.meta?.total || products.length,
      };
    },
    staleTime: 30_000,
  });
}
