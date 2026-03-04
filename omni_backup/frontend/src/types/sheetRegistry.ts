/**
 * Sheet Registry Types
 * Interfaces for Google Sheets Registry operations
 */

export interface SpreadsheetData {
  id: string;
  spreadsheetId: string;
  spreadsheetName: string;
  spreadsheetUrl: string;
  sheets: SheetInfo[];
  registeredBy: string;
  registeredAt: string;
  lastUsedAt?: string;
  isLocked: boolean;
  editingLockedUntil?: string;
  lastModifiedBy?: string;
  purpose: string;
  syncSettings: SyncSettings;
}

export interface SheetInfo {
  name: string;
  sheetId: number;
  columnCount: number;
  rowCount: number;
}

export interface SyncSettings {
  autoSync: boolean;
  syncInterval: number;
}

export interface RegistrationResult {
  spreadsheetId: string;
  spreadsheetName: string;
  sheets: SheetInfo[];
}

export interface RegistryState {
  sheets: SpreadsheetData[];
  loading: boolean;
  error: string | null;
  selectedSheet: SpreadsheetData | null;
  syncStatus: Record<string, any>;
}

export interface ApiResponse<T> {
  success: boolean;
  data: T;
  error?: string;
}

export interface SyncStatusData {
  lastSync: string;
  status: string;
  changes: number;
}
