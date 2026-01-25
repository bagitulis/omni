<template>
  <Transition name="alert">
    <div v-if="show && alert" class="alert" :class="`alert-${alert.type}`">
      <div class="alert-icon">{{ getAlertIcon }}</div>
      <div class="alert-content">
        <div class="alert-title">{{ alert.title }}</div>
        <div class="alert-message">{{ alert.message }}</div>
        <div v-if="alert.details" class="alert-details">
          {{ alert.details }}
        </div>
      </div>
      <button
        @click="$emit('close-alert')"
        class="btn-close"
        type="button"
        aria-label="Close alert"
      >
        ✕
      </button>
    </div>
  </Transition>
</template>

<script setup lang="ts">
import { computed } from "vue";

interface Alert {
  type: "success" | "error" | "warning" | "info";
  title: string;
  message: string;
  details?: string;
}

const props = defineProps<{
  alert: Alert | null;
  show: boolean;
}>();

defineEmits<{
  "close-alert": [];
}>();

const getAlertIcon = computed(() => {
  switch (props.alert?.type) {
    case "success":
      return "✅";
    case "error":
      return "❌";
    case "warning":
      return "⚠️";
    case "info":
      return "ℹ️";
    default:
      return "📢";
  }
});
</script>

<style scoped>
@import "./InventoryContent.styles.css";

.alert {
  display: flex;
  gap: 12px;
  padding: 12px 15px;
  margin-bottom: 15px;
  border-radius: 6px;
  border-left: 4px solid;
  background-color: white;
  animation: slideIn 0.3s ease-out;
}

.alert-icon {
  font-size: 18px;
  flex-shrink: 0;
}

.alert-content {
  flex: 1;
}

.alert-title {
  font-size: 13px;
  font-weight: 600;
  margin-bottom: 4px;
}

.alert-message {
  font-size: 12px;
  line-height: 1.4;
  margin-bottom: 4px;
}

.alert-details {
  font-size: 11px;
  color: #6b7280;
  margin-top: 4px;
  font-family: monospace;
}

.btn-close {
  background: none;
  border: none;
  font-size: 16px;
  cursor: pointer;
  color: #6b7280;
  transition: color 0.2s;
  flex-shrink: 0;
}

.btn-close:hover {
  color: #333;
}

.alert-success {
  border-color: #27ae60;
  background-color: #f0fdf4;
}

.alert-success .alert-title {
  color: #27ae60;
}

.alert-success .alert-message {
  color: #166534;
}

.alert-error {
  border-color: #e74c3c;
  background-color: #fef2f2;
}

.alert-error .alert-title {
  color: #e74c3c;
}

.alert-error .alert-message {
  color: #991b1b;
}

.alert-warning {
  border-color: #f39c12;
  background-color: #fffbeb;
}

.alert-warning .alert-title {
  color: #f39c12;
}

.alert-warning .alert-message {
  color: #b45309;
}

.alert-info {
  border-color: #3498db;
  background-color: #eff6ff;
}

.alert-info .alert-title {
  color: #3498db;
}

.alert-info .alert-message {
  color: #1e40af;
}

.alert-enter-active,
.alert-leave-active {
  transition: all 0.3s ease;
}

.alert-enter-from,
.alert-leave-to {
  opacity: 0;
  transform: translateY(-10px);
}

@keyframes slideIn {
  from {
    opacity: 0;
    transform: translateY(-10px);
  }
  to {
    opacity: 1;
    transform: translateY(0);
  }
}

@media (max-width: 768px) {
  .alert {
    padding: 10px 12px;
    gap: 10px;
  }

  .alert-title {
    font-size: 12px;
  }

  .alert-message {
    font-size: 11px;
  }
}
</style>
