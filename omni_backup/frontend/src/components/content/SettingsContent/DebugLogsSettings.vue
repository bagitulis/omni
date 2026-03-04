<template>
  <div class="logs-section">
    <div class="logs-header">
      <h2>📋 Debug Logs</h2>
      <div class="logs-controls">
        <label for="log-filter" class="sr-only">Filter by log level</label>
        <select
          id="log-filter"
          v-model="logFilter"
          class="log-filter"
          aria-label="Filter by log level"
        >
          <option value="">All Levels</option>
          <option value="DEBUG">DEBUG</option>
          <option value="INFO">INFO</option>
          <option value="WARNING">WARNING</option>
          <option value="ERROR">ERROR</option>
          <option value="CRITICAL">CRITICAL</option>
        </select>
        <button
          @click="refreshLogs"
          class="btn btn-refresh"
          :disabled="logsLoading"
        >
          <span v-if="!logsLoading">🔄 Refresh</span>
          <span v-else>Loading...</span>
        </button>
        <button @click="clearLogs" class="btn btn-danger">🗑️ Clear Logs</button>
      </div>
    </div>

    <div class="logs-info">
      <span>📊 Total Logs: {{ logsData.total_count }}</span>
      <span>🔍 Ditampilkan: {{ filteredLogs.length }}</span>
      <span v-if="lastLogsUpdate">⏰ Updated: {{ lastLogsUpdate }}</span>
    </div>

    <div class="logs-viewer">
      <div v-if="filteredLogs.length === 0" class="logs-empty">
        Tidak ada logs untuk ditampilkan
      </div>
      <div v-else class="logs-list">
        <div
          v-for="(log, index) in filteredLogs.slice().reverse()"
          :key="index"
          class="log-entry"
          :class="'log-' + log.level.toLowerCase()"
        >
          <span class="log-time">{{ formatTime(log.timestamp) }}</span>
          <span class="log-level" :class="'level-' + log.level.toLowerCase()">
            {{ log.level }}
          </span>
          <span class="log-logger">{{ log.logger }}</span>
          <span class="log-message">{{ log.message }}</span>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from "vue";
import { formatTime } from "@/utils/helpers";
import { getAuthHeaders, getApiBaseUrl } from "@/utils/apiHeaders";

interface LogEntry {
  level: string;
  timestamp: string;
  logger: string;
  message: string;
  [key: string]: any;
}

interface LogsData {
  logs: LogEntry[];
  total_count: number;
  filtered_count: number;
}

const logFilter = ref("");
const logsLoading = ref(false);
const logsData = ref<LogsData>({ logs: [], total_count: 0, filtered_count: 0 });
const lastLogsUpdate = ref("");

const filteredLogs = computed(() => {
  if (!logFilter.value) return logsData.value.logs;
  return logsData.value.logs.filter((log) => log.level === logFilter.value);
});

const refreshLogs = async () => {
  logsLoading.value = true;
  try {
    const response = await fetch(getApiBaseUrl("/settings/debug-logs"), {
      headers: getAuthHeaders(),
    });
    const data = await response.json();
    if (data.success) {
      logsData.value = data.data;
      lastLogsUpdate.value = new Date().toLocaleTimeString();
    }
  } catch {
    logsData.value = { logs: [], total_count: 0, filtered_count: 0 };
  } finally {
    logsLoading.value = false;
  }
};

const clearLogs = async () => {
  if (!confirm("Are you sure you want to delete all logs?")) return;

  try {
    const response = await fetch(getApiBaseUrl("/settings/debug-logs/clear"), {
      method: "POST",
      headers: getAuthHeaders(),
    });
    const result = await response.json();
    if (result.success) {
      logsData.value = { logs: [], total_count: 0, filtered_count: 0 };
      alert(result.message);
    }
  } catch {
    alert("Failed to delete logs");
  }
};

onMounted(() => {
  refreshLogs();
});
</script>

<style scoped>
.logs-section {
  display: flex;
  flex-direction: column;
  gap: 16px;
}

.logs-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  gap: 16px;
  flex-wrap: wrap;
}

.logs-header h2 {
  margin: 0;
  color: #333;
}

.logs-controls {
  display: flex;
  gap: 12px;
  align-items: center;
}

.log-filter {
  padding: 8px 12px;
  border: 2px solid #e0eaf7;
  border-radius: 6px;
  font-size: 0.95em;
  background: white;
  cursor: pointer;
  color: #1e293b;
  transition: all 0.3s ease;
}

.log-filter:hover {
  border-color: #3b82f6;
  background: #f8fafc;
}

.btn {
  padding: 8px 16px;
  border: none;
  border-radius: 6px;
  cursor: pointer;
  font-size: 0.95em;
  font-weight: 500;
  transition: all 0.3s ease;
}

.btn-refresh {
  background: #3b82f6;
  color: white;
}

.btn-refresh:hover:not(:disabled) {
  background: #2563eb;
  transform: translateY(-2px);
  box-shadow: 0 4px 12px rgba(59, 130, 246, 0.4);
}

.btn-refresh:disabled {
  opacity: 0.6;
  cursor: not-allowed;
}

.btn-danger {
  background: #ef4444;
  color: white;
}

.btn-danger:hover {
  background: #dc2626;
  transform: translateY(-2px);
  box-shadow: 0 4px 12px rgba(239, 68, 68, 0.4);
}

.logs-info {
  display: flex;
  gap: 24px;
  padding: 12px;
  background: #f8f9fa;
  border-radius: 8px;
  font-size: 0.95em;
  color: #666;
}

.logs-viewer {
  border: 1px solid #e0e0e0;
  border-radius: 8px;
  max-height: 600px;
  overflow-y: auto;
  background: #fafafa;
}

.logs-empty {
  padding: 48px 24px;
  text-align: center;
  color: #6b7280;
  font-size: 1.1em;
}

.logs-list {
  display: flex;
  flex-direction: column;
  gap: 0;
}

.log-entry {
  padding: 12px;
  border-bottom: 1px solid #e0e0e0;
  display: flex;
  gap: 12px;
  font-size: 0.9em;
  align-items: flex-start;
  flex-wrap: wrap;
  background: white;
  transition: background 0.3s ease;
}

.log-entry:hover {
  background: #f5f5f5;
}

.log-time {
  color: #6b7280;
  min-width: 100px;
  font-family: monospace;
  flex-shrink: 0;
}

.log-level {
  padding: 2px 8px;
  border-radius: 4px;
  font-weight: bold;
  min-width: 70px;
  text-align: center;
  flex-shrink: 0;
}

.level-debug {
  background: #e8f5e9;
  color: #2e7d32;
}

.level-info {
  background: #e3f2fd;
  color: #1565c0;
}

.level-warning {
  background: #fff3e0;
  color: #e65100;
}

.level-error {
  background: #ffebee;
  color: #c62828;
}

.level-critical {
  background: #fce4ec;
  color: #880e4f;
}

.log-logger {
  color: #666;
  font-weight: 500;
  min-width: 150px;
  flex-shrink: 0;
}

.log-message {
  color: #333;
  flex: 1;
  word-break: break-word;
  min-width: 200px;
}
</style>
