<template>
  <div class="filter-section-embedded">
    <div class="filter-subsection">
      <div class="filter-title">
        <h3>🔍 Column Filter</h3>
      </div>

      <div v-if="selectedColumns.length === 0" class="empty-state">
        <p>Select columns above to show filters.</p>
      </div>

      <div v-else class="column-filters">
        <div
          v-for="column in selectedColumns"
          :key="column"
          class="filter-item"
        >
          <label>{{ column }}</label>
          <input
            :value="columnFilters[column] || ''"
            @input="
              handleFilterInput(
                column,
                ($event.target as HTMLInputElement).value
              )
            "
            type="text"
            :placeholder="`Filter ${column}...`"
            class="filter-input"
          />
        </div>
      </div>

      <button
        v-if="selectedColumns.length > 0"
        @click="clearAllFilters"
        class="btn-clear-filters"
      >
        Clear All Filters
      </button>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed } from "vue";
import { useInventoryConfig } from "./composables/useInventoryConfig";

const props = defineProps<{
  selectedColumns: string[];
}>();

const { setColumnFilter, getColumnFilters } = useInventoryConfig();

const columnFilters = computed(() => getColumnFilters());

const handleFilterInput = (column: string, value: string) => {
  setColumnFilter(column, value);
};

const clearAllFilters = () => {
  props.selectedColumns.forEach((col) => {
    setColumnFilter(col, "");
  });
};
</script>

<style scoped>
.filter-section-embedded {
  border-top: 2px solid #f0f0f0;
  padding-top: 15px;
}

.filter-subsection {
  display: flex;
  flex-direction: column;
  gap: 10px;
}

.filter-title {
  display: flex;
  align-items: center;
  justify-content: space-between;
}

.filter-title h3 {
  margin: 0;
  font-size: 12px;
  font-weight: 600;
  color: #333;
}

.btn-clear-filters {
  align-self: flex-start;
  padding: 4px 8px;
  font-size: 10px;
  border: 1px solid #ddd;
  border-radius: 3px;
  background-color: white;
  cursor: pointer;
  color: #6b7280;
  transition: all 0.2s;
  white-space: nowrap;
}

.btn-clear-filters:hover {
  background-color: #f5f5f5;
  border-color: #6b7280;
}

.empty-state {
  padding: 10px;
  text-align: center;
  color: #6b7280;
  font-size: 11px;
}

.empty-state p {
  margin: 0;
}

.column-filters {
  display: grid;
  grid-template-columns: repeat(2, 1fr);
  gap: 10px;
}

.filter-item {
  display: flex;
  flex-direction: column;
  gap: 4px;
}

.filter-item label {
  font-size: 11px;
  color: #666;
  font-weight: 500;
}

.filter-input {
  padding: 6px 8px;
  border: 1px solid #ddd;
  border-radius: 4px;
  font-size: 12px;
  transition: border-color 0.2s;
  background-color: white;
}

.filter-input:focus {
  outline: none;
  border-color: #3498db;
  box-shadow: 0 0 0 2px rgba(52, 152, 219, 0.1);
}

.filter-input::placeholder {
  color: #9ca3af; /* Improved from #ccc for WCAG AA contrast */
}
</style>
