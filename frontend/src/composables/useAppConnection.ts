import { ref, Ref } from "vue";

export function useAppConnection() {
  const connectionStatus: Ref<"connecting" | "connected" | "error" | "disconnected"> =
    ref("connecting");
  const isReconnecting: Ref<boolean> = ref(false);
  const statusPollingInterval: Ref<ReturnType<typeof setInterval> | null> = ref(null);

  const isConnected = () => connectionStatus.value === "connected";

  const startStatusPolling = (loadStatusFn: () => Promise<void>) => {
    if (statusPollingInterval.value) clearInterval(statusPollingInterval.value);
    statusPollingInterval.value = setInterval(async () => {
      if (
        document.visibilityState === "visible" &&
        connectionStatus.value === "connected"
      ) {
        await loadStatusFn();
      }
    }, 30000);
  };

  const cleanup = () => {
    if (statusPollingInterval.value) {
      clearInterval(statusPollingInterval.value);
      statusPollingInterval.value = null;
    }
  };

  return {
    connectionStatus,
    isReconnecting,
    statusPollingInterval,
    isConnected,
    startStatusPolling,
    cleanup,
  };
}
