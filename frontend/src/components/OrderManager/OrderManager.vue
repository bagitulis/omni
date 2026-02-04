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

      <!-- Filters + Actions -->
      <div class="order-controls">
        <OrderFilterBar
          :searchQuery="searchQuery"
          :selectedPlatform="selectedPlatform"
          :platforms="uniquePlatforms"
          @update:searchQuery="searchQuery = $event"
          @update:selectedPlatform="selectedPlatform = $event"
          @apply-filters="applyFilters"
          @reset-filters="resetFilters"
        />

        <OrderActionsBar
          :activeTab="activeTab"
          :loading="loading"
          :hasData="filteredOrders.length > 0"
          @bulk-ship="openBulkShipModal"
          @refresh="refreshData"
          @export-n8n="handleExportToN8N"
          @export-csv="handleExportToCSV"
        />
      </div>

      <!-- Tabs -->
      <OrderTabs
        :activeTab="activeTab"
        :tabCounts="tabCounts"
        :orderTabs="ORDER_TABS"
        @tab-changed="changeTab"
      />
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

    <!-- Detail Modal -->
    <OrderDetailModal
      :visible="showDetailModal"
      :order="selectedOrder"
      @close="closeDetailModal"
    />
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from "vue";
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
import OrderActionsBar from "./OrderActionsBar.vue";
import OrderDetailModal from "./OrderDetailModal.vue";

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

// Bulk Shipment state
const showBulkShipModal = ref(false);

// Modal state
const showShipModal = ref(false);
const showCancelModal = ref(false);
const showDetailModal = ref(false);
const selectedOrder = ref<any>(null);

// Export handlers
const handleExportToN8N = async () => {
  await exportToN8NOnly(filteredOrders.value as Order[], activeTab.value);
};

const handleExportToCSV = async () => {
  await exportToCSVOnly(filteredOrders.value as Order[], activeTab.value);
};

// Filter handlers
const applyFilters = () => {
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
  const platform = (selectedOrder.value?.platform || "shopee").toLowerCase();
  const params: ShipOrderParams = {
    order_no: data.order_no,
    platform,
    shipping_provider: data.shipping_provider,
    tracking_number: data.tracking_number,
    address_id: data.address_id,
    pickup_time_id: data.pickup_time_id,
  };

  if (platform === "lazada") {
    const orderItemIds = selectedOrder.value?.items
      ?.map((item: any) => String(item.order_item_id || item.item_id || ""))
      .filter((id: string) => id !== "");
    params.order_item_ids = orderItemIds?.length
      ? orderItemIds
      : [data.order_no];
  }

  if (platform === "tiktok") {
    params.package_id = data.package_id || "";
  }

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
  const platform = (selectedOrder.value?.platform || "shopee").toLowerCase();
  const params: CancelOrderParams = {
    order_no: data.order_no,
    platform,
    cancel_reason: data.cancel_reason,
    reason_detail: data.reason_detail,
  };

  if (platform === "lazada") {
    const orderItemId =
      selectedOrder.value?.items?.[0]?.order_item_id ||
      selectedOrder.value?.items?.[0]?.item_id ||
      selectedOrder.value?.order_item_id ||
      selectedOrder.value?.orderItemId ||
      data.order_no;
    params.order_item_id = String(orderItemId);
  }

  const result = await cancelOrder(params);
  if (result.success) {
    closeCancelModal();
    await refreshData();
  }
};

// View detail handler
const handleViewDetail = (order: any) => {
  selectedOrder.value = order;
  showDetailModal.value = true;
};

const closeDetailModal = () => {
  showDetailModal.value = false;
  selectedOrder.value = null;
};

// Copy handler
const handleCopyOrderNumber = (orderNo: string) => {
  console.log("Copied order number:", orderNo);
};

onMounted(() => {
  initializeFromUrl();
});
</script>

<style scoped>
@import "./OrderManager.styles.css";
@import "./styles/variables.css";
@import "./styles/components.css";
@import "./styles/layout.css";
</style>
