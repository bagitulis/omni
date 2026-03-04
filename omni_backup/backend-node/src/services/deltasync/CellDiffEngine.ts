/**
 * CellDiffEngine - Detect cell-level changes between Database and Snapshot
 *
 * SINGLE RESPONSIBILITY: Compare data and identify changed cells
 * - Compare database records vs snapshot (last known sheet state)
 * - Detect changes at cell level (not row level)
 * - Respect column lock/active settings
 * - Generate minimal update operations
 */

import { SnapshotRecord } from "./SnapshotService";

export interface CellChange {
  sku: string;
  rowIndex: number; // Row position in Google Sheet (1-based after header)
  column: string; // Column name (e.g., "Total", "Shopee")
  columnIndex: number; // Column index in sheet (0-based)
  oldValue: any; // Value in snapshot
  newValue: any; // Value in database
  cellRange: string; // e.g., "D5" for direct API call
}

export interface NewRowChange {
  sku: string;
  rowData: Record<string, any>;
  rowArray: any[]; // Values in column order
}

export interface DiffResult {
  cellChanges: CellChange[];
  newRows: NewRowChange[];
  unchangedCount: number;
  skippedCount: number;
  summary: {
    totalCells: number;
    changedCells: number;
    newRows: number;
  };
}

export interface ColumnConfig {
  name: string;
  isActive: boolean; // Column is visible/selected
  isLocked: boolean; // Column should not be modified
}

export class CellDiffEngine {
  /**
   * Convert column index to Excel-style letter (0=A, 1=B, ..., 26=AA)
   */
  private columnIndexToLetter(index: number): string {
    let result = "";
    let temp = index;

    while (temp >= 0) {
      result = String.fromCharCode((temp % 26) + 65) + result;
      temp = Math.floor(temp / 26) - 1;
    }

    return result;
  }

  /**
   * Build cell range string (e.g., "D5")
   */
  private buildCellRange(columnIndex: number, rowIndex: number): string {
    const colLetter = this.columnIndexToLetter(columnIndex);
    return `${colLetter}${rowIndex + 1}`; // +1 because sheets are 1-indexed
  }

  /**
   * Normalize value for comparison
   */
  private normalizeValue(value: any): string {
    if (value === null || value === undefined) return "";
    if (typeof value === "boolean") return value ? "TRUE" : "FALSE";
    return String(value).trim();
  }

  /**
   * Check if two values are different
   */
  private valuesAreDifferent(oldVal: any, newVal: any): boolean {
    return this.normalizeValue(oldVal) !== this.normalizeValue(newVal);
  }

  /**
   * Compare database records against snapshots and detect cell-level changes
   */
  detectChanges(
    dbRecords: Array<{ sku: string; data: Record<string, any> }>,
    snapshots: Map<string, SnapshotRecord>,
    headers: string[],
    columnConfigs: ColumnConfig[]
  ): DiffResult {
    const cellChanges: CellChange[] = [];
    const newRows: NewRowChange[] = [];
    let unchangedCount = 0;
    let skippedCount = 0;

    // Build column index map and filter configs
    const columnIndexMap = new Map<string, number>();
    const editableColumns = new Set<string>();

    for (let i = 0; i < headers.length; i++) {
      columnIndexMap.set(headers[i], i);
    }

    // Determine which columns are editable (active AND not locked)
    for (const config of columnConfigs) {
      if (config.isActive && !config.isLocked) {
        editableColumns.add(config.name);
      }
    }

    // Process each database record
    for (const record of dbRecords) {
      const { sku, data } = record;

      if (!sku) {
        skippedCount++;
        continue;
      }

      const snapshot = snapshots.get(sku);

      if (!snapshot) {
        // New row - not in snapshot, append entire row
        const rowArray = headers.map((h) => data[h] ?? "");
        newRows.push({ sku, rowData: data, rowArray });
        continue;
      }

      // Existing row - compare cell by cell
      let hasChanges = false;

      for (const column of editableColumns) {
        const columnIndex = columnIndexMap.get(column);
        if (columnIndex === undefined) continue;

        const oldValue = snapshot.data[column];
        const newValue = data[column];

        if (this.valuesAreDifferent(oldValue, newValue)) {
          hasChanges = true;
          cellChanges.push({
            sku,
            rowIndex: snapshot.rowIndex,
            column,
            columnIndex,
            oldValue,
            newValue,
            cellRange: this.buildCellRange(columnIndex, snapshot.rowIndex),
          });
        }
      }

      if (!hasChanges) {
        unchangedCount++;
      }
    }

    return {
      cellChanges,
      newRows,
      unchangedCount,
      skippedCount,
      summary: {
        totalCells: cellChanges.length + newRows.length * headers.length,
        changedCells: cellChanges.length,
        newRows: newRows.length,
      },
    };
  }

  /**
   * Group cell changes by row for batch update optimization
   */
  groupChangesByRow(changes: CellChange[]): Map<number, CellChange[]> {
    const grouped = new Map<number, CellChange[]>();

    for (const change of changes) {
      const existing = grouped.get(change.rowIndex) || [];
      existing.push(change);
      grouped.set(change.rowIndex, existing);
    }

    return grouped;
  }

  /**
   * Build batch update ranges for Google Sheets API
   * Groups adjacent cells into ranges for efficiency
   */
  buildBatchRanges(
    changes: CellChange[],
    sheetName: string
  ): Array<{ range: string; values: any[][] }> {
    const rangeMap = new Map<string, any>();

    for (const change of changes) {
      const range = `${sheetName}!${change.cellRange}`;
      rangeMap.set(range, [[change.newValue]]);
    }

    return Array.from(rangeMap.entries()).map(([range, values]) => ({
      range,
      values,
    }));
  }
}

// Singleton instance
let instance: CellDiffEngine;

export function getCellDiffEngine(): CellDiffEngine {
  if (!instance) {
    instance = new CellDiffEngine();
  }
  return instance;
}
