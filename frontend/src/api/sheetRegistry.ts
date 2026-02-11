import apiClient from "./client";

export interface SheetInfo {
  name: string;
  sheet_id: number;
  column_count: number;
  row_count: number;
}

export interface SyncSettings {
  auto_sync: boolean;
  sync_interval: number;
}

export interface SpreadsheetData {
  id: string;
  spreadsheet_id: string;
  spreadsheet_name: string;
  spreadsheet_url: string;
  sheets: SheetInfo[];
  registered_by: string;
  registered_at: string;
  last_used_at?: string;
  is_locked: boolean;
  locked_by?: string;
  locked_at?: string;
  purpose: string;
  sync_settings: SyncSettings;
  is_active: boolean;
  created_at: string;
  updated_at: string;
}

export interface InventoryExportResponse {
  stats?: Record<string, unknown>;
}

export interface InventorySyncStatus {
  last_sync?: string;
  status?: string;
  changes?: number;
  [key: string]: unknown;
}

const REGISTRY_BASE_PATH = "/google/registry";

export async function registerSheet(
  url: string,
  purpose: string,
): Promise<SpreadsheetData> {
  const response = await apiClient.post<SpreadsheetData>(
    `${REGISTRY_BASE_PATH}/register`,
    {
      url,
      purpose,
    },
  );

  if (!response.success || !response.data) {
    throw new Error(response.error || "Registration failed");
  }

  return response.data;
}

export async function getSheets(): Promise<SpreadsheetData[]> {
  const response = await apiClient.get<SpreadsheetData[]>(REGISTRY_BASE_PATH);

  if (!response.success) {
    throw new Error(response.error || "Failed to fetch sheets");
  }

  return response.data ?? [];
}

export async function getSheet(id: string): Promise<SpreadsheetData | null> {
  const response = await apiClient.get<SpreadsheetData>(
    `${REGISTRY_BASE_PATH}/${id}`,
  );

  if (!response.success) {
    throw new Error(response.error || "Failed to fetch sheet");
  }

  return response.data ?? null;
}

export async function lockSheet(id: string): Promise<boolean> {
  const response = await apiClient.post(`${REGISTRY_BASE_PATH}/${id}/lock`);
  return response.success;
}

export async function unlockSheet(id: string): Promise<boolean> {
  const response = await apiClient.post(`${REGISTRY_BASE_PATH}/${id}/unlock`);
  return response.success;
}

export async function deleteSheet(id: string): Promise<boolean> {
  const response = await apiClient.delete(`${REGISTRY_BASE_PATH}/${id}`);
  return response.success;
}

export async function exportInventoryToSheet(
  sheet_id: string,
  data: Array<Record<string, unknown>>,
): Promise<InventoryExportResponse> {
  const response = await apiClient.post<InventoryExportResponse>(
    "/inventory/export-to-sheet",
    {
      sheet_id,
      data,
    },
  );

  if (!response.success) {
    throw new Error(response.error || "Failed to export inventory to sheet");
  }

  return response.data ?? {};
}

export async function importInventoryFromSheet(
  sheet_id: string,
  sheet_name?: string,
): Promise<Array<Record<string, unknown>>> {
  const response = await apiClient.post<Array<Record<string, unknown>>>(
    "/inventory/import-from-sheet",
    {
      sheet_id,
      sheet_name,
    },
  );

  if (!response.success) {
    throw new Error(response.error || "Import failed");
  }

  return response.data ?? [];
}

export async function checkSyncStatus(
  sheet_id: string,
): Promise<InventorySyncStatus> {
  const response = await apiClient.post<InventorySyncStatus>(
    "/inventory/sync-status",
    {
      sheet_id,
    },
  );

  if (!response.success) {
    throw new Error(response.error || "Failed to check sync status");
  }

  return response.data ?? {};
}
