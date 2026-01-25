/**
 * InventoryCheckboxHandler.ts
 * Handles checkbox-related logic for inventory table
 * - Manages selected rows state
 * - Converts boolean values to/from checkbox display
 * - Bulk selection operations
 */

export interface CheckboxState {
  selectedRows: Set<number>;
  allSelected: boolean;
}

export class InventoryCheckboxHandler {
  private state: CheckboxState;

  constructor() {
    this.state = {
      selectedRows: new Set(),
      allSelected: false,
    };
  }

  /**
   * Toggle selection for a single row
   */
  toggleRowSelection(rowIndex: number): void {
    if (this.state.selectedRows.has(rowIndex)) {
      this.state.selectedRows.delete(rowIndex);
      this.state.allSelected = false;
    } else {
      this.state.selectedRows.add(rowIndex);
    }
  }

  /**
   * Toggle select all rows
   */
  toggleSelectAll(totalRows: number): void {
    if (this.state.allSelected) {
      this.state.selectedRows.clear();
      this.state.allSelected = false;
    } else {
      this.state.selectedRows.clear();
      for (let i = 0; i < totalRows; i++) {
        this.state.selectedRows.add(i);
      }
      this.state.allSelected = true;
    }
  }

  /**
   * Check if row is selected
   */
  isRowSelected(rowIndex: number): boolean {
    return this.state.selectedRows.has(rowIndex);
  }

  /**
   * Get all selected row indices
   */
  getSelectedRowIndices(): number[] {
    return Array.from(this.state.selectedRows).sort((a, b) => a - b);
  }

  /**
   * Get count of selected rows
   */
  getSelectionCount(): number {
    return this.state.selectedRows.size;
  }

  /**
   * Clear all selections
   */
  clearSelection(): void {
    this.state.selectedRows.clear();
    this.state.allSelected = false;
  }

  /**
   * Convert boolean value to checkbox display text (✓ or -)
   */
  static booleanToCheckbox(value: any): string {
    if (value === null || value === undefined) return "-";
    const boolValue = String(value).toLowerCase();
    return boolValue === "true" || boolValue === "1" ? "✓" : "-";
  }

  /**
   * Convert checkbox display to boolean value for database
   */
  static checkboxToBoolean(checkboxValue: string): boolean {
    return checkboxValue === "✓";
  }

  /**
   * Check if value is boolean-like
   */
  static isBooleanLike(value: any): boolean {
    if (value === null || value === undefined) return false;
    const str = String(value).toLowerCase();
    return ["true", "false", "0", "1", "yes", "no", "✓", "-"].includes(str);
  }
}
