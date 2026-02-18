import type { ColumnConfig } from "@/types/shared";

/** Convert string-based column arrays to ColumnConfig[] for the shared ColumnManager. */
export function toColumnConfigs(
  availableColumns: string[],
  visibleColumns: string[],
  lockedColumns: string[],
): ColumnConfig[] {
  const visibleSet = new Set(visibleColumns);
  const lockedSet = new Set(lockedColumns);

  const ordered: ColumnConfig[] = [];
  let order = 0;

  for (const col of visibleColumns) {
    ordered.push({
      key: col,
      title: col,
      visible: true,
      locked: lockedSet.has(col),
      order: order++,
    });
  }

  for (const col of availableColumns) {
    if (!visibleSet.has(col)) {
      ordered.push({
        key: col,
        title: col,
        visible: false,
        locked: false,
        order: order++,
      });
    }
  }

  return ordered;
}

/** Convert ColumnConfig[] back to visible/locked string arrays. */
export function fromColumnConfigs(configs: ColumnConfig[]): {
  visibleColumns: string[];
  lockedColumns: string[];
} {
  const sorted = [...configs].sort((a, b) => a.order - b.order);
  const visibleColumns = sorted.filter((c) => c.visible).map((c) => c.key);
  const lockedColumns = sorted.filter((c) => c.locked).map((c) => c.key);
  return { visibleColumns, lockedColumns };
}
