<template>
  <div class="order-manager">
    <!-- Header Section -->
    <div class="order-header">
      <!-- Platform Stats Cards -->
      <OrderHeader
        :activeTab="activeTab"
        :searchQuery="searchQuery"
        :selectedPlatform="selectedPlatform"
        :orders="orders"
        @search-changed="searchQuery = $event"
        @platform-changed="selectedPlatform = $event"
      />

      <!-- Tabs & Actions -->
      <div class="tabs-actions-container">
        <!-- Tabs -->
        <OrderTabs
          :activeTab="activeTab"
          :tabCounts="tabCounts"
          :orderTabs="ORDER_TABS"
          @tab-changed="changeTab"
        />

        <!-- Actions -->
        <div class="order-actions">
          <button
            @click="refreshData"
            :disabled="loading"
            class="btn-refresh"
            title="Refresh data terbaru"
          >
            <i :class="loading ? 'pi pi-spin pi-spinner' : 'pi pi-refresh'"></i>
            <span class="btn-text">Refresh</span>
          </button>

          <!-- Export Dropdown -->
          <div class="export-dropdown" ref="exportDropdownRef">
            <button
              @click="toggleExportMenu"
              :disabled="loading || filteredOrders.length === 0"
              class="btn-export"
              title="Export options"
              aria-label="Export options"
            >
              <i class="pi pi-download" aria-hidden="true"></i>
              <span class="btn-text">Export</span>
              <i
                class="pi pi-chevron-down export-chevron"
                aria-hidden="true"
              ></i>
            </button>

            <div v-if="showExportMenu" class="export-menu">
              <button
                class="export-menu-item"
                @click="handleExportToN8N"
                title="Kirim data ke N8N untuk otomasi"
              >
                <i class="pi pi-send" aria-hidden="true"></i>
                <span>Export ke N8N</span>
              </button>
              <button
                class="export-menu-item"
                @click="handleExportToCSV"
                title="Download file CSV ke komputer"
              >
                <i class="pi pi-file" aria-hidden="true"></i>
                <span>Download CSV</span>
              </button>
            </div>
          </div>
        </div>
      </div>
    </div>

    <!-- States (loading, error, empty) -->
    <OrderStates
      :loading="loading"
      :error="error"
      :isNoData="filteredOrders.length === 0"
      :searchActive="!!searchQuery"
      @close-error="error = null"
    />

    <!-- Data table with summary -->
    <template v-if="!loading && !error && filteredOrders.length > 0">
      <OrderTable :filteredOrders="filteredOrders" :activeTab="activeTab" />

      <OrderSummary
        :filteredOrders="filteredOrders"
        :uniqueOrders="uniqueOrders"
        :uniquePlatforms="uniquePlatforms"
      />
    </template>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, onUnmounted } from "vue";
import {
  useOrderManager,
  ORDER_TABS,
  type Order,
} from "./composables/useOrderManager";
import { useOrderExport } from "./composables/useOrderExport";
import OrderHeader from "./OrderHeader.vue";
import OrderTabs from "./OrderTabs.vue";
import OrderTable from "./OrderTable.vue";
import OrderSummary from "./OrderSummary.vue";
import OrderStates from "./OrderStates.vue";

const {
  activeTab,
  orders,
  loading,
  error,
  searchQuery,
  selectedPlatform,
  tabCounts,
  uniquePlatforms,
  uniqueOrders,
  filteredOrders,
  changeTab,
  refreshData,
  initializeFromUrl,
} = useOrderManager();

const { exportToCSVOnly, exportToN8NOnly } = useOrderExport();

// Export dropdown state
const showExportMenu = ref(false);
const exportDropdownRef = ref<HTMLElement | null>(null);

const toggleExportMenu = () => {
  showExportMenu.value = !showExportMenu.value;
};

const closeExportMenu = () => {
  showExportMenu.value = false;
};

const handleExportToN8N = async () => {
  closeExportMenu();
  await exportToN8NOnly(filteredOrders.value as Order[], activeTab.value);
};

const handleExportToCSV = async () => {
  closeExportMenu();
  await exportToCSVOnly(filteredOrders.value as Order[], activeTab.value);
};

// Close dropdown when clicking outside
const handleClickOutside = (event: MouseEvent) => {
  if (
    exportDropdownRef.value &&
    !exportDropdownRef.value.contains(event.target as Node)
  ) {
    closeExportMenu();
  }
};

onMounted(() => {
  initializeFromUrl();
  document.addEventListener("click", handleClickOutside);
});

onUnmounted(() => {
  document.removeEventListener("click", handleClickOutside);
});
</script>

<style scoped>
@import "./OrderManager.styles.css";

/* Export Dropdown Styles */
.export-dropdown {
  position: relative;
  display: inline-block;
}

.btn-export {
  display: flex;
  align-items: center;
  gap: 0.5rem;
}

.export-chevron {
  font-size: 0.7rem;
  margin-left: 0.25rem;
}

.export-menu {
  position: absolute;
  top: 100%;
  right: 0;
  margin-top: 0.25rem;
  background: white;
  border: 1px solid #e2e8f0;
  border-radius: 0.5rem;
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.15);
  min-width: 180px;
  z-index: 100;
  overflow: hidden;
}

.export-menu-item {
  display: flex;
  align-items: center;
  gap: 0.75rem;
  width: 100%;
  padding: 0.75rem 1rem;
  border: none;
  background: transparent;
  cursor: pointer;
  font-size: 0.875rem;
  color: #374151;
  text-align: left;
  transition: background-color 0.15s ease;
}

.export-menu-item:hover {
  background-color: #f3f4f6;
}

.export-menu-item:first-child {
  border-bottom: 1px solid #e5e7eb;
}

.export-menu-item i {
  color: #6b7280;
  font-size: 1rem;
}

.export-menu-item:hover i {
  color: #374151;
}
</style>
