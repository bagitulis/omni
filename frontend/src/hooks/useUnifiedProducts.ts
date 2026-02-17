import { useQuery } from "@tanstack/react-query";
import { getProducts } from "@/api/products";
import type { MasterProduct } from "@/types/product";
import type {
  UnifiedProductRow,
  ProductFilterValues,
  PlatformLinkStatus,
  Platform,
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

    // Map status: inactive → archived, keep active|draft as-is
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
        platform_links: (sku.platform_links || []).map((link) => ({
          platform: link.platform,
          platform_product_id: link.platform_product_id,
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
 * Backend: active | inactive | draft
 * Output: active | archived | draft
 */
function mapProductStatus(
  status: "active" | "inactive" | "draft",
): "active" | "archived" | "draft" {
  if (status === "inactive") return "archived";
  return status;
}

/**
 * Map backend sync_status to output sync_status
 * Backend: synced | pending | failed | not_synced
 * Output: pending | synced | error | outdated
 */
function mapSyncStatus(
  backendStatus: "synced" | "pending" | "failed" | "not_synced",
): "pending" | "synced" | "error" | "outdated" {
  switch (backendStatus) {
    case "synced":
      return "synced";
    case "pending":
      return "pending";
    case "failed":
      return "error";
    case "not_synced":
      return "pending"; // Treat not_synced as pending
    default:
      return "pending";
  }
}

/**
 * Aggregate platform link status across all SKUs
 * Priority: linked > pending > error > not_linked
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
      let linkStatus: PlatformLinkStatus = "not_linked";
      if (link.sync_status === "synced") {
        linkStatus = "linked";
      } else if (
        link.sync_status === "pending" ||
        link.sync_status === "not_synced"
      ) {
        linkStatus = "pending";
      } else if (link.sync_status === "failed") {
        linkStatus = "error";
      }

      // Priority: linked > pending > error > not_linked
      statusMap[platform] = getHigherPriorityStatus(currentStatus, linkStatus);
    });
  });

  return statusMap;
}

/**
 * Determine higher priority status
 * Priority: linked > pending > error > not_linked
 */
function getHigherPriorityStatus(
  current: PlatformLinkStatus,
  incoming: PlatformLinkStatus,
): PlatformLinkStatus {
  const priority: PlatformLinkStatus[] = [
    "linked",
    "pending",
    "error",
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
  const { search, platform, status } = filters;

  return useQuery({
    queryKey: ["unified-products", filters, page, pageSize],
    queryFn: async () => {
      const response = await getProducts({
        page,
        limit: pageSize,
        search: search || undefined,
        status: status !== "all" ? status : undefined,
        platform: platform !== "all" ? platform : undefined,
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
