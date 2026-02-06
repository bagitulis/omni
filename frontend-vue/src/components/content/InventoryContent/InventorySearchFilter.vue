<template>
  <div class="filter-section">
    <div class="section-header" @click="isExpanded = !isExpanded">
      <h4>🔍 Column Filter</h4>
      <span class="toggle-icon" :class="{ 'is-expanded': isExpanded }">▼</span>
    </div>

    <div v-if="isExpanded" class="section-content">
      <div class="column-filters">
        <div
          v-for="column in visibleSchemaColumns"
          :key="column.column_name"
          class="filter-item"
        >
          <label>{{ column.column_name }}</label>
          <input
            :value="columnFilters[column.column_name] || ''"
            @input="
              handleFilterInput(
                column.column_name,
                ($event.target as HTMLInputElement).value
              )
            "
            type="text"
            :placeholder="`Filter ${column.column_name}...`"
            class="filter-input"
          />
        </div>
      </div>
      <div v-if="visibleSchemaColumns.length === 0" class="empty-filters">
        <p>
          No columns displayed. Select columns in Column Configuration above.
        </p>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed } from "vue";
import { useInventoryConfig } from "./composables/useInventoryConfig";

interface Column {
  column_name: string;
  column_type: string;
  is_key: boolean;
}

const props = defineProps<{
  schemaColumns: Column[];
  columnVisibility: Record<string, boolean>;
  columnFilters: Record<string, string>;
}>();

const emits = defineEmits<{
  "filter-changed": [data: { column: string; value: string }];
}>();

const isExpanded = ref(true);

const { setColumnFilter } = useInventoryConfig();

const visibleSchemaColumns = computed(() => {
  return props.schemaColumns.filter(
    (col) => props.columnVisibility[col.column_name] !== false
  );
});

const handleFilterInput = (column: string, value: string) => {
  setColumnFilter(column, value);
  emits("filter-changed", { column, value });
};
</script>

<style scoped>
@import "./InventoryContent.styles.css";

.filter-section {
  background-color: #fff;
  padding: 20px;
  border-radius: 8px;
  border: 1px solid #e0e0e0;
  margin-bottom: 15px;
}

.section-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  cursor: pointer;
  user-select: none;
}

.section-header h4 {
  margin: 0;
  color: #333;
  font-size: 14px;
  font-weight: 600;
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
  margin-top: 15px;
}

.column-filters {
  display: grid;
  grid-template-columns: repeat(2, 1fr);
  gap: 12px;
}

.filter-item {
  display: flex;
  flex-direction: column;
  gap: 4px;
}

.filter-item label {
  font-size: 12px;
  color: #555;
  font-weight: 500;
}

.filter-input {
  padding: 8px 10px;
  border: 1px solid #ddd;
  border-radius: 4px;
  font-size: 13px;
  transition: border-color 0.2s;
}

.filter-input:focus {
  outline: none;
  border-color: #3498db;
  box-shadow: 0 0 0 2px rgba(52, 152, 219, 0.1);
}

.empty-filters {
  text-align: center;
  padding: 20px;
  color: #6b7280;
  font-size: 12px;
}

.empty-filters p {
  margin: 0;
}
</style>
