<template>
  <div class="history-filters">
    <div class="filter-row">
      <div class="filter-group search-group">
        <label>🔍 Search</label>
        <input
          :value="filters.search"
          type="text"
          placeholder="Search job ID, type, error..."
          class="filter-input"
          @input="
            $emit('update:search', ($event.target as HTMLInputElement).value)
          "
        />
      </div>

      <div class="filter-group">
        <label for="filter-status">📊 Status</label>
        <select
          id="filter-status"
          :value="filters.status"
          class="filter-select"
          @change="
            $emit('update:status', ($event.target as HTMLSelectElement).value)
          "
        >
          <option value="all">All Status</option>
          <option value="completed">✅ Completed</option>
          <option value="failed">❌ Failed</option>
          <option value="pending">⏳ Pending</option>
        </select>
      </div>

      <div class="filter-group">
        <label for="filter-job-type">📋 Job Type</label>
        <select
          id="filter-job-type"
          :value="filters.jobType"
          class="filter-select"
          @change="
            $emit('update:jobType', ($event.target as HTMLSelectElement).value)
          "
        >
          <option value="all">All Types</option>
          <option v-for="type in jobTypes" :key="type" :value="type">
            {{ type }}
          </option>
        </select>
      </div>

      <div class="filter-group">
        <label for="filter-page-size">📄 Per Page</label>
        <select
          id="filter-page-size"
          :value="filters.pageSize"
          class="filter-select"
          @change="
            $emit(
              'update:pageSize',
              parseInt(($event.target as HTMLSelectElement).value)
            )
          "
        >
          <option :value="10">10</option>
          <option :value="20">20</option>
          <option :value="50">50</option>
          <option :value="100">100</option>
        </select>
      </div>

      <div class="filter-group filter-actions">
        <button
          @click="$emit('clear')"
          :disabled="isClearing || isLoading"
          class="btn btn-danger btn-sm"
          title="Delete all job history records"
        >
          {{ isClearing ? "Clearing..." : "🗑️ Clear All" }}
        </button>
        <button
          @click="$emit('refresh')"
          :disabled="isLoading"
          class="btn btn-primary btn-sm"
          title="Refresh history"
        >
          {{ isLoading ? "..." : "🔄" }}
        </button>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import type { HistoryFilters } from "./composables/useHistoryTab";

defineProps<{
  filters: HistoryFilters;
  jobTypes: string[];
  isLoading: boolean;
  isClearing: boolean;
}>();

defineEmits<{
  "update:search": [value: string];
  "update:status": [value: string];
  "update:jobType": [value: string];
  "update:pageSize": [value: number];
  refresh: [];
  clear: [];
}>();
</script>

<style src="./HistoryTab.styles.css" scoped></style>
