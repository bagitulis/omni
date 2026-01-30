<template>
  <div class="logs-section">
    <div class="logs-header">
      <h3>🔐 OAuth Connection Logs</h3>
      <button @click="$emit('refresh')" class="refresh-btn" :disabled="loading">
        {{ loading ? "Loading..." : "🔄 Refresh" }}
      </button>
    </div>

    <div class="logs-table-container" v-if="logs.length > 0">
      <table class="logs-table" aria-label="OAuth connection logs">
        <thead>
          <tr>
            <th scope="col">Time</th>
            <th scope="col">Platform</th>
            <th scope="col">Event Type</th>
            <th scope="col">Status</th>
            <th scope="col">Details</th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="log in logs" :key="log.id">
            <td>{{ formatDate(log.created_at) }}</td>
            <td>
              <span :class="['platform-badge', log.platform]">
                {{ log.platform }}
              </span>
            </td>
            <td>{{ formatEventType(log.event_type) }}</td>
            <td>
              <span :class="['status-badge', log.status]">
                {{ log.status }}
              </span>
            </td>
            <td class="details-cell">
              <button
                @click="toggleDetails(log.id)"
                class="details-btn"
                v-if="log.metadata"
              >
                {{ expandedIds.has(log.id) ? "▼" : "▶" }}
              </button>
            </td>
          </tr>
          <tr v-for="log in logs" :key="`details-${log.id}`">
            <td colspan="5" v-if="expandedIds.has(log.id)" class="details-row">
              <div class="metadata">
                <pre>{{ JSON.stringify(log.metadata, null, 2) }}</pre>
              </div>
            </td>
          </tr>
        </tbody>
      </table>
    </div>

    <div class="empty-logs" v-else>
      <p>No OAuth logs yet. Logs will appear when you authorize platforms.</p>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref } from "vue";

interface OAuthLog {
  id: string;
  platform: string;
  event_type: string;
  status: string;
  created_at: string;
  processed_at?: string;
  metadata?: Record<string, any>;
}

defineProps<{
  logs: OAuthLog[];
  loading: boolean;
}>();

defineEmits<{
  (e: "refresh"): void;
}>();

const expandedIds = ref(new Set<string>());

function formatDate(dateStr: string) {
  return new Date(dateStr).toLocaleString("id-ID", {
    day: "2-digit",
    month: "short",
    hour: "2-digit",
    minute: "2-digit",
    second: "2-digit",
  });
}

function formatEventType(event: string) {
  const map: Record<string, string> = {
    callback_received: "Callback Received",
    token_exchanged: "Token Exchanged",
    token_refreshed: "Token Refreshed",
    error: "Error",
  };
  return map[event] || event;
}

function toggleDetails(id: string) {
  if (expandedIds.value.has(id)) {
    expandedIds.value.delete(id);
  } else {
    expandedIds.value.add(id);
  }
}
</script>

<style scoped>
.logs-section {
  background: hsl(var(--b2));
  border-radius: 12px;
  padding: 24px;
  border: 1px solid hsl(var(--bc) / 0.1);
}

.logs-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 16px;
}

.logs-header h3 {
  font-size: 1.1rem;
  margin: 0;
}

.refresh-btn {
  padding: 8px 16px;
  background: hsl(var(--b3));
  color: hsl(var(--bc));
  border: 1px solid hsl(var(--bc) / 0.2);
  border-radius: 8px;
  cursor: pointer;
  transition: background 0.2s;
}

.refresh-btn:hover:not(:disabled) {
  background: hsl(var(--bc) / 0.1);
}

.refresh-btn:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

.logs-table-container {
  overflow-x: auto;
}

.logs-table {
  width: 100%;
  border-collapse: collapse;
}

.logs-table th,
.logs-table td {
  padding: 12px;
  text-align: left;
  border-bottom: 1px solid hsl(var(--bc) / 0.1);
}

.logs-table th {
  font-weight: 600;
  color: hsl(var(--bc) / 0.7);
  font-size: 0.85rem;
  text-transform: uppercase;
}

.platform-badge {
  padding: 4px 8px;
  border-radius: 4px;
  font-size: 0.8rem;
  font-weight: 500;
  text-transform: capitalize;
}

.platform-badge.shopee {
  background: #ee4d2d20;
  color: #ee4d2d;
}

.platform-badge.tiktok {
  background: #00000020;
  color: #000;
}

.platform-badge.lazada {
  background: #0d47a120;
  color: #0d47a1;
}

.status-badge {
  padding: 4px 8px;
  border-radius: 4px;
  font-size: 0.8rem;
  font-weight: 500;
  text-transform: capitalize;
}

.status-badge.success {
  background: #10b98120;
  color: #059669;
}

.status-badge.failed {
  background: #ef444420;
  color: #dc2626;
}

.status-badge.received {
  background: #f59e0b20;
  color: #d97706;
}

.details-cell {
  text-align: center;
}

.details-btn {
  background: none;
  border: none;
  cursor: pointer;
  padding: 4px 8px;
  color: hsl(var(--bc) / 0.7);
  font-size: 0.9rem;
}

.details-btn:hover {
  color: hsl(var(--bc));
}

.details-row {
  background: hsl(var(--b3) / 0.3);
}

.metadata {
  padding: 12px;
  background: hsl(var(--b1));
  border-radius: 8px;
  margin: 8px 0;
  overflow-x: auto;
}

.metadata pre {
  margin: 0;
  font-size: 0.85rem;
  color: hsl(var(--bc) / 0.8);
}

.empty-logs {
  text-align: center;
  padding: 32px 24px;
  color: hsl(var(--bc) / 0.6);
}

.empty-logs p {
  margin: 0;
}
</style>
