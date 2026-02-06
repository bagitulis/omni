<template>
  <span class="action-badge" :class="badgeClass">
    <span class="badge-icon" aria-hidden="true">{{ icon }}</span>
    <span class="badge-text">{{ label }}</span>
  </span>
</template>

<script setup lang="ts">
import { computed } from "vue";

const props = defineProps<{
  action: string;
}>();

const badgeClass = computed(() => {
  switch (props.action) {
    case "SCALE_UP":
      return "action-scale-up";
    case "MAINTAIN":
      return "action-maintain";
    case "REDUCE":
    case "MONITOR":
      return "action-reduce";
    case "STOP":
      return "action-stop";
    default:
      return "action-default";
  }
});

const icon = computed(() => {
  switch (props.action) {
    case "SCALE_UP":
      return "↑";
    case "MAINTAIN":
      return "―";
    case "REDUCE":
    case "MONITOR":
      return "↓";
    case "STOP":
      return "✕";
    default:
      return "?";
  }
});

const label = computed(() => {
  switch (props.action) {
    case "SCALE_UP":
      return "Scale Up";
    case "MAINTAIN":
      return "Maintain";
    case "REDUCE":
    case "MONITOR":
      return "Reduce";
    case "STOP":
      return "Stop";
    default:
      return props.action;
  }
});
</script>

<style scoped>
.action-badge {
  display: inline-flex;
  align-items: center;
  gap: 4px;
  padding: 4px 10px;
  border-radius: 12px;
  font-size: 12px;
  font-weight: 600;
  white-space: nowrap;
}

.badge-icon {
  font-weight: bold;
}

.action-scale-up {
  background: #d1fae5;
  color: #059669;
}

.action-maintain {
  background: #dbeafe;
  color: #2563eb;
}

.action-reduce {
  background: #fef3c7;
  color: #d97706;
}

.action-stop {
  background: #fee2e2;
  color: #dc2626;
}

.action-default {
  background: #f3f4f6;
  color: #6b7280;
}
</style>
