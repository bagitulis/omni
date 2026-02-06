<template>
  <Teleport to="body">
    <div
      v-if="modelValue"
      class="modal-overlay"
      @click.self="$emit('close')"
      role="dialog"
      aria-modal="true"
      aria-labelledby="change-password-title"
    >
      <div class="modal-container">
        <div class="modal-header">
          <h2 id="change-password-title" class="modal-title">
            Change Password
          </h2>
          <button
            class="modal-close-btn"
            type="button"
            aria-label="Close modal"
            @click="$emit('close')"
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
              />
            </svg>
          </button>
        </div>

        <form @submit.prevent="$emit('change-password', passwordForm)">
          <div class="modal-body">
            <div class="form-group">
              <label class="form-label" for="current-password"
                >Current Password</label
              >
              <input
                id="current-password"
                v-model="passwordForm.currentPassword"
                type="password"
                class="form-control"
                required
              />
            </div>
            <div class="form-group">
              <label class="form-label" for="new-password">New Password</label>
              <input
                id="new-password"
                v-model="passwordForm.newPassword"
                type="password"
                class="form-control"
                required
              />
            </div>
            <div class="form-group">
              <label class="form-label" for="confirm-password"
                >Confirm Password</label
              >
              <input
                id="confirm-password"
                v-model="passwordForm.confirmPassword"
                type="password"
                class="form-control"
                required
              />
            </div>

            <div v-if="message" :class="['message', status]">
              {{ message }}
            </div>
          </div>

          <div class="modal-footer">
            <button
              type="button"
              class="modal-btn modal-btn-secondary"
              @click="$emit('close')"
            >
              Cancel
            </button>
            <button type="submit" class="modal-btn modal-btn-primary">
              Change Password
            </button>
          </div>
        </form>
      </div>
    </div>
  </Teleport>
</template>

<script setup lang="ts">
import { ref, watch, onMounted, onUnmounted } from "vue";

interface Props {
  modelValue: boolean;
  message?: string;
  status?: "success" | "error" | "";
}

const props = defineProps<Props>();
defineEmits(["close", "change-password"]);

const passwordForm = ref({
  currentPassword: "",
  newPassword: "",
  confirmPassword: "",
});

// ESC key to close modal
const handleKeydown = (e: KeyboardEvent) => {
  if (e.key === "Escape" && props.modelValue) {
    // emit close handled by parent
  }
};

// Body scroll lock
watch(
  () => props.modelValue,
  (visible) => {
    document.body.style.overflow = visible ? "hidden" : "";
  },
);

onMounted(() => {
  document.addEventListener("keydown", handleKeydown);
});

onUnmounted(() => {
  document.removeEventListener("keydown", handleKeydown);
  document.body.style.overflow = "";
});
</script>

<style scoped>
/* Modal Overlay - Tokopedia Style */
.modal-overlay {
  position: fixed;
  inset: 0;
  background: rgba(0, 0, 0, 0.5);
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
  width: 100%;
  max-width: 420px;
  max-height: 90vh;
  overflow: hidden;
  display: flex;
  flex-direction: column;
  animation: slideUp 250ms cubic-bezier(0.4, 0, 0.2, 1);
}

/* Modal Header */
.modal-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 16px 20px;
  border-bottom: 1px solid #e0e0e0;
  background: #ffffff;
  flex-shrink: 0;
}

.modal-title {
  margin: 0;
  font-size: 1.125rem;
  font-weight: 700;
  color: #212121;
}

.modal-close-btn {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 32px;
  height: 32px;
  padding: 0;
  background: transparent;
  border: none;
  border-radius: 6px;
  color: #5f656e;
  cursor: pointer;
  transition: all 150ms ease;
}

.modal-close-btn:hover {
  background: #f3f4f5;
  color: #212121;
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
  background: #ffffff;
}

/* Form Elements */
.form-group {
  margin-bottom: 16px;
}

.form-label {
  display: block;
  font-weight: 600;
  margin-bottom: 6px;
  color: #212121;
  font-size: 0.875rem;
}

.form-control {
  width: 100%;
  padding: 10px 12px;
  border: 1px solid #e0e0e0;
  border-radius: 8px;
  background: #ffffff;
  color: #212121;
  font-size: 0.875rem;
  transition:
    border-color 150ms ease,
    box-shadow 150ms ease;
}

.form-control:focus {
  outline: none;
  border-color: #027a0a;
  box-shadow: 0 0 0 3px rgba(2, 122, 10, 0.1);
}

/* Message */
.message {
  margin-top: 16px;
  padding: 12px 16px;
  border-radius: 8px;
  font-size: 0.875rem;
  font-weight: 500;
}

.message.success {
  background-color: #e5f9e6;
  color: #027a0a;
  border: 1px solid #b3ebc5;
}

.message.error {
  background-color: #fef2f2;
  color: #dc2626;
  border: 1px solid #fecaca;
}

/* Modal Footer */
.modal-footer {
  display: flex;
  justify-content: flex-end;
  gap: 12px;
  padding: 16px 20px;
  border-top: 1px solid #e0e0e0;
  background: #ffffff;
  flex-shrink: 0;
}

/* Buttons */
.modal-btn {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  gap: 8px;
  padding: 10px 20px;
  border-radius: 8px;
  font-weight: 600;
  font-size: 0.875rem;
  cursor: pointer;
  transition: all 150ms ease;
  border: 1px solid transparent;
}

.modal-btn-secondary {
  background: #ffffff;
  color: #212121;
  border-color: #e0e0e0;
}

.modal-btn-secondary:hover {
  background: #f3f4f5;
  border-color: #bdbdbd;
}

.modal-btn-primary {
  background: #027a0a;
  color: white;
  border-color: #027a0a;
}

.modal-btn-primary:hover {
  background: #026208;
  border-color: #026208;
}

.modal-btn:focus-visible {
  outline: 2px solid #027a0a;
  outline-offset: 2px;
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

/* Responsive */
@media (max-width: 640px) {
  .modal-overlay {
    padding: 0;
  }

  .modal-container {
    max-width: 100%;
    max-height: 100%;
    height: 100%;
    border-radius: 0;
  }

  .modal-footer {
    flex-direction: column;
  }

  .modal-btn {
    width: 100%;
  }
}

/* Accessibility */
@media (prefers-reduced-motion: reduce) {
  .modal-overlay,
  .modal-container {
    animation: none;
  }
}
</style>
