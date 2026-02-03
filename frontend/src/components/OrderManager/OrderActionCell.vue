<template>
  <div class="action-col">
    <div class="action-buttons">
      <button
        v-if="showShipButton"
        @click="$emit('ship-order')"
        class="btn-action primary"
        title="Arrange Shipment"
      >
        Arrange Shipment
      </button>
      <button
        v-else-if="showResponseButton"
        @click="$emit('ship-order')"
        class="btn-action primary"
        title="Respond"
      >
        Respond
      </button>
      <button
        v-if="showCancelButton"
        @click="$emit('cancel-order')"
        class="btn-action secondary"
        title="Cancel Order"
      >
        <Icon name="close" size="sm" />
      </button>
      <button
        @click="$emit('view-detail')"
        class="btn-action secondary"
        title="View Details"
      >
        <Icon name="eye" size="sm" />
      </button>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed } from "vue";
import Icon from "@/components/ui/Icon.vue";

interface Props {
  activeTab: string;
  orderStatus: string;
}

const props = defineProps<Props>();

defineEmits<{
  "ship-order": [];
  "cancel-order": [];
  "view-detail": [];
}>();

const showShipButton = computed(
  () => props.activeTab === "unprocess" && props.orderStatus !== "CANCELLED",
);

const showResponseButton = computed(
  () => props.activeTab === "unpaid" && props.orderStatus !== "CANCELLED",
);

const showCancelButton = computed(
  () =>
    ["unprocess", "unpaid"].includes(props.activeTab) &&
    props.orderStatus !== "CANCELLED",
);
</script>

<style scoped>
@import "./OrderManager.theme.css";

/* Action Column */
.action-col {
  display: flex;
  align-items: center;
  justify-content: flex-start;
  padding: 8px 0;
}

.action-buttons {
  display: flex;
  gap: 8px;
  flex-wrap: nowrap;
  align-items: center;
}

.btn-action {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  gap: var(--om-spacing-xs);
  padding: 6px 12px;
  font-size: var(--om-font-xs);
  font-weight: 500;
  border-radius: var(--om-radius-sm);
  cursor: pointer;
  transition: all var(--om-transition-fast);
  border: 1px solid transparent;
  white-space: nowrap;
  flex: 1;
  min-width: 60px;
}

.btn-action.primary {
  background: #ee4d2d !important;
  color: #ffffff !important;
  border: 1px solid #ee4d2d !important;
  box-shadow: 0 2px 4px rgba(238, 77, 45, 0.25);
  font-weight: 600;
  padding: 8px 16px;
  min-width: auto;
  flex: none;
}

.btn-action.primary:hover:not(:disabled) {
  background: #d73211 !important;
  color: #ffffff !important;
  border-color: #d73211 !important;
  box-shadow: 0 4px 8px rgba(238, 77, 45, 0.35);
  transform: translateY(-1px);
}

.btn-action.secondary {
  background: #ffffff !important;
  color: #757575 !important;
  border: 1px solid #e0e0e0 !important;
  box-shadow: 0 1px 2px rgba(0, 0, 0, 0.05);
  padding: 8px 10px;
  min-width: 36px;
  flex: none;
}

.btn-action.secondary:hover:not(:disabled) {
  background: #f5f5f5 !important;
  color: #212121 !important;
  border-color: #bdbdbd !important;
  transform: translateY(-1px);
}

.btn-action:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

.btn-action i {
  font-size: 1rem;
}

@media (max-width: 1200px) {
  .btn-action {
    padding: 6px 8px;
    font-size: var(--om-font-xs);
  }
}

@media (max-width: 768px) {
  .action-col {
    border: 1px solid var(--om-border);
    border-radius: var(--om-radius-sm);
    padding: var(--om-spacing-sm);
    background: var(--om-bg-secondary);
  }

  .action-buttons {
    flex-direction: column;
  }

  .btn-action {
    width: 100%;
    justify-content: flex-start;
  }
}
</style>
