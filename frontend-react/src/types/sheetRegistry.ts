/**
 * Sheet Registry Types
 * Interfaces for Google Sheets Registry operations
 * API types use snake_case to match backend JSON response
 */

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

export interface RegistrationResult {
  spreadsheet_id: string;
  spreadsheet_name: string;
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
  last_sync: string;
  status: string;
  changes: number;
}
