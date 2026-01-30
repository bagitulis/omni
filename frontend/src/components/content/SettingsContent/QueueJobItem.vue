<template>
  <div class="queue-item">
    <div class="queue-number">{{ index + 1 }}</div>
    <div class="queue-info">
      <div class="queue-title">{{ job.type }}</div>
      <div class="queue-meta">
        <span class="queue-id">{{ job.id.substring(0, 12) }}...</span>
        <span class="queue-created">{{ formatTime(job.created_at) }}</span>
      </div>
    </div>
    <div class="queue-priority">
      <span class="badge" :class="`priority-${job.priority}`">
        {{ job.priority }}
      </span>
    </div>
    <button @click="$emit('cancel')" class="btn btn-danger btn-sm">
      ⚡ Force Cancel
    </button>
  </div>
</template>

<script setup lang="ts">
import { formatTime, type Job } from "./queueUtils";

defineProps<{
  job: Job;
  index: number;
}>();

defineEmits<{
  cancel: [];
}>();
</script>

<style scoped lang="css">
.queue-item {
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

.queue-item:hover {
  background: #ffffff;
  border-color: #0066cc;
  box-shadow: 0 2px 8px rgba(0, 102, 204, 0.1);
}

.queue-number {
  display: flex;
  align-items: center;
  justify-content: center;
  min-width: 32px;
  height: 32px;
  background: #e3f2fd;
  color: #0066cc;
  border-radius: 50%;
  font-weight: 700;
  font-size: 14px;
  flex-shrink: 0;
}

.queue-info {
  flex: 1;
  min-width: 0;
}

.queue-title {
  font-size: 14px;
  font-weight: 600;
  color: #1a1a1a;
  margin-bottom: 4px;
}

.queue-meta {
  display: flex;
  align-items: center;
  gap: 12px;
  font-size: 12px;
  color: #6b7280;
}

.queue-id {
  font-family: "Courier New", monospace;
}

.queue-created::before {
  content: "⏰";
  margin-right: 4px;
}

.queue-priority {
  display: flex;
  align-items: center;
}

.badge {
  padding: 4px 10px;
  border-radius: 4px;
  font-size: 11px;
  font-weight: 700;
  text-transform: uppercase;
}

.priority-high {
  background: #fff3cd;
  color: #856404;
  animation: pulse 1.5s ease-in-out infinite;
}

.priority-normal {
  background: #d1ecf1;
  color: #0c5460;
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
  .queue-item {
    flex-direction: column;
    align-items: flex-start;
  }

  .queue-priority {
    width: 100%;
  }
}
</style>
