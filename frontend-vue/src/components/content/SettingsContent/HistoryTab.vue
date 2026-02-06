<template>
  <div v-if="activeTab === 'history'" class="tab-content">
    <HistoryFilters
      :filters="filters"
      :job-types="jobTypes"
      :is-loading="isLoading"
      :is-clearing="isClearing"
      @update:search="handleSearchUpdate"
      @update:status="handleStatusUpdate"
      @update:job-type="handleJobTypeUpdate"
      @update:page-size="handlePageSizeUpdate"
      @refresh="fetchHistory"
      @clear="$emit('clear-history')"
    />

    <div v-if="isLoading && historyData.length === 0" class="loading-state">
      <p>Loading history...</p>
    </div>

    <div v-else-if="historyData.length === 0" class="empty-state">
      <p>No history found</p>
      <p class="text-muted">
        {{
          hasActiveFilters
            ? "Try adjusting your filters"
            : "Completed jobs will appear here"
        }}
      </p>
    </div>

    <div v-else class="history-list">
      <HistoryTable
        :items="historyData"
        :columns="columns"
        :sort-by="sortBy"
        :sort-order="sortOrder"
        @sort="toggleSort"
      />

      <HistoryPagination
        :current-page="currentPage"
        :total-pages="totalPages"
        :pagination-start="paginationStart"
        :pagination-end="paginationEnd"
        :total-items="totalItems"
        :visible-pages="visiblePages"
        @go-to-page="goToPage"
      />
    </div>
  </div>
</template>

<script setup lang="ts">
import { onMounted, toRef } from "vue";
import HistoryFilters from "./HistoryFilters.vue";
import HistoryTable from "./HistoryTable.vue";
import HistoryPagination from "./HistoryPagination.vue";
import { useHistoryTab, type JobHistory } from "./composables/useHistoryTab";

const props = defineProps<{
  activeTab: string;
  recentHistory: JobHistory[];
  isClearing: boolean;
}>();

defineEmits<{
  "clear-history": [];
}>();

const activeTabRef = toRef(props, "activeTab");

const {
  isLoading,
  historyData,
  jobTypes,
  currentPage,
  totalItems,
  totalPages,
  sortBy,
  sortOrder,
  filters,
  columns,
  hasActiveFilters,
  paginationStart,
  paginationEnd,
  visiblePages,
  debouncedFetch,
  fetchHistory,
  fetchJobTypes,
  toggleSort,
  goToPage,
} = useHistoryTab(activeTabRef);

function handleSearchUpdate(value: string) {
  filters.value.search = value;
  debouncedFetch();
}

function handleStatusUpdate(value: string) {
  filters.value.status = value;
  fetchHistory();
}

function handleJobTypeUpdate(value: string) {
  filters.value.jobType = value;
  fetchHistory();
}

function handlePageSizeUpdate(value: number) {
  filters.value.pageSize = value;
  fetchHistory();
}

onMounted(() => {
  if (props.activeTab === "history") {
    fetchHistory();
    fetchJobTypes();
  }
});
</script>

<style src="./HistoryTab.styles.css" scoped></style>
