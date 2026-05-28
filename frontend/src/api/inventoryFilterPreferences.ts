import apiClient from "./client";

export interface InventoryFilterPreferences {
  visible_columns: string[];
  column_filters: Record<string, string>;
  search_query: string;
  locked_columns: string[];
}

interface FilterPreferencesResponse {
  success: boolean;
  error?: string;
  data?: {
    visible_columns?: unknown;
    visibleColumns?: unknown;
    column_filters?: unknown;
    columnFilters?: unknown;
    search_query?: unknown;
    searchQuery?: unknown;
    locked_columns?: unknown;
    lockedColumns?: unknown;
  };
}

function normalizeStringArray(value: unknown): string[] {
  if (Array.isArray(value)) {
    return value
      .filter((item): item is string => typeof item === "string")
      .map((item) => item.trim())
      .filter(Boolean);
  }

  if (typeof value === "string") {
    try {
      const parsed = JSON.parse(value) as unknown;
      return normalizeStringArray(parsed);
    } catch (err) { logger.warn("Operation failed:", { err: err });
      return [];
    }
  }

  return [];
}

function normalizeColumnFilters(value: unknown): Record<string, string> {
  if (!value || typeof value !== "object" || Array.isArray(value)) {
    return {};
  }

  const entries = Object.entries(value as Record<string, unknown>)
    .map(([column, columnValue]) => {
      if (typeof columnValue === "string") {
        return [column, columnValue.trim()] as const;
      }

      if (typeof columnValue === "number" || typeof columnValue === "boolean") {
        return [column, String(columnValue)] as const;
      }

      return [column, ""] as const;
    })
    .filter(([, columnValue]) => columnValue !== "");

  return Object.fromEntries(entries);
}

function normalizePreferences(
  payload: FilterPreferencesResponse["data"],
): InventoryFilterPreferences {
  const visible_columns = normalizeStringArray(
    payload?.visible_columns ?? payload?.visibleColumns,
  );

  const locked_columns = normalizeStringArray(
    payload?.locked_columns ?? payload?.lockedColumns,
  );

  const column_filters = normalizeColumnFilters(
    payload?.column_filters ?? payload?.columnFilters,
  );

  const search_query =
    typeof payload?.search_query === "string"
      ? payload.search_query
      : typeof payload?.searchQuery === "string"
        ? payload.searchQuery
        : "";

  return {
    visible_columns,
    column_filters,
    search_query,
    locked_columns,
  };
}

export async function getInventoryFilterPreferences(): Promise<InventoryFilterPreferences> {
  const response = await apiClient.client.get<FilterPreferencesResponse>(
    "/filter-preferences",
    {
      params: {
        platform: "inventory",
        page: "inventory",
      },
    },
  );

  const payload = response.data;
  if (!payload.success) {
    throw new Error(
      payload.error || "Failed to load inventory filter preferences",
    );
  }

  return normalizePreferences(payload.data);
}

export async function saveInventoryFilterPreferences(
  preferences: InventoryFilterPreferences,
): Promise<void> {
  const response = await apiClient.post("/filter-preferences", {
    platform: "inventory",
    page: "inventory",
    tab: "inventory",
    visible_columns: preferences.visible_columns,
    column_filters: preferences.column_filters,
    filters: preferences.column_filters,
    search_query: preferences.search_query,
    locked_columns: preferences.locked_columns,
  });

  if (!response.success) {
    throw new Error(
      response.error || "Failed to save inventory filter preferences",
    );
  }
}
