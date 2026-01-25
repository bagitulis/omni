<template>
  <div class="filter-dropdown-wrapper">
    <!-- Sort Options -->
    <div class="sort-section">
      <button
        @click="applySortAscending"
        :class="{ active: currentSort === 'asc' }"
        class="sort-btn"
        title="Sort A to Z"
      >
        ▲ A → Z
      </button>
      <button
        @click="applySortDescending"
        :class="{ active: currentSort === 'desc' }"
        class="sort-btn"
        title="Sort Z to A"
      >
        ▼ Z → A
      </button>
    </div>

    <div class="filter-header">
      <label for="filter-search-input" class="visually-hidden"
        >Search filter values</label
      >
      <input
        id="filter-search-input"
        v-model="searchQuery"
        type="text"
        placeholder="Search values..."
        class="search-input"
      />
    </div>

    <div class="filter-options">
      <!-- Select All -->
      <label class="filter-option">
        <input
          type="checkbox"
          :checked="isAllSelected"
          @change="toggleSelectAll"
          class="filter-checkbox"
        />
        <span class="option-label">(Select All)</span>
      </label>

      <!-- Individual Options -->
      <label
        v-for="option in filteredOptions"
        :key="option"
        class="filter-option"
      >
        <input
          type="checkbox"
          :checked="isOptionSelected(option)"
          @change="toggleOption(option)"
          class="filter-checkbox"
        />
        <span class="option-label">{{ option }}</span>
      </label>

      <!-- No Results -->
      <div v-if="filteredOptions.length === 0" class="no-results">
        No matching values
      </div>
    </div>

    <!-- Action Buttons -->
    <div class="filter-actions">
      <button @click="applyFilter" class="btn-apply">OK</button>
      <button @click="clearFilter" class="btn-cancel">Clear</button>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed } from "vue";

const props = defineProps<{
  columnName: string;
  options: string[];
  selectedValues: string[];
  currentSort?: "asc" | "desc" | null;
}>();

const emits = defineEmits<{
  apply: [values: string[]];
  clear: [];
  sort: [data: { column: string; direction: "asc" | "desc" | null }];
}>();

const searchQuery = ref("");
const selectedOptions = ref<Set<string>>(new Set(props.selectedValues));
// Initialize from parent prop if provided
const currentSort = ref<"asc" | "desc" | null>(props.currentSort || null);

const sortedOptions = computed(() => {
  const sorted = [...props.options];
  if (currentSort.value === "asc") {
    sorted.sort((a, b) => String(a).localeCompare(String(b)));
  } else if (currentSort.value === "desc") {
    sorted.sort((a, b) => String(b).localeCompare(String(a)));
  }
  return sorted;
});

const filteredOptions = computed(() => {
  if (!searchQuery.value) return sortedOptions.value;
  const query = searchQuery.value.toLowerCase();
  return sortedOptions.value.filter((opt) =>
    String(opt).toLowerCase().includes(query)
  );
});

const isAllSelected = computed(() => {
  return (
    selectedOptions.value.size > 0 &&
    selectedOptions.value.size === props.options.length
  );
});

const isOptionSelected = (option: string) => {
  return selectedOptions.value.has(String(option));
};

const toggleOption = (option: string) => {
  const key = String(option);
  if (selectedOptions.value.has(key)) {
    selectedOptions.value.delete(key);
  } else {
    selectedOptions.value.add(key);
  }
};

const toggleSelectAll = () => {
  if (isAllSelected.value) {
    selectedOptions.value.clear();
  } else {
    selectedOptions.value.clear();
    props.options.forEach((opt) => selectedOptions.value.add(String(opt)));
  }
};

const applySortAscending = () => {
  currentSort.value = currentSort.value === "asc" ? null : "asc";
  emits("sort", { column: props.columnName, direction: currentSort.value });
};

const applySortDescending = () => {
  currentSort.value = currentSort.value === "desc" ? null : "desc";
  emits("sort", { column: props.columnName, direction: currentSort.value });
};

const applyFilter = () => {
  // Emit the sort event along with filter to ensure sort persists in main table
  emits("sort", { column: props.columnName, direction: currentSort.value });
  emits("apply", Array.from(selectedOptions.value));
};

const clearFilter = () => {
  selectedOptions.value.clear();
  searchQuery.value = "";
  // Don't reset currentSort here - let it persist if user set it
  // Only emit clear for the filter values, not the sort
  emits("clear");
};
</script>

<style scoped>
.filter-dropdown-wrapper {
  display: flex;
  flex-direction: column;
  background: white;
  border: 1px solid #d9d9d9;
  border-radius: 6px;
  box-shadow: 0 3px 12px rgba(0, 0, 0, 0.15);
  min-width: 260px;
  max-width: 320px;
  max-height: 420px;
  z-index: 10001;
  overflow: hidden;
}

.sort-section {
  display: flex;
  gap: 6px;
  padding: 10px;
  border-bottom: 1px solid #f0f0f0;
  background-color: #fafafa;
  flex-wrap: wrap;
}

.sort-btn {
  flex: 1;
  min-width: 100px;
  padding: 6px 10px;
  border: 1px solid #ddd;
  border-radius: 4px;
  background: white;
  font-size: 12px;
  font-weight: 500;
  cursor: pointer;
  transition: all 0.2s;
  color: #666;
  white-space: nowrap;
}

.sort-btn:hover:not(.active) {
  border-color: #6b7280;
  background: #f5f5f5;
}

.sort-btn.active {
  border-color: #3b82f6;
  background: #dbeafe;
  color: #3b82f6;
}

.filter-header {
  padding: 10px;
  border-bottom: 1px solid #f0f0f0;
}

.search-input {
  width: 100%;
  padding: 8px 12px;
  border: 1px solid #d9d9d9;
  border-radius: 4px;
  font-size: 13px;
  box-sizing: border-box;
  transition: all 0.2s;
}

.search-input:focus {
  outline: none;
  border-color: #3b82f6;
  box-shadow: 0 0 0 2px rgba(59, 130, 246, 0.1);
}

.search-input::placeholder {
  color: #6b7280;
}

.filter-options {
  flex: 1;
  overflow-y: auto;
  padding: 6px 0;
  max-height: 250px;
}

.filter-options::-webkit-scrollbar {
  width: 6px;
}

.filter-options::-webkit-scrollbar-track {
  background: transparent;
}

.filter-options::-webkit-scrollbar-thumb {
  background: #ccc;
  border-radius: 3px;
}

.filter-options::-webkit-scrollbar-thumb:hover {
  background: #999;
}

.filter-option {
  display: flex;
  align-items: center;
  padding: 8px 12px;
  cursor: pointer;
  gap: 8px;
  transition: background-color 0.15s;
  user-select: none;
}

.filter-option:hover {
  background-color: #f5f5f5;
}

.filter-checkbox {
  width: 16px;
  height: 16px;
  cursor: pointer;
  accent-color: #3b82f6;
  flex-shrink: 0;
}

.option-label {
  font-size: 13px;
  color: #333;
  flex: 1;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.no-results {
  padding: 16px 12px;
  text-align: center;
  color: #6b7280;
  font-size: 12px;
  background: #fafafa;
}

.filter-actions {
  display: flex;
  gap: 8px;
  padding: 10px;
  border-top: 1px solid #f0f0f0;
  background-color: #fafafa;
}

.btn-apply,
.btn-cancel {
  flex: 1;
  padding: 7px 14px;
  border: 1px solid #ddd;
  border-radius: 4px;
  font-size: 12px;
  font-weight: 500;
  cursor: pointer;
  transition: all 0.2s;
}

.btn-apply {
  background-color: #3b82f6;
  color: white;
  border-color: #3b82f6;
}

.btn-apply:hover {
  background-color: #2563eb;
  border-color: #2563eb;
}

.btn-cancel {
  background-color: white;
  color: #666;
  border-color: #ddd;
}

.btn-cancel:hover {
  background-color: #f5f5f5;
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
