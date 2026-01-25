/**
 * CheckboxDataConverter.ts
 * Backend utility to convert between checkbox values from Google Sheets and boolean values
 * Use for Google Sheets import/export
 *
 * Google Sheets checkbox → "TRUE"/"FALSE" string
 * Database storage → "true"/"false" string or boolean
 * Frontend display → checkbox input element
 */

export class CheckboxDataConverter {
  /**
   * Convert database boolean value to Google Sheets checkbox representation
   * Database "true"/"false" → true/false boolean for Sheets checkbox validation
   */
  static booleanToSheetsCheckbox(value: any): boolean {
    if (value === null || value === undefined) return false;

    // Direct boolean type
    if (typeof value === "boolean") return value;

    // String representation
    const str = String(value).toLowerCase();
    return str === "true" || str === "1" || str === "yes";
  }

  /**
   * Convert Google Sheets checkbox value to boolean string
   * Google Sheets checkbox comes as true/false boolean or "TRUE"/"FALSE"
   * Use when importing from Google Sheets
   */
  static sheetsCheckboxToBoolean(value: any): string {
    if (value === null || value === undefined) return "false";

    // Direct boolean type
    if (typeof value === "boolean") return value ? "true" : "false";

    // String representation from Google Sheets checkbox
    const str = String(value).trim().toUpperCase();
    return ["TRUE", "1", "YES", "✓"].includes(str) ? "true" : "false";
  }

  /**
   * Check if value is from checkbox (Google Sheets or boolean)
   */
  static isBooleanLike(value: any): boolean {
    // Direct boolean type check
    if (typeof value === "boolean") return true;

    if (value === null || value === undefined) return false;

    // Check for "true" and "false" strings (case-insensitive)
    const str = String(value).toLowerCase();
    return str === "true" || str === "false";
  }

  /**
   * Transform row data for sheets export
   * Convert all "true"/"false" to "TRUE"/"FALSE" for Google Sheets
   */
  static transformRowForSheetsExport(
    row: Record<string, any>
  ): Record<string, any> {
    const transformed = { ...row };

    for (const [key, value] of Object.entries(transformed)) {
      if (this.isBooleanLike(value)) {
        // Convert to Google Sheets checkbox representation
        transformed[key] = this.booleanToSheetsCheckbox(value);
      }
    }

    return transformed;
  }

  /**
   * Transform row data from sheets import
   * Convert all checkbox values to "true"/"false" strings for database
   */
  static transformRowFromSheetsImport(
    row: Record<string, any>
  ): Record<string, any> {
    const transformed = { ...row };

    for (const [key, value] of Object.entries(transformed)) {
      if (this.isBooleanLike(value)) {
        // Convert checkbox to "true"/"false" string
        transformed[key] = this.sheetsCheckboxToBoolean(value);
      }
    }

    return transformed;
  }

  /**
   * Transform entire dataset for sheets export
   */
  static transformDatasetForSheetsExport(
    data: Record<string, any>[]
  ): Record<string, any>[] {
    return data.map((row) => this.transformRowForSheetsExport(row));
  }

  /**
   * Transform entire dataset from sheets import
   */
  static transformDatasetFromSheetsImport(
    data: Record<string, any>[]
  ): Record<string, any>[] {
    return data.map((row) => this.transformRowFromSheetsImport(row));
  }
}
