<template>
  <div
    ref="dropdownRef"
    class="header-filter-dropdown"
    :style="dropdownPosition"
    @click.stop
  >
    <!-- Search Box -->
    <label for="header-filter-search" class="visually-hidden"
      >Search filter values</label
    >
    <input
      id="header-filter-search"
      :value="searchQuery"
      type="text"
      placeholder="Search values..."
      class="search-input"
      @input="
        $emit('update:searchQuery', ($event.target as HTMLInputElement).value)
      "
      @click.stop
    />

    <!-- Sort Options -->
    <div class="sort-section">
      <div class="section-title">Urutkan</div>
      <button
        :class="{
          active: headerFilter.getSortDirection(columnName) === 'asc',
        }"
        class="sort-btn"
        @click="handleSort('asc')"
      >
        ↑ Ascending
      </button>
      <button
        :class="{
          active: headerFilter.getSortDirection(columnName) === 'desc',
        }"
        class="sort-btn"
        @click="handleSort('desc')"
      >
        ↓ Descending
      </button>
      <button
        :class="{
          active: headerFilter.getSortDirection(columnName) === null,
        }"
        class="sort-btn"
        @click="handleSort(null)"
      >
        × Clear Sort
      </button>
    </div>

    <!-- Filter Options -->
    <div class="filter-section">
      <div class="section-title">Filter</div>
      <div v-if="!columnValues || columnValues.length === 0" class="no-values">
        ⚠️ Tidak ada nilai untuk difilter
      </div>
      <div v-else class="filter-options">
        <label
          v-for="value in filteredValues"
          :key="value"
          class="filter-option"
        >
          <input
            type="checkbox"
            :checked="isValueSelected(value)"
            @change="toggleValue(value)"
          />
          <span>{{ value }}</span>
        </label>
      </div>
    </div>

    <!-- Action Buttons -->
    <div class="actions">
      <button class="btn-apply" @click="applyFilters">Terapkan</button>
      <button class="btn-clear" @click="clearFilters">Bersihkan</button>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, ref, onMounted, onUnmounted, nextTick } from "vue";

interface Props {
  columnName: string;
  columnValues: string[];
  selectedValues: Set<string>;
  searchQuery: string;
  headerFilter: any;
  triggerElement?: HTMLElement;
}

const props = defineProps<Props>();

const emits = defineEmits<{
  "update:searchQuery": [value: string];
  "update:selectedValues": [values: Set<string>];
  "sort-apply": [
    data: { columnName: string; direction: "asc" | "desc" | null },
  ];
  "filter-apply": [];
  "filter-clear": [];
}>();

const dropdownRef = ref<HTMLElement | null>(null);
const dropdownPosition = ref({});

const calculatePosition = () => {
  if (!props.triggerElement) return;

  const triggerRect = props.triggerElement.getBoundingClientRect();
  const dropdownWidth = 340;
  const desiredMaxHeight = 420;

  // Always anchor dropdown below the trigger
  const top = triggerRect.bottom + 6;
  let left = triggerRect.left + window.scrollX;

  // Keep within right viewport edge
  if (left + dropdownWidth > window.innerWidth + window.scrollX) {
    left = window.innerWidth + window.scrollX - dropdownWidth - 10;
  }

  if (left < 10) {
    left = 10;
  }

  // Constrain height to available space below trigger (no upward flip)
  const availableHeight = window.innerHeight - triggerRect.bottom - 16;
  const maxHeight = Math.max(260, Math.min(desiredMaxHeight, availableHeight));

  dropdownPosition.value = {
    top: `${top + window.scrollY}px`,
    left: `${left}px`,
    maxHeight: `${maxHeight}px`,
  };
};

onMounted(() => {
  nextTick(() => {
    calculatePosition();
    window.addEventListener("resize", calculatePosition);
    window.addEventListener("scroll", calculatePosition, true);
  });
});

onUnmounted(() => {
  window.removeEventListener("resize", calculatePosition);
  window.removeEventListener("scroll", calculatePosition, true);
});

const filteredValues = computed<string[]>(() => {
  const query = props.searchQuery.toLowerCase();
  if (!query) {
    return props.columnValues || [];
  }

  return props.columnValues.filter((v: string) =>
    v.toLowerCase().includes(query)
  );
});

const isValueSelected = (value: string): boolean => {
  return props.selectedValues.has(value);
};

const toggleValue = (value: string) => {
  const newSet = new Set(props.selectedValues);
  if (newSet.has(value)) {
    newSet.delete(value);
  } else {
    newSet.add(value);
  }
  emits("update:selectedValues", newSet);
};

const handleSort = (direction: "asc" | "desc" | null) => {
  // Emit to parent FIRST
  emits("sort-apply", {
    columnName: props.columnName,
    direction,
  });

  // Then update headerFilter state (for badge/UI)
  props.headerFilter.applySort(props.columnName, direction);
};

const applyFilters = () => {
  emits("filter-apply");
};

const clearFilters = () => {
  emits("filter-clear");
};
</script>

<style scoped>
.header-filter-dropdown {
  position: fixed;
  z-index: 10000;
  background-color: white;
  border: 1px solid #ddd;
  border-radius: 6px;
  box-shadow: 0 3px 10px rgba(0, 0, 0, 0.12);
  min-width: 320px;
  max-width: 360px;
  max-height: 420px;
  overflow-y: auto;
  padding: 10px;
  display: flex;
  flex-direction: column;
  gap: 10px;
}

.search-input {
  padding: 6px 10px;
  border: 1px solid #ddd;
  border-radius: 4px;
  font-size: 12px;
  font-family: inherit;
}

.search-input:focus {
  outline: none;
  border-color: #3b82f6;
  box-shadow: 0 0 0 3px rgba(59, 130, 246, 0.1);
}

.section-title {
  font-size: 11px;
  font-weight: 600;
  color: #666;
  text-transform: uppercase;
  margin-bottom: 6px;
}

.sort-section {
  display: flex;
  flex-direction: column;
  gap: 6px;
  border-bottom: 1px solid #e5e7eb;
  padding-bottom: 10px;
}

.sort-btn {
  padding: 6px 7px;
  background-color: #f9fafb;
  border: 1px solid #e5e7eb;
  border-radius: 4px;
  font-size: 12px;
  cursor: pointer;
  transition: all 0.2s;
  text-align: left;
}

.sort-btn:hover {
  background-color: #f3f4f6;
  border-color: #d1d5db;
}

.sort-btn.active {
  background-color: #dbeafe;
  border-color: #3b82f6;
  color: #1e40af;
  font-weight: 500;
}

.filter-section {
  display: flex;
  flex-direction: column;
  gap: 6px;
}
.no-values {
  padding: 12px;
  text-align: center;
  color: #6b7280;
  font-size: 12px;
  font-style: italic;
}
.filter-options {
  display: flex;
  flex-direction: column;
  gap: 4px;
  max-height: 260px;
  overflow-y: auto;
}

.filter-option {
  display: flex;
  align-items: center;
  gap: 6px;
  padding: 4px 6px;
  cursor: pointer;
  border-radius: 4px;
  transition: background-color 0.2s;
  font-size: 12px;
}

.filter-option:hover {
  background-color: #f3f4f6;
}

.filter-option input {
  cursor: pointer;
  accent-color: #3b82f6;
}

.actions {
  display: flex;
  gap: 8px;
  padding-top: 10px;
  border-top: 1px solid #e5e7eb;
}

.btn-apply {
  flex: 1;
  padding: 7px 10px;
  background-color: #3b82f6;
  color: white;
  border: none;
  border-radius: 4px;
  font-size: 12px;
  font-weight: 500;
  cursor: pointer;
  transition: background-color 0.2s;
}

.btn-apply:hover {
  background-color: #2563eb;
}

.btn-clear {
  flex: 1;
  padding: 7px 10px;
  background-color: transparent;
  color: #666;
  border: 1px solid #ddd;
  border-radius: 4px;
  font-size: 12px;
  cursor: pointer;
  transition: all 0.2s;
}

.btn-clear:hover {
  background-color: #f9fafb;
  border-color: #6b7280;
}

.visually-hidden {
  position: absolute;
  width: 1px;
  height: 1px;
  padding: 0;
  margin: -1px;
  overflow: hidden;
  clip: rect(0, 0, 0, 0);
  white-space: nowrap;
  border: 0;
}
</style>
