<template>
  <Teleport to="body">
    <div v-if="show" class="modal-overlay" @click.self="$emit('close')">
      <div
        class="batch-modal"
        role="dialog"
        aria-modal="true"
        :aria-labelledby="titleId"
      >
        <div class="modal-header">
          <h3 :id="titleId">
            {{
              type === "price"
                ? "💰 Update Harga Batch"
                : "📦 Update Stok Batch"
            }}
          </h3>
          <button class="modal-close" @click="$emit('close')">✕</button>
        </div>
        <div class="modal-body">
          <p class="modal-info">
            Anda akan mengupdate <strong>{{ count }}</strong> varian sekaligus
          </p>
          <div class="form-group">
            <label>{{
              type === "price" ? "Harga Baru (Rp)" : "Stok Baru"
            }}</label>
            <input
              type="number"
              :value="modelValue"
              @input="
                $emit(
                  'update:modelValue',
                  ($event.target as HTMLInputElement).valueAsNumber,
                )
              "
              :placeholder="
                type === 'price' ? 'Masukkan harga...' : 'Masukkan stok...'
              "
              :min="0"
              :step="type === 'price' ? 1000 : 1"
              class="batch-input"
              ref="inputRef"
            />
          </div>
        </div>
        <div class="modal-footer">
          <button class="btn-cancel" @click="$emit('close')">Batal</button>
          <button
            class="btn-confirm"
            @click="$emit('confirm')"
            :disabled="updating || modelValue === null || modelValue < 0"
          >
            <span v-if="updating" class="spinner-small"></span>
            {{ updating ? "Mengupdate..." : "Update Semua" }}
          </button>
        </div>
      </div>
    </div>
  </Teleport>
</template>

<script setup lang="ts">
import { onMounted, onUnmounted, ref, watch, nextTick } from "vue";

const props = defineProps<{
  show: boolean;
  type: "price" | "stock";
  count: number;
  updating: boolean;
  modelValue: number | null;
}>();

const emit = defineEmits<{
  "update:modelValue": [value: number];
  close: [];
  confirm: [];
}>();

const inputRef = ref<HTMLInputElement | null>(null);
const titleId = `modal-title-${Math.random().toString(36).substr(2, 9)}`;

watch(
  () => props.show,
  (newVal) => {
    if (newVal) {
      document.body.style.overflow = "hidden";
      nextTick(() => {
        inputRef.value?.focus();
      });
    } else {
      document.body.style.overflow = "";
    }
  },
);

const handleKeydown = (e: KeyboardEvent) => {
  if (e.key === "Escape" && props.show) {
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
.modal-overlay {
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background: rgba(0, 0, 0, 0.5);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 1000;
}

.batch-modal {
  background: white;
  border-radius: 0.75rem;
  width: 100%;
  max-width: 400px;
  box-shadow: 0 20px 40px rgba(0, 0, 0, 0.2);
  animation: modal-in 0.2s ease-out;
}

@keyframes modal-in {
  from {
    opacity: 0;
    transform: scale(0.95) translateY(-10px);
  }
  to {
    opacity: 1;
    transform: scale(1) translateY(0);
  }
}

.modal-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 1rem 1.25rem;
  border-bottom: 1px solid #e5e7eb;
}

.modal-header h3 {
  margin: 0;
  font-size: 1.125rem;
  font-weight: 600;
  color: #1f2937;
}

.modal-close {
  background: none;
  border: none;
  font-size: 1.25rem;
  color: #9ca3af;
  cursor: pointer;
  padding: 0.25rem;
  border-radius: 0.25rem;
  transition: all 0.2s;
}

.modal-close:hover {
  background: #f3f4f6;
  color: #1f2937;
}

.modal-body {
  padding: 1.25rem;
}

.modal-info {
  margin: 0 0 1rem 0;
  color: #6b7280;
  font-size: 0.9375rem;
}

.form-group {
  display: flex;
  flex-direction: column;
  gap: 0.5rem;
}

.form-group label {
  font-weight: 500;
  color: #374151;
  font-size: 0.875rem;
}

.batch-input {
  padding: 0.75rem 1rem;
  border: 2px solid #e5e7eb;
  border-radius: 0.5rem;
  font-size: 1rem;
  font-weight: 500;
  transition: all 0.2s;
}

.batch-input:focus {
  outline: none;
  border-color: #3b82f6;
  box-shadow: 0 0 0 3px rgba(59, 130, 246, 0.1);
}

.modal-footer {
  display: flex;
  justify-content: flex-end;
  gap: 0.75rem;
  padding: 1rem 1.25rem;
  border-top: 1px solid #e5e7eb;
  background: #f9fafb;
  border-radius: 0 0 0.75rem 0.75rem;
}

.btn-cancel {
  padding: 0.625rem 1.25rem;
  border: 1px solid #d1d5db;
  background: white;
  color: #374151;
  border-radius: 0.375rem;
  font-weight: 500;
  cursor: pointer;
  transition: all 0.2s;
}

.btn-cancel:hover {
  background: #f3f4f6;
}

.btn-confirm {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  padding: 0.625rem 1.25rem;
  border: none;
  background: #3b82f6;
  color: white;
  border-radius: 0.375rem;
  font-weight: 500;
  cursor: pointer;
  transition: all 0.2s;
}

.btn-confirm:hover:not(:disabled) {
  background: #2563eb;
}

.btn-confirm:disabled {
  background: #9ca3af;
  cursor: not-allowed;
}

.spinner-small {
  width: 16px;
  height: 16px;
  border: 2px solid rgba(255, 255, 255, 0.3);
  border-top-color: white;
  border-radius: 50%;
  animation: spin 1s linear infinite;
}

@keyframes spin {
  to {
    transform: rotate(360deg);
  }
}
</style>
