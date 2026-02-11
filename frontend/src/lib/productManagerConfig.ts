export interface PlatformConfig {
  platform: "lazada" | "shopee" | "tiktok";
  columnFields: string[];
  columnLabels: Record<string, string>;
  availableStatuses: string[];
  filterableFields: string[];
  searchFields: string[];
  apiEndpoint: string;
}

const lazadaConfig: PlatformConfig = {
  platform: "lazada",
  columnFields: [
    "item_id",
    "sku_id",
    "sku_name",
    "item_name",
    "variant_name",
    "price",
    "quantity",
    "status",
    "updated_at",
  ],
  columnLabels: {
    item_id: "Item ID",
    sku_id: "SKU ID",
    sku_name: "SKU Name",
    item_name: "Product Name",
    variant_name: "Variant Name",
    price: "Price",
    quantity: "Stock",
    status: "Status",
    updated_at: "Updated",
  },
  availableStatuses: ["ACTIVE", "INACTIVE", "DELISTED", "UNKNOWN"],
  filterableFields: [
    "item_id",
    "sku_id",
    "sku_name",
    "item_name",
    "variant_name",
    "price",
    "quantity",
    "status",
  ],
  searchFields: ["item_id", "sku_id", "sku_name", "item_name", "variant_name"],
  apiEndpoint: "/product",
};

const shopeeConfig: PlatformConfig = {
  platform: "shopee",
  columnFields: [
    "item_id",
    "model_id",
    "sku",
    "item_name",
    "sku_name",
    "price",
    "stock",
    "status",
    "updated_at",
  ],
  columnLabels: {
    item_id: "Item ID",
    model_id: "Model ID",
    sku: "SKU ID",
    item_name: "Product Name",
    sku_name: "Model Name",
    price: "Current Price",
    stock: "Seller Stock",
    status: "Status",
    updated_at: "Updated",
  },
  availableStatuses: ["NORMAL", "BANNED", "DELETED", "UNLIST"],
  filterableFields: [
    "item_id",
    "model_id",
    "sku",
    "item_name",
    "sku_name",
    "price",
    "stock",
    "status",
  ],
  searchFields: ["item_id", "model_id", "sku", "item_name", "sku_name"],
  apiEndpoint: "/product",
};

const tiktokConfig: PlatformConfig = {
  platform: "tiktok",
  columnFields: [
    "product_id",
    "sku_id",
    "seller_sku",
    "item_name",
    "variant_name",
    "price",
    "quantity",
    "status",
    "updated_at",
  ],
  columnLabels: {
    product_id: "Product ID",
    sku_id: "SKU ID",
    seller_sku: "Seller SKU",
    item_name: "Product Name",
    variant_name: "Variant Name",
    price: "Price",
    quantity: "Stock",
    status: "Status",
    updated_at: "Updated",
  },
  availableStatuses: [
    "DRAFT",
    "ACTIVATE",
    "SELLER_DEACTIVATED",
    "PLATFORM_DEACTIVATED",
    "FREEZE",
    "DELETED",
  ],
  filterableFields: [
    "product_id",
    "sku_id",
    "seller_sku",
    "item_name",
    "variant_name",
    "price",
    "quantity",
    "status",
  ],
  searchFields: [
    "product_id",
    "sku_id",
    "seller_sku",
    "item_name",
    "variant_name",
  ],
  apiEndpoint: "/product",
};

export const platformConfigs: Record<string, PlatformConfig> = {
  lazada: lazadaConfig,
  shopee: shopeeConfig,
  tiktok: tiktokConfig,
};

export function getPlatformConfig(platform: string): PlatformConfig {
  return platformConfigs[platform] || lazadaConfig;
}
