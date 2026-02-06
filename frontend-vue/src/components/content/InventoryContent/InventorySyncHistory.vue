<template>
  <div v-if="show" class="sync-history">
    <div class="history-header">
      <h4>📜 Sync History</h4>
      <button
        @click="$emit('close-history')"
        class="btn-close"
        type="button"
        aria-label="Close history"
      >
        ✕
      </button>
    </div>

    <div v-if="syncHistory.length === 0" class="empty-history">
      <p>No sync history available</p>
    </div>

    <div v-else class="history-list">
      <div
        v-for="(entry, index) in syncHistory"
        :key="index"
        class="history-item"
        :class="`status-${entry.status}`"
      >
        <div class="history-icon">
          {{ getStatusIcon(entry.status) }}
        </div>
        <div class="history-content">
          <div class="history-title">
            <span
              class="direction-badge"
              :class="`direction-${entry.direction}`"
            >
              {{ entry.direction === "from" ? "📥" : "📤" }}
              {{ entry.direction === "from" ? "From" : "To" }}
              {{ entry.source || "Sheet" }}
            </span>
            <span class="status-badge" :class="`badge-${entry.status}`">
              {{ capitalizeStatus(entry.status) }}
            </span>
          </div>
          <div class="history-timestamp">
            {{ formatTime(entry.timestamp) }}
          </div>
          <div v-if="entry.message" class="history-message">
            {{ entry.message }}
          </div>
          <div v-if="entry.details" class="history-details">
            <small>{{ entry.details }}</small>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
interface SyncHistoryEntry {
  direction: "from" | "to";
  source: string;
  status: "success" | "error" | "pending";
  timestamp: string;
  message?: string;
  details?: string;
}

defineProps<{
  syncHistory: SyncHistoryEntry[];
  show: boolean;
}>();

defineEmits<{
  "close-history": [];
}>();

const getStatusIcon = (status: string): string => {
  switch (status) {
    case "success":
      return "✅";
    case "error":
      return "❌";
    case "pending":
      return "⏳";
    default:
      return "❓";
  }
};

const capitalizeStatus = (status: string): string => {
  return status.charAt(0).toUpperCase() + status.slice(1);
};

const formatTime = (timestamp: string): string => {
  try {
    const date = new Date(timestamp);
    return date.toLocaleString("id-ID", {
      year: "numeric",
      month: "2-digit",
      day: "2-digit",
      hour: "2-digit",
      minute: "2-digit",
      second: "2-digit",
    });
  } catch {
    return timestamp;
  }
};
</script>

<style scoped>
@import "./InventoryContent.styles.css";

.sync-history {
  background-color: white;
  border: 1px solid #e0e0e0;
  border-radius: 8px;
  margin-top: 15px;
  overflow: hidden;
  max-height: 400px;
  display: flex;
  flex-direction: column;
}

.history-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 12px 15px;
  background-color: #f9f9f9;
  border-bottom: 1px solid #e0e0e0;
}

.history-header h4 {
  margin: 0;
  font-size: 13px;
  font-weight: 600;
  color: #333;
}

.btn-close {
  background: none;
  border: none;
  font-size: 16px;
  cursor: pointer;
  color: #6b7280;
  transition: color 0.2s;
}

.btn-close:hover {
  color: #333;
}

.empty-history {
  padding: 20px;
  text-align: center;
  color: #6b7280;
  font-size: 13px;
}

.history-list {
  overflow-y: auto;
  flex: 1;
}

.history-item {
  display: flex;
  gap: 12px;
  padding: 12px 15px;
  border-bottom: 1px solid #f0f0f0;
  transition: background-color 0.2s;
}

.history-item:hover {
  background-color: #fafafa;
}

.history-item.status-success {
  border-left: 3px solid #27ae60;
}

.history-item.status-error {
  border-left: 3px solid #e74c3c;
}

.history-item.status-pending {
  border-left: 3px solid #f39c12;
}

.history-icon {
  font-size: 18px;
  flex-shrink: 0;
}

.history-content {
  flex: 1;
  min-width: 0;
}

.history-title {
  display: flex;
  gap: 8px;
  align-items: center;
  margin-bottom: 4px;
  flex-wrap: wrap;
}

.direction-badge {
  font-size: 11px;
  font-weight: 600;
  color: #555;
  background-color: #f0f0f0;
  padding: 2px 6px;
  border-radius: 3px;
}

.direction-from {
  background-color: #e8f4f8;
  color: #0099cc;
}

.direction-to {
  background-color: #e8f8e8;
  color: #00aa00;
}

.status-badge {
  font-size: 11px;
  font-weight: 600;
  padding: 2px 6px;
  border-radius: 3px;
}

.badge-success {
  background-color: #e8f8e8;
  color: #27ae60;
}

.badge-error {
  background-color: #ffe8e8;
  color: #e74c3c;
}

.badge-pending {
  background-color: #fff8e8;
  color: #f39c12;
}

.history-timestamp {
  font-size: 11px;
  color: #6b7280;
  margin-bottom: 4px;
}

.history-message {
  font-size: 12px;
  color: #666;
  margin-bottom: 4px;
}

.history-details {
  font-size: 11px;
  color: #6b7280;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

@media (max-width: 768px) {
  .sync-history {
    max-height: 300px;
  }

  .history-item {
    gap: 8px;
    padding: 10px 12px;
  }
}
</style>
