<template>
  <div class="stats-container">
    <div class="stat-card">
      <div class="stat-icon">📊</div>
      <div class="stat-info">
        <div class="stat-label">Total Items</div>
        <div class="stat-value">{{ stats.total_records }}</div>
      </div>
    </div>

    <div class="stat-card">
      <div class="stat-icon">📋</div>
      <div class="stat-info">
        <div class="stat-label">Column Count</div>
        <div class="stat-value">{{ stats.total_columns }}</div>
      </div>
    </div>

    <div class="stat-card">
      <div class="stat-icon">⏰</div>
      <div class="stat-info">
        <div class="stat-label">Sinkron Terakhir</div>
        <div class="stat-value">{{ formatLastSync }}</div>
      </div>
    </div>

    <div class="stat-card" :class="`status-${syncStatus}`">
      <div class="stat-icon">{{ statusIcon }}</div>
      <div class="stat-info">
        <div class="stat-label">Status</div>
        <div class="stat-value">{{ syncStatus || "Unknown" }}</div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed } from "vue";

interface Stats {
  total_records: number;
  total_columns: number;
  last_sync?: string;
}

const props = defineProps<{
  stats: Stats;
  syncStatus: string;
}>();

const formatLastSync = computed(() => {
  if (!props.stats.last_sync) return "Never synced";
  try {
    const dateStr = props.stats.last_sync;
    // Handle different date formats
    const date = new Date(dateStr);

    // Check if date is valid
    if (isNaN(date.getTime())) {
      return "Never synced";
    }

    const now = new Date();
    const diffMs = now.getTime() - date.getTime();
    const diffMins = Math.floor(diffMs / 60000);
    const diffHours = Math.floor(diffMs / 3600000);
    const diffDays = Math.floor(diffMs / 86400000);

    if (diffMins < 1) return "Baru saja";
    if (diffMins < 60) return `${diffMins} menit lalu`;
    if (diffHours < 24) return `${diffHours} jam lalu`;
    if (diffDays < 7) return `${diffDays} hari lalu`;
    return date.toLocaleDateString("id-ID");
  } catch {
    return "Belum pernah";
  }
});

const statusIcon = computed(() => {
  const status = props.syncStatus?.toLowerCase() || "unknown";
  switch (status) {
    case "success":
      return "✅";
    case "partial":
      return "⚠️";
    case "error":
      return "❌";
    case "syncing":
      return "🔄";
    default:
      return "⏳";
  }
});
</script>

<style scoped>
@import "./InventoryContent.styles.css";

.stats-container {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(150px, 1fr));
  gap: 8px;
  margin-bottom: 10px;
}

.stat-card {
  background-color: white;
  border: 1px solid #e0e0e0;
  border-radius: 8px;
  padding: 12px;
  display: flex;
  align-items: center;
  gap: 8px;
  transition: all 0.2s;
}

.stat-card:hover {
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.1);
}

.stat-icon {
  font-size: 20px;
  width: 34px;
  height: 34px;
  display: flex;
  align-items: center;
  justify-content: center;
  background-color: #f5f5f5;
  border-radius: 6px;
}

.stat-info {
  flex: 1;
}

.stat-label {
  font-size: 11px;
  color: #6b7280;
  font-weight: 500;
  margin-bottom: 4px;
}

.stat-value {
  font-size: 15px;
  font-weight: 700;
  color: #333;
}

.stat-card.status-success {
  border-left: 4px solid #27ae60;
}

.stat-card.status-partial {
  border-left: 4px solid #f39c12;
}

.stat-card.status-error {
  border-left: 4px solid #e74c3c;
}

.stat-card.status-syncing {
  border-left: 4px solid #3498db;
}

@media (max-width: 768px) {
  .stats-container {
    grid-template-columns: repeat(2, 1fr);
  }

  .stat-value {
    font-size: 14px;
  }
}
</style>
