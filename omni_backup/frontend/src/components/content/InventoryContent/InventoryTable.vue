<template>
  <div class="table-wrapper">
    <div v-if="loading" class="loading-message">
      <div class="spinner"></div>
      <p>📡 Memuat data...</p>
    </div>
    <div
      v-else-if="!filteredList || filteredList.length === 0"
      class="empty-state"
    >
      <p>📭 Tidak ada data ditemukan</p>
      <p v-if="!filteredList" class="debug-text">
        ⚠️ filteredList is undefined
      </p>
    </div>
    <table v-else class="data-table">
      <thead>
        <tr>
          <th class="th-header checkbox-header">
            <input
              type="checkbox"
              :checked="isAllSelected"
              @change="handleToggleSelectAll"
              class="header-checkbox"
              aria-label="Select all rows"
            />
          </th>
          <TableHeaderCell
            v-for="column in visibleColumns"
            :key="column.column_name"
            :column="column"
            :header-filter="headerFilter"
            :full-inventory-list="fullInventoryList"
            @filter-apply="
              (col: string, vals: string[]) =>
                emits('filter-changed', { column: col, values: vals })
            "
            @filter-clear="
              (e: string) => emits('filter-changed', { column: e, values: [] })
            "
            @sort-apply="handleSortEvent"
          />
          <th class="th-header lock-stock-header">Lock Stock</th>
          <th class="th-header available-stock-header">Available Stock</th>
          <th class="th-header marketplace-header shopee-header">
            <span class="header-text">Shopee</span>
          </th>
          <th class="th-header marketplace-header tiktok-header">
            <span class="header-text">Tiktok</span>
          </th>
          <th class="th-header marketplace-header lazada-header">
            <span class="header-text">Lazada</span>
            <button
              class="settings-btn"
              @click.stop="$emit('open-marketplace-settings')"
              title="Marketplace Settings"
            >
              ⚙️
            </button>
          </th>
        </tr>
      </thead>
      <tbody>
        <tr
          v-for="(item, idx) in filteredList"
          :key="idx"
          class="table-row"
          :class="{ 'row-selected': isRowSelected(idx) }"
          @mouseenter="hoveredRow = idx"
          @mouseleave="hoveredRow = -1"
        >
          <td class="table-cell checkbox-cell">
            <input
              type="checkbox"
              :checked="isRowSelected(idx)"
              @change="handleToggleRowSelection(idx)"
              class="row-checkbox"
              :aria-label="`Select row ${idx}`"
            />
          </td>
          <td
            v-for="column in visibleColumns"
            :key="`${idx}-${column.column_name}`"
            class="table-cell"
            :class="{ 'cell-locked': !isColumnEditable(column.column_name) }"
            @click="
              !isBooleanLike(getCellValueRaw(item, column.column_name)) &&
              handleStartEdit(idx, column.column_name)
            "
          >
            <div
              v-if="isEditingCell(idx, column.column_name)"
              class="inline-edit"
            >
              <input
                v-model="editValue"
                @keyup.enter="handleSaveEdit(idx, column.column_name)"
                @keyup.escape="cancelEdit"
                @blur="cancelEdit"
                type="text"
                class="edit-input"
                autoFocus
              />
            </div>
            <div v-else class="cell-content">
              <input
                v-if="isBooleanLike(getCellValueRaw(item, column.column_name))"
                type="checkbox"
                :checked="isCellChecked(item, column.column_name)"
                @change="toggleCellCheckbox(idx, column.column_name)"
                class="cell-checkbox"
                :disabled="!isColumnEditable(column.column_name)"
                :aria-label="`Toggle ${column.column_name} for row ${idx + 1}`"
              />
              <span v-else>{{
                getItemCellValue(item, column.column_name)
              }}</span>
              <span
                v-if="
                  hoveredRow === idx &&
                  isColumnEditable(column.column_name) &&
                  !isBooleanLike(getCellValueRaw(item, column.column_name))
                "
                class="edit-hint"
                >✎</span
              >
              <span
                v-else-if="
                  hoveredRow === idx && !isColumnEditable(column.column_name)
                "
                class="lock-hint"
                >🔒</span
              >
            </div>
          </td>
          <td class="table-cell lock-stock-cell">
            <span class="lock-stock-value">{{
              getItemLockedQty(extractSKU(item))
            }}</span>
          </td>
          <td class="table-cell available-stock-cell">
            <span class="available-stock-value">{{
              getItemAvailableStock(extractSKU(item), getTotalQty(item))
            }}</span>
          </td>
          <MarketplaceCell
            platform="shopee"
            :value="getMarketplaceAllocation(item).shopee"
            :batch-result="getBatchResult(extractSKU(item))"
            @clone="(p) => handleClone(extractSKU(item), p)"
          />
          <MarketplaceCell
            platform="tiktok"
            :value="getMarketplaceAllocation(item).tiktok"
            :batch-result="getBatchResult(extractSKU(item))"
            @clone="(p) => handleClone(extractSKU(item), p)"
          />
          <MarketplaceCell
            platform="lazada"
            :value="getMarketplaceAllocation(item).lazada"
            :batch-result="getBatchResult(extractSKU(item))"
            @clone="(p) => handleClone(extractSKU(item), p)"
          />
        </tr>
      </tbody>
    </table>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, watch } from "vue";
import { useInventoryConfig } from "./composables/useInventoryConfig";
import { useTableSelection } from "./composables/useTableSelection";
import { useCellEdit } from "./composables/useCellEdit";
import { useTableHeaderFilter } from "./composables/useTableHeaderFilter";
import { isCellChecked, isBooleanLike } from "./composables/useBooleanField";
import { useMarketplaceAllocation } from "./composables/useMarketplaceAllocation";
import { useMarketplaceSettings } from "./composables/useMarketplaceSettings";
import {
  extractSKU,
  getCellValueRaw,
  getCellValue,
  getLockedQty,
  calculateAvailableStock,
  getTotalQty,
} from "./composables/useInventoryItemHelpers";
import TableHeaderCell from "./TableHeaderCell.vue";
import MarketplaceCell from "./MarketplaceCell.vue";

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

const props = defineProps<{
  filteredList: any[];
  fullInventoryList: any[];
  schemaColumns: Column[];
  columnVisibility: Record<string, boolean>;
  loading?: boolean;
  batchResults?: BatchCheckResult[];
  lockStockMap?: Record<string, number>;
}>();

const emits = defineEmits<{
  "cell-edit-save": [
    data: { itemIndex: number; column: string; value: string; item?: any },
  ];
  "edit-item": [item: any];
  "delete-item": [item: any];
  "filter-changed": [data: { column: string; values: string[] }];
  sort: [data: { column: string; direction: "asc" | "desc" | null }];
  "open-marketplace-settings": [];
  "clone-product": [data: { sku: string; platform: string }];
}>();

const hoveredRow = ref(-1);
const { lockedColumns, columnOrder } = useInventoryConfig();
const totalRows = computed(() => props.filteredList?.length || 0);
const { isAllSelected, isRowSelected, toggleRowSelection, toggleSelectAll } =
  useTableSelection(() => totalRows.value);
const { editingCell, editValue, isEditingCell, cancelEdit } = useCellEdit();
const headerFilter = useTableHeaderFilter();
const { getAllocationForItem } = useMarketplaceAllocation();
const { totalColumn, autoColumn } = useMarketplaceSettings();

const handleSortEvent = (data: {
  columnName: string;
  direction: "asc" | "desc" | null;
}) => {
  emits("sort", { column: data.columnName, direction: data.direction });
};

watch(
  lockedColumns,
  () => {
    editingCell.value = null;
  },
  { deep: true },
);

const isColumnEditable = (columnName: string) =>
  !lockedColumns.value.has(columnName);

const visibleColumns = computed(() => {
  const visible = props.schemaColumns.filter(
    (col) => props.columnVisibility[col.column_name] !== false,
  );
  const savedOrder = columnOrder.value;
  if (savedOrder?.length > 0) {
    return visible.sort((a, b) => {
      const orderA = savedOrder.indexOf(a.column_name);
      const orderB = savedOrder.indexOf(b.column_name);
      return (orderA === -1 ? 9999 : orderA) - (orderB === -1 ? 9999 : orderB);
    });
  }
  return visible;
});

const getMarketplaceAllocation = (item: any) => {
  const sku = extractSKU(item);
  const batchResult = props.batchResults?.find((r) => r.sku === sku);
  const platformAvailability = batchResult
    ? {
        shopee: batchResult.shopee,
        tiktok: batchResult.tiktok,
        lazada: batchResult.lazada,
      }
    : undefined;
  return getAllocationForItem(
    item,
    totalColumn.value,
    autoColumn.value,
    platformAvailability,
  );
};

const syncSelectionToData = (rowIndex: number, selected: boolean) => {
  const item = props.filteredList?.[rowIndex];
  if (!item) return;
  item.checkbox = selected;
  item.selected = selected;
  const fullIndex = props.fullInventoryList?.indexOf(item) ?? -1;
  if (fullIndex >= 0)
    Object.assign(props.fullInventoryList[fullIndex], {
      checkbox: selected,
      selected,
    });
  else if (item.sku) {
    const match = props.fullInventoryList?.find((inv) => inv.sku === item.sku);
    if (match) Object.assign(match, { checkbox: selected, selected });
  }
};

const handleToggleRowSelection = (rowIndex: number) => {
  toggleRowSelection(rowIndex);
  syncSelectionToData(rowIndex, isRowSelected(rowIndex));
};
const handleToggleSelectAll = () => {
  toggleSelectAll();
  props.filteredList?.forEach((_, idx) =>
    syncSelectionToData(idx, isRowSelected(idx)),
  );
};

watch(
  [visibleColumns, () => props.fullInventoryList],
  ([newCols, newData]) => {
    if (newCols.length > 0 && newData?.length > 0)
      headerFilter.initializeColumnValues(newCols, newData);
  },
  { immediate: true },
);

const getItemLockedQty = (sku: string) => getLockedQty(sku, props.lockStockMap);
const getItemAvailableStock = (sku: string, total: number) =>
  calculateAvailableStock(sku, total, props.lockStockMap);
const getItemCellValue = (item: any, column: string) =>
  getCellValue(item, column, isBooleanLike);
const getBatchResult = (sku: string) =>
  props.batchResults?.find((r) => r.sku === sku) || null;

const toggleCellCheckbox = (itemIndex: number, column: string) => {
  const item = props.filteredList[itemIndex];
  emits("cell-edit-save", {
    itemIndex,
    column,
    value: (!isCellChecked(item, column)).toString(),
    item,
  });
};

const handleSaveEdit = (itemIndex: number, column: string) => {
  emits("cell-edit-save", {
    itemIndex,
    column,
    value: editValue.value,
    item: props.filteredList[itemIndex],
  });
  editingCell.value = null;
};

const handleStartEdit = (itemIndex: number, column: string) => {
  if (!isColumnEditable(column)) return;
  editingCell.value = { itemIndex, column };
  editValue.value = String(props.filteredList[itemIndex][column] || "");
};

const handleClone = (sku: string, platform: string) => {
  emits("clone-product", { sku, platform });
};
</script>

<style scoped>
@import "./InventoryTable.styles.css";
</style>
