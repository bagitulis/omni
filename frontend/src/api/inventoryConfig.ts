import type { InventoryConfig } from "@/types/inventory";
import apiClient from "./client";
import { throwIfFailed } from "./inventoryCore";

export type RawInventoryConfig = Partial<InventoryConfig> & {
  selected_columns?: string | string[] | null;
  key_column?: string | null;
  key_column_name?: string | null;
};

export function normalizeSelectedColumns(
  value: RawInventoryConfig["selected_columns"],
): string {
  if (Array.isArray(value)) {
    return JSON.stringify(
      value.filter((column): column is string => typeof column === "string"),
    );
  }

  if (typeof value === "string") {
    return value;
  }

  return "";
}

export function normalizeInventoryConfig(
  config: RawInventoryConfig | null | undefined,
): InventoryConfig | null {
  if (!config) {
    return null;
  }

  const keyColumn =
    typeof config.key_column === "string"
      ? config.key_column
      : typeof config.key_column_name === "string"
        ? config.key_column_name
        : "";

  return {
    id: typeof config.id === "string" ? config.id : "",
    tenant_id: typeof config.tenant_id === "string" ? config.tenant_id : "",
    spreadsheet_id:
      typeof config.spreadsheet_id === "string" ? config.spreadsheet_id : "",
    sheet_name: typeof config.sheet_name === "string" ? config.sheet_name : "",
    selected_columns: normalizeSelectedColumns(config.selected_columns),
    all_columns:
      typeof config.all_columns === "string" ? config.all_columns : "",
    header_row: typeof config.header_row === "number" ? config.header_row : 1,
    data_start_row:
      typeof config.data_start_row === "number" ? config.data_start_row : 2,
    key_column: keyColumn,
    auto_sync: Boolean(config.auto_sync),
    sync_interval_seconds:
      typeof config.sync_interval_seconds === "number"
        ? config.sync_interval_seconds
        : 300,
    last_sync_timestamp: config.last_sync_timestamp ?? null,
    last_headers_hash:
      typeof config.last_headers_hash === "string"
        ? config.last_headers_hash
        : "",
    last_sync_status:
      typeof config.last_sync_status === "string"
        ? config.last_sync_status
        : "",
    low_stock_threshold:
      typeof config.low_stock_threshold === "number"
        ? config.low_stock_threshold
        : undefined,
    created_at: typeof config.created_at === "string" ? config.created_at : "",
    updated_at: typeof config.updated_at === "string" ? config.updated_at : "",
  };
}

/**
 * Get inventory configuration
 * Backend route: GET /api/inventory/config
 */
export async function getInventoryConfig(): Promise<InventoryConfig | null> {
  const response = await apiClient.get<RawInventoryConfig>("/inventory/config");
  if (!response.success) {
    throw new Error(response.error || "Failed to fetch inventory config");
  }
  return normalizeInventoryConfig(response.data);
}

/**
 * Update inventory configuration
 * Backend route: PUT /api/inventory/config
 */
export async function updateInventoryConfig(
  config: Partial<InventoryConfig>,
): Promise<InventoryConfig> {
  const response = await apiClient.put<InventoryConfig>(
    "/inventory/config",
    config,
  );
  throwIfFailed(response, "Failed to update inventory config");
  if (!response.data) {
    throw new Error("Updated inventory config response is empty");
  }
  return response.data;
}
