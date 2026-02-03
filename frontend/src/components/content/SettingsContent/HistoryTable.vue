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
              <Icon
                v-if="sortBy === col.key"
                :name="sortOrder === 'asc' ? 'sort-asc' : 'sort-desc'"
                size="xs"
              />
              <Icon v-else name="chevron-up" size="xs" class="sort-inactive" />
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
        <td class="job-type">{{ record.job_type || "Unknown" }}</td>
        <td class="job-id" :title="record.job_id">
          {{ truncateId(record.job_id) }}
        </td>
        <td>
          <span class="badge" :class="`status-${record.status}`">
            {{ getStatusEmoji(record.status) }} {{ record.status }}
          </span>
        </td>
        <td class="duration">
          {{ formatDuration(record.duration_ms) }}
        </td>
        <td class="completed-time">
          {{ formatDateTime(record.completed_at || record.created_at) }}
        </td>
        <td class="error-msg">
          <span
            v-if="record.error_message"
            class="error-text"
            :title="record.error_message"
          >
            {{ truncateError(record.error_message) }}
          </span>
          <span v-else class="text-muted">-</span>
        </td>
      </tr>
    </tbody>
  </table>
</template>

<script setup lang="ts">
import Icon from "@/components/ui/Icon.vue";
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
