/**
 * Column Grouping Utilities
 * Pure functions for categorizing inventory columns into groups.
 * Extracted from InventoryColumnsModal per SRP.
 */

const GROUPS = ["Product Info", "Stock", "Price", "Platform", "Other"] as const;
export type ColumnGroup = (typeof GROUPS)[number];

export interface GroupedColumnEntry {
  key: ColumnGroup;
  label: string;
  children: string[];
}

/** Detect which group a column belongs to based on its name. */
export function detectGroup(column: string): ColumnGroup {
  const n = column.toLowerCase();

  if (
    n === "name" ||
    n === "product_name" ||
    n === "sku" ||
    n === "key_value" ||
    n.includes("name") ||
    n.includes("sku")
  ) {
    return "Product Info";
  }

  if (
    n === "stock" ||
    n === "stok" ||
    n.includes("stock") ||
    n.includes("stok")
  ) {
    return "Stock";
  }

  if (
    n === "price" ||
    n === "harga" ||
    n.includes("price") ||
    n.includes("harga")
  ) {
    return "Price";
  }

  if (
    n.startsWith("shopee_") ||
    n.startsWith("tiktok_") ||
    n.startsWith("lazada_")
  ) {
    return "Platform";
  }

  return "Other";
}

/** Group columns by category, returning only non-empty groups. */
export function buildGroups(columns: string[]): GroupedColumnEntry[] {
  const grouped: Record<ColumnGroup, string[]> = {
    "Product Info": [],
    Stock: [],
    Price: [],
    Platform: [],
    Other: [],
  };

  for (const column of columns) {
    grouped[detectGroup(column)].push(column);
  }

  return GROUPS.map((group) => ({
    key: group,
    label: `${group} (${grouped[group].length})`,
    children: grouped[group],
  })).filter((g) => g.children.length > 0);
}
