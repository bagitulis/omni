<template>
  <div class="header-controls">
    <!-- Row 1: Search & Filter -->
    <div class="controls-row">
      <div class="search-and-filter">
        <div class="search-box">
          <label for="inventory-search" class="visually-hidden"
            >Search products</label
          >
          <input
            id="inventory-search"
            :value="searchQuery"
            @input="
              $emit('search-changed', ($event.target as HTMLInputElement).value)
            "
            type="text"
            placeholder="🔍 Search products..."
            @keyup.enter="$emit('search-submit')"
          />
          <button @click="$emit('search-submit')" class="btn-search">
            Search
          </button>
        </div>
        <button
          v-if="hasActiveFilters"
          @click="$emit('clear-filters')"
          class="btn-clear"
          title="Clear all filters"
        >
          ✕ Clear
        </button>
      </div>
    </div>

    <!-- Row 2: Action Buttons -->
    <div class="controls-row">
      <div class="action-buttons">
        <button
          @click="$emit('check-platform-status')"
          class="btn-action btn-sync-status"
          title="Batch check SKU to all platforms"
        >
          <span class="btn-icon">✔</span>
          <span class="btn-text">Sync Status</span>
        </button>
        <button
          @click="$emit('sync-from')"
          class="btn-action btn-sync-from"
          :disabled="syncing"
          title="Load data from Google Sheets to Database"
        >
          <span class="btn-icon">📥</span>
          <span v-if="!syncing" class="btn-text">Sync from Sheet</span>
          <span v-else class="btn-text">Syncing...</span>
        </button>
        <button
          @click="$emit('sync-to')"
          class="btn-action btn-sync-to"
          :disabled="syncing"
          title="Export data from Database to Google Sheets"
        >
          <span class="btn-icon">📤</span>
          <span v-if="!syncing" class="btn-text">Export to Sheet</span>
          <span v-else class="btn-text">Exporting...</span>
        </button>
        <button
          @click="handleUpdateStockClick"
          class="btn-action btn-update-stock"
          :disabled="updatingStock"
          title="Update stock to all platforms (Shopee, Lazada, TikTok)"
        >
          <span class="btn-icon">🔄</span>
          <span v-if="!updatingStock" class="btn-text">Update Stock</span>
          <span v-else class="btn-text">Updating...</span>
        </button>
        <button
          @click="handleUpdatePriceClick"
          class="btn-action btn-update-price"
          :disabled="updatingPrice"
          title="Update price to all platforms (Shopee, Lazada, TikTok)"
        >
          <span class="btn-icon">💰</span>
          <span v-if="!updatingPrice" class="btn-text">Update Price</span>
          <span v-else class="btn-text">Updating...</span>
        </button>
        <button
          @click="$emit('open-mpq-modal')"
          class="btn-action btn-mpq"
          title="Bulk Pricing Settings (Shopee + TikTok)"
        >
          <span class="btn-icon">⚙️</span>
          <span class="btn-text">Bulk Pricing</span>
        </button>
        <button
          @click="$emit('add-item')"
          class="btn-action btn-add"
          title="Add new item to inventory"
        >
          <span class="btn-icon">➕</span>
          <span class="btn-text">Add Item</span>
        </button>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
defineProps<{
  searchQuery: string;
  syncing: boolean;
  updatingStock: boolean;
  updatingPrice: boolean;
  hasActiveFilters: boolean;
}>();

const emit = defineEmits<{
  "search-changed": [query: string];
  "search-submit": [];
  "clear-filters": [];
  "check-platform-status": [];
  "sync-from": [];
  "sync-to": [];
  "update-stock": [];
  "update-price": [];
  "open-mpq-modal": [];
  "add-item": [];
}>();

function handleUpdateStockClick() {
  console.log("[InventoryHeader] Update Stock clicked");
  emit("update-stock");
}

function handleUpdatePriceClick() {
  console.log("[InventoryHeader] Update Price clicked");
  emit("update-price");
}
</script>

<style scoped>
@import "./InventoryContent.styles.css";

.inventory-header {
  display: flex;
  flex-direction: column;
  gap: 10px;
  margin-bottom: 14px;
}

.header-top {
  display: flex;
  align-items: center;
}

.header-top h2 {
  margin: 0;
  color: #333;
  font-size: 18px;
}

.header-controls {
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.controls-row {
  display: flex;
  gap: 8px;
  align-items: center;
  flex-wrap: wrap;
}

.search-and-filter {
  display: flex;
  gap: 8px;
  align-items: center;
  flex: 1;
  min-width: 260px;
}

.search-box {
  display: flex;
  gap: 6px;
  flex: 1;
  min-width: 220px;
}

.search-box input {
  flex: 1;
  min-width: 180px;
  padding: 8px 10px;
  border: 1px solid #ddd;
  border-radius: 5px;
  font-size: 13px;
}

.search-box input:focus {
  outline: none;
  border-color: #3498db;
  box-shadow: 0 0 0 3px rgba(52, 152, 219, 0.1);
}

.action-buttons {
  display: flex;
  gap: 10px;
  flex-wrap: wrap;
  align-items: center;
}

/* Responsive */
@media (max-width: 768px) {
  .controls-row {
    flex-direction: column;
    align-items: stretch;
  }

  .search-and-filter {
    flex-direction: column;
  }

  .action-buttons {
    flex-direction: column;
  }

  .btn-filter-toggle,
  .btn-sync-from,
  .btn-sync-to,
  .btn-add,
  .btn-wholesale,
  .btn-settings {
    width: 100%;
  }
}

.btn-wholesale {
  background: linear-gradient(135deg, #e74c3c 0%, #c0392b 100%);
  color: white;
}

.btn-wholesale:hover {
  background: linear-gradient(135deg, #c0392b 0%, #a93226 100%);
}

.btn-mpq {
  background: linear-gradient(135deg, #9b59b6 0%, #8e44ad 100%);
  color: white;
}

.btn-mpq:hover {
  background: linear-gradient(135deg, #8e44ad 0%, #7d3c98 100%);
}

/* Visually hidden but accessible to screen readers */
.visually-hidden {
  position: absolute;
  width: 1px;
  height: 1px;
  padding: 0;
  margin: -1px;
  overflow: hidden;
  clip: rect(0, 0, 0, 0);
  white-space: nowrap;
  border: 0;
}
</style>
