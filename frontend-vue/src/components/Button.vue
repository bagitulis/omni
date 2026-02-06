<template>
  <button
    :class="buttonClasses"
    :disabled="disabled || loading"
    @click="handleClick"
  >
    <!-- Loading Spinner -->
    <span v-if="loading" class="btn-spinner">
      <svg class="spinner-icon" viewBox="0 0 24 24" fill="none">
        <circle
          cx="12"
          cy="12"
          r="10"
          stroke="currentColor"
          stroke-width="3"
          stroke-dasharray="32"
          stroke-linecap="round"
        />
      </svg>
    </span>

    <!-- Icon Left -->
    <span v-if="iconLeft && !loading" class="btn-icon btn-icon-left">
      <slot name="icon-left">{{ iconLeft }}</slot>
    </span>

    <!-- Label & Slot -->
    <span :class="{ 'btn-content-hidden': loading }">
      {{ label }}<slot />
    </span>

    <!-- Icon Right -->
    <span v-if="iconRight" class="btn-icon btn-icon-right">
      <slot name="icon-right">{{ iconRight }}</slot>
    </span>
  </button>
</template>

<script setup lang="ts">
import { computed } from "vue";

interface Props {
  label?: string;
  loading?: boolean;
  disabled?: boolean;
  variant?:
    | "primary"
    | "secondary"
    | "ghost"
    | "danger"
    | "success"
    | "outline";
  size?: "sm" | "md" | "lg";
  fullWidth?: boolean;
  iconLeft?: string;
  iconRight?: string;
}

const props = withDefaults(defineProps<Props>(), {
  label: "",
  loading: false,
  disabled: false,
  variant: "primary",
  size: "md",
  fullWidth: false,
  iconLeft: "",
  iconRight: "",
});

const emit = defineEmits<{ click: [event: MouseEvent] }>();

const handleClick = (event: MouseEvent) => {
  if (!props.loading && !props.disabled) {
    emit("click", event);
  }
};

const buttonClasses = computed(() => [
  "btn-base",
  `btn-${props.variant}`,
  `btn-size-${props.size}`,
  {
    "btn-loading": props.loading,
    "btn-disabled": props.disabled,
    "btn-full-width": props.fullWidth,
  },
]);
</script>

<style scoped>
.btn-base {
  position: relative;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  gap: 0.5rem;
  font-weight: 600;
  border-radius: var(--radius-md, 0.5rem);
  border: none;
  cursor: pointer;
  transition:
    background-color var(--transition-base, 200ms ease),
    transform var(--transition-fast, 150ms ease),
    box-shadow var(--transition-base, 200ms ease);
}

.btn-base:active:not(:disabled) {
  transform: scale(0.98);
}
.btn-base:focus-visible {
  outline: 2px solid var(--color-primary, #3b82f6);
  outline-offset: 2px;
}
.btn-base:focus:not(:focus-visible) {
  outline: none;
}

/* Sizes */
.btn-size-sm {
  padding: 0.375rem 0.75rem;
  font-size: 0.8125rem;
}
.btn-size-md {
  padding: 0.5rem 1rem;
  font-size: 0.875rem;
}
.btn-size-lg {
  padding: 0.75rem 1.5rem;
  font-size: 1rem;
}

/* Variants */
.btn-primary {
  background: var(--color-primary-600, #2563eb);
  color: white;
}
.btn-primary:hover:not(:disabled) {
  background: var(--color-primary-700, #1d4ed8);
  box-shadow: 0 4px 12px rgba(37, 99, 235, 0.3);
}

.btn-secondary {
  background: var(--color-secondary, #8b5cf6);
  color: white;
}
.btn-secondary:hover:not(:disabled) {
  background: #7c3aed;
  box-shadow: 0 4px 12px rgba(139, 92, 246, 0.3);
}

.btn-ghost {
  background: transparent;
  color: var(--color-text-secondary, #6b7280);
}
.btn-ghost:hover:not(:disabled) {
  background: var(--color-bg-secondary, #f8fafc);
  color: var(--color-text-primary, #1f2937);
}

.btn-outline {
  background: transparent;
  color: var(--color-primary, #3b82f6);
  border: 1px solid var(--color-primary, #3b82f6);
}
.btn-outline:hover:not(:disabled) {
  background: var(--color-primary-50, #eff6ff);
}

.btn-danger {
  background: var(--color-error, #ef4444);
  color: white;
}
.btn-danger:hover:not(:disabled) {
  background: #dc2626;
  box-shadow: 0 4px 12px rgba(239, 68, 68, 0.3);
}

.btn-success {
  background: var(--color-success, #10b981);
  color: white;
}
.btn-success:hover:not(:disabled) {
  background: #059669;
  box-shadow: 0 4px 12px rgba(16, 185, 129, 0.3);
}

/* States */
.btn-disabled,
.btn-base:disabled {
  opacity: 0.6;
  cursor: not-allowed;
}

.btn-full-width {
  width: 100%;
}

.btn-loading {
  cursor: wait;
}
.btn-content-hidden {
  opacity: 0;
}

/* Spinner */
.btn-spinner {
  position: absolute;
  display: flex;
  align-items: center;
  justify-content: center;
}
.spinner-icon {
  width: 1.25rem;
  height: 1.25rem;
  animation: spin 1s linear infinite;
}
@keyframes spin {
  from {
    transform: rotate(0deg);
  }
  to {
    transform: rotate(360deg);
  }
}

/* Icons */
.btn-icon {
  display: flex;
  align-items: center;
  font-size: 1rem;
}
</style>
