import { useCallback, useEffect, useMemo } from "react";
import { useMutation } from "@tanstack/react-query";
import {
  useAvailableColumns,
  useInventoryFilterPreferences,
  useSaveInventoryFilterPreferences,
  useSelectedColumns,
} from "@/hooks/useInventory";
import { saveSelectedColumns } from "@/api/inventory";
import { useInventoryFilterStore } from "@/stores/inventoryFilterStore";
import {
  toColumnConfigs,
  fromColumnConfigs,
} from "../utils/inventoryColumnConfigs";
import { useInventoryFilterPreferenceSync } from "./useInventoryFilterPreferenceSync";
import type { ColumnConfig } from "@/types/shared";

/**
 * Manages column visibility, locking, and preference persistence for inventory pages.
 * Extracted from SimplifiedInventoryPage per SRP.
 */
export function useInventoryColumns() {
  const {
    visibleColumns,
    lockedColumns,
    columnFilters,
    search,
    preferencesLoaded,
    setVisibleColumns,
    setLockedColumns,
    hydratePreferences,
    markPreferencesLoaded,
  } = useInventoryFilterStore();

  const { data: availableColumns = [] } = useAvailableColumns();
  const { data: selectedColumns = [] } = useSelectedColumns();
  const { data: filterPreferences, isFetched: isFilterPreferencesFetched } =
    useInventoryFilterPreferences();
  const { mutate: saveFilterPreferences } = useSaveInventoryFilterPreferences();
  const saveSelectedColumnsMutation = useMutation({
    mutationFn: saveSelectedColumns,
  });

  // --- Resolved columns ---

  const resolvedVisibleColumns = useMemo(() => {
    if (visibleColumns.length > 0) return visibleColumns;
    if (selectedColumns.length > 0) return selectedColumns;
    return availableColumns;
  }, [availableColumns, selectedColumns, visibleColumns]);

  const resolvedLockedColumns = useMemo(() => {
    if (resolvedVisibleColumns.length === 0) return lockedColumns;
    const visibleSet = new Set(resolvedVisibleColumns);
    return lockedColumns.filter((col) => visibleSet.has(col));
  }, [lockedColumns, resolvedVisibleColumns]);

  // --- Preference hydration ---

  useEffect(() => {
    if (preferencesLoaded || !isFilterPreferencesFetched) return;
    if (filterPreferences) {
      hydratePreferences(filterPreferences);
    } else if (selectedColumns.length > 0) {
      setVisibleColumns(selectedColumns);
    }
    markPreferencesLoaded();
  }, [
    filterPreferences,
    hydratePreferences,
    isFilterPreferencesFetched,
    markPreferencesLoaded,
    preferencesLoaded,
    selectedColumns,
    setVisibleColumns,
  ]);

  useEffect(() => {
    if (!preferencesLoaded) return;
    if (resolvedVisibleColumns.length === 0 && selectedColumns.length > 0) {
      setVisibleColumns(selectedColumns);
    }
  }, [preferencesLoaded, resolvedVisibleColumns.length, selectedColumns, setVisibleColumns]);

  // --- Sync preferences ---

  useInventoryFilterPreferenceSync({
    preferencesLoaded,
    visibleColumns: resolvedVisibleColumns,
    lockedColumns: resolvedLockedColumns,
    columnFilters,
    searchQuery: search,
    saveFilterPreferences,
  });

  // --- ColumnManager integration ---

  const columnConfigs = useMemo(
    () =>
      toColumnConfigs(availableColumns, resolvedVisibleColumns, resolvedLockedColumns),
    [availableColumns, resolvedVisibleColumns, resolvedLockedColumns],
  );

  const handleColumnChange = useCallback(
    (nextConfigs: ColumnConfig[]) => {
      const { visibleColumns: nextVisible, lockedColumns: nextLocked } =
        fromColumnConfigs(nextConfigs);
      setVisibleColumns(nextVisible);
      setLockedColumns(nextLocked);
      saveSelectedColumnsMutation.mutate(nextVisible);
    },
    [setVisibleColumns, setLockedColumns, saveSelectedColumnsMutation],
  );

  const handleColumnReset = useCallback(() => {
    setVisibleColumns(availableColumns);
    setLockedColumns([]);
    saveSelectedColumnsMutation.mutate(availableColumns);
  }, [availableColumns, setVisibleColumns, setLockedColumns, saveSelectedColumnsMutation]);

  return {
    availableColumns,
    resolvedVisibleColumns,
    resolvedLockedColumns,
    columnConfigs,
    handleColumnChange,
    handleColumnReset,
    columnFilters,
  };
}
