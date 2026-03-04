/**
 * Inventory Validation Service
 *
 * SINGLE RESPONSIBILITY: Validate sheet headers and structure
 */

export interface HeaderValidationResult {
  is_valid: boolean;
  message: string;
  missing_columns?: string[];
  extra_columns?: string[];
}

export class InventoryValidationService {
  async validateHeaders(
    headers: string[],
    settings: any
  ): Promise<HeaderValidationResult> {
    try {
      // Validate headers from sheet:
      // 1. Must not be empty
      // 2. Must have at least 1 header
      // 3. Header values cannot be only whitespace

      if (!headers || headers.length === 0) {
        return {
          is_valid: false,
          message: "Sheet has no headers",
        };
      }

      // Filter valid headers (not empty/whitespace)
      const validHeaders = headers.filter((h: string) => h && h.trim());

      if (validHeaders.length === 0) {
        return {
          is_valid: false,
          message: "Sheet headers are all empty",
        };
      }

      // If allColumns is set in settings, validate headers match
      let allColumns: string[] = [];

      if (
        Array.isArray(settings?.allColumns) &&
        settings.allColumns.length > 0
      ) {
        allColumns = settings.allColumns;
      } else if (
        typeof settings?.allColumns === "string" &&
        settings.allColumns.length > 0
      ) {
        // Try to parse as JSON first
        try {
          const parsed = JSON.parse(settings.allColumns);
          if (Array.isArray(parsed) && parsed.length > 0) {
            allColumns = parsed;
          } else {
            // Not JSON, try comma-separated
            allColumns = settings.allColumns
              .split(",")
              .map((col: string) => col.trim())
              .filter(Boolean);
          }
        } catch {
          // Not JSON, treat as comma-separated string
          allColumns = settings.allColumns
            .split(",")
            .map((col: string) => col.trim())
            .filter(Boolean);
        }
      }

      // If allColumns is set, validate that headers match
      if (allColumns.length > 0) {
        const missingColumns = allColumns.filter(
          (col: string) => !validHeaders.includes(col)
        );
        const extraColumns = validHeaders.filter(
          (col: string) => !allColumns.includes(col)
        );

        const isValid =
          missingColumns.length === 0 && extraColumns.length === 0;

        return {
          is_valid: isValid,
          message: isValid ? "Headers are valid" : "Headers mismatch detected",
          missing_columns: missingColumns,
          extra_columns: extraColumns,
        };
      }

      // If allColumns is empty, accept all headers from sheet
      return {
        is_valid: true,
        message:
          "Headers are valid (no stored column definition to validate against)",
      };
    } catch (error: any) {
      return {
        is_valid: false,
        message: `Validation error: ${error.message}`,
      };
    }
  }
}
