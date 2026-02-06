import { ref, computed, Ref } from "vue";
import { getAuthHeaders, getApiBaseUrl } from "@/utils/apiHeaders";

interface LogEntry {
  timestamp: string;
  level: "DEBUG" | "INFO" | "WARNING" | "ERROR" | "CRITICAL";
  logger: string;
  message: string;
}

interface LogsData {
  logs: LogEntry[];
  total_count: number;
  filtered_count: number;
}

export function useSettingsLogs() {
  const logsData: Ref<LogsData> = ref({
    logs: [],
    total_count: 0,
    filtered_count: 0,
  });
  const logsLoading: Ref<boolean> = ref(false);
  const logFilter: Ref<string> = ref("");
  const lastLogsUpdate: Ref<string> = ref("");

  const filteredLogs = computed((): LogEntry[] => {
    if (!logFilter.value) {
      return logsData.value.logs;
    }
    return logsData.value.logs.filter(
      (log: LogEntry) => log.level === logFilter.value
    );
  });

  const refreshLogs = async (): Promise<void> => {
    logsLoading.value = true;
    try {
      const response = await fetch(
        `${getApiBaseUrl("/settings/debug-logs")}?limit=500&level=${logFilter.value || ""}`,
        { headers: getAuthHeaders() }
      );
      const result = await response.json();
      if (result.success) {
        logsData.value = result.data;
        lastLogsUpdate.value = new Date().toLocaleTimeString();
      }
    } catch (error) {
      console.error("Error fetching logs:", error);
    } finally {
      logsLoading.value = false;
    }
  };

  const clearLogs = async (): Promise<void> => {
    if (!confirm("Are you sure you want to delete all logs?")) return;

    try {
      const response = await fetch(
        getApiBaseUrl("/settings/debug-logs/clear"),
        {
          method: "POST",
          headers: getAuthHeaders(),
        }
      );
      const result = await response.json();
      if (result.success) {
        logsData.value = { logs: [], total_count: 0, filtered_count: 0 };
        alert(result.message);
      }
    } catch (error) {
      console.error("Error clearing logs:", error);
      alert("Failed to delete logs");
    }
  };

  return {
    logsData,
    logsLoading,
    logFilter,
    lastLogsUpdate,
    filteredLogs,
    refreshLogs,
    clearLogs,
  };
}
