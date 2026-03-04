/**
 * Sheet Registry API Service
 * Handles all API calls for Google Sheets Registry
 */

import { getAuthHeaders } from "@/utils/apiHeaders";
import type { SpreadsheetData, ApiResponse } from "@/types/sheetRegistry";

const BASE_URL = "/api/google/registry";

interface FetchOptions {
  method?: string;
  body?: string;
  headers?: Record<string, string>;
  credentials?: "include" | "omit" | "same-origin";
}

/**
 * Generic fetch wrapper with auth headers
 */
async function fetchWithAuth<T>(
  url: string,
  options: FetchOptions = {}
): Promise<ApiResponse<T>> {
  const response = await fetch(url, {
    ...options,
    headers: getAuthHeaders(),
    credentials: "include",
  });

  if (!response.ok) {
    const errorData = await response.json();
    throw new Error(errorData.error || "Request failed");
  }

  return response.json();
}

/**
 * Register a new spreadsheet by URL
 */
export async function registerSheet(
  url: string,
  purpose: string
): Promise<SpreadsheetData> {
  const response = await fetchWithAuth<SpreadsheetData>(
    `${BASE_URL}/register`,
    {
      method: "POST",
      body: JSON.stringify({ url, purpose }),
    }
  );

  if (!response.success) {
    throw new Error("Registration failed");
  }

  return response.data;
}

/**
 * Get all registered sheets for current user
 */
export async function getSheets(): Promise<SpreadsheetData[]> {
  const response = await fetchWithAuth<SpreadsheetData[]>(BASE_URL, {
    method: "GET",
  });

  if (!response.success) {
    throw new Error("Failed to fetch sheets");
  }

  return response.data;
}

/**
 * Get specific sheet by ID
 */
export async function getSheet(id: string): Promise<SpreadsheetData> {
  const response = await fetchWithAuth<SpreadsheetData>(`${BASE_URL}/${id}`, {
    method: "GET",
  });

  if (!response.success) {
    throw new Error("Failed to fetch sheet");
  }

  return response.data;
}

/**
 * Lock sheet for editing
 */
export async function lockSheet(id: string): Promise<boolean> {
  const response = await fetchWithAuth<null>(`${BASE_URL}/${id}/lock`, {
    method: "POST",
  });

  return response.success;
}

/**
 * Unlock sheet after editing
 */
export async function unlockSheet(id: string): Promise<boolean> {
  const response = await fetchWithAuth<null>(`${BASE_URL}/${id}/unlock`, {
    method: "POST",
  });

  return response.success;
}

/**
 * Delete registered sheet
 */
export async function deleteSheet(id: string): Promise<boolean> {
  const response = await fetchWithAuth<null>(`${BASE_URL}/${id}`, {
    method: "DELETE",
  });

  return response.success;
}

/**
 * Export inventory data to sheet
 */
export async function exportInventoryToSheet(
  sheetId: string,
  data: Record<string, any>[]
): Promise<boolean> {
  const response = await fetchWithAuth<{ stats: any }>(
    "/api/inventory/export-to-sheet",
    {
      method: "POST",
      body: JSON.stringify({ sheetId, data }),
    }
  );

  return response.success;
}

/**
 * Import inventory data from sheet
 */
export async function importInventoryFromSheet(
  sheetId: string,
  sheetName?: string
): Promise<Record<string, any>[]> {
  const response = await fetchWithAuth<Record<string, any>[]>(
    "/api/inventory/import-from-sheet",
    {
      method: "POST",
      body: JSON.stringify({ sheetId, sheetName }),
    }
  );

  if (!response.success) {
    throw new Error("Import failed");
  }

  return response.data;
}

/**
 * Check sync status for sheet
 */
export async function checkSyncStatus(sheetId: string): Promise<any> {
  const response = await fetchWithAuth<any>("/api/inventory/sync-status", {
    method: "POST",
    body: JSON.stringify({ sheetId }),
  });

  if (!response.success) {
    throw new Error("Failed to check sync status");
  }

  return response.data;
}
