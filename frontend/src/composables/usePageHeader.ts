/**
 * Page Header Integration Composable
 * Syncs page state with the unified header in Navbar
 * Use in any page that needs connection status, loading, and refresh
 */
import { computed, watch, onMounted, onBeforeUnmount } from "vue";
import { useUnifiedHeader, type ConnectionStatus } from "./useUnifiedHeader";
import { useAppStore } from "../store/app";

interface PageHeaderOptions {
  refreshCallback?: () => void;
  showRefresh?: boolean;
  pageTitle?: string;
}

export function usePageHeader(options: PageHeaderOptions = {}) {
  const appStore = useAppStore();
  const unifiedHeader = useUnifiedHeader();

  // Computed from app store
  const connectionStatus = computed(
    () => appStore.connectionStatus as ConnectionStatus
  );
  const loading = computed(() => appStore.loading);
  const status = computed(() => appStore.status);
  const requestCount = computed(() => appStore.requestCount);
  const errorCount = computed(() => appStore.errorCount);

  // Sync state with unified header
  const syncHeaderState = () => {
    unifiedHeader.setConnectionStatus(connectionStatus.value);
    unifiedHeader.setLoading(loading.value);

    if (options.showRefresh !== undefined) {
      unifiedHeader.setShowRefresh(options.showRefresh);
    }

    if (options.pageTitle) {
      unifiedHeader.setPageTitle(options.pageTitle);
    }

    if (options.refreshCallback) {
      unifiedHeader.registerRefreshCallback(options.refreshCallback);
    }
  };

  // Watch for state changes
  watch([connectionStatus, loading], () => {
    unifiedHeader.setConnectionStatus(connectionStatus.value);
    unifiedHeader.setLoading(loading.value);
  });

  onMounted(() => {
    syncHeaderState();
  });

  onBeforeUnmount(() => {
    // Clean up callback when page unmounts
    unifiedHeader.unregisterRefreshCallback();
  });

  return {
    // State
    connectionStatus,
    loading,
    status,
    requestCount,
    errorCount,

    // Actions
    setLoading: unifiedHeader.setLoading,
    setConnectionStatus: unifiedHeader.setConnectionStatus,
    triggerRefresh: unifiedHeader.triggerRefresh,
  };
}
