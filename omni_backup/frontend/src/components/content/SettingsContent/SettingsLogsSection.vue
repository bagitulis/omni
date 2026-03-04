<template>
  <div class="logs-section">
    <div class="logs-header">
      <h2>📋 Debug Logs</h2>
      <div class="logs-controls">
        <select v-model="logFilter" class="log-filter">
          <option value="">All Levels</option>
          <option value="DEBUG">DEBUG</option>
          <option value="INFO">INFO</option>
          <option value="WARNING">WARNING</option>
          <option value="ERROR">ERROR</option>
          <option value="CRITICAL">CRITICAL</option>
        </select>
        <button
          class="btn btn-refresh"
          @click="refreshLogs"
          :disabled="logsLoading"
        >
          {{ logsLoading ? "Loading..." : "Refresh" }}
        </button>
        <button
          class="btn btn-danger"
          @click="clearLogs"
          :disabled="logsLoading"
        >
          Clear Logs
        </button>
      </div>
    </div>

    <div class="logs-info">
      <div>📊 Total: {{ logsData.total_count }}</div>
      <div>🔍 Filtered: {{ logsData.filtered_count }}</div>
      <div>⏰ Updated: {{ lastLogsUpdate }}</div>
    </div>

    <div class="logs-viewer">
      <div v-if="filteredLogs.length === 0" class="logs-empty">
        No logs found
      </div>
      <div v-else class="logs-list">
        <div
          v-for="(log, index) in filteredLogs"
          :key="index"
          class="log-entry"
        >
          <span class="log-time">{{ log.timestamp }}</span>
          <span class="log-level" :class="`level-${log.level.toLowerCase()}`">
            {{ log.level }}
          </span>
          <span class="log-logger">{{ log.logger }}</span>
          <span class="log-message">{{ log.message }}</span>
        </div>
      </div>
    </div>
  </div>
</template>

<script lang="ts">
import { defineComponent } from "vue";
import { useSettingsLogs } from "./composables/useSettingsLogs";

export default defineComponent({
  name: "SettingsLogsSection",
  setup() {
    const {
      logsData,
      logsLoading,
      logFilter,
      lastLogsUpdate,
      filteredLogs,
      refreshLogs,
      clearLogs,
    } = useSettingsLogs();

    return {
      logsData,
      logsLoading,
      logFilter,
      lastLogsUpdate,
      filteredLogs,
      refreshLogs,
      clearLogs,
    };
  },
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
  font-size: 0.9em;
  background: white;
  cursor: pointer;
  color: #1e293b;
}

.btn {
  padding: 8px 16px;
  border: none;
  border-radius: 6px;
  cursor: pointer;
  font-size: 0.9em;
  font-weight: 500;
  transition: all 0.3s ease;
}

.btn-refresh {
  background: #3b82f6;
  color: white;
}

.btn-refresh:hover:not(:disabled) {
  background: #2563eb;
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
}

.logs-info {
  display: flex;
  gap: 24px;
  padding: 12px;
  background: #f8f9fa;
  border-radius: 8px;
  font-size: 0.9em;
  color: #666;
}

.logs-viewer {
  border: 1px solid #e0e0e0;
  border-radius: 8px;
  max-height: 400px;
  overflow-y: auto;
  background: #fafafa;
}

.logs-empty {
  padding: 40px 24px;
  text-align: center;
  color: #6b7280;
  font-size: 1em;
}

.logs-list {
  display: flex;
  flex-direction: column;
}

.log-entry {
  padding: 12px;
  border-bottom: 1px solid #e0e0e0;
  display: flex;
  gap: 12px;
  font-size: 0.85em;
  align-items: flex-start;
  flex-wrap: wrap;
  background: white;
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
}

@media (max-width: 768px) {
  .logs-controls {
    flex-direction: column;
    width: 100%;
  }

  .logs-controls select,
  .logs-controls button {
    width: 100%;
  }
}
</style>
