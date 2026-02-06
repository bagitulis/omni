<template>
  <th
    :key="column.column_name"
    class="th-header"
    :class="{ 'has-active-filter': headerFilter.hasFilter(column.column_name) }"
  >
    <div class="header-content">
      <span class="column-name">{{ column.column_name }}</span>

      <!-- Filter & Sort Button -->
      <button
        ref="filterButtonRef"
        class="filter-sort-btn"
        :class="{ active: isDropdownOpen(column.column_name) }"
        :data-column="column.column_name"
        @click.stop="toggleDropdown"
        :title="
          headerFilter.hasFilter(column.column_name)
            ? 'Filter active'
            : 'Filter & Sort'
        "
      >
        <span class="icon">⇅</span>
        <span
          v-if="headerFilter.getFilterCount(column.column_name) > 0"
          class="badge"
        >
          {{ headerFilter.getFilterCount(column.column_name) }}
        </span>
      </button>

      <!-- Dropdown Menu -->
      <HeaderFilterDropdown
        v-if="isDropdownOpen(column.column_name)"
        :column-name="column.column_name"
        :column-values="columnValues"
        :selected-values="selectedValues"
        :search-query="searchQuery"
        :header-filter="headerFilter"
        :trigger-element="filterButtonRef"
        @update:search-query="searchQuery = $event"
        @update:selected-values="selectedValues = $event"
        @sort-apply="handleSort"
        @filter-apply="applyFilters"
        @filter-clear="clearFilters"
      />
    </div>
  </th>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, onBeforeUnmount } from "vue";
import { useTableHeaderFilter } from "./composables/useTableHeaderFilter";
import { useHeaderFilterDropdown } from "./composables/useHeaderFilterDropdown";
import HeaderFilterDropdown from "./HeaderFilterDropdown.vue";

interface Column {
  column_name: string;
  column_type: string;
  is_key: boolean;
}

interface Props {
  column: Column;
  headerFilter: ReturnType<typeof useTableHeaderFilter>;
  fullInventoryList: any[]; // REQUIRED: pass current inventory data
}

const props = defineProps<Props>();

const emits = defineEmits<{
  "filter-apply": [columnName: string, selectedValues: string[]];
  "filter-clear": [columnName: string];
  "sort-apply": [
    data: { columnName: string; direction: "asc" | "desc" | null },
  ];
}>();

const {
  isDropdownOpen,
  toggleDropdown: toggleGlobalDropdown,
  closeDropdown,
} = useHeaderFilterDropdown();

const searchQuery = ref("");
const selectedValues = ref<Set<string>>(new Set());
const filterButtonRef = ref<HTMLElement | undefined>(undefined);

const columnValues = computed<string[]>(() => {
  // Unwrap ref to avoid reading meta fields on the Ref object itself
  const values = props.headerFilter.columnValues.value || {};

  const rawValues =
    (values as Record<string, unknown[]>)[props.column.column_name] || [];

  const uniqueValues = [...new Set(rawValues)].filter(
    (v): v is string => typeof v === "string"
  );

  return uniqueValues.sort();
});

const toggleDropdown = () => {
  toggleGlobalDropdown(props.column.column_name);

  if (isDropdownOpen(props.column.column_name)) {
    // CRITICAL FIX: Pass fullInventoryList explicitly with data parameter
    if (props.fullInventoryList && props.fullInventoryList.length > 0) {
      props.headerFilter.initializeColumnValues(
        [props.column],
        props.fullInventoryList
      );
    }

    loadCurrentValues();
    searchQuery.value = "";
  }
};

const loadCurrentValues = () => {
  const currentFilter = JSON.parse(
    localStorage.getItem(`column_filter_${props.column.column_name}`) || "[]"
  );
  selectedValues.value = new Set(currentFilter);
};

const handleSort = (data: {
  columnName: string;
  direction: "asc" | "desc" | null;
}) => {
  // Update headerFilter state first
  props.headerFilter.applySort(data.columnName, data.direction);

  // Then emit to parent
  emits("sort-apply", data);
};

const applyFilters = () => {
  const values = Array.from(selectedValues.value);
  props.headerFilter.applyFilter(props.column.column_name, values);
  emits("filter-apply", props.column.column_name, values);
  toggleGlobalDropdown(props.column.column_name);
};

const clearFilters = () => {
  selectedValues.value.clear();
  props.headerFilter.clearFilter(props.column.column_name);
  emits("filter-clear", props.column.column_name);
  toggleGlobalDropdown(props.column.column_name);
};

const handleClickOutside = (event: MouseEvent) => {
  const target = event.target as HTMLElement;
  const isClickInDropdown = target.closest(".header-filter-dropdown");
  const isClickOnButton = target.closest(".filter-sort-btn");

  if (!isClickInDropdown && !isClickOnButton) {
    closeDropdown();
  }
};

onMounted(() => {
  document.addEventListener("click", handleClickOutside);
});

onBeforeUnmount(() => {
  document.removeEventListener("click", handleClickOutside);
});
</script>

<style scoped>
.th-header {
  background-color: #f5f5f5;
  border-bottom: 2px solid #ddd;
  padding: 12px;
  text-align: left;
  font-weight: 600;
  color: #333;
  position: relative;
}

.th-header.has-active-filter {
  background-color: #fffbf0;
}

.header-content {
  display: flex;
  align-items: center;
  gap: 8px;
  position: relative;
}

.column-name {
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
  max-width: 150px;
}

.filter-sort-btn {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 4px;
  padding: 4px 6px;
  background-color: transparent;
  border: 1px solid #ddd;
  border-radius: 4px;
  cursor: pointer;
  font-size: 12px;
  transition: all 0.2s;
  color: #666;
}

.filter-sort-btn:hover {
  background-color: #e8e8e8;
  border-color: #6b7280;
}

.filter-sort-btn.active {
  background-color: #3b82f6;
  border-color: #3b82f6;
  color: white;
}

.filter-sort-btn .icon {
  font-size: 10px;
}

.badge {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 16px;
  height: 16px;
  background-color: #ef4444;
  color: white;
  border-radius: 50%;
  font-size: 11px;
  font-weight: 600;
}
</style>
