<template>
  <div class="warning-box">
    <Icon name="warning" size="lg" />
    <span>
      This action cannot be undone. Please confirm you want to cancel this
      order.
    </span>
  </div>
</template>

<script setup lang="ts">
import Icon from "@/components/ui/Icon.vue";
</script>

<style scoped>
@import "./OrderCancelModal.styles.css";
</style>
