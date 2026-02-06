/**
 * useColumnConfiguration Composable
 * RESPONSIBILITY: Handle column configuration API calls and state management
 * - Load available columns from backend
 * - Load and save selected columns
 * - Manage column selection state
 */

import { getAuthHeaders, getApiBaseUrl } from "../../../../utils/apiHeaders";

export function useColumnConfiguration() {
  const API_BASE_URL = getApiBaseUrl("/inventory");

  /**
   * Load all available columns from Google Sheets
   */
  async function loadAvailableColumns(): Promise<any[]> {
    try {
      const response = await fetch(`${API_BASE_URL}/columns/available`, {
        headers: getAuthHeaders(),
      });
      const data = await response.json();

      if (data.success === true) {
        console.log(
          "📋 Available columns:",
          (data.columns || []).map(
            (c: any) => `${c.spreadsheet_column}: ${c.name}`,
          ),
        );
        return data.columns || [];
      }

      throw new Error(
        data.error || data.message || "Failed to load available columns",
      );
    } catch (error: any) {
      console.error("❌ Error loading available columns:", error);
      throw error;
    }
  }

  /**
   * Load currently selected columns
   */
  async function loadSelectedColumns(): Promise<string[]> {
    try {
      const response = await fetch(`${API_BASE_URL}/columns/selected`, {
        headers: getAuthHeaders(),
      });
      const data = await response.json();

      if (data.success === true) {
        console.log("📋 Selected columns:", data.selected_columns);
        return data.selected_columns || [];
      }

      throw new Error(
        data.error || data.message || "Failed to load selected columns",
      );
    } catch (error: any) {
      console.error("❌ Error loading selected columns:", error);
      throw error;
    }
  }

  /**
   * Save column selection to backend
   */
  async function saveColumnConfiguration(
    selectedColumns: string[],
  ): Promise<void> {
    try {
      const response = await fetch(`${API_BASE_URL}/columns/selected`, {
        method: "POST",
        headers: {
          ...getAuthHeaders(),
          "Content-Type": "application/json",
        },
        body: JSON.stringify({ selected_columns: selectedColumns }),
      });

      const data = await response.json();

      if (!data.success) {
        throw new Error(
          data.error || data.message || "Failed to save column configuration",
        );
      }

      console.log("✅ Column configuration saved");
    } catch (error: any) {
      console.error("❌ Error saving column configuration:", error);
      throw error;
    }
  }

  /**
   * Get the last column letter based on total columns count
   */
  function getLastColumnLetter(totalColumns: number): string {
    if (totalColumns === 0) return "A";
    return String.fromCharCode(64 + totalColumns); // 65 = 'A', so 64+count = last letter
  }

  return {
    loadAvailableColumns,
    loadSelectedColumns,
    saveColumnConfiguration,
    getLastColumnLetter,
  };
}
