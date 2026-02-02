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
import { platformFormatter } from "./composables/useModalForm";
import type { Order } from "./composables/useOrderManager";

defineProps<{ order: Order | null }>();

const fmt = (platform?: string): string => platformFormatter.format(platform);
</script>

<style scoped>
@import "./OrderCancelModal.styles.css";
</style>
