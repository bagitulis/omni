export type InventoryColumnMoveDirection = "up" | "down";

function uniquePreserveOrder(columns: string[]): string[] {
  const seen = new Set<string>();
  const result: string[] = [];

  for (const column of columns) {
    const trimmed = column.trim();
    if (!trimmed || seen.has(trimmed)) {
      continue;
    }
    seen.add(trimmed);
    result.push(trimmed);
  }

  return result;
}

export function toggleInventoryColumn(
  columns: string[],
  target: string,
  checked: boolean,
): string[] {
  const normalizedTarget = target.trim();
  if (!normalizedTarget) {
    return uniquePreserveOrder(columns);
  }

  const normalizedColumns = uniquePreserveOrder(columns);
  const exists = normalizedColumns.includes(normalizedTarget);

  if (checked) {
    return exists
      ? normalizedColumns
      : [...normalizedColumns, normalizedTarget];
  }

  if (!exists) {
    return normalizedColumns;
  }

  return normalizedColumns.filter((column) => column !== normalizedTarget);
}

export function moveInventoryColumn(
  columns: string[],
  target: string,
  direction: InventoryColumnMoveDirection,
): string[] {
  const normalizedTarget = target.trim();
  const normalizedColumns = uniquePreserveOrder(columns);
  const currentIndex = normalizedColumns.indexOf(normalizedTarget);

  if (currentIndex === -1) {
    return normalizedColumns;
  }

  const nextIndex = direction === "up" ? currentIndex - 1 : currentIndex + 1;
  if (nextIndex < 0 || nextIndex >= normalizedColumns.length) {
    return normalizedColumns;
  }

  const nextColumns = [...normalizedColumns];
  [nextColumns[currentIndex], nextColumns[nextIndex]] = [
    nextColumns[nextIndex],
    nextColumns[currentIndex],
  ];
  return nextColumns;
}

export function buildInventoryColumnControlOrder(
  availableColumns: string[],
  visibleColumns: string[],
): string[] {
  const normalizedVisible = uniquePreserveOrder(visibleColumns);
  const normalizedAvailable = uniquePreserveOrder(availableColumns);
  const visibleSet = new Set(normalizedVisible);
  const hiddenColumns = normalizedAvailable.filter(
    (column) => !visibleSet.has(column),
  );

  return [...normalizedVisible, ...hiddenColumns];
}
