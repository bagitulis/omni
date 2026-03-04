/**
 * Inventory Table Logic Composable
 * Handles table data management, selection sync, and marketplace allocation
 */

import { ref, computed, watch } from "vue";
import { useInventoryConfig } from "./useInventoryConfig";
import { useTableSelection } from "./useTableSelection";
import { useCellEdit } from "./useCellEdit";
import { useTableHeaderFilter } from "./useTableHeaderFilter";
import { useMarketplaceAllocation } from "./useMarketplaceAllocation";
import { useMarketplaceSettings } from "./useMarketplaceSettings";
import { extractSKU } from "./useInventoryItemHelpers";

interface Column {
  column_name: string;
  column_type: string;
  is_key: boolean;
}

interface BatchCheckResult {
  sku: string;
  lazada: boolean;
  shopee: boolean;
  tiktok: boolean;
}

interface TableProps {
  filteredList: any[];
  fullInventoryList: any[];
  schemaColumns: Column[];
  columnVisibility: Record<string, boolean>;
  batchResults?: BatchCheckResult[];
  lockStockMap?: Record<string, number>;
}

export function useInventoryTableLogic(props: TableProps) {
  const hoveredRow = ref(-1);
  const { lockedColumns, columnOrder } = useInventoryConfig();
  const totalRows = computed(() => props.filteredList?.length || 0);

  const { isAllSelected, isRowSelected, toggleRowSelection, toggleSelectAll } =
    useTableSelection(() => totalRows.value);
  const { editingCell, editValue, isEditingCell, cancelEdit } = useCellEdit();
  const headerFilter = useTableHeaderFilter();

  const { getAllocationForItem } = useMarketplaceAllocation();
  const { totalColumn, autoColumn } = useMarketplaceSettings();

  // Watch for lock changes - to trigger re-render
  watch(
    lockedColumns,
    () => {
      editingCell.value = null;
    },
    { deep: true }
  );

  const isColumnEditableReactive = (columnName: string): boolean => {
    return !lockedColumns.value.has(columnName);
  };

  const visibleColumns = computed(() => {
    const visible = props.schemaColumns.filter(
      (col) => props.columnVisibility[col.column_name] !== false
    );

    const savedOrder = columnOrder.value;
    if (savedOrder && savedOrder.length > 0) {
      return visible.sort((a, b) => {
        const indexA = savedOrder.indexOf(a.column_name);
        const indexB = savedOrder.indexOf(b.column_name);
        const orderA = indexA === -1 ? 9999 : indexA;
        const orderB = indexB === -1 ? 9999 : indexB;
        return orderA - orderB;
      });
    }

    return visible;
  });

  const getMarketplaceAllocation = (item: any) => {
    const sku = extractSKU(item);
    const batchResult = props.batchResults?.find((r) => r.sku === sku);

    const platformAvailability = batchResult
      ? {
          shopee: batchResult.shopee === true,
          tiktok: batchResult.tiktok === true,
          lazada: batchResult.lazada === true,
        }
      : undefined;

    return getAllocationForItem(
      item,
      totalColumn.value,
      autoColumn.value,
      platformAvailability
    );
  };

  const getBatchResult = (sku: string): BatchCheckResult | null => {
    return props.batchResults?.find((r) => r.sku === sku) || null;
  };

  // Sync selection to underlying data
  const syncSelectionToData = (rowIndex: number, selected: boolean) => {
    const item = props.filteredList?.[rowIndex];
    if (!item) return;

    item.checkbox = selected;
    item.selected = selected;

    const fullIndex = props.fullInventoryList?.indexOf(item) ?? -1;
    if (fullIndex >= 0) {
      props.fullInventoryList[fullIndex].checkbox = selected;
      props.fullInventoryList[fullIndex].selected = selected;
    } else if (item.sku) {
      const match = props.fullInventoryList?.find(
        (inv) => inv.sku === item.sku
      );
      if (match) {
        match.checkbox = selected;
        match.selected = selected;
      }
    }
  };

  const handleToggleRowSelection = (rowIndex: number) => {
    toggleRowSelection(rowIndex);
    syncSelectionToData(rowIndex, isRowSelected(rowIndex));
  };

  const handleToggleSelectAll = () => {
    toggleSelectAll();
    props.filteredList?.forEach((_, idx) => {
      syncSelectionToData(idx, isRowSelected(idx));
    });
  };

  // Initialize header filter when data changes
  const initHeaderFilter = () => {
    watch(
      [visibleColumns, () => props.fullInventoryList],
      ([newCols, newData]) => {
        if (newCols.length > 0 && newData && newData.length > 0) {
          headerFilter.initializeColumnValues(newCols, newData);
        }
      },
      { immediate: true }
    );
  };

  return {
    hoveredRow,
    editingCell,
    editValue,
    isEditingCell,
    cancelEdit,
    headerFilter,
    isAllSelected,
    isRowSelected,
    isColumnEditableReactive,
    visibleColumns,
    getMarketplaceAllocation,
    getBatchResult,
    handleToggleRowSelection,
    handleToggleSelectAll,
    initHeaderFilter,
  };
}
