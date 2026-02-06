<template>
  <div class="columns-manager">
    <div class="config-section">
      <div class="section-header" @click="isExpanded = !isExpanded">
        <h2>⚙️ Column Configuration & Lock</h2>
        <span class="toggle-icon" :class="{ 'is-expanded': isExpanded }"
          >▼</span
        >
      </div>

      <div v-if="isExpanded" class="section-content">
        <div v-if="loading" class="loading-state">
          <div class="spinner"></div>
          <span>Loading...</span>
        </div>

        <template v-else>
          <!-- Summary Info -->
          <ColumnsSummaryInfo
            :total-columns="availableColumns.length"
            :visible-count="visibleColumnCount"
            :locked-count="lockedCount"
          />

          <!-- Action Buttons -->
          <ColumnsActionsBar
            @select-all="selectAll"
            @clear-all="clearAll"
            @lock-all="lockAll"
            @unlock-all="unlockAll"
          />

          <!-- Columns Grid with Lock Toggle -->
          <div v-if="availableColumns.length === 0" class="empty-state">
            <p>No columns available.</p>
          </div>

          <ColumnsConfigTable
            v-else
            :available-columns="availableColumns"
            :is-selected="isSelected"
            :is-column-locked="isColumnLocked"
            :inventory-list="fullInventoryList"
            @toggle-visibility="toggleColumn"
          />

          <!-- Status Messages -->
          <div v-if="errorMessage" class="alert alert-error">
            {{ errorMessage }}
          </div>
        </template>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from "vue";
import { useColumnManager } from "./composables/useColumnManager";
import ColumnsConfigTable from "./ColumnsConfigTable.vue";
import ColumnsSummaryInfo from "./ColumnsSummaryInfo.vue";
import ColumnsActionsBar from "./ColumnsActionsBar.vue";

interface Column {
  column_name: string;
  column_type: string;
  is_key: boolean;
}

const props = defineProps<{
  schemaColumns?: Column[];
  fullInventoryList?: Record<string, any>[];
}>();

const isExpanded = ref(false);

const emits = defineEmits<{
  "visibility-changed": [data: { column: string; visible: boolean }];
  "filter-changed": [data: { column: string; values: string[] }];
  "columns-saved": [];
  sort: [data: { column: string; direction: "asc" | "desc" | null }];
}>();

const {
  loading,
  availableColumns,
  errorMessage,
  visibleColumnCount,
  lockedCount,
  loadData,
  isSelected,
  toggleColumn: cmToggleColumn,
  selectAll,
  clearAll,
  lockAll,
  unlockAll,
  isColumnLocked,
} = useColumnManager();

// Wrapper to emit events
const toggleColumn = (columnName: string) => {
  const wasSelected = isSelected(columnName);
  cmToggleColumn(columnName);
  emits("visibility-changed", { column: columnName, visible: !wasSelected });
};

onMounted(async () => {
  await loadData();
});
</script>

<style scoped>
.columns-manager {
  display: flex;
  flex-direction: column;
  width: 100%;
}

.config-section {
  background-color: #fff;
  border-radius: 8px;
  border: 1px solid #e0e0e0;
  overflow: visible;
  width: 100%;
}

.section-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  cursor: pointer;
  user-select: none;
  padding: 12px 20px;
  transition: background-color 0.2s;
}

.section-header:hover {
  background-color: #f9f9f9;
}

.section-header h2 {
  margin: 0;
  color: #333;
  font-size: 14px;
  font-weight: 600;
  flex: 1;
}

.toggle-icon {
  display: inline-block;
  transition: transform 0.2s;
  color: #6b7280;
}

.toggle-icon.is-expanded {
  transform: rotate(180deg);
}

.section-content {
  padding: 15px 20px;
  display: flex;
  flex-direction: column;
  gap: 15px;
  width: 100%;
  box-sizing: border-box;
}

.loading-state {
  display: flex;
  align-items: center;
  gap: 10px;
  padding: 20px;
  text-align: center;
  color: #6b7280;
  font-size: 13px;
}

.spinner {
  width: 16px;
  height: 16px;
  border: 2px solid #e0e0e0;
  border-top-color: #3498db;
  border-radius: 50%;
  animation: spin 0.6s linear infinite;
}

@keyframes spin {
  to {
    transform: rotate(360deg);
  }
}

.empty-state {
  padding: 20px;
  text-align: center;
  color: #6b7280;
  font-size: 12px;
}

.alert {
  padding: 10px 12px;
  border-radius: 4px;
  font-size: 12px;
}

.alert-error {
  background-color: #ffe0e0;
  color: #c0392b;
  border: 1px solid #e74c3c;
}
</style>
