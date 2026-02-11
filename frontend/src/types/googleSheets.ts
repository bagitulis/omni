/**
 * Google Sheets Settings Types
 * All types use snake_case to match backend JSON
 */

export interface SpreadsheetLinks {
  inventory_url: string;
  wallet_url: string;
  shipping_url: string;
  order_url: string;
}

export interface SheetMetadata {
  name: string;
  sheet_id: number;
  index: number;
  column_count: number;
  row_count: number;
}

export interface ValidationResult {
  spreadsheet_id: string;
  name: string;
  sheets: SheetMetadata[];
  type: string;
}

export interface GoogleSheetsSettings {
  links: SpreadsheetLinks;
  sheets_metadata: Record<string, SheetMetadata[]>;
  last_updated: string;
}

export interface ValidateLinkPayload {
  spreadsheet_url: string;
  type: string;
}

export interface SaveLinksPayload {
  inventory_url: string | null;
  wallet_url: string | null;
  shipping_url: string | null;
  order_url: string | null;
}

export interface UpdateSettingsPayload {
  inventory_spreadsheet_id?: string;
  inventory_sheet_name?: string;
  inventory_available_worksheets?: SheetMetadata[];
  wallet_spreadsheet_id?: string;
  wallet_sheet_name?: string;
  wallet_available_worksheets?: SheetMetadata[];
  shipping_spreadsheet_id?: string;
  shipping_sheet_name?: string;
  shipping_available_worksheets?: SheetMetadata[];
  order_spreadsheet_id?: string;
  order_sheet_name?: string;
  order_available_worksheets?: SheetMetadata[];
}
