<template>
  <teleport to="body">
    <transition name="modal-fade">
      <div v-if="isOpen" class="modal-backdrop" @click="closeOnBackdrop">
        <div
          ref="containerRef"
          class="modal-container"
          role="dialog"
          aria-modal="true"
          :aria-labelledby="titleId"
          @click.stop
        >
          <!-- Header -->
          <div class="modal-header">
            <h2 :id="titleId" class="modal-title">
              {{ title }}
            </h2>
            <button
              @click.stop="$emit('close')"
              type="button"
              class="modal-close-btn"
              aria-label="Close modal"
            >
              <svg
                width="20"
                height="20"
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
          <div v-if="$slots.footer" class="modal-footer">
            <slot name="footer" />
          </div>
        </div>
      </div>
    </transition>
  </teleport>
</template>

<script setup lang="ts">
import { ref, watch, computed, onMounted, onUnmounted, toRef } from "vue";
import { useFocusTrap } from "@/composables/useFocusTrap";

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

// Container ref for focus trap
const containerRef = ref<HTMLElement | null>(null);
const isOpenRef = toRef(props, "isOpen");

// Focus trap for accessibility
useFocusTrap(containerRef, isOpenRef);

// Generate unique ID for aria-labelledby
const titleId = computed(
  () => `modal-title-${Math.random().toString(36).substr(2, 9)}`,
);

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
  },
);

// ESC key to close modal
const handleKeydown = (e: KeyboardEvent) => {
  if (e.key === "Escape" && props.isOpen) {
    emit("close");
  }
};

onMounted(() => {
  document.addEventListener("keydown", handleKeydown);
});

onUnmounted(() => {
  document.removeEventListener("keydown", handleKeydown);
  document.body.style.overflow = "";
});
</script>

<style scoped>
/* Modal Backdrop - Tokopedia Style */
.modal-backdrop {
  position: fixed;
  inset: 0;
  background-color: rgba(0, 0, 0, 0.5);
  backdrop-filter: blur(4px);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 9999;
  padding: 16px;
  animation: fadeIn 200ms ease-out;
}

/* Modal Container */
.modal-container {
  background: #ffffff;
  border-radius: 12px;
  box-shadow: 0 20px 60px rgba(0, 0, 0, 0.15);
  max-width: 500px;
  width: 100%;
  display: flex;
  flex-direction: column;
  max-height: 90vh;
  overflow: hidden;
  animation: slideUp 250ms cubic-bezier(0.4, 0, 0.2, 1);
}

/* Modal Header */
.modal-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 16px 20px;
  border-bottom: 1px solid #e0e0e0;
  flex-shrink: 0;
  background: #ffffff;
}

.modal-title {
  font-size: 1.125rem;
  font-weight: 700;
  color: #212121;
  margin: 0;
  line-height: 1.4;
}

.modal-close-btn {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 32px;
  height: 32px;
  border-radius: 6px;
  color: #6c727c;
  background: transparent;
  border: none;
  cursor: pointer;
  transition: all 150ms ease;
  padding: 0;
}

.modal-close-btn:hover {
  background-color: #f3f4f5;
  color: #212121;
}

.modal-close-btn:active {
  transform: scale(0.95);
}

.modal-close-btn:focus-visible {
  outline: 2px solid #027a0a;
  outline-offset: 2px;
}

/* Modal Body */
.modal-body {
  flex: 1;
  overflow-y: auto;
  padding: 20px;
  color: #212121;
  background: #ffffff;
}

.modal-body::-webkit-scrollbar {
  width: 6px;
}

.modal-body::-webkit-scrollbar-track {
  background: transparent;
}

.modal-body::-webkit-scrollbar-thumb {
  background: #e0e0e0;
  border-radius: 3px;
}

.modal-body::-webkit-scrollbar-thumb:hover {
  background: #bdbdbd;
}

/* Modal Footer */
.modal-footer {
  display: flex;
  align-items: center;
  justify-content: flex-end;
  gap: 12px;
  padding: 16px 20px;
  border-top: 1px solid #e0e0e0;
  flex-shrink: 0;
  background: #ffffff;
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
  transition: all 250ms cubic-bezier(0.4, 0, 0.2, 1);
}

.modal-fade-enter-from,
.modal-fade-leave-to {
  opacity: 0;
}

.modal-fade-enter-to,
.modal-fade-leave-from {
  opacity: 1;
}

/* Responsive */
@media (max-width: 640px) {
  .modal-backdrop {
    padding: 0;
  }

  .modal-container {
    max-width: 100%;
    max-height: 100%;
    height: 100%;
    border-radius: 0;
  }

  .modal-header {
    padding: 12px 16px;
  }

  .modal-title {
    font-size: 1rem;
  }

  .modal-body {
    padding: 16px;
  }

  .modal-footer {
    padding: 12px 16px;
    flex-direction: column;
  }

  .modal-footer > * {
    width: 100%;
  }
}

/* Accessibility */
@media (prefers-reduced-motion: reduce) {
  .modal-backdrop,
  .modal-container {
    animation: none;
  }
}
</style>
