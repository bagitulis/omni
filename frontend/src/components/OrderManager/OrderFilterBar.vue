<template>
  <div class="filter-bar">
    <!-- Search Input -->
    <div class="filter-search" role="search">
      <Icon name="search" size="sm" class="search-icon" />
      <input
        type="text"
        :value="searchQuery"
        @input="onSearchInput"
        placeholder="Search by Order No., SKU, or Product..."
        class="search-input"
        aria-label="Search orders by order number, SKU, or product name"
      />
      <button
        v-if="searchQuery"
        @click="clearSearch"
        class="clear-btn"
        title="Clear search"
        aria-label="Clear search"
      >
        <Icon name="close" size="sm" />
      </button>
    </div>

    <div class="filter-controls">
      <!-- Platform Filter -->
      <div class="filter-platform">
        <select
          :value="selectedPlatform"
          @change="onPlatformChange"
          class="platform-select"
          aria-label="Filter by platform"
        >
          <option value="">All Platforms</option>
          <option
            v-for="platform in platforms"
            :key="platform"
            :value="platform"
          >
            {{ formatPlatformName(platform) }}
          </option>
        </select>
      </div>

      <!-- Shipping Filter -->
      <div class="filter-shipping">
        <select
          :value="selectedShipping"
          @change="onShippingChange"
          class="shipping-select"
          aria-label="Filter by shipping provider"
        >
          <option value="">All Shipping</option>
          <option
            v-for="provider in shippingProviders"
            :key="provider"
            :value="provider"
          >
            {{ provider }}
          </option>
        </select>
      </div>

      <!-- Reset Button Only - filters apply automatically -->
      <button
        @click="resetFilters"
        class="reset-btn"
        title="Reset all filters"
        aria-label="Reset all filters"
      >
        Reset
      </button>
    </div>
  </div>
</template>

<script setup lang="ts">
import Icon from "@/components/ui/Icon.vue";

interface Props {
  searchQuery: string;
  selectedPlatform: string;
  platforms: string[];
  selectedShipping?: string;
  shippingProviders?: string[];
}

const props = defineProps<Props>();

const emit = defineEmits<{
  "update:searchQuery": [value: string];
  "update:selectedPlatform": [value: string];
  "update:selectedShipping": [value: string];
  "apply-filters": [];
  "reset-filters": [];
}>();

// Debounce timer
let debounceTimer: ReturnType<typeof setTimeout>;

const onSearchInput = (event: Event) => {
  const value = (event.target as HTMLInputElement).value;

  // Debounce search input
  clearTimeout(debounceTimer);
  debounceTimer = setTimeout(() => {
    emit("update:searchQuery", value);
  }, 300);
};

const clearSearch = () => {
  emit("update:searchQuery", "");
};

const onPlatformChange = (event: Event) => {
  const value = (event.target as HTMLSelectElement).value;
  emit("update:selectedPlatform", value);
};

const onShippingChange = (event: Event) => {
  const value = (event.target as HTMLSelectElement).value;
  emit("update:selectedShipping", value);
};

const formatPlatformName = (platform: string): string => {
  const names: Record<string, string> = {
    shopee: "Shopee",
    lazada: "Lazada",
    tiktok: "TikTok Shop",
  };
  return names[platform.toLowerCase()] || platform;
};

const applyFilters = () => {
  emit("apply-filters");
};

const resetFilters = () => {
  emit("update:searchQuery", "");
  emit("update:selectedPlatform", "");
  emit("update:selectedShipping", "");
  emit("reset-filters");
};
</script>

<style scoped>
@import "./OrderManager.theme.css";

.filter-bar {
  display: flex;
  align-items: center;
  gap: var(--om-spacing-sm);
  padding: var(--om-spacing-sm);
  background: transparent;
  border: 1px solid transparent;
  border-radius: var(--om-radius-md);
  flex: 1;
  min-width: 0;
}

.filter-search {
  flex: 1 1 320px;
  min-width: 240px;
  position: relative;
}

.search-icon {
  position: absolute;
  left: var(--om-spacing-sm);
  top: 50%;
  transform: translateY(-50%);
  color: var(--om-text-secondary);
  font-size: var(--om-font-sm);
}

.search-input {
  width: 100%;
  padding: 0.65rem var(--om-spacing-md);
  padding-left: 2.25rem;
  padding-right: 2rem;
  border: 1px solid var(--om-border);
  border-radius: var(--om-radius-md);
  font-size: var(--om-font-sm);
  background: var(--om-bg-primary);
  transition: border-color var(--om-transition-fast);
}

.search-input:focus {
  outline: none;
  border-color: var(--om-primary);
}

.search-input::placeholder {
  color: var(--om-text-disabled);
}

.clear-btn {
  position: absolute;
  right: var(--om-spacing-sm);
  top: 50%;
  transform: translateY(-50%);
  background: none;
  border: none;
  color: var(--om-text-secondary);
  cursor: pointer;
  padding: 10px;
  display: flex;
  align-items: center;
  justify-content: center;
  min-width: 44px;
  min-height: 44px;
}

.clear-btn:hover {
  color: var(--om-text-primary);
}

.filter-controls {
  display: flex;
  align-items: center;
  gap: var(--om-spacing-sm);
  flex-wrap: wrap;
}

.filter-platform,
.filter-shipping {
  min-width: 160px;
}

.platform-select,
.shipping-select {
  width: 100%;
  padding: 0.6rem 0.9rem;
  border: 1px solid var(--om-border);
  border-radius: var(--om-radius-md);
  font-size: var(--om-font-sm);
  background: var(--om-bg-primary);
  cursor: pointer;
  transition:
    border-color var(--om-transition-fast),
    box-shadow var(--om-transition-fast);
}

.platform-select:focus,
.shipping-select:focus {
  outline: none;
  border-color: var(--om-primary);
  box-shadow: 0 0 0 2px var(--om-primary-light);
}

.reset-btn {
  padding: 0.55rem 0.9rem;
  border-radius: var(--om-radius-md);
  border: 1px solid var(--om-border);
  background: var(--om-bg-secondary);
  font-size: var(--om-font-xs);
  font-weight: 600;
  color: var(--om-text-secondary);
  cursor: pointer;
  transition: all var(--om-transition-fast);
}

.reset-btn:hover {
  border-color: var(--om-primary);
  color: var(--om-primary);
  background: var(--om-bg-primary);
}

/* Responsive */
@media (max-width: 768px) {
  .filter-bar {
    flex-direction: column;
    align-items: stretch;
  }

  .filter-search {
    min-width: 100%;
  }

  .filter-platform {
    width: 100%;
  }

  .filter-controls {
    width: 100%;
    flex-direction: column;
    align-items: stretch;
  }
}
</style>
