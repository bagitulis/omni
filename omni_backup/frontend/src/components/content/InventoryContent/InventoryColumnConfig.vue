<template>
  <div class="filter-section">
    <div class="section-header" @click="isExpanded = !isExpanded">
      <h4>⚙️ Inventory Column Configuration</h4>
      <span class="toggle-icon" :class="{ 'is-expanded': isExpanded }">▼</span>
    </div>

    <div v-if="isExpanded" class="section-content">
      <div v-if="loadingColumns" class="loading-state">
        <div class="spinner"></div>
        <span>Loading columns...</span>
      </div>

      <div v-else class="columns-config">
        <div class="config-header">
          <div class="header-info">
            <p>Select columns from Google Sheets to display</p>
            <p class="text-sm">
              Total: {{ availableColumns.length }} (A-{{
                getLastColumnLetter(availableColumns.length)
              }}) | Selected: {{ selectedColumns.length }}
            </p>
          </div>
          <div class="header-actions">
            <button @click="handleSelectAll" class="btn-small">
              Select All
            </button>
            <button @click="handleClearAll" class="btn-small btn-secondary">
              Clear All
            </button>
          </div>
        </div>

        <!-- Column Checkboxes Grid -->
        <div v-if="availableColumns.length === 0" class="empty-state">
          <p>No columns available. Configure spreadsheet first.</p>
        </div>

        <div v-else class="columns-grid">
          <ColumnCheckbox
            v-for="column in availableColumns"
            :key="column.name"
            :column="column"
            :is-selected="isSelected(column.name)"
            @toggle="handleToggleColumn(column.name)"
          />
        </div>

        <div v-if="errorMessage" class="alert alert-error">
          {{ errorMessage }}
        </div>
      </div>

      <div v-if="savingColumns" class="auto-saving">
        <span class="spinner-small"></span> Menyimpan...
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from "vue";
import { useInventoryColumnManager } from "./composables/useInventoryColumnManager";
import ColumnCheckbox from "./ColumnCheckbox.vue";

// eslint-disable-next-line @typescript-eslint/no-unused-vars
const props = defineProps<{
  schemaColumns: any[];
}>();

const emits = defineEmits<{
  "visibility-changed": [data: { column: string; visible: boolean }];
  "columns-saved": [];
}>();

const isExpanded = ref(true);

const {
  loadingColumns,
  savingColumns,
  availableColumns,
  selectedColumns,
  errorMessage,
  loadData,
  isSelected,
  toggleColumn,
  selectAll,
  clearAll,
  getLastColumnLetter,
} = useInventoryColumnManager();

onMounted(async () => {
  await loadData();
});

const emitVisibilityChange = (column: string, visible: boolean) => {
  emits("visibility-changed", { column, visible });
};

const handleToggleColumn = (column: string) => {
  toggleColumn(column);
  emitVisibilityChange(column, isSelected(column));
};

const handleSelectAll = () => {
  const previous = new Set(selectedColumns.value);
  selectAll();
  availableColumns.value.forEach((col: any) => {
    if (!previous.has(col.name)) {
      emitVisibilityChange(col.name, true);
    }
  });
  emits("columns-saved");
};

const handleClearAll = () => {
  clearAll();
  selectedColumns.value.forEach((col: string) => {
    emitVisibilityChange(col, false);
  });
  emits("columns-saved");
};
</script>

<style scoped>
@import "./InventoryColumnConfig.styles.css";
</style>
