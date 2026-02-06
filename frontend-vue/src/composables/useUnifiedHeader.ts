/**
 * Unified Header State Composable
 * Provides centralized state management for header actions across all pages
 * Follows SRP - single responsibility for header state only
 */
import { ref, readonly, computed, type Ref, type ComputedRef } from "vue";

export type ConnectionStatus =
  | "connecting"
  | "connected"
  | "error"
  | "disconnected";

interface HeaderState {
  connectionStatus: Ref<ConnectionStatus>;
  isLoading: Ref<boolean>;
  pageTitle: Ref<string>;
  showRefresh: Ref<boolean>;
}

interface HeaderActions {
  setConnectionStatus: (status: ConnectionStatus) => void;
  setLoading: (loading: boolean) => void;
  setPageTitle: (title: string) => void;
  setShowRefresh: (show: boolean) => void;
  triggerRefresh: () => void;
}

interface HeaderCallbacks {
  onRefresh: Ref<(() => void) | null>;
  registerRefreshCallback: (callback: () => void) => void;
  unregisterRefreshCallback: () => void;
}

// Singleton state for header
const connectionStatus = ref<ConnectionStatus>("connecting");
const isLoading = ref(false);
const pageTitle = ref("Dashboard");
const showRefresh = ref(true);
const onRefreshCallback = ref<(() => void) | null>(null);

export function useUnifiedHeader(): HeaderState &
  HeaderActions &
  HeaderCallbacks {
  const setConnectionStatus = (status: ConnectionStatus) => {
    connectionStatus.value = status;
  };

  const setLoading = (loading: boolean) => {
    isLoading.value = loading;
  };

  const setPageTitle = (title: string) => {
    pageTitle.value = title;
  };

  const setShowRefresh = (show: boolean) => {
    showRefresh.value = show;
  };

  const triggerRefresh = () => {
    if (onRefreshCallback.value) {
      onRefreshCallback.value();
    }
  };

  const registerRefreshCallback = (callback: () => void) => {
    onRefreshCallback.value = callback;
  };

  const unregisterRefreshCallback = () => {
    onRefreshCallback.value = null;
  };

  return {
    // State (readonly refs)
    connectionStatus: readonly(connectionStatus) as Ref<ConnectionStatus>,
    isLoading: readonly(isLoading) as Ref<boolean>,
    pageTitle: readonly(pageTitle) as Ref<string>,
    showRefresh: readonly(showRefresh) as Ref<boolean>,

    // Actions
    setConnectionStatus,
    setLoading,
    setPageTitle,
    setShowRefresh,
    triggerRefresh,

    // Callbacks
    onRefresh: onRefreshCallback,
    registerRefreshCallback,
    unregisterRefreshCallback,
  };
}

// Connection status text helper
export function useConnectionStatusText(
  status: Ref<ConnectionStatus>
): ComputedRef<string> {
  return computed(() => {
    const statusMap: Record<ConnectionStatus, string> = {
      connecting: "Connecting...",
      connected: "Connected",
      error: "Disconnected",
      disconnected: "Offline",
    };
    return statusMap[status.value] || "Unknown";
  });
}
