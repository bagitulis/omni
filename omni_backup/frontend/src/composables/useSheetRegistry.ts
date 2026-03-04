/**
 * Composable for Google Sheets Registry operations
 * Provides reactive state and methods for spreadsheet management
 */

import { reactive, computed } from "vue";
import type { SpreadsheetData, RegistryState } from "@/types/sheetRegistry";
import * as sheetApi from "@/services/sheetRegistryApi";

// Re-export types for backwards compatibility
export type { SpreadsheetData, RegistrationResult } from "@/types/sheetRegistry";

// Shared reactive state
const state = reactive<RegistryState>({
  sheets: [],
  loading: false,
  error: null,
  selectedSheet: null,
  syncStatus: {},
});

// Computed getters
const isLoading = computed(() => state.loading);
const hasError = computed(() => state.error !== null);
const sheets = computed(() => state.sheets);
const selectedSheet = computed(() => state.selectedSheet);

/**
 * Wrapper to handle loading state and errors
 */
async function withLoadingState<T>(
  operation: () => Promise<T>
): Promise<T | null> {
  state.loading = true;
  state.error = null;

  try {
    return await operation();
  } catch (err: any) {
    state.error = err.message;
    return null;
  } finally {
    state.loading = false;
  }
}

/**
 * Register a new spreadsheet by URL
 */
async function registerSheet(
  url: string,
  purpose: string
): Promise<SpreadsheetData | null> {
  return withLoadingState(async () => {
    const data = await sheetApi.registerSheet(url, purpose);
    state.sheets.push(data);
    return data;
  });
}

/**
 * Get all registered sheets for current user
 */
async function getSheets(): Promise<SpreadsheetData[]> {
  const result = await withLoadingState(async () => {
    const data = await sheetApi.getSheets();
    state.sheets = data;
    return data;
  });
  return result || [];
}

/**
 * Get specific sheet by ID
 */
async function getSheet(id: string): Promise<SpreadsheetData | null> {
  return withLoadingState(async () => {
    const data = await sheetApi.getSheet(id);
    state.selectedSheet = data;
    return data;
  });
}

/**
 * Lock sheet for editing
 */
async function lockSheet(id: string): Promise<boolean> {
  const result = await withLoadingState(async () => {
    const success = await sheetApi.lockSheet(id);
    if (success) {
      updateSheetLockState(id, true);
    }
    return success;
  });
  return result || false;
}

/**
 * Unlock sheet after editing
 */
async function unlockSheet(id: string): Promise<boolean> {
  const result = await withLoadingState(async () => {
    const success = await sheetApi.unlockSheet(id);
    if (success) {
      updateSheetLockState(id, false);
    }
    return success;
  });
  return result || false;
}

/**
 * Helper to update lock state in local state
 */
function updateSheetLockState(id: string, isLocked: boolean): void {
  const sheet = state.sheets.find((s) => s.id === id);
  if (sheet) {
    sheet.isLocked = isLocked;
  }
  if (state.selectedSheet?.id === id) {
    state.selectedSheet.isLocked = isLocked;
  }
}

/**
 * Delete registered sheet
 */
async function deleteSheet(id: string): Promise<boolean> {
  const result = await withLoadingState(async () => {
    const success = await sheetApi.deleteSheet(id);
    if (success) {
      state.sheets = state.sheets.filter((s) => s.id !== id);
      if (state.selectedSheet?.id === id) {
        state.selectedSheet = null;
      }
    }
    return success;
  });
  return result || false;
}

/**
 * Export inventory data to sheet
 */
async function exportInventoryToSheet(
  sheetId: string,
  data: Record<string, any>[]
): Promise<boolean> {
  const result = await withLoadingState(async () => {
    return sheetApi.exportInventoryToSheet(sheetId, data);
  });
  return result || false;
}

/**
 * Import inventory data from sheet
 */
async function importInventoryFromSheet(
  sheetId: string,
  sheetName?: string
): Promise<Record<string, any>[] | null> {
  return withLoadingState(async () => {
    return sheetApi.importInventoryFromSheet(sheetId, sheetName);
  });
}

/**
 * Check sync status for sheet
 */
async function checkSyncStatus(sheetId: string): Promise<any | null> {
  return withLoadingState(async () => {
    const data = await sheetApi.checkSyncStatus(sheetId);
    state.syncStatus[sheetId] = data;
    return data;
  });
}

/**
 * Clear error message
 */
function clearError(): void {
  state.error = null;
}

/**
 * Clear all state
 */
function reset(): void {
  state.sheets = [];
  state.loading = false;
  state.error = null;
  state.selectedSheet = null;
  state.syncStatus = {};
}

export function useSheetRegistry() {
  return {
    // State
    state,
    sheets,
    selectedSheet,
    isLoading,
    hasError,

    // Sheet Management
    registerSheet,
    getSheets,
    getSheet,
    lockSheet,
    unlockSheet,
    deleteSheet,

    // Import/Export
    exportInventoryToSheet,
    importInventoryFromSheet,
    checkSyncStatus,

    // Utilities
    clearError,
    reset,
  };
}
