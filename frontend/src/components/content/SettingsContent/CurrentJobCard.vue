<template>
  <div
    class="job-card job-running"
    :class="{ 'job-timeout-warning': isJobStuck }"
  >
    <div class="job-header">
      <div class="job-title">
        <h3>🎬 {{ job.type }}</h3>
        <span class="job-id">ID: {{ job.id.substring(0, 8) }}...</span>
        <span v-if="isJobStuck" class="timeout-badge"
          >⏱️ STUCK ({{ durationMinutes }}m)</span
        >
      </div>
      <div class="btn-group">
        <button @click="$emit('cancel')" class="btn btn-danger btn-sm">
          ❌ Cancel
        </button>
        <button
          v-if="isJobStuck"
          @click="$emit('force-cancel')"
          class="btn btn-urgent btn-sm"
          title="Force cancel stuck job"
        >
          ⚡ Force Cancel
        </button>
      </div>
    </div>

    <div class="job-details">
      <div class="detail-row">
        <span class="label">Status:</span>
        <span
          class="badge badge-running"
          :class="{ 'badge-timeout': isJobStuck }"
        >
          {{ job.status }}
        </span>
      </div>
      <div class="detail-row">
        <span class="label">Priority:</span>
        <span class="badge" :class="`priority-${job.priority}`">
          {{ job.priority }}
        </span>
      </div>
      <div class="detail-row">
        <span class="label">Started:</span>
        <span>{{ formatDateTime(job.started_at) }}</span>
      </div>
      <div class="detail-row">
        <span class="label">Duration:</span>
        <span :class="{ 'duration-warning': isJobStuck }">
          {{ duration }}
          <span v-if="isJobStuck" class="duration-exceeded">
            (Timeout: 5 min)</span
          >
        </span>
      </div>
      <div v-if="job.data" class="detail-row">
        <span class="label">Data:</span>
        <code class="data-preview">{{ JSON.stringify(job.data) }}</code>
      </div>
    </div>

    <div class="job-progress" :class="{ 'progress-stuck': isJobStuck }">
      <div class="progress-bar">
        <div
          class="progress-fill"
          :style="{ width: progressPercent + '%' }"
        ></div>
      </div>
      <span class="progress-text">
        {{
          isJobStuck
            ? "STUCK - Waiting for timeout or force cancel"
            : "Processing..."
        }}
      </span>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed } from "vue";
import {
  formatDateTime,
  calculateDuration,
  getJobDurationMinutes,
  getJobProgressPercent,
  type Job,
} from "./currentJobUtils";

interface Props {
  job: Job;
  isJobStuck: boolean;
}

const props = defineProps<Props>();

defineEmits<{
  cancel: [];
  "force-cancel": [];
}>();

const duration = computed(() => calculateDuration(props.job.started_at));
const durationMinutes = computed(() =>
  getJobDurationMinutes(props.job.started_at),
);
const progressPercent = computed(() =>
  getJobProgressPercent(props.job.started_at),
);
</script>

<style src="./CurrentJobCard.styles.css" scoped></style>
