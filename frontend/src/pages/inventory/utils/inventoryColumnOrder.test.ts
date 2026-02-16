import { describe, expect, it } from "vitest";
import {
  buildInventoryColumnControlOrder,
  moveInventoryColumn,
  toggleInventoryColumn,
} from "./inventoryColumnOrder";

describe("inventoryColumnOrder utilities", () => {
  it("adds toggled column to the end without duplicates", () => {
    expect(toggleInventoryColumn(["SKU", "Stock"], "Price", true)).toEqual([
      "SKU",
      "Stock",
      "Price",
    ]);

    expect(toggleInventoryColumn(["SKU", "Stock"], "Stock", true)).toEqual([
      "SKU",
      "Stock",
    ]);
  });

  it("removes toggled column while preserving remaining order", () => {
    expect(
      toggleInventoryColumn(["SKU", "Stock", "Price"], "Stock", false),
    ).toEqual(["SKU", "Price"]);
  });

  it("moves column up and down with boundary no-op", () => {
    expect(
      moveInventoryColumn(["SKU", "Stock", "Price"], "Stock", "up"),
    ).toEqual(["Stock", "SKU", "Price"]);
    expect(
      moveInventoryColumn(["SKU", "Stock", "Price"], "Stock", "down"),
    ).toEqual(["SKU", "Price", "Stock"]);

    expect(moveInventoryColumn(["SKU", "Stock"], "SKU", "up")).toEqual([
      "SKU",
      "Stock",
    ]);
    expect(moveInventoryColumn(["SKU", "Stock"], "Stock", "down")).toEqual([
      "SKU",
      "Stock",
    ]);
  });

  it("orders controls with visible first and hidden next", () => {
    expect(
      buildInventoryColumnControlOrder(
        ["SKU", "Stock", "Price", "Category"],
        ["Price", "SKU", "Unknown"],
      ),
    ).toEqual(["Price", "SKU", "Unknown", "Stock", "Category"]);
  });

  it("keeps visible columns when available columns endpoint is empty", () => {
    expect(buildInventoryColumnControlOrder([], ["SKU", "Price"])).toEqual([
      "SKU",
      "Price",
    ]);
  });
});
