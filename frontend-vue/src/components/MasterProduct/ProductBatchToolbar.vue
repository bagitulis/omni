<template>
  <transition name="slide-down">
    <div v-if="selectedCount > 0" class="batch-toolbar">
      <div class="batch-info">
        <span class="batch-count">{{ selectedCount }} varian dipilih</span>
        <button class="btn-clear-selection" @click="$emit('clear')">
          ✕ Batal
        </button>
      </div>
      <div class="batch-actions">
        <button
          class="btn-batch btn-batch-price"
          @click="$emit('open-price-modal')"
        >
          💰 Update Harga
        </button>
        <button
          class="btn-batch btn-batch-stock"
          @click="$emit('open-stock-modal')"
        >
          📦 Update Stok
        </button>
      </div>
    </div>
  </transition>
</template>

<script setup lang="ts">
defineProps<{
  selectedCount: number;
}>();

defineEmits<{
  clear: [];
  "open-price-modal": [];
  "open-stock-modal": [];
}>();
</script>

<style scoped>
.batch-toolbar {
  position: sticky;
  top: 0;
  z-index: 100;
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 0.75rem 1rem;
  background: linear-gradient(135deg, #3b82f6 0%, #2563eb 100%);
  color: white;
  border-radius: 0.5rem;
  box-shadow: 0 4px 12px rgba(59, 130, 246, 0.3);
}

.batch-info {
  display: flex;
  align-items: center;
  gap: 1rem;
}

.batch-count {
  font-weight: 600;
  font-size: 0.9375rem;
}

.btn-clear-selection {
  background: rgba(255, 255, 255, 0.2);
  border: none;
  color: white;
  padding: 0.375rem 0.75rem;
  border-radius: 0.25rem;
  cursor: pointer;
  font-size: 0.8125rem;
  transition: background 0.2s;
}

.btn-clear-selection:hover {
  background: rgba(255, 255, 255, 0.3);
}

.batch-actions {
  display: flex;
  gap: 0.5rem;
}

.btn-batch {
  display: flex;
  align-items: center;
  gap: 0.375rem;
  padding: 0.5rem 1rem;
  border: none;
  border-radius: 0.375rem;
  font-weight: 500;
  font-size: 0.875rem;
  cursor: pointer;
  transition: all 0.2s;
}

.btn-batch-price {
  background: #10b981;
  color: white;
}

.btn-batch-price:hover {
  background: #059669;
}

.btn-batch-stock {
  background: #f59e0b;
  color: white;
}

.btn-batch-stock:hover {
  background: #d97706;
}

.slide-down-enter-active,
.slide-down-leave-active {
  transition: all 0.3s ease;
}

.slide-down-enter-from,
.slide-down-leave-to {
  opacity: 0;
  transform: translateY(-10px);
}
</style>
