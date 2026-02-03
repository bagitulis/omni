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
  align-items: flex-start;
  /* Ensure it takes full width/height of its grid area if needed, 
     but flex-start aligns it top-left usually */
}

.action-buttons {
  display: flex;
  gap: var(--om-spacing-xs);
  flex-wrap: wrap;
  width: 100%;
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
  background: transparent;
  color: #ee4d2d;
  border: 1px solid #ee4d2d;
}

.btn-action.primary:hover:not(:disabled) {
  background: #fff0ed;
  color: #d73211;
  border-color: #d73211;
}

.btn-action.secondary {
  background: transparent;
  color: var(--om-text-secondary);
  border: 1px solid var(--om-border);
}

.btn-action.secondary:hover:not(:disabled) {
  background: var(--om-bg-secondary);
  color: var(--om-text-primary);
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
