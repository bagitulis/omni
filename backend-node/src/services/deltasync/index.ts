/**
 * Delta Sync Module Exports
 *
 * Cell-level differential sync for Google Sheets
 * Minimizes API quota usage by only updating changed cells
 */

export { SnapshotService, getSnapshotService } from "./SnapshotService";
export type { SnapshotRecord, SnapshotSaveResult } from "./SnapshotService";

export { CellDiffEngine, getCellDiffEngine } from "./CellDiffEngine";
export type {
  CellChange,
  NewRowChange,
  DiffResult,
  ColumnConfig,
} from "./CellDiffEngine";

export { DeltaSyncExporter, getDeltaSyncExporter } from "./DeltaSyncExporter";
export type { DeltaExportResult } from "./DeltaSyncExporter";
