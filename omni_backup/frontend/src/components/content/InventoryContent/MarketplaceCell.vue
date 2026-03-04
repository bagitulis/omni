<template>
  <td :class="['table-cell', 'marketplace-cell', `${platform}-cell`]">
    <div class="cell-wrapper">
      <div class="allocation-value" :class="statusClasses">
        {{ value }}
      </div>
      <!-- Clone button appears when product EXISTS (green status) -->
      <button
        v-if="isProductFound"
        class="clone-btn"
        :title="`Clone from ${platformName} to other platforms`"
        @click.stop="handleCloneClick"
      >
        <i class="pi pi-copy" aria-hidden="true"></i>
      </button>
    </div>
  </td>
</template>

<script setup lang="ts">
import { computed } from "vue";

interface BatchResult {
  shopee?: boolean;
  tiktok?: boolean;
  lazada?: boolean;
}

type Platform = "shopee" | "tiktok" | "lazada";

const props = defineProps<{
  platform: Platform;
  value: number;
  batchResult: BatchResult | null;
}>();

const emit = defineEmits<{
  clone: [platform: Platform];
}>();

const platformName = computed(() => {
  const names: Record<Platform, string> = {
    shopee: "Shopee",
    tiktok: "TikTok",
    lazada: "Lazada",
  };
  return names[props.platform];
});

const isProductFound = computed(() => {
  return props.batchResult?.[props.platform] === true;
});

const statusClasses = computed(() => {
  if (!props.batchResult) return {};

  const platformStatus = props.batchResult[props.platform];
  return {
    "has-status": true,
    "status-found": platformStatus === true,
    "status-not-found": platformStatus === false,
  };
});

function handleCloneClick() {
  emit("clone", props.platform);
}
</script>

<style scoped>
.table-cell {
  padding: 8px 12px;
  text-align: center;
  font-size: 13px;
  color: #333;
  background-color: #fff;
  vertical-align: middle;
  border-bottom: 1px solid #e0e0e0;
}

.marketplace-cell {
  width: 70px;
  min-width: 70px;
  padding: 8px;
  text-align: center;
}

.cell-wrapper {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 4px;
}

.allocation-value {
  display: inline-block;
  padding: 4px 12px;
  border-radius: 4px;
  font-size: 12px;
  font-weight: 600;
  min-width: 40px;
  text-align: center;
  transition: all 0.2s ease;
}

/* Platform-specific base colors */
.shopee-cell .allocation-value {
  background-color: #fff5f5;
  color: #ee4d2d;
}

.tiktok-cell .allocation-value {
  background-color: #e8e8e8;
  color: #333;
}

.lazada-cell .allocation-value {
  background-color: #e0e0ff;
  color: #0f146d;
}

/* Status-based styles (when batch check has results) */
.allocation-value.has-status.status-found {
  box-shadow: 0 2px 4px rgba(46, 204, 113, 0.3);
  border: 2px solid #2ecc71;
}

.allocation-value.has-status.status-not-found {
  box-shadow: 0 2px 4px rgba(231, 76, 60, 0.3);
  border: 2px solid #e74c3c;
}

.allocation-value:hover {
  transform: scale(1.05);
}

/* Clone button */
.clone-btn {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 22px;
  height: 22px;
  padding: 0;
  border: none;
  border-radius: 4px;
  background: #3b82f6;
  color: white;
  cursor: pointer;
  opacity: 0;
  transition: all 0.2s ease;
}

.clone-btn i {
  font-size: 10px;
}

.cell-wrapper:hover .clone-btn {
  opacity: 1;
}

.clone-btn:hover {
  background: #2563eb;
  transform: scale(1.1);
}

@media (max-width: 768px) {
  .marketplace-cell {
    width: 60px;
    min-width: 60px;
    padding: 6px 4px;
  }

  .allocation-value {
    font-size: 11px;
    min-width: 32px;
    padding: 3px 8px;
  }

  .clone-btn {
    width: 18px;
    height: 18px;
    opacity: 1;
  }

  .clone-btn i {
    font-size: 8px;
  }
}
</style>

