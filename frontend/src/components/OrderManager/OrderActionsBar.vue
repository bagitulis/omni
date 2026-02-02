<template>
  <div class="order-actions">
    <!-- Bulk Shipment Button -->
    <button
      v-if="activeTab === 'unprocess'"
      @click="$emit('bulk-ship')"
      :disabled="loading"
      class="btn-bulk-ship"
      title="Ship multiple orders at once"
    >
      <i class="pi pi-truck"></i>
      <span class="btn-text">Bulk Shipment</span>
    </button>

    <button
      @click="$emit('refresh')"
      :disabled="loading"
      class="btn-refresh"
      title="Refresh latest data"
    >
      <i :class="loading ? 'pi pi-spin pi-spinner' : 'pi pi-refresh'"></i>
      <span class="btn-text">Refresh</span>
    </button>

    <OrderExportMenu
      :loading="loading"
      :disabled="!hasData"
      @export-n8n="$emit('export-n8n')"
      @export-csv="$emit('export-csv')"
    />
  </div>
</template>

<script setup lang="ts">
import OrderExportMenu from "./OrderExportMenu.vue";

defineProps<{
  activeTab: string;
  loading: boolean;
  hasData: boolean;
}>();

defineEmits<{
  (e: "bulk-ship"): void;
  (e: "refresh"): void;
  (e: "export-n8n"): void;
  (e: "export-csv"): void;
}>();
</script>

<style scoped>
@import "./OrderManager.styles.css";

.order-actions {
  display: flex;
  gap: 0.75rem;
  align-items: center;
}

.btn-bulk-ship {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  padding: 0.5rem 1rem;
  background: #ee4d2d;
  color: white;
  border: none;
  border-radius: var(--om-radius-sm);
  font-size: var(--om-font-sm);
  font-weight: 500;
  cursor: pointer;
  transition: background-color var(--om-transition-fast);
}

.btn-bulk-ship:hover:not(:disabled) {
  background: #d73211;
}

.btn-bulk-ship:disabled {
  opacity: 0.6;
  cursor: not-allowed;
}
</style>
