<template>
  <div class="info-section">
    <div class="info-row">
      <span>Order No.</span><span>{{ order?.order_no }}</span>
    </div>
    <div class="info-row">
      <span>Platform</span>
      <span
        :class="[
          'om-platform-badge',
          `om-platform-${order?.platform?.toLowerCase()}`,
        ]"
      >
        {{ fmt(order?.platform) }}
      </span>
    </div>
    <div class="info-row">
      <span>Buyer</span><span>{{ order?.buyer_username }}</span>
    </div>
  </div>
</template>

<script setup lang="ts">
import { platformFormatter } from "../composables/useModalForm";

defineProps<{ order: any }>();

const fmt = (platform?: string): string => platformFormatter.format(platform);
</script>

<style scoped>
@import "../OrderManager.theme.css";

.info-section {
  background: var(--om-bg-secondary);
  border-radius: var(--om-radius-md);
  padding: var(--om-spacing-md);
  margin-bottom: var(--om-spacing-lg);
}

.info-row {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: var(--om-spacing-xs) 0;
  font-size: var(--om-font-sm);
}

.info-row:not(:last-child) {
  border-bottom: 1px solid var(--om-border);
  padding-bottom: var(--om-spacing-sm);
  margin-bottom: var(--om-spacing-sm);
}

.info-row span:first-child {
  color: var(--om-text-secondary);
}

.info-row span:last-child {
  font-weight: 500;
  color: var(--om-text-primary);
}
</style>
