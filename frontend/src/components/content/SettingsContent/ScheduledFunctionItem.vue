<template>
  <div class="scheduled-item" :class="{ overdue: isScheduleOverdue(config) }">
    <div class="scheduled-number">{{ index + 1 }}</div>
    <div class="scheduled-info">
      <div class="scheduled-title">{{ config.name }}</div>
      <div class="scheduled-meta">
        <span class="scheduled-interval"
          >{{ config.interval_minutes }}m interval</span
        >
        <span
          class="scheduled-status"
          :style="{ color: config.enabled ? '#28a745' : '#dc3545' }"
        >
          {{ config.enabled ? "✓ Enabled" : "✗ Disabled" }}
        </span>
      </div>
    </div>
    <div class="scheduled-time" v-if="config.next_scheduled_execution">
      <div class="scheduled-next">
        <span class="label">Next:</span>
        <span class="time" :class="{ overdue: isScheduleOverdue(config) }">
          {{ formatDateTime(config.next_scheduled_execution) }}
        </span>
        <span class="countdown">{{ getTimeUntilTrigger(config) }}</span>
      </div>
    </div>
    <button
      @click="$emit('cancel')"
      class="btn btn-danger btn-sm"
      title="Remove scheduled execution"
    >
      Cancel
    </button>
  </div>
</template>

<script setup lang="ts">
import {
  formatDateTime,
  getTimeUntilTrigger,
  isScheduleOverdue,
  type AutoFunctionConfig,
} from "./queueUtils";

defineProps<{
  config: AutoFunctionConfig;
  index: number;
}>();

defineEmits<{
  cancel: [];
}>();
</script>

<style scoped lang="css">
.scheduled-item {
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 12px;
  background: #f8f9fa;
  border: 1px solid #e0e0e0;
  border-radius: 8px;
  transition: all 0.2s ease;
  animation: slideInLeft 0.3s ease-out;
}

@keyframes slideInLeft {
  from {
    opacity: 0;
    transform: translateX(-12px);
  }
  to {
    opacity: 1;
    transform: translateX(0);
  }
}

.scheduled-item.overdue {
  border-color: #f39c12;
  background: linear-gradient(135deg, #fffbf0 0%, #ffffff 100%);
}

.scheduled-item:hover {
  background: #ffffff;
  border-color: #0066cc;
  box-shadow: 0 2px 8px rgba(0, 102, 204, 0.1);
}

.scheduled-number {
  display: flex;
  align-items: center;
  justify-content: center;
  min-width: 32px;
  height: 32px;
  background: #fff3cd;
  color: #856404;
  border-radius: 50%;
  font-weight: 700;
  font-size: 14px;
  flex-shrink: 0;
}

.scheduled-info {
  flex: 1;
  min-width: 0;
}

.scheduled-title {
  font-size: 14px;
  font-weight: 600;
  color: #1a1a1a;
  margin-bottom: 4px;
}

.scheduled-meta {
  display: flex;
  align-items: center;
  gap: 12px;
  font-size: 12px;
  color: #6b7280;
}

.scheduled-interval::before {
  content: "⏱️";
  margin-right: 4px;
}

.scheduled-status {
  font-weight: 700;
}

.scheduled-time {
  display: flex;
  flex-direction: column;
  align-items: flex-end;
  gap: 4px;
  text-align: right;
}

.scheduled-next {
  display: flex;
  align-items: center;
  gap: 8px;
}

.scheduled-next .label {
  font-size: 11px;
  color: #6b7280;
  font-weight: 600;
}

.scheduled-next .time {
  font-size: 13px;
  font-weight: 600;
  color: #1a1a1a;
  font-family: "Courier New", monospace;
}

.scheduled-next .time.overdue {
  color: #f39c12;
  animation: pulse 1s ease-in-out infinite;
}

.countdown {
  font-size: 11px;
  color: #6b7280;
}

@keyframes pulse {
  0%,
  100% {
    opacity: 1;
  }
  50% {
    opacity: 0.7;
  }
}

.btn {
  padding: 6px 12px;
  background: #e74c3c;
  color: white;
  border: none;
  border-radius: 6px;
  font-size: 12px;
  font-weight: 600;
  cursor: pointer;
  transition: all 0.2s ease;
  white-space: nowrap;
}

.btn:hover {
  background: #c0392b;
  box-shadow: 0 2px 8px rgba(231, 76, 60, 0.2);
  transform: translateY(-1px);
}

@media (max-width: 768px) {
  .scheduled-item {
    flex-direction: column;
    align-items: flex-start;
  }

  .scheduled-time {
    width: 100%;
    align-items: flex-start;
    text-align: left;
  }
}
</style>
