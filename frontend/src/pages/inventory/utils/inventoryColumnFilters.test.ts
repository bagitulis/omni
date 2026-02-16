import { describe, expect, it } from "vitest";
import type { InventoryRecord } from "@/types/inventory";
import { applyInventoryColumnFilters } from "./inventoryColumnFilters";

const records: InventoryRecord[] = [
  {
    id: "1",
    key_value: "SKU-1",
    key_column_name: "SKU",
    data: {
      SKU: "SKU-1",
      Name: "Blue Shirt",
      Category: "Fashion",
      Stock: 12,
      Price: 120000,
    },
    sync_status: "synced",
    platform_status: [],
    created_at: "2026-02-01T00:00:00Z",
    updated_at: "2026-02-01T00:00:00Z",
  },
  {
    id: "2",
    key_value: "SKU-2",
    key_column_name: "SKU",
    data: {
      SKU: "SKU-2",
      Name: "Black Pants",
      Category: "Fashion",
      Stock: 5,
      Price: 150000,
    },
    sync_status: "not_synced",
    platform_status: [],
    created_at: "2026-02-01T00:00:00Z",
    updated_at: "2026-02-01T00:00:00Z",
  },
];

describe("applyInventoryColumnFilters", () => {
  it("returns all records when no active filters", () => {
    expect(applyInventoryColumnFilters(records, {})).toHaveLength(2);
    expect(applyInventoryColumnFilters(records, { Name: "   " })).toHaveLength(
      2,
    );
  });

  it("applies case-insensitive text matching per column", () => {
    const filtered = applyInventoryColumnFilters(records, {
      Name: "blue",
    });

    expect(filtered).toHaveLength(1);
    expect(filtered[0]?.key_value).toBe("SKU-1");
  });

  it("applies AND logic across multiple column filters", () => {
    const filtered = applyInventoryColumnFilters(records, {
      Category: "fashion",
      Stock: "12",
    });

    expect(filtered).toHaveLength(1);
    expect(filtered[0]?.key_value).toBe("SKU-1");
  });
});
