<template>
  <div class="tiktok-ads-virtual-table">
    <!-- Header Toolbar -->
    <div class="table-toolbar">
      <div class="left">
        <h3>TikTok Ads Data ({{ total.toLocaleString() }} records)</h3>
      </div>
      <div class="right">
        <select v-model="selectedPeriod" class="period-select">
          <option value="">All Periods</option>
          <option v-for="period in periods" :key="period" :value="period">
            {{ period }}
          </option>
        </select>
      </div>
    </div>

    <!-- Loading State -->
    <div v-if="isLoading" class="loading-state">
      <div class="spinner"></div>
      <p>Loading data...</p>
    </div>

    <!-- Error State -->
    <div v-else-if="isError" class="error-state">
      <p>Error loading data: {{ error?.message }}</p>
    </div>

    <!-- Virtual Scrolling Table -->
    <div v-else class="table-container" ref="containerRef">
      <div
        class="table-wrapper"
        :style="{ height: `${virtualizer.getTotalSize()}px` }"
      >
        <div
          v-for="virtualRow in virtualizer.getVirtualItems()"
          :key="virtualRow.index"
          class="table-row"
          :style="{
            position: 'absolute',
            top: 0,
            left: 0,
            width: '100%',
            height: `${virtualRow.size}px`,
            transform: `translateY(${virtualRow.start}px)`,
          }"
        >
          <TiktokAdsRow :data="items[virtualRow.index]" />
        </div>
      </div>

      <!-- Load More Trigger -->
      <div v-if="hasMore" class="load-more-trigger" @click="loadMore">
        <button :disabled="isFetchingMore">
          {{ isFetchingMore ? "Loading..." : "Load More" }}
        </button>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, watchEffect } from "vue";
import { useVirtualizer } from "@tanstack/vue-virtual";
import { useVirtualScroll } from "@/composables/useVirtualScroll";
import { useTiktokAdsAnalytics } from "@/composables/useTiktokAdsAnalytics";
import TiktokAdsRow from "./TiktokAdsRow.vue";

interface Props {
  periods?: string[];
}

const props = withDefaults(defineProps<Props>(), {
  periods: () => [],
});

const selectedPeriod = ref("");
const containerRef = ref<HTMLElement | null>(null);
const { getDataWithCursor } = useTiktokAdsAnalytics();

// Virtual scroll setup
const {
  items,
  total,
  isLoading,
  isError,
  error,
  isFetchingMore,
  hasMore,
  loadMore,
} = useVirtualScroll({
  queryKey: ["tiktok-ads-data", selectedPeriod],
  queryFn: (cursor) => getDataWithCursor(selectedPeriod.value, cursor, 100),
  limit: 100,
});

// Virtualizer for efficient rendering
const virtualizer = useVirtualizer({
  count: computed(() => items.value.length),
  getScrollElement: () => containerRef.value,
  estimateSize: () => 48, // Row height in pixels
  overscan: 10, // Render 10 extra items
});

// Auto-load more when scrolled to bottom
watchEffect(() => {
  const lastItem = virtualizer.getVirtualItems().at(-1);
  if (!lastItem) return;

  if (
    lastItem.index >= items.value.length - 1 &&
    hasMore.value &&
    !isFetchingMore.value
  ) {
    loadMore();
  }
});
</script>

<script setup lang="ts">
import { ref, computed, watchEffect } from "vue";
import { useVirtualizer } from "@tanstack/vue-virtual";
import { useVirtualScroll } from "@/composables/useVirtualScroll";
import { useTiktokAdsAnalytics } from "@/composables/useTiktokAdsAnalytics";
import TiktokAdsRow from "./TiktokAdsRow.vue";

interface Props {
  periods?: string[];
}

const props = withDefaults(defineProps<Props>(), {
  periods: () => [],
});

const selectedPeriod = ref("");
const containerRef = ref<HTMLElement | null>(null);
const { getDataWithCursor } = useTiktokAdsAnalytics();

// Virtual scroll setup
const {
  items,
  total,
  isLoading,
  isError,
  error,
  isFetchingMore,
  hasMore,
  loadMore,
} = useVirtualScroll({
  queryKey: ["tiktok-ads-data", selectedPeriod],
  queryFn: (cursor) => getDataWithCursor(selectedPeriod.value, cursor, 100),
  limit: 100,
});

// Virtualizer for efficient rendering
const virtualizer = useVirtualizer({
  count: computed(() => items.value.length),
  getScrollElement: () => containerRef.value,
  estimateSize: () => 48, // Row height in pixels
  overscan: 10, // Render 10 extra items
});

// Auto-load more when scrolled to bottom
watchEffect(() => {
  const lastItem = virtualizer.getVirtualItems().at(-1);
  if (!lastItem) return;

  if (
    lastItem.index >= items.value.length - 1 &&
    hasMore.value &&
    !isFetchingMore.value
  ) {
    loadMore();
  }
});
</script>

<style scoped>
.tiktok-ads-virtual-table {
  display: flex;
  flex-direction: column;
  height: 100%;
  border: 1px solid #e0e0e0;
  border-radius: 8px;
  overflow: hidden;
}

.table-toolbar {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 16px;
  background: #f5f5f5;
  border-bottom: 1px solid #e0e0e0;
}

.table-toolbar h3 {
  margin: 0;
  font-size: 16px;
  font-weight: 600;
}

.period-select {
  padding: 8px 12px;
  border: 1px solid #ccc;
  border-radius: 4px;
  font-size: 14px;
}

.table-container {
  flex: 1;
  overflow-y: auto;
  position: relative;
}

.table-wrapper {
  position: relative;
  width: 100%;
}

.table-row {
  border-bottom: 1px solid #f0f0f0;
}

.loading-state,
.error-state {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  padding: 48px;
  text-align: center;
}

.spinner {
  width: 40px;
  height: 40px;
  border: 4px solid #f3f3f3;
  border-top: 4px solid #3498db;
  border-radius: 50%;
  animation: spin 1s linear infinite;
}

@keyframes spin {
  0% {
    transform: rotate(0deg);
  }
  100% {
    transform: rotate(360deg);
  }
}

.load-more-trigger {
  text-align: center;
  padding: 16px;
}

.load-more-trigger button {
  padding: 8px 24px;
  background: #3498db;
  color: white;
  border: none;
  border-radius: 4px;
  cursor: pointer;
  font-size: 14px;
}

.load-more-trigger button:disabled {
  opacity: 0.6;
  cursor: not-allowed;
}
</style>
