<template>
  <div class="dashboard">
    <div class="main-layout">
      <!-- Left Sidebar -->
      <LeftSidebar
        :collapsed="uiStore.leftSidebarCollapsed"
        :active-tab="uiStore.activeTab"
        :active-platform="uiStore.activePlatform"
        @toggle="uiStore.toggleLeftSidebar"
        @tab-change="handleTabChange"
        @platform-change="handlePlatformChange"
      />

      <!-- Main Content -->
      <main class="main-content">
        <!-- Settings Content -->
        <SettingsContent
          v-if="uiStore.activeTab === 'settings'"
          :loading="loading"
          @refresh-all="refreshAll"
        />

        <!-- Product Manager Content -->
        <ProductManager
          v-if="uiStore.activeTab === 'product-management'"
          :platform="uiStore.activePlatform"
          :status="status"
          :loading="loading"
        />

        <!-- Order Manager Content -->
        <OrderManager
          v-else-if="uiStore.activeTab === 'order-management'"
          :status="status"
          :loading="loading"
        />

        <!-- Empty State -->
        <div v-else class="empty-state">
          <i class="pi pi-inbox" aria-hidden="true"></i>
          <h3>No Content Selected</h3>
          <p>Please select a tab from the left sidebar</p>
        </div>
      </main>
    </div>

    <!-- Modals -->
    <DashboardModals
      :modals="modals"
      :wallet-params="walletParams"
      :shipping-params="shippingParams"
      :shipping-files="shippingFiles"
      :show-shipping-form="showShippingForm"
      :show-shipping-file-list="showShippingFileList"
      :current-price-option="currentPriceOption"
      @handle-token-operation="handleTokenOperation"
      @handle-update-wallet-params="handleUpdateWalletParams"
      @handle-export-wallet="exportWalletToSheets"
      @handle-close-shipping="handleCloseShippingModal"
      @handle-select-shipping-option="selectShippingOption"
      @handle-load-shipping-files="loadShippingFiles"
      @handle-process-shipping-file="processShippingFile"
      @handle-update-shipping-params="handleUpdateShippingParams"
      @handle-export-shipping="exportShippingToSheets"
      @handle-select-price-option="handleSelectPriceOption"
      @handle-execute-price-update="executePriceUpdate"
      @handle-execute-order-export="executeExportOrder"
      @close-modal="closeModal"
    />
  </div>
</template>

<script lang="ts">
import { computed, defineAsyncComponent, onMounted, watch } from "vue";
import { useAppStore } from "../store/app";
import { useUIStore } from "../store/ui";
import { useGoogleSheetsStore } from "../store/googleSheets";
import { useModalsStore } from "../store/modals";
import { useDashboardModals } from "@/composables/useDashboardModals";
import { useDashboardOperations } from "@/composables/useDashboardOperations";
import { useDashboardNavigation } from "@/composables/useDashboardNavigation";
import { useDashboardPlatformOps } from "@/composables/useDashboardPlatformOps";
import { useDashboardShipping } from "@/composables/useDashboardShipping";
import { useDashboardWallet } from "@/composables/useDashboardWallet";
import { useDashboardHandlers } from "@/composables/useDashboardHandlers";
import { useUnifiedHeader } from "@/composables/useUnifiedHeader";

// Layout Components
import LeftSidebar from "../components/layout/LeftSidebar.vue";
import DashboardModals from "../components/layout/DashboardModals.vue";

const SettingsContent = defineAsyncComponent(
  () =>
    import(
      /* webpackChunkName: "settings-chunk" */ "../components/content/SettingsContent"
    ),
);
const ProductManager = defineAsyncComponent(
  () =>
    import(
      /* webpackChunkName: "product-managers" */ "../components/ProductManager/ProductManager.vue"
    ),
);
const OrderManager = defineAsyncComponent(
  () =>
    import(
      /* webpackChunkName: "order-managers" */ "../components/OrderManager/OrderManager.vue"
    ),
);
const InventoryContent = defineAsyncComponent(
  () =>
    import(
      /* webpackChunkName: "inventory-chunk" */ "../components/content/InventoryContent/"
    ),
);

export default {
  name: "AppDashboard",
  components: {
    LeftSidebar,
    DashboardModals,
    SettingsContent,
    ProductManager,
    OrderManager,
    InventoryContent,
  },
  setup() {
    const appStore = useAppStore();
    const uiStore = useUIStore();
    const googleSheetsStore = useGoogleSheetsStore();
    const modalsStore = useModalsStore();

    // Type-safe wrapper for closeModal to use with composables
    const closeModalWrapper = (modal: string) => {
      modalsStore.closeModal(modal as keyof typeof modalsStore.modals);
    };

    // Composables
    const {
      currentPriceOption,
      showShippingForm,
      showShippingFileList,
      shippingFiles,
    } = useDashboardModals();

    const {
      lastUpdated,
      walletParams,
      shippingParams,
      refreshAll,
      loadStatus,
      executeOperation,
    } = useDashboardOperations();

    const {
      handleTabChange,
      handleNavigateToSettings,
      handlePlatformChange,
      initRouteSync,
    } = useDashboardNavigation();

    const {
      handlePlatformOperation,
      executePriceUpdate,
      executeExportOrder,
      handleTokenOperation,
    } = useDashboardPlatformOps();

    const {
      selectShippingOption,
      loadShippingFiles,
      processShippingFile,
      exportShippingToSheets,
      initializeShippingForm,
    } = useDashboardShipping(
      shippingParams,
      showShippingForm,
      showShippingFileList,
      closeModalWrapper,
    );

    const { exportWalletToSheets } = useDashboardWallet(
      walletParams,
      closeModalWrapper,
    );

    const {
      handleCloseShippingModal,
      handleSelectPriceOption,
      handleUpdateWalletParams,
      handleUpdateShippingParams,
    } = useDashboardHandlers(
      walletParams,
      shippingParams,
      currentPriceOption,
      showShippingForm,
      showShippingFileList,
      modalsStore,
    );

    // Computed
    const status = computed(() => appStore.status);
    const loading = computed(() => appStore.loading);
    const connectionStatus = computed(() => appStore.connectionStatus);
    const requestCount = computed(() => appStore.requestCount);
    const errorCount = computed(() => appStore.errorCount);

    // Unified header integration
    const unifiedHeader = useUnifiedHeader();

    // Sync app state with unified header
    const syncHeaderState = () => {
      unifiedHeader.setConnectionStatus(connectionStatus.value);
      unifiedHeader.setLoading(loading.value);
      unifiedHeader.registerRefreshCallback(refreshAll);
    };

    // Lifecycle - Non-blocking initialization for faster FCP/LCP
    onMounted(() => {
      console.log(
        "[Dashboard] Component mounted, activeTab:",
        uiStore.activeTab,
      );
      // Sync header immediately for UI feedback
      syncHeaderState();
      initRouteSync();

      // Defer app initialization to after first paint
      // This reduces critical path latency significantly
      requestAnimationFrame(() => {
        appStore.initializeApp();
      });
    });

    // Watch for changes to sync with header
    watch([connectionStatus, loading], () => {
      unifiedHeader.setConnectionStatus(connectionStatus.value);
      unifiedHeader.setLoading(loading.value);
    });

    const handleShowTokenModal = (platform: string) => {
      modalsStore.showModal("token", platform);
    };

    const handleShowExportOrdersModal = (platform: string) => {
      modalsStore.showModal("exportOrders", platform);
    };

    const handleShowPriceModal = (platform: string) => {
      modalsStore.showModal("price", platform);
    };

    const handleShowWalletModal = () => {
      modalsStore.showModal("wallet");
    };

    const handleShowShippingModal = () => {
      modalsStore.showModal("shipping");
      initializeShippingForm();
    };

    return {
      // Stores
      uiStore,
      googleSheetsStore,
      appStore,
      modalsStore,

      // State
      lastUpdated,
      currentPriceOption,
      showShippingForm,
      showShippingFileList,
      walletParams,
      shippingParams,

      // Computed
      status,
      loading,
      connectionStatus,
      requestCount,
      errorCount,
      modals: computed(() => modalsStore.modals),
      shippingFiles,

      // Methods
      handleTabChange,
      handleNavigateToSettings,
      handlePlatformChange,
      handleShowTokenModal,
      handleShowExportOrdersModal,
      handleShowPriceModal,
      handleShowWalletModal,
      handleShowShippingModal,
      handleCloseShippingModal,
      handleSelectPriceOption,
      handleUpdateWalletParams,
      handleUpdateShippingParams,
      closeModal: closeModalWrapper,
      exportWalletToSheets,
      selectShippingOption,
      loadShippingFiles,
      processShippingFile,
      exportShippingToSheets,
      initializeShippingForm,
      handlePlatformOperation,
      executePriceUpdate,
      executeExportOrder,
      handleTokenOperation,
      executeOperation,
      refreshAll,
      loadStatus,
    };
  },
};
</script>

<style scoped>
@import "./Dashboard.module.css";
</style>
