<template>
  <Teleport to="body">
    <div
      v-if="config"
      class="modal-overlay"
      @click.self="$emit('close-editor')"
      @keydown.esc="$emit('close-editor')"
      role="dialog"
      aria-modal="true"
      aria-labelledby="config-editor-title"
    >
      <div class="modal-container">
        <!-- Header -->
        <div class="modal-header">
          <h3 id="config-editor-title" class="modal-title">
            {{
              config.id ? "Configure " + config.name : "Add New Auto-Function"
            }}
          </h3>
          <button
            @click="$emit('close-editor')"
            class="modal-close-btn"
            type="button"
            aria-label="Close modal"
          >
            &times;
          </button>
        </div>

        <!-- Body -->
        <div class="modal-body">
          <!-- Function Name Select -->
          <div v-if="!config.id" class="form-group">
            <label for="function-name">Function Name</label>
            <select
              id="function-name"
              v-model="config.name"
              class="form-select"
            >
              <option value="">-- Select a function --</option>
              <option
                v-for="func in availableFunctions"
                :key="func.value"
                :value="func.value"
              >
                {{ func.label }}
              </option>
            </select>
          </div>

          <!-- Interval Input -->
          <div class="form-group">
            <label for="interval-input">Interval (minutes)</label>
            <div class="input-group-number">
              <button
                @click="decrementInterval"
                class="btn-spin"
                type="button"
                aria-label="Decrease interval"
              >
                −
              </button>
              <input
                id="interval-input"
                v-model.number="config.interval_minutes"
                type="number"
                min="1"
                class="form-input"
              />
              <button
                @click="incrementInterval"
                class="btn-spin"
                type="button"
                aria-label="Increase interval"
              >
                +
              </button>
            </div>
          </div>

          <!-- Time Window -->
          <div class="time-window-row">
            <div class="form-group">
              <label for="start-time">Start Time (Optional)</label>
              <div class="time-picker">
                <input
                  id="start-time"
                  v-model="config.start_time"
                  type="time"
                  class="form-input"
                />
                <div class="time-controls">
                  <button
                    @click="$emit('increment-time', 'start_time', 15)"
                    class="btn-time-control"
                    title="Add 15 min"
                    type="button"
                    aria-label="Increase start time by 15 minutes"
                  >
                    ↑
                  </button>
                  <button
                    @click="$emit('decrement-time', 'start_time', 15)"
                    class="btn-time-control"
                    title="Sub 15 min"
                    type="button"
                    aria-label="Decrease start time by 15 minutes"
                  >
                    ↓
                  </button>
                </div>
              </div>
            </div>

            <div class="form-group">
              <label for="end-time">End Time (Optional)</label>
              <div class="time-picker">
                <input
                  id="end-time"
                  v-model="config.end_time"
                  type="time"
                  class="form-input"
                />
                <div class="time-controls">
                  <button
                    @click="$emit('increment-time', 'end_time', 15)"
                    class="btn-time-control"
                    title="Add 15 min"
                    type="button"
                    aria-label="Increase end time by 15 minutes"
                  >
                    ↑
                  </button>
                  <button
                    @click="$emit('decrement-time', 'end_time', 15)"
                    class="btn-time-control"
                    title="Sub 15 min"
                    type="button"
                    aria-label="Decrease end time by 15 minutes"
                  >
                    ↓
                  </button>
                </div>
              </div>
            </div>
          </div>

          <!-- Enable Checkbox -->
          <div class="form-group checkbox-group">
            <label class="checkbox-label">
              <input
                type="checkbox"
                v-model="config.enabled"
                class="checkbox"
              />
              <span>Enable this function</span>
            </label>
          </div>
        </div>

        <!-- Footer -->
        <div class="modal-footer">
          <button
            @click="$emit('close-editor')"
            class="modal-btn modal-btn-secondary"
            type="button"
          >
            Cancel
          </button>
          <button
            @click="$emit('save-config')"
            class="modal-btn modal-btn-primary"
            type="button"
          >
            {{ config.id ? "Save" : "Create" }}
          </button>
        </div>
      </div>
    </div>
  </Teleport>
</template>

<script setup lang="ts">
/**
 * Config Editor Modal
 * Single Responsibility: Handle configuration form only
 * JSON uses snake_case as per AGENTS.md standard
 */
import { computed, onMounted, onUnmounted, watch } from "vue";
import { type AutoFunctionConfig } from "./configTableUtils";

const props = defineProps<{
  config: AutoFunctionConfig | null;
  availableFunctions: Array<{ value: string; label: string }>;
}>();

const emit = defineEmits<{
  "close-editor": [];
  "save-config": [];
  "increment-time": [field: "start_time" | "end_time", minutes: number];
  "decrement-time": [field: "start_time" | "end_time", minutes: number];
}>();

const config = computed(() => props.config);

const incrementInterval = () => {
  if (config.value) {
    config.value.interval_minutes++;
  }
};

const decrementInterval = () => {
  if (config.value) {
    config.value.interval_minutes = Math.max(
      1,
      config.value.interval_minutes - 1,
    );
  }
};

// ESC key handler
const handleKeydown = (e: KeyboardEvent) => {
  if (e.key === "Escape" && props.config) {
    emit("close-editor");
  }
};

// Body scroll lock
watch(
  () => props.config,
  (newConfig) => {
    if (newConfig) {
      document.body.style.overflow = "hidden";
    } else {
      document.body.style.overflow = "";
    }
  },
  { immediate: true },
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
  max-width: 480px;
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
  font-size: 1.5rem;
  line-height: 1;
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
  padding: 24px;
  background: #ffffff;
}

/* Form Styles */
.form-group {
  margin-bottom: 20px;
}

.form-group label {
  display: block;
  margin-bottom: 8px;
  font-weight: 600;
  font-size: 0.875rem;
  color: #212121;
}

.form-input,
.form-select {
  width: 100%;
  padding: 10px 12px;
  border: 1px solid #e0e0e0;
  border-radius: 8px;
  font-size: 0.875rem;
  color: #212121;
  background: #ffffff;
  transition: all 150ms ease;
}

.form-input:focus,
.form-select:focus {
  outline: none;
  border-color: #027a0a;
  box-shadow: 0 0 0 3px rgba(2, 122, 10, 0.1);
}

.form-select {
  cursor: pointer;
  appearance: none;
  background-image: url("data:image/svg+xml,%3Csvg xmlns='http://www.w3.org/2000/svg' width='12' height='12' viewBox='0 0 12 12'%3E%3Cpath fill='%235f656e' d='M6 8L1 3h10z'/%3E%3C/svg%3E");
  background-repeat: no-repeat;
  background-position: right 12px center;
  padding-right: 36px;
}

/* Number Input Group */
.input-group-number {
  display: flex;
  align-items: center;
  gap: 8px;
}

.input-group-number .form-input {
  flex: 1;
  text-align: center;
  font-weight: 600;
  font-size: 1rem;
}

.btn-spin {
  width: 40px;
  height: 40px;
  padding: 0;
  border: 1px solid #e0e0e0;
  background: #f3f4f5;
  border-radius: 8px;
  cursor: pointer;
  font-weight: bold;
  font-size: 1.25rem;
  color: #212121;
  transition: all 150ms ease;
  display: flex;
  align-items: center;
  justify-content: center;
}

.btn-spin:hover {
  background: #e0e0e0;
  border-color: #027a0a;
  color: #027a0a;
}

.btn-spin:focus-visible {
  outline: 2px solid #027a0a;
  outline-offset: 2px;
}

/* Time Window */
.time-window-row {
  display: flex;
  gap: 16px;
}

.time-window-row .form-group {
  flex: 1;
}

.time-picker {
  display: flex;
  gap: 8px;
  align-items: center;
}

.time-picker .form-input {
  flex: 1;
  font-weight: 600;
  text-align: center;
}

.time-controls {
  display: flex;
  flex-direction: column;
  gap: 4px;
}

.btn-time-control {
  width: 28px;
  height: 22px;
  padding: 0;
  border: 1px solid #e0e0e0;
  background: #f3f4f5;
  border-radius: 4px;
  cursor: pointer;
  font-weight: bold;
  font-size: 0.75rem;
  color: #5f656e;
  transition: all 150ms ease;
  display: flex;
  align-items: center;
  justify-content: center;
}

.btn-time-control:hover {
  background: #e0e0e0;
  border-color: #027a0a;
  color: #027a0a;
}

.btn-time-control:focus-visible {
  outline: 2px solid #027a0a;
  outline-offset: 2px;
}

/* Checkbox */
.checkbox-group {
  margin-top: 8px;
}

.checkbox-label {
  display: flex;
  align-items: center;
  gap: 10px;
  cursor: pointer;
  font-weight: 500;
  color: #212121;
}

.checkbox {
  width: 18px;
  height: 18px;
  cursor: pointer;
  accent-color: #027a0a;
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

  .time-window-row {
    flex-direction: column;
    gap: 0;
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
    transition: none;
  }
}
</style>
