import apiClient from "./client";

/**
 * Get available columns for inventory table
 * Backend route: GET /api/inventory/columns/available
 */
export async function getAvailableColumns(): Promise<string[]> {
  const response = await apiClient.client.get("/inventory/columns/available");
  const data = response.data as {
    success?: boolean;
    error?: string;
    columns?: Array<{ name?: string; key?: string; label?: string }>;
    data?: Array<{ name?: string; key?: string; label?: string }>;
  };
  if (!data.success) {
    throw new Error(data.error || "Failed to fetch available columns");
  }

  const columns = Array.isArray(data.columns)
    ? data.columns
    : Array.isArray(data.data)
      ? data.data
      : [];

  const normalizedColumns = columns
    .map((col) => {
      if (typeof col.name === "string" && col.name.trim().length > 0) {
        return col.name;
      }
      if (typeof col.label === "string" && col.label.trim().length > 0) {
        return col.label;
      }
      if (typeof col.key === "string" && col.key.trim().length > 0) {
        return col.key;
      }
      return "";
    })
    .filter((column): column is string => column.length > 0);

  return Array.from(new Set(normalizedColumns));
}

/**
 * Get selected columns for inventory table
 * Backend route: GET /api/inventory/columns/selected
 */
export async function getSelectedColumns(): Promise<string[]> {
  const response = await apiClient.client.get("/inventory/columns/selected");
  const data = response.data;
  if (!data.success) {
    throw new Error("Failed to fetch selected columns");
  }

  if (Array.isArray(data.selected_columns)) {
    return data.selected_columns;
  }

  // Modular handler returns { data: [...] }
  if (Array.isArray(data.data)) {
    return data.data;
  }

  return [];
}

/**
 * Save selected columns
 * Backend route: POST /api/inventory/columns/selected
 */
export async function saveSelectedColumns(columns: string[]): Promise<void> {
  const response = await apiClient.post("/inventory/columns/selected", {
    selected_columns: columns,
    columns,
  });
  if (!response.success) {
    throw new Error(response.error || "Failed to save column selection");
  }
}
