<template>
  <div class="order-header">
    <!-- Title -->
    <div class="header-top">
      <h1 class="page-title">
        <i class="pi pi-shopping-bag" aria-hidden="true"></i> Order Manager
      </h1>
      <p class="page-subtitle">
        Manage and monitor all orders from various platforms
      </p>
    </div>

    <!-- Platform Stats Cards -->
    <div class="platform-stats">
      <div
        v-for="platform in ['SHOPEE', 'LAZADA', 'TIKTOK']"
        :key="platform"
        :class="['stat-card', platform.toLowerCase()]"
      >
        <div class="stat-icon">
          <PlatformBadge :platform="platform.toLowerCase()" size="md" />
        </div>
        <div class="stat-content">
          <p class="stat-platform">{{ platform }}</p>
          <p class="stat-number">{{ getPlatformCount(platform) }}</p>
          <p class="stat-label">
            {{ activeTab === "locked" ? "Produk" : "Order" }}
          </p>
        </div>
      </div>
    </div>

    <!-- Filter Section -->
    <div class="filter-section">
      <div class="filter-group">
        <div class="search-wrapper">
          <i class="pi pi-search" aria-hidden="true"></i>
          <label for="order-search" class="visually-hidden"
            >Search orders</label
          >
          <input
            id="order-search"
            :value="searchQuery"
            @input="
              $emit('search-changed', ($event.target as HTMLInputElement).value)
            "
            type="text"
            placeholder="Search: Order No, SKU, Product Name..."
            class="search-input"
          />
        </div>
        <div class="platform-filter">
          <label for="platform-filter">
            <i class="pi pi-filter" aria-hidden="true"></i>
            Platform:
          </label>
          <select
            id="platform-filter"
            :value="selectedPlatform"
            @change="
              $emit(
                'platform-changed',
                ($event.target as HTMLSelectElement).value
              )
            "
            class="filter-select"
          >
            <option value="">All Platforms</option>
            <option value="SHOPEE">🛍️ Shopee</option>
            <option value="LAZADA">📦 Lazada</option>
            <option value="TIKTOK">🎵 TikTok</option>
          </select>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import PlatformBadge from "@/components/PlatformBadge.vue";

interface Order {
  platform: string;
  [key: string]: any;
}

const props = defineProps<{
  activeTab: string;
  searchQuery: string;
  selectedPlatform: string;
  orders: Order[];
}>();

defineEmits<{
  "search-changed": [query: string];
  "platform-changed": [platform: string];
}>();

const getPlatformCount = (platform: string): number => {
  if (!props.orders || props.orders.length === 0) return 0;

  const platformLower = platform.toLowerCase();

  const platformItems = props.orders.filter((o) => {
    if (props.activeTab === "locked") {
      // For locked: check if platform is in the platforms array (lowercase)
      return (
        o.platforms &&
        o.platforms.some((p: string) => p.toLowerCase() === platformLower)
      );
    } else {
      // For other tabs: check platform field directly (compare lowercase)
      return o.platform && o.platform.toLowerCase() === platformLower;
    }
  });

  if (platformItems.length === 0) return 0;

  // Different tabs have different field names for counting
  if (props.activeTab === "locked") {
    // For locked: count rows (each row is a unique product with multiple platforms)
    return platformItems.length;
  } else if (props.activeTab === "today") {
    // For today: count unique orders by orderSn
    const uniqueOrderSns = new Set(platformItems.map((o) => o.orderSn));
    return uniqueOrderSns.size;
  } else {
    // For other tabs (unpaid, unprocess, processed): count unique orders by order_no
    const uniqueOrderNos = new Set(platformItems.map((o) => o.order_no));
    return uniqueOrderNos.size;
  }
};
</script>

<style scoped>
@import "./OrderManager.styles.css";
</style>
