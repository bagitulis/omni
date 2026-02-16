import type { InventoryRecord } from "@/types/inventory";

export type InventoryColumnFilters = Record<string, string>;

function normalizeFilterValue(value: unknown): string {
  if (value == null) {
    return "";
  }

  if (typeof value === "string") {
    return value;
  }

  if (typeof value === "number" || typeof value === "boolean") {
    return String(value);
  }

  if (Array.isArray(value)) {
    return value.map((item) => normalizeFilterValue(item)).join(" ");
  }

  if (typeof value === "object") {
    return JSON.stringify(value);
  }

  return "";
}

export function applyInventoryColumnFilters(
  records: InventoryRecord[],
  columnFilters: InventoryColumnFilters,
): InventoryRecord[] {
  const activeFilters = Object.entries(columnFilters)
    .map(([column, value]) => [column, value.trim()] as const)
    .filter(([, value]) => value !== "");

  if (activeFilters.length === 0) {
    return records;
  }

  return records.filter((record) =>
    activeFilters.every(([column, value]) => {
      const recordValue = normalizeFilterValue(record.data?.[column]);
      return recordValue.toLowerCase().includes(value.toLowerCase());
    }),
  );
}
