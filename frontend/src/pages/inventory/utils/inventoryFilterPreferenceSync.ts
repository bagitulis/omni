import type { InventoryFilterPreferences } from "@/api/inventoryFilterPreferences";

function normalizeColumns(columns: string[]): string[] {
  const seen = new Set<string>();
  const normalized: string[] = [];

  for (const column of columns) {
    const trimmed = column.trim();
    if (!trimmed || seen.has(trimmed)) {
      continue;
    }

    seen.add(trimmed);
    normalized.push(trimmed);
  }

  return normalized;
}

function normalizeColumnFilters(
  filters: Record<string, string>,
): Record<string, string> {
  const entries = Object.entries(filters)
    .map(([column, value]) => [column.trim(), value.trim()] as const)
    .filter(([column, value]) => column !== "" && value !== "")
    .sort(([a], [b]) => a.localeCompare(b));

  return Object.fromEntries(entries);
}

export function normalizeInventoryFilterPreferencesForSync(
  payload: InventoryFilterPreferences,
): InventoryFilterPreferences {
  return {
    visible_columns: normalizeColumns(payload.visible_columns),
    locked_columns: normalizeColumns(payload.locked_columns),
    column_filters: normalizeColumnFilters(payload.column_filters),
    search_query: payload.search_query,
  };
}

export function buildInventoryFilterPreferencesSignature(
  payload: InventoryFilterPreferences,
): string {
  return JSON.stringify(normalizeInventoryFilterPreferencesForSync(payload));
}
