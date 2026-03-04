<template>
  <div class="platform-status-container">
    <!-- Loading state -->
    <div v-if="loading" class="status-loading">
      <span class="spinner"></span>
    </div>

    <!-- Status bullet -->
    <div v-else 
      class="status-bullet"
      :class="{ 'found': platformStatus, 'not-found': !platformStatus }"
      :title="`${platformLabel}: ${platformStatus ? 'Found' : 'Not Found'}`"
    >
      {{ platformCode }}
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue';
import type { Ref } from 'vue';
import { usePlatformSkuCheck } from '@/composables/usePlatformSkuCheck';

interface Props {
  sku: string;
  platform: 'lazada' | 'shopee' | 'tiktok';
}

const props = defineProps<Props>();

// Use composable untuk check SKU
const skuCheckResult: Ref<any> = usePlatformSkuCheck(props.sku);

// Get status untuk platform tertentu
const platformStatus = computed(() => {
  if (!skuCheckResult.value) return false;
  
  switch (props.platform) {
    case 'lazada':
      return skuCheckResult.value.lazada ?? false;
    case 'shopee':
      return skuCheckResult.value.shopee ?? false;
    case 'tiktok':
      return skuCheckResult.value.tiktok ?? false;
    default:
      return false;
  }
});

const platformCode = computed(() => {
  const codes: Record<string, string> = {
    lazada: 'L',
    shopee: 'S',
    tiktok: 'T',
  };
  return codes[props.platform] || '?';
});

const platformLabel = computed(() => {
  const labels: Record<string, string> = {
    lazada: 'Lazada',
    shopee: 'Shopee',
    tiktok: 'TikTok',
  };
  return labels[props.platform] || 'Unknown';
});

const loading = computed(() => skuCheckResult.value?.loading ?? true);
</script>

<style scoped>
.platform-status-container {
  display: flex;
  align-items: center;
  justify-content: center;
}

.status-loading {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 30px;
  height: 30px;
}

.spinner {
  display: inline-block;
  width: 12px;
  height: 12px;
  border: 2px solid #e0e0e0;
  border-top-color: #3498db;
  border-radius: 50%;
  animation: spin 0.8s linear infinite;
}

@keyframes spin {
  to {
    transform: rotate(360deg);
  }
}

.status-bullet {
  width: 24px;
  height: 24px;
  border-radius: 50%;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 12px;
  font-weight: 700;
  cursor: pointer;
  transition: all 0.2s ease;
}

.status-bullet.found {
  background-color: #2ecc71;
  color: white;
  box-shadow: 0 2px 4px rgba(46, 204, 113, 0.3);
}

.status-bullet.not-found {
  background-color: #e74c3c;
  color: white;
  box-shadow: 0 2px 4px rgba(231, 76, 60, 0.3);
}

.status-bullet:hover {
  transform: scale(1.15);
}

@media (max-width: 768px) {
  .status-bullet {
    width: 22px;
    height: 22px;
    font-size: 11px;
  }
}
</style>
