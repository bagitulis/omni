<template>
  <div class="order-actions">
    <!-- Bulk Shipment Button -->
    <button
      v-if="activeTab === 'unprocess'"
      @click="$emit('bulk-ship')"
      :disabled="loading"
      class="btn-action btn-bulk-ship"
      title="Ship multiple orders at once"
    >
      <Icon name="truck" size="sm" />
      <span>Bulk Shipment</span>
    </button>

    <button
      @click="$emit('refresh')"
      :disabled="loading"
      class="btn-action btn-refresh"
      title="Refresh latest data"
    >
      <Icon :name="loading ? 'spinner' : 'refresh'" :spin="loading" size="sm" />
      <span>Refresh</span>
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
import Icon from "@/components/ui/Icon.vue";

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
  gap: var(--om-spacing-md); /* Increased gap */
  align-items: center;
  flex-wrap: wrap;
}

.btn-action {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 0.5rem;
  padding: 0.55rem 0.95rem;
  border-radius: var(--om-radius-md);
  font-size: var(--om-font-xs);
  font-weight: 700;
  letter-spacing: 0.02em;
  text-transform: uppercase;
  min-height: 36px; /* Use min-height */
  cursor: pointer;
  transition: all var(--om-transition-fast);
  border: 1px solid transparent;
}

.btn-action:disabled {
  opacity: 0.6;
  cursor: not-allowed;
}

.btn-bulk-ship {
  background: #ee4d2d !important;
  color: #ffffff !important;
  border: 1px solid #ee4d2d !important;
  box-shadow: 0 2px 4px rgba(238, 77, 45, 0.2);
}

.btn-bulk-ship:hover:not(:disabled) {
  background: #d73211 !important;
  border-color: #d73211 !important;
  box-shadow: 0 4px 8px rgba(238, 77, 45, 0.3);
}

.btn-refresh {
  background: #ffffff !important;
  color: #212121 !important;
  border: 1px solid #e0e0e0 !important;
  box-shadow: 0 1px 3px rgba(0, 0, 0, 0.08);
}

.btn-refresh:hover:not(:disabled) {
  border-color: #ee4d2d !important;
  color: #ee4d2d !important;
  background: #fff7f4 !important;
  box-shadow: 0 2px 6px rgba(238, 77, 45, 0.15);
}
</style>
