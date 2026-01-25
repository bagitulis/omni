<template>
  <table class="history-table">
    <thead>
      <tr>
        <th
          v-for="col in columns"
          :key="col.key"
          :class="['sortable', { active: sortBy === col.key }]"
          @click="col.sortable && $emit('sort', col.key)"
        >
          <div class="th-content">
            <span>{{ col.label }}</span>
            <span v-if="col.sortable" class="sort-icon">
              <span v-if="sortBy === col.key">
                {{ sortOrder === "asc" ? "↑" : "↓" }}
              </span>
              <span v-else class="sort-inactive">↕</span>
            </span>
          </div>
        </th>
      </tr>
    </thead>
    <tbody>
      <tr
        v-for="record in items"
        :key="record.id"
        :class="`history-${record.status}`"
      >
        <td class="job-type">{{ record.jobType || "Unknown" }}</td>
        <td class="job-id" :title="record.jobId">
          {{ truncateId(record.jobId) }}
        </td>
        <td>
          <span class="badge" :class="`status-${record.status}`">
            {{ getStatusEmoji(record.status) }} {{ record.status }}
          </span>
        </td>
        <td class="duration">
          {{ formatDuration(record.durationMs) }}
        </td>
        <td class="completed-time">
          {{ formatDateTime(record.completedAt || record.createdAt) }}
        </td>
        <td class="error-msg">
          <span
            v-if="record.errorMessage"
            class="error-text"
            :title="record.errorMessage"
          >
            {{ truncateError(record.errorMessage) }}
          </span>
          <span v-else class="text-muted">-</span>
        </td>
      </tr>
    </tbody>
  </table>
</template>

<script setup lang="ts">
import type { JobHistory, ColumnDef } from "./composables/useHistoryTab";
import {
  truncateId,
  truncateError,
  getStatusEmoji,
  formatDuration,
  formatDateTime,
} from "./utils/historyFormatters";

defineProps<{
  items: JobHistory[];
  columns: ColumnDef[];
  sortBy: string;
  sortOrder: "asc" | "desc";
}>();

defineEmits<{
  sort: [key: string];
}>();
</script>

<style src="./HistoryTab.styles.css" scoped></style>
