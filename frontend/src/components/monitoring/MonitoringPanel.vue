<template>
  <div class="monitoring-panel">
    <!-- Header -->
    <div class="panel-header">
      <h3>📊 Monitoring Metrics</h3>
      <div class="header-actions">
        <button
          @click="toggleAutoRefresh"
          :class="['btn-refresh', { active: autoRefresh }]"
        >
          <span :class="{ spin: isLoading }">🔄</span>
          {{ autoRefresh ? "Auto" : "Manual" }}
        </button>
        <button @click="fetchMetrics" :disabled="isLoading" class="btn-refresh">
          Refresh
        </button>
      </div>
    </div>

    <!-- Error Display -->
    <div v-if="error" class="error-banner">⚠️ {{ error }}</div>

    <!-- Loading State -->
    <div v-if="isLoading && !metrics" class="loading">
      ⏳ Loading metrics...
    </div>

    <!-- Metrics Display -->
    <div v-else-if="metrics" class="metrics-grid">
      <!-- Cache Stats -->
      <div class="metric-card">
        <div class="card-icon">💾</div>
        <div class="card-content">
          <div class="card-title">Cache</div>
          <div class="card-value">
            {{ metrics.metrics.cache.hitRate.toFixed(1) }}%
          </div>
          <div class="card-subtitle">
            {{ metrics.metrics.cache.hits }} hits /
            {{ metrics.metrics.cache.misses }} misses
          </div>
        </div>
      </div>

      <!-- Queue Stats -->
      <div class="metric-card">
        <div class="card-icon">📦</div>
        <div class="card-content">
          <div class="card-title">Queue</div>
          <div class="card-value">{{ metrics.queue.queued }}</div>
          <div class="card-subtitle">
            {{ metrics.metrics.queue.completed }} completed /
            {{ metrics.metrics.queue.failed }} failed
          </div>
        </div>
      </div>

      <!-- Alerts Stats -->
      <div
        class="metric-card"
        :class="{ 'alert-card': metrics.alerts.bySeverity.critical > 0 }"
      >
        <div class="card-icon">🚨</div>
        <div class="card-content">
          <div class="card-title">Alerts</div>
          <div class="card-value">{{ metrics.alerts.total }}</div>
          <div class="card-subtitle">
            {{ metrics.alerts.bySeverity.critical }} critical /
            {{ metrics.alerts.bySeverity.warning }} warning
          </div>
        </div>
      </div>

      <!-- Performance Stats -->
      <div class="metric-card">
        <div class="card-icon">⚡</div>
        <div class="card-content">
          <div class="card-title">Avg Time</div>
          <div class="card-value">
            {{ metrics.metrics.queue.avgProcessingTime.toFixed(0) }}ms
          </div>
          <div class="card-subtitle">
            {{ metrics.metrics.queue.enqueued }} total jobs
          </div>
        </div>
      </div>
    </div>

    <!-- Actions -->
    <div v-if="metrics" class="panel-actions">
      <button @click="exportMetrics" class="btn-action">📥 Export</button>
      <button @click="clearMetrics" class="btn-action btn-danger">
        🗑️ Clear
      </button>
    </div>
  </div>
</template>

<script setup lang="ts">
import { useMonitoring } from "@/composables/useMonitoring";

const {
  metrics,
  isLoading,
  error,
  autoRefresh,
  fetchMetrics,
  clearMetrics,
  startAutoRefresh,
  stopAutoRefresh,
  exportMetrics,
} = useMonitoring();

const toggleAutoRefresh = () => {
  if (autoRefresh.value) {
    stopAutoRefresh();
  } else {
    startAutoRefresh(5000); // Refresh every 5 seconds
  }
};
</script>

<style scoped>
.monitoring-panel {
  padding: 20px;
  background: #ffffff;
  border-radius: 8px;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.1);
}

.panel-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 20px;
}

.panel-header h3 {
  margin: 0;
  font-size: 18px;
  color: #333;
}

.header-actions {
  display: flex;
  gap: 10px;
}

.btn-refresh {
  padding: 6px 12px;
  border: 1px solid #ddd;
  border-radius: 4px;
  background: white;
  cursor: pointer;
  font-size: 14px;
  transition: all 0.2s;
}

.btn-refresh:hover {
  background: #f5f5f5;
}

.btn-refresh.active {
  background: #4caf50;
  color: white;
  border-color: #4caf50;
}

.btn-refresh:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

.spin {
  display: inline-block;
  animation: spin 1s linear infinite;
}

@keyframes spin {
  from {
    transform: rotate(0deg);
  }
  to {
    transform: rotate(360deg);
  }
}

.error-banner {
  padding: 12px;
  background: #ffebee;
  border-left: 4px solid #f44336;
  border-radius: 4px;
  color: #c62828;
  margin-bottom: 20px;
}

.loading {
  text-align: center;
  padding: 40px;
  color: #666;
}

.metrics-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(250px, 1fr));
  gap: 16px;
  margin-bottom: 20px;
}

.metric-card {
  display: flex;
  align-items: center;
  padding: 16px;
  background: #f9f9f9;
  border-radius: 8px;
  border: 1px solid #e0e0e0;
  transition: all 0.2s;
}

.metric-card:hover {
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.1);
  transform: translateY(-2px);
}

.metric-card.alert-card {
  background: #fff3e0;
  border-color: #ff9800;
}

.card-icon {
  font-size: 32px;
  margin-right: 16px;
}

.card-content {
  flex: 1;
}

.card-title {
  font-size: 12px;
  color: #666;
  text-transform: uppercase;
  font-weight: 600;
  margin-bottom: 4px;
}

.card-value {
  font-size: 24px;
  font-weight: 700;
  color: #333;
  margin-bottom: 4px;
}

.card-subtitle {
  font-size: 12px;
  color: #6b7280;
}

.panel-actions {
  display: flex;
  gap: 10px;
  justify-content: flex-end;
  padding-top: 16px;
  border-top: 1px solid #e0e0e0;
}

.btn-action {
  padding: 8px 16px;
  border: 1px solid #ddd;
  border-radius: 4px;
  background: white;
  cursor: pointer;
  font-size: 14px;
  transition: all 0.2s;
}

.btn-action:hover {
  background: #f5f5f5;
}

.btn-action.btn-danger {
  color: #f44336;
  border-color: #f44336;
}

.btn-action.btn-danger:hover {
  background: #ffebee;
}
</style>
