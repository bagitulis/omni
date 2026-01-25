<template>
  <div v-if="showFilterPanel" class="filter-panel">
    <div class="filter-panel-header">
      <h4>Filters</h4>
      <button
        @click="$emit('clear-filters')"
        class="btn-close-filter"
        title="Clear filters"
        aria-label="Clear filters"
      >
        <i class="pi pi-times" aria-hidden="true"></i>
      </button>
    </div>

    <div class="filter-section">
      <div
        class="section-header"
        @click="visibilityExpanded = !visibilityExpanded"
      >
        <h4>Column Visibility</h4>
        <span
          class="expand-icon"
          :class="{ expanded: visibilityExpanded }"
          aria-hidden="true"
        >
          <i class="pi pi-chevron-down" aria-hidden="true"></i>
        </span>
      </div>
      <div v-if="visibilityExpanded" class="column-visibility-grid">
        <label
          v-for="(label, key) in columnLabels"
          :key="key"
          class="checkbox-label"
        >
          <input
            type="checkbox"
            :checked="visibleColumns[key]"
            @change="
              $emit('update:visibleColumns', {
                ...visibleColumns,
                [key]: !visibleColumns[key],
              })
            "
            class="checkbox-input"
          />
          <span>{{ label }}</span>
        </label>
      </div>
      <button
        v-if="visibilityExpanded"
        @click="$emit('reset-visibility')"
        class="btn-small"
      >
        <i class="pi pi-refresh" aria-hidden="true"></i> Reset
      </button>
    </div>

    <div class="filter-section">
      <h4>Column Filters</h4>
      <div class="filter-inputs">
        <div
          v-for="field in filterableFields"
          :key="field"
          class="filter-group"
        >
          <label>{{ columnLabels[field] }}</label>
          <input
            v-if="field !== 'status'"
            :value="columnFilters[field]"
            @input="
              $emit('update:columnFilters', {
                ...columnFilters,
                [field]: ($event.target as HTMLInputElement).value,
              })
            "
            type="text"
            placeholder="Contains..."
            class="filter-input"
          />
          <select
            v-else
            :value="columnFilters[field]"
            @change="
              $emit('update:columnFilters', {
                ...columnFilters,
                [field]: ($event.target as HTMLSelectElement).value,
              })
            "
            class="filter-input"
          >
            <option value="">All</option>
            <option
              v-for="status in availableStatuses"
              :key="status"
              :value="status"
            >
              {{ status }}
            </option>
          </select>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref } from "vue";

defineProps({
  showFilterPanel: {
    type: Boolean,
    required: true,
  },
  columnLabels: {
    type: Object,
    required: true,
  },
  columnFilters: {
    type: Object,
    required: true,
  },
  visibleColumns: {
    type: Object,
    required: true,
  },
  filterableFields: {
    type: Array as () => string[],
    required: true,
  },
  availableStatuses: {
    type: Array as () => string[],
    default: () => [],
  },
});

defineEmits([
  "update:columnFilters",
  "update:visibleColumns",
  "reset-visibility",
  "clear-filters",
]);

const visibilityExpanded = ref(false);
</script>

<style scoped>
.filter-panel {
  /* OPTIMIZED: Changed from inline horizontal layout to dropdown modal */
  position: absolute;
  top: 100%;
  right: 0;
  background: white;
  border: 1px solid #e0e0e0;
  border-radius: 8px;
  padding: 16px;
  margin-top: 8px;
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.15);
  width: 380px;
  max-height: 600px;
  overflow-y: auto;
  z-index: 100;
}

.filter-panel-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 8px;
  padding-bottom: 8px;
  border-bottom: 1px solid #f0f0f0;
}

.filter-panel-header h4 {
  margin: 0;
  font-size: 0.9rem;
  font-weight: 600;
  color: #2c3e50;
}

.btn-close-filter {
  background: none;
  border: none;
  cursor: pointer;
  color: #6b7280;
  font-size: 1.2rem;
  padding: 4px;
  transition: color 0.3s ease;
  flex-shrink: 0;
}

.btn-close-filter:hover {
  color: #333;
}

.filter-section {
  margin-bottom: 12px;
}

.filter-section:last-child {
  margin-bottom: 0;
}

.section-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  cursor: pointer;
  user-select: none;
  padding: 6px 0;
  margin-bottom: 6px;
}

.section-header:hover h4 {
  color: #1976d2;
}

.section-header h4 {
  margin: 0;
  font-size: 0.9rem;
  font-weight: 600;
  color: #2c3e50;
  transition: color 0.2s ease;
}

.expand-icon {
  display: flex;
  align-items: center;
  color: #6b7280;
  transition: transform 0.3s ease;
  font-size: 0.8rem;
}

.expand-icon.expanded {
  transform: rotate(180deg);
}

.column-visibility-grid {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 8px;
  margin-bottom: 8px;
  padding: 4px 0;
}

.checkbox-label {
  display: flex;
  align-items: center;
  gap: 6px;
  cursor: pointer;
  font-size: 0.85rem;
  user-select: none;
  padding: 6px;
  border-radius: 4px;
  transition: background 0.2s ease;
  white-space: nowrap;
}

.checkbox-label:hover {
  background: #f9f9f9;
}

.checkbox-input {
  cursor: pointer;
  width: 16px;
  height: 16px;
  flex-shrink: 0;
}

.btn-small {
  padding: 6px 10px;
  border: 1px solid #ddd;
  background: white;
  border-radius: 4px;
  cursor: pointer;
  font-size: 0.8rem;
  display: flex;
  align-items: center;
  gap: 5px;
  transition: all 0.3s ease;
  white-space: nowrap;
  flex-shrink: 0;
}

.btn-small:hover {
  background: #f5f5f5;
  border-color: #6b7280;
}

.filter-inputs {
  display: grid;
  grid-template-columns: 1fr;
  gap: 12px;
  align-items: flex-start;
  padding-top: 4px;
  width: 100%;
}

.filter-group {
  display: flex;
  flex-direction: column;
  gap: 4px;
  min-width: 0;
  flex: 1;
}

.filter-group label {
  font-size: 0.75rem;
  font-weight: 600;
  color: #666;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.filter-input {
  padding: 6px 8px;
  border: 1px solid #ddd;
  border-radius: 4px;
  font-size: 0.8rem;
  transition: border-color 0.3s ease;
  width: 100%;
  box-sizing: border-box;
  height: 32px;
}

.filter-input:focus {
  outline: none;
  border-color: #1976d2;
  box-shadow: 0 0 0 2px rgba(25, 118, 210, 0.1);
}

.filter-input::-webkit-outer-spin-button,
.filter-input::-webkit-inner-spin-button {
  -webkit-appearance: none;
  margin: 0;
}

/* Scrollbar styling */
.filter-panel::-webkit-scrollbar {
  height: 6px;
}

.filter-panel::-webkit-scrollbar-track {
  background: #f1f1f1;
  border-radius: 3px;
}

.filter-panel::-webkit-scrollbar-thumb {
  background: #ccc;
  border-radius: 3px;
}

.filter-panel::-webkit-scrollbar-thumb:hover {
  background: #999;
}

/* Mobile responsiveness */
@media (max-width: 768px) {
  .filter-panel {
    position: fixed;
    top: auto;
    bottom: 0;
    right: 0;
    left: 0;
    width: 100%;
    max-height: 80vh;
    border-radius: 12px 12px 0 0;
    border: 1px solid #e0e0e0;
  }

  .filter-inputs {
    flex-direction: column;
  }

  .filter-group {
    min-width: auto;
    width: 100%;
  }
}
</style>
