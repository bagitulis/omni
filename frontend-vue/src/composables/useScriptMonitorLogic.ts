import { ref, onMounted, onUnmounted } from "vue";
import apiClient from "@/services/api";

/**
 * Script Monitor Logic Composable
 * JSON uses snake_case as per AGENTS.md standard
 */

interface Job {
  id: string;
  type: string;
  status: string;
  priority: string;
  data: Record<string, any>;
  started_at?: Date;
  created_at: Date;
}

interface JobHistory {
  id: number;
  job_id: string;
  job_type?: string;
  status: string;
  duration_ms?: number;
  error_message?: string;
  created_at: Date;
  started_at?: Date;
  completed_at?: Date;
}

interface AutoFunctionConfig {
  id: number;
  name: string;
  enabled: boolean;
  interval_minutes: number;
  start_time?: string;
  end_time?: string;
  last_executed?: string | Date;
  next_scheduled_execution?: string | Date;
  created_at?: string | Date;
  updated_at?: string | Date;
}

const JOB_TIMEOUT_MINUTES = 5;
let cancelledScheduledIds = new Set<number>(); // Track cancelled scheduled items

export function useScriptMonitorLogic() {
  const isLoading = ref(false);
  const isClearing = ref(false);
  const error = ref<string | null>(null);
  const lastUpdate = ref<Date | null>(null);
  const currentJob = ref<Job | null>(null);
  const pendingQueue = ref<Job[]>([]);
  const recentHistory = ref<JobHistory[]>([]);
  const autoFunctionConfigs = ref<AutoFunctionConfig[]>([]);
  const editingConfig = ref<AutoFunctionConfig | null>(null);
  const isJobStuck = ref(false);
  const queueCount = ref(0);

  const stats = ref({
    completed: 0,
    failed: 0,
    pending: 0,
    running: 0,
  });

  let refreshInterval: number | null = null;

  onMounted(() => {
    refreshMonitor();
    refreshInterval = window.setInterval(refreshMonitor, 2000);
  });

  onUnmounted(() => {
    if (refreshInterval) clearInterval(refreshInterval);
    cancelledScheduledIds.clear();
  });

  async function refreshMonitor() {
    isLoading.value = true;
    error.value = null;
    let monitorData: any = null;

    try {
      const response = await apiClient.client.get("/jobs/monitor");

      if (response.data.success) {
        monitorData = response.data.data;
        currentJob.value = monitorData.currentJob || monitorData.current_job;
        pendingQueue.value =
          monitorData.pendingQueue || monitorData.pending_queue;
        queueCount.value =
          monitorData.totalPending || monitorData.total_pending;
        recentHistory.value =
          monitorData.recentHistory || monitorData.recent_history;
        stats.value.completed =
          monitorData.totalCompleted || monitorData.total_completed;
        lastUpdate.value = new Date();
        updateJobStuckStatus();
      }

      const configResponse = await apiClient.client.get("/jobs/auto-functions");

      if (configResponse.data.success) {
        let configs = configResponse.data.data.configs || [];

        // Filter out any cancelled scheduled items
        configs = configs.map((config: AutoFunctionConfig) => {
          if (
            cancelledScheduledIds.has(config.id) &&
            config.next_scheduled_execution
          ) {
            return { ...config, next_scheduled_execution: null };
          }
          return config;
        });

        autoFunctionConfigs.value = configs;

        const totalPending =
          monitorData?.totalPending || monitorData?.total_pending || 0;
        const scheduledCount =
          configs?.filter((c: any) => c.next_scheduled_execution).length || 0;
        queueCount.value = totalPending + scheduledCount;
      }
    } catch (err: any) {
      error.value = err.message || "Failed to load monitor data";
      console.error("❌ Error in refreshMonitor:", err);
    } finally {
      isLoading.value = false;
    }
  }

  async function cancelJob(jobId: string) {
    try {
      const response = await apiClient.client.post(`/jobs/cancel/${jobId}`);
      if (response.data.success) {
        refreshMonitor();
      }
    } catch (err: any) {
      error.value = `Failed to cancel job: ${err.message}`;
    }
  }

  async function forceCancel(jobId: string) {
    if (!confirm("Are you sure you want to force cancel this stuck job?"))
      return;
    try {
      const response = await apiClient.client.post(
        `/jobs/force-cancel/${jobId}`,
      );
      if (response.data.success) refreshMonitor();
      else error.value = "Failed to force cancel job";
    } catch (err: any) {
      error.value = `Failed to force cancel job: ${err.message}`;
    }
  }

  async function clearHistory() {
    if (!confirm("🗑️ Are you sure you want to delete ALL job history?")) return;
    isClearing.value = true;
    try {
      const response = await apiClient.client.delete(`/jobs/history`);
      if (response.data.success) {
        recentHistory.value = [];
        await refreshMonitor();
      }
    } catch (err: any) {
      error.value = `Failed to clear history: ${err.message}`;
    } finally {
      isClearing.value = false;
    }
  }

  async function toggleAutoFunction(name: string, enabled: boolean) {
    try {
      const config = autoFunctionConfigs.value.find((c) => c.name === name);
      if (!config) return;
      const response = await apiClient.client.put(
        `/jobs/auto-functions/${name}`,
        {
          enabled,
          interval_minutes: config.interval_minutes,
          start_time: config.start_time,
          end_time: config.end_time,
        },
      );
      if (response.data.success) refreshMonitor();
    } catch (err: any) {
      error.value = `Failed to toggle auto-function: ${err.message}`;
    }
  }

  function showConfigEditor(config: AutoFunctionConfig) {
    editingConfig.value = { ...config };
  }

  async function saveConfigEditor() {
    if (!editingConfig.value) return;
    if (
      !editingConfig.value.interval_minutes ||
      editingConfig.value.interval_minutes < 1
    ) {
      error.value = "Interval must be at least 1 minute";
      return;
    }

    try {
      const isNewFunction = !editingConfig.value.id;
      const payload = {
        name: editingConfig.value.name,
        enabled: editingConfig.value.enabled,
        interval_minutes: editingConfig.value.interval_minutes,
        start_time: editingConfig.value.start_time,
        end_time: editingConfig.value.end_time,
      };

      const response = isNewFunction
        ? await apiClient.client.post(`/jobs/auto-functions`, payload)
        : await apiClient.client.put(
            `/jobs/auto-functions/${editingConfig.value.name}`,
            payload,
          );

      if (response.data.success) {
        editingConfig.value = null;
        refreshMonitor();
      }
    } catch (err: any) {
      error.value = `Failed to save config: ${err.message}`;
    }
  }

  function showAddNewFunction() {
    editingConfig.value = {
      id: 0,
      name: "",
      enabled: true,
      interval_minutes: 30,
      start_time: "08:00",
      end_time: "22:00",
    };
  }

  function incrementTime(field: "start_time" | "end_time", minutes: number) {
    if (!editingConfig.value) return;
    const time = editingConfig.value[field];
    const [h, m] = (time || "00:00").split(":").map(Number);
    const totalMinutes = h * 60 + m + minutes;
    const newH = Math.floor(totalMinutes / 60) % 24;
    const newM = totalMinutes % 60;
    editingConfig.value[field] =
      `${String(newH).padStart(2, "0")}:${String(newM).padStart(2, "0")}`;
  }

  function decrementTime(field: "start_time" | "end_time", minutes: number) {
    if (!editingConfig.value) return;
    const time = editingConfig.value[field];
    const [h, m] = (time || "23:45").split(":").map(Number);
    let totalMinutes = h * 60 + m - minutes;
    if (totalMinutes < 0) totalMinutes = 24 * 60 + totalMinutes;
    const newH = Math.floor(totalMinutes / 60) % 24;
    const newM = totalMinutes % 60;
    editingConfig.value[field] =
      `${String(newH).padStart(2, "0")}:${String(newM).padStart(2, "0")}`;
  }

  async function deleteAutoFunction(name: string) {
    if (!confirm(`Are you sure you want to delete "${name}"?`)) return;
    try {
      const response = await apiClient.client.delete(
        `/jobs/auto-functions/${name}`,
      );
      if (response.data.success) refreshMonitor();
    } catch (err: any) {
      error.value = `Failed to delete auto-function: ${err.message}`;
    }
  }

  function formatTime(date: Date | string | undefined): string {
    if (!date) return "-";
    const d = typeof date === "string" ? new Date(date) : date;
    const jakartaFormatter = new Intl.DateTimeFormat("en-US", {
      timeZone: "Asia/Jakarta",
      hour: "2-digit",
      minute: "2-digit",
      second: "2-digit",
      hour12: true,
    });
    return jakartaFormatter.format(d);
  }

  function updateJobStuckStatus(): void {
    if (!currentJob.value) {
      isJobStuck.value = false;
      return;
    }
    const startedAt = currentJob.value.started_at;
    if (!startedAt) {
      isJobStuck.value = false;
      return;
    }
    const startTime =
      typeof startedAt === "string"
        ? new Date(startedAt).getTime()
        : (startedAt as Date).getTime();
    isJobStuck.value =
      (Date.now() - startTime) / (1000 * 60) > JOB_TIMEOUT_MINUTES;
  }

  async function cancelScheduledExecution(
    configId: number,
    _configName: string,
  ) {
    try {
      // Track this as cancelled so it won't reappear on refresh
      cancelledScheduledIds.add(configId);

      // Remove from UI immediately for instant feedback
      autoFunctionConfigs.value = autoFunctionConfigs.value.map(
        (config: AutoFunctionConfig) => {
          if (config.id === configId) {
            return { ...config, next_scheduled_execution: undefined };
          }
          return config;
        },
      );

      // Then cancel on backend
      const response = await apiClient.client.post(
        `/jobs/auto-functions/${configId}/cancel-scheduled`,
      );

      if (response.data.success) {
        console.log(`✅ Cancelled scheduled execution for config ${configId}`);
      } else {
        // If backend cancellation fails, remove from cancelled tracking
        cancelledScheduledIds.delete(configId);
        await refreshMonitor();
      }
    } catch (err: any) {
      console.error("Error cancelling scheduled execution:", err);
      // Remove from cancelled tracking on error
      cancelledScheduledIds.delete(configId);
      await refreshMonitor();
    }
  }

  return {
    // State
    isLoading,
    isClearing,
    error,
    lastUpdate,
    currentJob,
    pendingQueue,
    recentHistory,
    autoFunctionConfigs,
    editingConfig,
    isJobStuck,
    queueCount,
    stats,

    // Methods
    refreshMonitor,
    cancelJob,
    forceCancel,
    clearHistory,
    toggleAutoFunction,
    showConfigEditor,
    saveConfigEditor,
    showAddNewFunction,
    incrementTime,
    decrementTime,
    deleteAutoFunction,
    formatTime,
    cancelScheduledExecution,
  };
}
