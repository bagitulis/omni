import type { Product } from "@/types/product";
import {
  type ColumnConfig,
  getImageUrl,
  type Platform,
  type ProductFilterValues,
  type UnifiedProductRow,
} from "@/types/shared";

export const DEFAULT_FILTERS: ProductFilterValues = {
  search: "",
  platform: "all",
  status: "all",
  category: "all",
};

export const DEFAULT_PRODUCT_PAGE_COLUMNS: ColumnConfig[] = [
  { key: "image", title: "Image", visible: true, order: 0, width: 160 },
  {
    key: "name",
    title: "Name",
    visible: true,
    order: 1,
    width: 300,
    locked: true,
  },
  { key: "price", title: "Price", visible: true, order: 2, width: 140 },
  { key: "stock", title: "Stock", visible: true, order: 3, width: 120 },
  { key: "platforms", title: "Platforms", visible: true, order: 4, width: 180 },
  { key: "category", title: "Category", visible: true, order: 5, width: 140 },
  {
    key: "actions",
    title: "Actions",
    visible: true,
    order: 6,
    width: 120,
    locked: true,
  },
];

export const PLATFORMS: Platform[] = ["shopee", "tiktok", "lazada"];

export function readFiltersFromUrl(
  searchParams: URLSearchParams,
): ProductFilterValues {
  const platform = searchParams.get("platform");
  const status = searchParams.get("status");

  return {
    search: searchParams.get("search") || DEFAULT_FILTERS.search,
    platform:
      platform === "shopee" || platform === "tiktok" || platform === "lazada"
        ? platform
        : (DEFAULT_FILTERS.platform as Platform | "all"),
    status:
      status === "active" || status === "draft" || status === "archived"
        ? status
        : (DEFAULT_FILTERS.status as "active" | "draft" | "archived" | "all"),
    category: searchParams.get("category") || DEFAULT_FILTERS.category,
  };
}

export function toImageSrc(images: string[]): string | undefined {
  const firstImage = images[0];
  if (!firstImage) {
    return undefined;
  }

  if (firstImage.startsWith("http://") || firstImage.startsWith("https://")) {
    return firstImage;
  }

  // Local paths need /uploads/ prefix for static file serving
  if (firstImage.startsWith("/")) {
    if (firstImage.startsWith("/uploads/")) {
      return firstImage;
    }
    return `/uploads${firstImage}`;
  }

  return getImageUrl(firstImage, "medium");
}

export function toLegacyProduct(row: UnifiedProductRow): Product {
  const firstLinkedPlatform = PLATFORMS.find(
    (platform) => row.platform_summary[platform] === "linked",
  );

  return {
    id: row.id,
    tenant_id: "",
    title: row.title,
    description: row.description,
    images: row.images,
    status: row.status,
    created_at: row.created_at,
    updated_at: row.updated_at,
    skus: row.skus.map((sku) => ({
      id: sku.id,
      tenant_id: "",
      master_product_id: row.id,
      seller_sku: sku.seller_sku,
      variant_name: sku.variant_name,
      variant_data: {},
      price: sku.price,
      stock: sku.stock,
      created_at: row.created_at,
      updated_at: row.updated_at,
      platform_links: sku.platform_links.map((link) => ({
        id: 0,
        master_product_id: row.id,
        master_sku_id: sku.id,
        platform: link.platform,
        platform_product_id: link.platform_product_id,
        platform_sku_id: link.platform_sku_id,
        sync_status:
          link.sync_status === "error"
            ? "failed"
            : link.sync_status === "outdated"
              ? "pending"
              : link.sync_status,
        last_synced_at: link.last_synced_at,
        error_message: link.error_message,
      })),
    })),
    item_id: String(row.id),
    item_name: row.title,
    item_sku: row.primary_sku,
    price: row.primary_price,
    stock: row.primary_stock,
    platform: firstLinkedPlatform || "master",
    image_url: toImageSrc(row.images) || "",
  };
}

export function areFiltersEqual(
  a: ProductFilterValues,
  b: ProductFilterValues,
): boolean {
  return (
    a.search === b.search &&
    a.platform === b.platform &&
    a.status === b.status &&
    a.category === b.category
  );
}

export function buildFilterSearchParams(
  filters: ProductFilterValues,
): URLSearchParams {
  const params = new URLSearchParams();
  if (filters.search.trim()) params.set("search", filters.search.trim());
  if (filters.platform !== "all") params.set("platform", filters.platform);
  if (filters.status !== "all") params.set("status", filters.status);
  if (filters.category !== "all") params.set("category", filters.category);
  return params;
}

export function getErrorMessage(error: unknown): string {
  if (error instanceof Error) {
    return error.message;
  }

  if (typeof error === "string") {
    return error;
  }

  return JSON.stringify(error);
}
