<template>
  <div class="filter-bar">
    <!-- Search Input -->
    <div class="filter-search">
      <i class="pi pi-search search-icon"></i>
      <input
        type="text"
        :value="searchQuery"
        @input="onSearchInput"
        placeholder="Search by Order No., SKU, or Product..."
        class="search-input"
      />
      <button
        v-if="searchQuery"
        @click="clearSearch"
        class="clear-btn"
        title="Clear search"
      >
        <i class="pi pi-times"></i>
      </button>
    </div>

    <!-- Platform Filter -->
    <div class="filter-platform">
      <select
        :value="selectedPlatform"
        @change="onPlatformChange"
        class="platform-select"
      >
        <option value="">All Platforms</option>
        <option v-for="platform in platforms" :key="platform" :value="platform">
          {{ formatPlatformName(platform) }}
        </option>
      </select>
    </div>

    <!-- Filter Actions -->
    <div class="filter-actions">
      <button @click="applyFilters" class="om-btn om-btn-primary">
        <i class="pi pi-filter"></i>
        <span>Apply</span>
      </button>
      <button @click="resetFilters" class="om-btn om-btn-secondary">
        <i class="pi pi-refresh"></i>
        <span>Reset</span>
      </button>
    </div>
  </div>
</template>

<script setup lang="ts">
interface Props {
  searchQuery: string;
  selectedPlatform: string;
  platforms: string[];
}

const props = defineProps<Props>();

const emit = defineEmits<{
  "update:searchQuery": [value: string];
  "update:selectedPlatform": [value: string];
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
  emit("reset-filters");
};
</script>

<style scoped>
@import "./OrderManager.theme.css";

.filter-bar {
  display: flex;
  align-items: center;
  gap: var(--om-spacing-md);
  padding: var(--om-spacing-md);
  background: var(--om-bg-primary);
  border: 1px solid var(--om-border);
  border-radius: var(--om-radius-md);
  flex-wrap: wrap;
}

.filter-search {
  flex: 1;
  min-width: 250px;
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
  padding: var(--om-spacing-sm) var(--om-spacing-md);
  padding-left: 2.25rem;
  padding-right: 2rem;
  border: 1px solid var(--om-border);
  border-radius: var(--om-radius-sm);
  font-size: var(--om-font-sm);
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
  padding: 4px;
  display: flex;
  align-items: center;
  justify-content: center;
}

.clear-btn:hover {
  color: var(--om-text-primary);
}

.filter-platform {
  min-width: 150px;
}

.platform-select {
  width: 100%;
  padding: var(--om-spacing-sm) var(--om-spacing-md);
  border: 1px solid var(--om-border);
  border-radius: var(--om-radius-sm);
  font-size: var(--om-font-sm);
  background: var(--om-bg-primary);
  cursor: pointer;
}

.platform-select:focus {
  outline: none;
  border-color: var(--om-primary);
}

.filter-actions {
  display: flex;
  gap: var(--om-spacing-sm);
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

  .filter-actions {
    justify-content: stretch;
  }

  .filter-actions .om-btn {
    flex: 1;
  }
}
</style>
