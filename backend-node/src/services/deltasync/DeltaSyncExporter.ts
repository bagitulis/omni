/**
 * DeltaSyncExporter - Execute cell-level exports to Google Sheets
 *
 * SINGLE RESPONSIBILITY: Execute delta sync exports with minimal API calls
 * - Update only changed cells (not entire rows)
 * - Append new rows
 * - Update snapshot after successful export
 * - Respect locked columns
 */

import { Logger } from "winston";
import { getLogger } from "../../utils/logger";
import { getGoogleSheetsService } from "../googleSheetsService";
import { getSnapshotService, SnapshotService } from "./SnapshotService";
import {
  getCellDiffEngine,
  CellDiffEngine,
  CellChange,
  DiffResult,
  ColumnConfig,
} from "./CellDiffEngine";

export interface DeltaExportResult {
  success: boolean;
  message: string;
  stats: {
    cellsUpdated: number;
    rowsAppended: number;
    unchanged: number;
    skipped: number;
    apiCalls: number;
  };
  errors: string[];
  duration: number;
}

export class DeltaSyncExporter {
  private logger: Logger;
  private diffEngine: CellDiffEngine;
  private snapshotService: SnapshotService;

  constructor() {
    this.logger = getLogger("DeltaSyncExporter");
    this.diffEngine = getCellDiffEngine();
    this.snapshotService = getSnapshotService();
  }

  /**
   * Execute delta sync export
   * 1. Get current sheet data (refresh snapshot if needed)
   * 2. Compare database vs snapshot
   * 3. Update only changed cells
   * 4. Append new rows
   * 5. Update snapshot
   */
  async exportDelta(
    spreadsheetId: string,
    sheetName: string,
    dbRecords: Array<{ sku: string; data: Record<string, any> }>,
    headers: string[],
    columnConfigs: ColumnConfig[],
    lockedColumns: string[] = []
  ): Promise<DeltaExportResult> {
    const startTime = Date.now();
    let apiCalls = 0;

    try {
      // 1. Get snapshot (or sync from sheet if stale/empty)
      const snapshots = await this.snapshotService.getAllSnapshots();
      const snapshotCount = await this.snapshotService.getSnapshotCount();

      if (snapshotCount === 0) {
        this.logger.info("📸 No snapshot found, fetching fresh from sheet...");
        const sheetsService = getGoogleSheetsService();
        const sheetData = await sheetsService.getSheetData(
          spreadsheetId,
          sheetName
        );
        apiCalls++;

        if (sheetData && sheetData.length > 0) {
          await this.snapshotService.saveSnapshot(
            sheetData,
            sheetData[0],
            "SKU"
          );
          // Re-fetch snapshots after save
          const refreshedSnapshots =
            await this.snapshotService.getAllSnapshots();
          return this.performDeltaExport(
            spreadsheetId,
            sheetName,
            dbRecords,
            headers,
            columnConfigs,
            lockedColumns,
            refreshedSnapshots,
            startTime,
            apiCalls
          );
        } else {
          // Sheet is empty - treat all db records as new rows
          this.logger.warn("⚠️ Sheet is empty, will append all records as new");
          return this.performDeltaExport(
            spreadsheetId,
            sheetName,
            dbRecords,
            headers,
            columnConfigs,
            lockedColumns,
            new Map(), // Empty snapshot = all records are new
            startTime,
            apiCalls
          );
        }
      }

      return this.performDeltaExport(
        spreadsheetId,
        sheetName,
        dbRecords,
        headers,
        columnConfigs,
        lockedColumns,
        snapshots,
        startTime,
        apiCalls
      );
    } catch (error: any) {
      const duration = (Date.now() - startTime) / 1000;
      this.logger.error(`❌ Delta export failed: ${error.message}`);
      return {
        success: false,
        message: `Export failed: ${error.message}`,
        stats: {
          cellsUpdated: 0,
          rowsAppended: 0,
          unchanged: 0,
          skipped: 0,
          apiCalls,
        },
        errors: [error.message],
        duration,
      };
    }
  }

  /**
   * Perform the actual delta export
   */
  private async performDeltaExport(
    spreadsheetId: string,
    sheetName: string,
    dbRecords: Array<{ sku: string; data: Record<string, any> }>,
    headers: string[],
    columnConfigs: ColumnConfig[],
    lockedColumns: string[],
    snapshots: Map<string, any>,
    startTime: number,
    initialApiCalls: number
  ): Promise<DeltaExportResult> {
    let apiCalls = initialApiCalls;
    const errors: string[] = [];

    // Apply locked columns to config
    const finalConfigs = columnConfigs.map((c) => ({
      ...c,
      isLocked: c.isLocked || lockedColumns.includes(c.name),
    }));

    // 2. Detect changes
    const diff = this.diffEngine.detectChanges(
      dbRecords,
      snapshots,
      headers,
      finalConfigs
    );

    this.logger.info(
      `📊 Diff result: ${diff.cellChanges.length} cells changed, ${diff.newRows.length} new rows`
    );

    // 3. Execute cell updates
    const cellUpdateResults = await this.executeCellUpdates(
      spreadsheetId,
      sheetName,
      diff.cellChanges
    );
    apiCalls += cellUpdateResults.apiCalls;
    errors.push(...cellUpdateResults.errors);

    // 4. Append new rows
    const appendResults = await this.appendNewRows(
      spreadsheetId,
      sheetName,
      diff.newRows,
      headers
    );
    apiCalls += appendResults.apiCalls;
    errors.push(...appendResults.errors);

    // 5. Update snapshot for changed cells
    await this.updateSnapshotAfterExport(diff);

    const duration = (Date.now() - startTime) / 1000;
    const success = errors.length === 0;

    const message = success
      ? `✅ Delta sync: ${cellUpdateResults.updated} cells updated, ${appendResults.appended} rows appended`
      : `⚠️ Delta sync completed with ${errors.length} errors`;

    return {
      success,
      message,
      stats: {
        cellsUpdated: cellUpdateResults.updated,
        rowsAppended: appendResults.appended,
        unchanged: diff.unchangedCount,
        skipped: diff.skippedCount,
        apiCalls,
      },
      errors,
      duration,
    };
  }

  /**
   * Execute cell-level updates
   */
  private async executeCellUpdates(
    spreadsheetId: string,
    sheetName: string,
    changes: CellChange[]
  ): Promise<{ updated: number; apiCalls: number; errors: string[] }> {
    if (changes.length === 0) {
      return { updated: 0, apiCalls: 0, errors: [] };
    }

    const sheetsService = getGoogleSheetsService();
    const errors: string[] = [];
    let updated = 0;
    let apiCalls = 0;

    // Build batch ranges
    const ranges = this.diffEngine.buildBatchRanges(changes, sheetName);

    // Execute updates (batching could be improved with batchUpdate API)
    for (const { range, values } of ranges) {
      try {
        const success = await sheetsService.updateRange(
          spreadsheetId,
          range,
          values
        );
        apiCalls++;

        if (success) {
          updated++;
        } else {
          errors.push(`Failed to update ${range}`);
        }
      } catch (error: any) {
        errors.push(`Error updating ${range}: ${error.message}`);
      }
    }

    return { updated, apiCalls, errors };
  }

  /**
   * Append new rows to sheet
   */
  private async appendNewRows(
    spreadsheetId: string,
    sheetName: string,
    newRows: Array<{
      sku: string;
      rowData: Record<string, any>;
      rowArray: any[];
    }>,
    _headers: string[]
  ): Promise<{ appended: number; apiCalls: number; errors: string[] }> {
    if (newRows.length === 0) {
      return { appended: 0, apiCalls: 0, errors: [] };
    }

    const sheetsService = getGoogleSheetsService();

    try {
      const values = newRows.map((r) => r.rowArray);
      const success = await sheetsService.appendRows(
        spreadsheetId,
        sheetName,
        values
      );

      if (success) {
        return { appended: newRows.length, apiCalls: 1, errors: [] };
      } else {
        return {
          appended: 0,
          apiCalls: 1,
          errors: ["Failed to append rows"],
        };
      }
    } catch (error: any) {
      return {
        appended: 0,
        apiCalls: 1,
        errors: [`Error appending rows: ${error.message}`],
      };
    }
  }

  /**
   * Update snapshot after successful export
   */
  private async updateSnapshotAfterExport(diff: DiffResult): Promise<void> {
    // Update snapshot for changed cells
    for (const change of diff.cellChanges) {
      await this.snapshotService.updateSnapshotCell(
        change.sku,
        change.column,
        change.newValue
      );
    }

    // Note: New rows will be captured on next sync from sheet
    // This is intentional to ensure row indices are correct
  }
}

// Singleton instance
let instance: DeltaSyncExporter;

export function getDeltaSyncExporter(): DeltaSyncExporter {
  if (!instance) {
    instance = new DeltaSyncExporter();
  }
  return instance;
}
