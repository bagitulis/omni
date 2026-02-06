<template>
  <div class="product-manager">
    <!-- Platform Content (hanya tampilkan konten sesuai platform) -->
    <div class="platform-content">
      <!-- Shopee Tab -->
      <ShopeeProductManager v-if="platform === 'shopee'" />

      <!-- Lazada Tab -->
      <LazadaProductManager v-if="platform === 'lazada'" />

      <!-- TikTok Tab -->
      <TiktokProductManager v-if="platform === 'tiktok'" />
    </div>
  </div>
</template>

<script setup lang="ts">
import { defineAsyncComponent } from "vue";

// Lazy load platform managers to reduce initial bundle
const ShopeeProductManager = defineAsyncComponent(
  () =>
    import(
      /* webpackChunkName: "shopee-chunk" */ "./ShopeeProductManager/ShopeeProductManager.vue"
    )
);
const LazadaProductManager = defineAsyncComponent(
  () =>
    import(/* webpackChunkName: "lazada-chunk" */ "./LazadaProductManager/LazadaProductManager.vue")
);
const TiktokProductManager = defineAsyncComponent(
  () =>
    import(/* webpackChunkName: "tiktok-chunk" */ "./TiktokProductManager/TiktokProductManager.vue")
);

defineProps({
  platform: {
    type: String,
    default: "shopee",
  },
});
</script>

<style scoped>
.product-manager {
  min-height: 100vh;
  background: #f8f9fa;
}

.platform-content {
  max-width: 1400px;
  margin: 0 auto;
}
</style>
