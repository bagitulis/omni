import { useEffect, useCallback } from "react";
import { useAppStore } from "@/stores/appStore";
import apiClient from "@/api/client";

export function useConnectionStatus() {
  const setConnectionStatus = useAppStore((state) => state.setConnectionStatus);
  const connectionStatus = useAppStore((state) => state.connectionStatus);
  const isConnected = useAppStore((state) => state.isConnected);

  const checkHealth = useCallback(async () => {
    try {
      const response = await apiClient.healthCheck();
      if (response && response.success) {
        setConnectionStatus("connected");
      } else {
        setConnectionStatus("disconnected");
      }
    } catch {
      setConnectionStatus("error");
    }
  }, [setConnectionStatus]);

  useEffect(() => {
    let intervalId: ReturnType<typeof setInterval> | null = null;

    // Initial check
    checkHealth();

    const startPolling = () => {
      // Clear any existing interval just in case
      if (intervalId) clearInterval(intervalId);

      intervalId = setInterval(() => {
        // Only poll if tab is visible
        if (document.visibilityState === "visible") {
          checkHealth();
        }
      }, 30000); // 30 seconds
    };

    const stopPolling = () => {
      if (intervalId) {
        clearInterval(intervalId);
        intervalId = null;
      }
    };

    const handleVisibilityChange = () => {
      if (document.visibilityState === "visible") {
        // Immediate check when becoming visible
        checkHealth();
        // Ensure polling is running
        startPolling();
      } else {
        // Pause polling when hidden
        stopPolling();
      }
    };

    // Start polling on mount
    startPolling();

    // Listen for visibility changes
    document.addEventListener("visibilitychange", handleVisibilityChange);

    // Cleanup
    return () => {
      stopPolling();
      document.removeEventListener("visibilitychange", handleVisibilityChange);
    };
  }, [checkHealth]);

  return { connectionStatus, isConnected };
}
