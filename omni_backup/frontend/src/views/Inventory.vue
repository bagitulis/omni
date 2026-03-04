<template>
  <div class="inventory-page">
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
        <div class="inventory-wrapper">
          <ErrorBoundary>
            <InventoryContent
              :spreadsheet-id="googleSheetsStore.inventorySpreadsheetId"
              :sheet-name="googleSheetsStore.inventorySheetName"
              @navigate-to-settings="navigateToSettings"
            />
          </ErrorBoundary>
        </div>
      </main>
    </div>
  </div>
</template>

<script lang="ts">
import { ref, Ref, onMounted, computed } from "vue";
import { useRouter } from "vue-router";
import { useUIStore } from "../store/ui";
import { useAppStore } from "../store/app";
import { useModalsStore } from "../store/modals";
import { useGoogleSheetsStore } from "../store/googleSheets";
import { defineAsyncComponent } from "vue";
import {
  useUnifiedHeader,
  type ConnectionStatus,
} from "@/composables/useUnifiedHeader";

// Layout Components
import LeftSidebar from "../components/layout/LeftSidebar.vue";
import ErrorBoundary from "../components/ErrorBoundary.vue";

// Lazy load InventoryContent
const InventoryContent = defineAsyncComponent(
  () =>
    import(
      /* webpackChunkName: "inventory-chunk" */ "../components/content/InventoryContent/"
    )
);

export default {
  name: "InventoryPage",
  components: {
    LeftSidebar,
    InventoryContent,
    ErrorBoundary,
  },
  setup() {
    const router = useRouter();
    const uiStore = useUIStore();
    const appStore = useAppStore();
    const modalsStore = useModalsStore();
    const googleSheetsStore = useGoogleSheetsStore();

    // Set active tab to inventory
    uiStore.setActiveTab("inventory");

    // Unified header integration
    const unifiedHeader = useUnifiedHeader();

    const connectionStatus = computed(() => appStore.connectionStatus);
    const loading = computed(() => appStore.loading);
    const status = computed(() => appStore.status);
    const lastUpdated: Ref<string> = ref(new Date().toLocaleString());
    const requestCount = computed(() => appStore.requestCount);
    const errorCount = computed(() => appStore.errorCount);

    // Sync with unified header
    const syncHeader = () => {
      unifiedHeader.setConnectionStatus(
        connectionStatus.value as ConnectionStatus
      );
      unifiedHeader.setLoading(loading.value);
      unifiedHeader.setPageTitle("Inventory");
    };

    // Load token status when page opens
    onMounted(() => {
      // Sync header immediately for UI feedback
      syncHeader();
      unifiedHeader.registerRefreshCallback(() => appStore.loadStatus(true));

      // Non-blocking: defer init to after first paint
      const initAndLoad = async () => {
        // Initialize app if not already initialized (for direct route access like /inventory)
        if (appStore.connectionStatus === "connecting") {
          await appStore.initializeApp();
        }
        // Load initial status
        await appStore.loadStatus(true);
      };

      requestAnimationFrame(() => initAndLoad());
    });

    const handleTabChange = (tab: string): void => {
      uiStore.setActiveTab(tab as any);
      if (tab !== "inventory") {
        router.push("/");
      }
    };

    const handlePlatformChange = (platform: string): void => {
      uiStore.setActivePlatform(platform as any);
    };

    const navigateToSettings = () => {
      uiStore.setActiveTab("settings");
      router.push("/");
    };

    const showTokenModal = () => {
      modalsStore.showModal("token", uiStore.activePlatform);
    };

    const loadStatus = async () => {
      await appStore.loadStatus(true);
    };

    return {
      uiStore,
      appStore,
      googleSheetsStore,
      connectionStatus,
      loading,
      lastUpdated,
      requestCount,
      errorCount,
      status,
      handleTabChange,
      handlePlatformChange,
      navigateToSettings,
      showTokenModal,
      loadStatus,
    };
  },
};
</script>

<style scoped>
.inventory-page {
  display: flex;
  flex-direction: column;
  height: 100vh;
  overflow: hidden;
}

.main-layout {
  display: flex;
  flex: 1;
  overflow: hidden;
}

.main-content {
  flex: 1;
  overflow-y: auto;
  display: flex;
  flex-direction: column;
}

.inventory-wrapper {
  display: flex;
  flex-direction: column;
  gap: 20px;
  padding: 20px;
}
</style>
