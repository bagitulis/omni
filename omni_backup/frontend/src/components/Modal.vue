<template>
  <teleport to="body">
    <transition name="modal-fade">
      <div
        v-if="isOpen"
        class="modal-backdrop"
        @click="closeOnBackdrop"
      >
        <div
          class="modal-container"
          @click.stop
        >
          <!-- Header -->
          <div class="modal-header">
            <h2 class="modal-title">
              {{ title }}
            </h2>
            <button
              @click.stop="$emit('close')"
              type="button"
              class="modal-close-btn"
              aria-label="Close modal"
            >
              <svg
                class="w-6 h-6"
                fill="none"
                stroke="currentColor"
                viewBox="0 0 24 24"
              >
                <path
                  stroke-linecap="round"
                  stroke-linejoin="round"
                  stroke-width="2"
                  d="M6 18L18 6M6 6l12 12"
                ></path>
              </svg>
            </button>
          </div>

          <!-- Body -->
          <div class="modal-body">
            <slot />
          </div>

          <!-- Footer -->
          <div
            v-if="$slots.footer"
            class="modal-footer"
          >
            <slot name="footer" />
          </div>
        </div>
      </div>
    </transition>
  </teleport>
</template>

<script setup lang="ts">
import { watch } from "vue";

interface Props {
  isOpen: boolean;
  title: string;
  closeOnBackdropClick?: boolean;
}

const props = withDefaults(defineProps<Props>(), {
  closeOnBackdropClick: true,
});

const emit = defineEmits<{
  close: [];
}>();

const closeOnBackdrop = () => {
  if (props.closeOnBackdropClick) {
    emit("close");
  }
};

// Prevent body scroll when modal is open
watch(
  () => props.isOpen,
  (isOpen) => {
    if (isOpen) {
      document.body.style.overflow = "hidden";
    } else {
      document.body.style.overflow = "";
    }
  }
);
</script>

<style scoped>
/* Modal Backdrop */
.modal-backdrop {
  position: fixed;
  inset: 0;
  background-color: rgba(0, 0, 0, 0.5);
  backdrop-filter: blur(4px);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 9999;
  padding: 1rem;
  animation: fadeIn 300ms ease-out;
}

/* Modal Container */
.modal-container {
  background: linear-gradient(135deg, #ffffff 0%, #f8fafb 100%);
  border-radius: 0.75rem;
  box-shadow: 
    0 20px 60px rgba(0, 0, 0, 0.3),
    0 0 1px rgba(148, 255, 255, 0.2);
  max-width: 32rem;
  width: 100%;
  display: flex;
  flex-direction: column;
  max-height: 90vh;
  overflow: hidden;
  border: 1px solid #e5e7eb;
  animation: slideUp 300ms cubic-bezier(0.4, 0, 0.2, 1);
}

/* Modal Header */
.modal-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 1.5rem 1.5rem;
  border-bottom: 1px solid #e5e7eb;
  flex-shrink: 0;
  background: linear-gradient(
    135deg,
    #f9fafb 0%,
    #f3f4f6 100%
  );
  border-top-left-radius: 0.75rem;
  border-top-right-radius: 0.75rem;
}

.modal-title {
  font-size: 1.125rem;
  font-weight: 700;
  color: #1f2937;
  letter-spacing: 0.3px;
}

.modal-close-btn {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 2rem;
  height: 2rem;
  border-radius: 0.375rem;
  color: #6b7280;
  background: transparent;
  border: none;
  cursor: pointer;
  transition: all 200ms ease;
  padding: 0;
}

.modal-close-btn:hover {
  background-color: #e5e7eb;
  color: #3b82f6;
}

.modal-close-btn:active {
  transform: scale(0.95);
}

/* Modal Body */
.modal-body {
  flex: 1;
  overflow-y: auto;
  padding: 1.5rem;
  color: #1f2937;
  background: #ffffff;
}

.modal-body::-webkit-scrollbar {
  width: 6px;
}

.modal-body::-webkit-scrollbar-track {
  background: #f9fafb;
  border-radius: 3px;
}

.modal-body::-webkit-scrollbar-thumb {
  background: #d1d5db;
  border-radius: 3px;
}

.modal-body::-webkit-scrollbar-thumb:hover {
  background: #9ca3af;
}

/* Modal Footer */
.modal-footer {
  display: flex;
  align-items: center;
  justify-content: flex-end;
  gap: 0.75rem;
  padding: 1rem 1.5rem;
  border-top: 1px solid #e5e7eb;
  flex-shrink: 0;
  background: linear-gradient(
    135deg,
    #f9fafb 0%,
    #f3f4f6 100%
  );
  border-bottom-left-radius: 0.75rem;
  border-bottom-right-radius: 0.75rem;
}

/* Animations */
@keyframes fadeIn {
  from {
    opacity: 0;
  }
  to {
    opacity: 1;
  }
}

@keyframes slideUp {
  from {
    opacity: 0;
    transform: translateY(20px);
  }
  to {
    opacity: 1;
    transform: translateY(0);
  }
}

.modal-fade-enter-active,
.modal-fade-leave-active {
  transition: all 300ms cubic-bezier(0.4, 0, 0.2, 1);
}

.modal-fade-enter-from,
.modal-fade-leave-to {
  opacity: 0;
}

.modal-fade-enter-to,
.modal-fade-leave-from {
  opacity: 1;
}

/* Dark Mode Adjustments */
@media (prefers-color-scheme: dark) {
  .modal-backdrop {
    background-color: rgba(0, 0, 0, 0.6);
  }

  .modal-container {
    box-shadow: 
      0 20px 60px rgba(0, 0, 0, 0.5),
      0 0 1px rgba(148, 255, 255, 0.3);
  }
}

/* Responsive */
@media (max-width: 640px) {
  .modal-container {
    max-width: calc(100% - 1rem);
    max-height: 95vh;
  }

  .modal-header {
    padding: 1rem 1.25rem;
  }

  .modal-title {
    font-size: 1rem;
  }

  .modal-body {
    padding: 1rem 1.25rem;
  }

  .modal-footer {
    padding: 0.75rem 1.25rem;
  }
}
</style>
