import { describe, expect, it } from "vitest";
import {
  buildInventoryFilterPreferencesSignature,
  normalizeInventoryFilterPreferencesForSync,
} from "./inventoryFilterPreferenceSync";

describe("inventoryFilterPreferenceSync", () => {
  it("normalizes payload by trimming and removing empty values", () => {
    const normalized = normalizeInventoryFilterPreferencesForSync({
      visible_columns: [" SKU ", "Stock", "SKU", ""],
      locked_columns: [" Stock", "", "Stock"],
      column_filters: {
        " Price ": " 120000 ",
        Name: "",
      },
      search_query: "shirt",
    });

    expect(normalized).toEqual({
      visible_columns: ["SKU", "Stock"],
      locked_columns: ["Stock"],
      column_filters: { Price: "120000" },
      search_query: "shirt",
    });
  });

  it("returns the same signature for equivalent filter objects", () => {
    const signatureA = buildInventoryFilterPreferencesSignature({
      visible_columns: ["SKU", "Stock"],
      locked_columns: ["Stock"],
      column_filters: { Price: "120000", Name: "Blue" },
      search_query: "shirt",
    });

    const signatureB = buildInventoryFilterPreferencesSignature({
      visible_columns: ["SKU", "Stock"],
      locked_columns: ["Stock"],
      column_filters: { Name: "Blue", Price: "120000" },
      search_query: "shirt",
    });

    expect(signatureA).toBe(signatureB);
  });

  it("returns a different signature when payload values differ", () => {
    const base = {
      visible_columns: ["SKU", "Stock"],
      locked_columns: ["Stock"],
      column_filters: { Price: "120000" },
      search_query: "shirt",
    };

    const baseSignature = buildInventoryFilterPreferencesSignature(base);
    const changedSignature = buildInventoryFilterPreferencesSignature({
      ...base,
      search_query: "pants",
    });

    expect(changedSignature).not.toBe(baseSignature);
  });
});
