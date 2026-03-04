<template>
  <transition-group
    name="toast"
    tag="div"
    class="fixed top-4 right-4 z-50 space-y-3 pointer-events-none"
  >
    <div
      v-for="toast in toasts"
      :key="toast.id"
      :class="[
        'alert shadow-xl min-w-72 max-w-sm pointer-events-auto',
        getAlertClass(toast.severity),
      ]"
    >
      <div class="flex items-start gap-3 w-full">
        <!-- Icon -->
        <div class="flex-shrink-0 text-lg mt-0.5">
          {{ getToastIcon(toast.severity) }}
        </div>

        <!-- Content -->
        <div class="flex-1 min-w-0">
          <h3 v-if="toast.summary" class="font-bold text-sm leading-tight">
            {{ toast.summary }}
          </h3>
          <p v-if="toast.detail" class="text-xs mt-1 opacity-90 leading-tight">
            {{ toast.detail }}
          </p>
        </div>

        <!-- Close Button -->
        <button
          @click="remove(toast.id)"
          class="btn btn-ghost btn-sm btn-circle flex-shrink-0"
          aria-label="Close"
        >
          <i class="pi pi-times text-base" aria-hidden="true"></i>
        </button>
      </div>
    </div>
  </transition-group>
</template>

<script setup lang="ts">
import { ref } from "vue";

interface Toast {
  id: string;
  summary: string;
  detail?: string;
  severity: "success" | "info" | "warning" | "error";
  life?: number;
  timeoutId?: NodeJS.Timeout;
}

const toasts = ref<Toast[]>([]);
let nextId = 0;

const add = (
  summary: string,
  detail?: string,
  severity: Toast["severity"] = "info",
  life: number = 3000,
) => {
  const id = `toast-${nextId++}`;
  const toast: Toast = { id, summary, detail, severity, life };

  toasts.value.push(toast);

  // Auto-remove after life duration
  if (life > 0) {
    const timeoutId = setTimeout(() => {
      remove(id);
    }, life);

    // Store timeout ID for cleanup
    const toastIndex = toasts.value.findIndex((t) => t.id === id);
    if (toastIndex > -1) {
      toasts.value[toastIndex].timeoutId = timeoutId;
    }
  }
};

const remove = (id: string) => {
  const index = toasts.value.findIndex((t) => t.id === id);
  if (index > -1) {
    // Clear timeout if exists
    if (toasts.value[index].timeoutId) {
      clearTimeout(toasts.value[index].timeoutId);
    }
    toasts.value.splice(index, 1);
  }
};

const clear = () => {
  // Clear all timeouts
  toasts.value.forEach((toast) => {
    if (toast.timeoutId) {
      clearTimeout(toast.timeoutId);
    }
  });
  toasts.value = [];
};

const getAlertClass = (severity: Toast["severity"]) => {
  const classes: Record<Toast["severity"], string> = {
    success: "alert-success",
    info: "alert-info",
    warning: "alert-warning",
    error: "alert-error",
  };
  return classes[severity];
};

const getToastIcon = (severity: Toast["severity"]) => {
  const icons: Record<Toast["severity"], string> = {
    success: "✓",
    info: "ℹ",
    warning: "⚠",
    error: "✕",
  };
  return icons[severity];
};

// Expose methods for global use
defineExpose({ add, remove, clear });
</script>

<style scoped>
.toast-enter-active,
.toast-leave-active {
  transition: all 300ms;
}

.toast-enter-from,
.toast-leave-to {
  opacity: 0;
  transform: translateX(2rem);
}

.toast-move {
  transition: all 300ms;
}
</style>
