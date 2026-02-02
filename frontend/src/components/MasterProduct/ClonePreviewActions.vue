<template>
  <div class="modal-footer">
    <button @click="$emit('cancel')" class="btn btn-secondary">Batal</button>
    <button
      v-if="hasConflict"
      @click="$emit('clone', true)"
      class="btn btn-warning"
      :disabled="cloning"
    >
      <span v-if="cloning" class="spinner-sm"></span>
      Update Existing
    </button>
    <button
      @click="$emit('clone', false)"
      class="btn btn-primary"
      :disabled="cloning"
    >
      <span v-if="cloning" class="spinner-sm"></span>
      {{ hasConflict ? "Clone Anyway" : "Clone ke " + targetPlatformLabel }}
    </button>
  </div>
</template>

<script setup lang="ts">
defineProps<{
  hasConflict: boolean;
  cloning: boolean;
  targetPlatformLabel: string;
}>();

defineEmits<{
  (e: "cancel"): void;
  (e: "clone", updateExisting: boolean): void;
}>();
</script>

<style scoped>
.modal-footer {
  display: flex;
  justify-content: flex-end;
  gap: 0.75rem;
  padding: 1.25rem 1.5rem;
  background: #f9fafb;
  border-top: 1px solid #e5e7eb;
  border-radius: 0 0 1rem 1rem;
}

.btn {
  padding: 0.75rem 1.5rem;
  border: none;
  border-radius: 0.5rem;
  font-weight: 600;
  cursor: pointer;
  transition: all 0.2s;
  display: flex;
  align-items: center;
  justify-content: center;
}

.btn:disabled {
  opacity: 0.6;
  cursor: not-allowed;
}

.btn-secondary {
  background: #e5e7eb;
  color: #374151;
}

.btn-secondary:hover:not(:disabled) {
  background: #d1d5db;
}

.btn-warning {
  background: #f59e0b;
  color: white;
}

.btn-warning:hover:not(:disabled) {
  background: #d97706;
}

.btn-primary {
  background: linear-gradient(135deg, #ff6b2c 0%, #ff5511 100%);
  color: white;
}

.btn-primary:hover:not(:disabled) {
  transform: translateY(-1px);
  box-shadow: 0 4px 12px rgba(255, 107, 44, 0.3);
}

.spinner-sm {
  width: 18px;
  height: 18px;
  border: 2px solid rgba(255, 255, 255, 0.3);
  border-top-color: white;
  border-radius: 50%;
  animation: spin 0.6s linear infinite;
  display: inline-block;
  margin-right: 0.5rem;
}

@keyframes spin {
  to {
    transform: rotate(360deg);
  }
}
</style>
