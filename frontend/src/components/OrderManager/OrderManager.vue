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

      <!-- Filter Bar -->
      <OrderFilterBar
        :searchQuery="searchQuery"
        :selectedPlatform="selectedPlatform"
        :platforms="uniquePlatforms"
        @update:searchQuery="searchQuery = $event"
        @update:selectedPlatform="selectedPlatform = $event"
        @apply-filters="applyFilters"
        @reset-filters="resetFilters"
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
          <!-- Bulk Shipment Button -->
          <button
            v-if="activeTab === 'unprocess'"
            @click="openBulkShipModal"
            :disabled="loading"
            class="btn-bulk-ship"
            title="Ship multiple orders at once"
          >
            <i class="pi pi-truck"></i>
            <span class="btn-text">Bulk Shipment</span>
          </button>

          <button
            @click="refreshData"
            :disabled="loading"
            class="btn-refresh"
            title="Refresh latest data"
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
                title="Send data to N8N for automation"
              >
                <i class="pi pi-send" aria-hidden="true"></i>
                <span>Export to N8N</span>
              </button>
              <button
                class="export-menu-item"
                @click="handleExportToCSV"
                title="Download CSV file"
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
      <OrderTable
        :filteredOrders="filteredOrders"
        :activeTab="activeTab"
        @ship-order="openShipModal"
        @cancel-order="openCancelModal"
        @view-detail="handleViewDetail"
        @copy-order-number="handleCopyOrderNumber"
      />

      <OrderSummary
        :filteredOrders="filteredOrders"
        :uniqueOrders="uniqueOrders"
        :uniquePlatforms="uniquePlatforms"
      />
    </template>

    <!-- Ship Modal -->
    <OrderShipModal
      :visible="showShipModal"
      :order="selectedOrder"
      @close="closeShipModal"
      @confirm="handleShipConfirm"
    />

    <!-- Cancel Modal -->
    <OrderCancelModal
      :visible="showCancelModal"
      :order="selectedOrder"
      @close="closeCancelModal"
      @confirm="handleCancelConfirm"
    />
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
import {
  useOrderActions,
  type ShipOrderParams,
  type CancelOrderParams,
} from "./composables/useOrderActions";
import OrderHeader from "./OrderHeader.vue";
import OrderTabs from "./OrderTabs.vue";
import OrderTable from "./OrderTable.vue";
import OrderSummary from "./OrderSummary.vue";
import OrderStates from "./OrderStates.vue";
import OrderFilterBar from "./OrderFilterBar.vue";
import OrderShipModal from "./OrderShipModal.vue";
import OrderCancelModal from "./OrderCancelModal.vue";

// Composables
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
const { shipOrder, cancelOrder } = useOrderActions();

// Export dropdown state
const showExportMenu = ref(false);
const exportDropdownRef = ref<HTMLElement | null>(null);

// Bulk Shipment state
const selectedOrders = ref<string[]>([]);
const showBulkShipModal = ref(false);

// Modal state
const showShipModal = ref(false);
const showCancelModal = ref(false);
const selectedOrder = ref<any>(null);

// Export handlers
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

// Filter handlers
const applyFilters = () => {
  // Filters are already reactive, this is for explicit apply action
  console.log("Filters applied:", {
    search: searchQuery.value,
    platform: selectedPlatform.value,
  });
};

const resetFilters = () => {
  searchQuery.value = "";
  selectedPlatform.value = "";
};

const openBulkShipModal = () => {
  showBulkShipModal.value = true;
};

// Ship modal handlers
const openShipModal = (order: any) => {
  selectedOrder.value = order;
  showShipModal.value = true;
};

const closeShipModal = () => {
  showShipModal.value = false;
  selectedOrder.value = null;
};

const handleShipConfirm = async (data: {
  order_no: string;
  shipping_provider: string;
  tracking_number?: string;
  address_id?: number;
}) => {
  const params: ShipOrderParams = {
    order_no: data.order_no,
    platform: selectedOrder.value?.platform || "shopee",
    shipping_provider: data.shipping_provider,
    tracking_number: data.tracking_number,
    address_id: data.address_id,
  };

  const result = await shipOrder(params);
  if (result.success) {
    closeShipModal();
    await refreshData();
  }
};

// Cancel modal handlers
const openCancelModal = (order: any) => {
  selectedOrder.value = order;
  showCancelModal.value = true;
};

const closeCancelModal = () => {
  showCancelModal.value = false;
  selectedOrder.value = null;
};

const handleCancelConfirm = async (data: {
  order_no: string;
  cancel_reason: string;
  reason_detail?: string;
}) => {
  const params: CancelOrderParams = {
    order_no: data.order_no,
    platform: selectedOrder.value?.platform || "shopee",
    cancel_reason: data.cancel_reason,
    reason_detail: data.reason_detail,
  };

  const result = await cancelOrder(params);
  if (result.success) {
    closeCancelModal();
    await refreshData();
  }
};

// View detail handler
const handleViewDetail = (order: any) => {
  console.log("View detail:", order);
  // TODO: Implement order detail view/modal
};

// Copy handler
const handleCopyOrderNumber = (orderNo: string) => {
  console.log("Copied order number:", orderNo);
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
@import "./OrderManager.theme.css";

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
  border: 1px solid var(--om-border);
  border-radius: var(--om-radius-md);
  box-shadow: var(--om-shadow-lg);
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
  font-size: var(--om-font-sm);
  color: var(--om-text-primary);
  text-align: left;
  transition: background-color var(--om-transition-fast);
}

.export-menu-item:hover {
  background-color: var(--om-bg-secondary);
}

.export-menu-item:first-child {
  border-bottom: 1px solid var(--om-border);
}

.export-menu-item i {
  color: var(--om-text-secondary);
  font-size: 1rem;
}

.export-menu-item:hover i {
  color: var(--om-text-primary);
}

.btn-bulk-ship {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  padding: 0.5rem 1rem;
  background: #ee4d2d;
  color: white;
  border: none;
  border-radius: var(--om-radius-sm);
  font-size: var(--om-font-sm);
  font-weight: 500;
  cursor: pointer;
  transition: background-color var(--om-transition-fast);
}

.btn-bulk-ship:hover:not(:disabled) {
  background: #d73211;
}

.btn-bulk-ship:disabled {
  opacity: 0.6;
  cursor: not-allowed;
}
</style>
