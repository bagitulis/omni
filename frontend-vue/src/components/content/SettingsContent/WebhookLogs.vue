<template>
  <div class="logs-section">
    <div class="logs-header">
      <h3><Icon name="document" size="sm" /> Recent Webhook Logs</h3>
      <button
        @click="$emit('refresh')"
        class="refresh-btn"
        :disabled="loading"
        aria-label="Refresh logs"
      >
        <Icon v-if="!loading" name="refresh" size="sm" />
        {{ loading ? "Loading..." : "Refresh" }}
      </button>
    </div>

    <div class="logs-table-container" v-if="logs.length > 0">
      <table class="logs-table" aria-label="Webhook logs">
        <thead>
          <tr>
            <th scope="col">Time</th>
            <th scope="col">Platform</th>
            <th scope="col">Event</th>
            <th scope="col">Status</th>
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
            <td>
              <span
                class="event-type"
                :style="{
                  borderLeftColor: getEventCategoryColor(
                    log.event_type,
                    log.platform,
                  ),
                }"
              >
                {{ getEventDisplay(log.event_type, log.platform) }}
              </span>
            </td>
            <td>
              <span
                :class="[
                  'status-badge',
                  log.status === 'processed' ? 'success' : 'pending',
                ]"
              >
                {{ log.status === "processed" ? "Processed" : "Pending" }}
              </span>
            </td>
          </tr>
        </tbody>
      </table>
    </div>

    <div class="empty-logs" v-else>
      <p>No webhook logs yet. Logs will appear when platforms send events.</p>
    </div>
  </div>
</template>

<script setup lang="ts">
import { formatEventType, getCategoryColor } from "@/constants/shopeePushCodes";
import Icon from "@/components/ui/Icon.vue";

interface LogEntry {
  id: string;
  platform: string;
  event_type: string | null;
  created_at: string;
  status: string;
}

defineProps<{
  logs: LogEntry[];
  loading: boolean;
}>();

defineEmits<{
  (e: "refresh"): void;
}>();

function formatDate(dateStr: string) {
  return new Date(dateStr).toLocaleString("id-ID", {
    day: "2-digit",
    month: "short",
    hour: "2-digit",
    minute: "2-digit",
  });
}

function getEventDisplay(eventType: string | null, platform: string) {
  if (platform === "shopee") {
    return formatEventType(eventType);
  }
  return eventType || "unknown";
}

function getEventCategoryColor(
  eventType: string | null,
  platform: string,
): string {
  if (platform === "shopee" && eventType) {
    return getCategoryColor(eventType);
  }
  return "#888";
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
  background: #fee2e2;
  color: #991b1b;
}

.platform-badge.tiktok {
  background: #f3f4f6;
  color: #1f2937;
}

.platform-badge.lazada {
  background: #dbeafe;
  color: #1e40af;
}

.status-badge {
  padding: 4px 8px;
  border-radius: 4px;
  font-size: 0.8rem;
  font-weight: 500;
}

.status-badge.success {
  background: hsl(var(--su) / 0.2);
  color: hsl(var(--su));
}

.status-badge.pending {
  background: hsl(var(--wa) / 0.2);
  color: hsl(var(--wa));
}

.event-type {
  padding: 4px 8px;
  border-left: 3px solid #888;
  background: hsl(var(--b3) / 0.5);
  border-radius: 0 4px 4px 0;
  font-size: 0.85rem;
  display: inline-block;
}

.empty-logs {
  text-align: center;
  padding: 32px;
  color: hsl(var(--bc) / 0.5);
}
</style>
