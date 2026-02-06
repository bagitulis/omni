<template>
  <div class="route-monitoring">
    <!-- Header Section -->
    <div class="monitoring-header">
      <div class="header-info">
        <h3>📊 Route Performance & Health</h3>
        <p class="header-subtitle">
          Real-time monitoring of route performance metrics
        </p>
      </div>
      <div class="header-actions">
        <button
          @click="refreshRouteMetrics"
          class="btn btn-primary"
          :disabled="isRefreshing"
        >
          <span class="btn-icon" :class="{ 'icon-spin': isRefreshing }"
            >🔄</span
          >
          <span class="btn-text">{{
            isRefreshing ? "Refreshing..." : "Refresh"
          }}</span>
        </button>
        <button @click="exportRouteData" class="btn btn-secondary">
          <span class="btn-icon">📥</span>
          <span class="btn-text">Export</span>
        </button>
        <button @click="clearRouteMetrics" class="btn btn-danger">
          <span class="btn-icon">🗑️</span>
          <span class="btn-text">Clear</span>
        </button>
      </div>
    </div>

    <!-- Health Summary Section -->
    <div class="section">
      <RouteHealthCard
        :healthData="routeHealthData"
        :errorStats="routeErrorStats"
      />
    </div>

    <!-- Performance Table Section -->
    <div class="performance-section">
      <div class="perf-header">
        <div class="perf-title">📈 Performance Metrics</div>
      </div>
      <RoutePerformanceTable :performanceData="routePerformanceData" />
    </div>

    <!-- Queue Status Section -->
    <RouteQueueSection :queueInfo="routeQueueInfo" />

    <!-- Slow Routes Alert Section -->
    <RouteSlowRoutesSection :slowRoutes="routeSlowRoutes" />

    <!-- Route States Section -->
    <RouteStatesSection :routeStates="routeStateData" />
  </div>
</template>

<script setup lang="ts">
import { ref, computed } from "vue";
import { useScriptMonitorDisplay } from "@/composables/useScriptMonitorDisplay";
import RouteHealthCard from "./RouteHealthCard.vue";
import RoutePerformanceTable from "./RoutePerformanceTable.vue";
import RouteQueueSection from "./RouteQueueSection.vue";
import RouteSlowRoutesSection from "./RouteSlowRoutesSection.vue";
import RouteStatesSection from "./RouteStatesSection.vue";

/**
 * Route Monitoring Tab
 * Single Responsibility: Manage tab layout and delegate to sub-components
 */

const isRefreshing = ref(false);

const {
  getRoutePerformanceSummary,
  getRouteStateSummary,
  getQueueStatus,
  getErrorStatistics,
  getHealthSummary,
  exportMonitoringData,
} = useScriptMonitorDisplay();

// Computed properties
const routePerformanceData = computed(() => getRoutePerformanceSummary.value);
const routeStateData = computed(() => getRouteStateSummary.value);
const routeQueueInfo = computed(() => getQueueStatus.value);
const routeErrorStats = computed(() => getErrorStatistics.value);
const routeHealthData = computed(() => getHealthSummary.value);
const routeSlowRoutes = computed(() => routeErrorStats.value.slowRoutes);

// Methods
const refreshRouteMetrics = async () => {
  isRefreshing.value = true;
  try {
    await new Promise((resolve) => setTimeout(resolve, 500)); // Simulate refresh
  } finally {
    isRefreshing.value = false;
  }
};

const exportRouteData = () => {
  const data = exportMonitoringData();
  const blob = new Blob([data], { type: "application/json" });
  const url = window.URL.createObjectURL(blob);
  const a = document.createElement("a");
  a.href = url;
  a.download = `route-monitoring-${new Date().toISOString()}.json`;
  a.click();
  window.URL.revokeObjectURL(url);
};

const clearRouteMetrics = () => {
  if (
    confirm(
      "Are you sure you want to clear all route metrics? This action cannot be undone."
    )
  ) {
    // Clear route metrics
  }
};
</script>

<style scoped lang="css">
@import "./ScriptMonitor.styles.css";

.route-monitoring {
  display: flex;
  flex-direction: column;
  gap: 24px;
  padding: 0;
}

@keyframes slideInDown {
  from {
    opacity: 0;
    transform: translateY(-12px);
  }
  to {
    opacity: 1;
    transform: translateY(0);
  }
}

/* Header Section */
.monitoring-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 24px;
  padding: 24px;
  background: #ffffff;
  border-radius: 12px;
  border: 1px solid #e0e0e0;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.05);
}

@keyframes fadeInScale {
  from {
    opacity: 0;
    transform: scale(0.95);
  }
  to {
    opacity: 1;
    transform: scale(1);
  }
}

.header-info h3 {
  margin: 0 0 8px 0;
  font-size: 20px;
  font-weight: 700;
  color: #1a1a1a;
  letter-spacing: -0.3px;
}

.header-subtitle {
  margin: 0;
  font-size: 14px;
  color: #6b7280;
  line-height: 1.4;
}

.header-actions {
  display: flex;
  align-items: center;
  gap: 12px;
  flex-wrap: wrap;
  justify-content: flex-end;
}

.btn {
  display: inline-flex;
  align-items: center;
  gap: 8px;
  padding: 8px 16px;
  border: none;
  border-radius: 6px;
  font-size: 14px;
  font-weight: 600;
  cursor: pointer;
  transition: all 0.2s ease;
  white-space: nowrap;
}

.btn:disabled {
  opacity: 0.6;
  cursor: not-allowed;
}

.btn-primary {
  background: #0066cc;
  color: white;
}

.btn-primary:hover:not(:disabled) {
  background: #0052a3;
  box-shadow: 0 2px 8px rgba(0, 102, 204, 0.2);
}

.btn-secondary {
  background: #f8f9fa;
  color: #1a1a1a;
  border: 1px solid #e0e0e0;
}

.btn-secondary:hover {
  background: #f0f1f3;
  border-color: #0066cc;
}

.btn-danger {
  background: #e74c3c;
  color: white;
}

.btn-danger:hover {
  background: #c0392b;
  box-shadow: 0 2px 8px rgba(231, 76, 60, 0.2);
  transform: translateY(-1px);
}

.btn-icon {
  font-size: 16px;
  display: inline-flex;
  align-items: center;
}

.btn-icon.icon-spin {
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

.btn-text {
  font-size: 14px;
}

/* Performance Section - Compact & Centered */
.performance-section {
  display: flex;
  flex-direction: column;
  gap: 0;
  background: #ffffff;
  border: 1px solid #e0e0e0;
  border-radius: 10px;
  overflow: hidden;
}

.perf-header {
  padding: 12px 16px;
  background: linear-gradient(90deg, #f8f9fa 0%, #ffffff 100%);
  border-bottom: 2px solid #0066cc;
}

.perf-title {
  font-size: 13px;
  font-weight: 700;
  color: #1a1a1a;
  margin: 0;
  letter-spacing: 0.3px;
}

/* Section Styles */
.section {
  display: flex;
  flex-direction: column;
  gap: 8px;
  padding: 12px;
  background: #ffffff;
  border: 1px solid #e0e0e0;
  border-radius: 10px;
}

@keyframes slideInUp {
  from {
    opacity: 0;
    transform: translateY(12px);
  }
  to {
    opacity: 1;
    transform: translateY(0);
  }
}

.section-title {
  font-size: 14px;
  font-weight: 700;
  color: #1a1a1a;
  margin: 0;
}

/* Responsive Design */
@media (max-width: 768px) {
  .route-monitoring {
    gap: 16px;
  }

  .monitoring-header {
    flex-direction: column;
    gap: 16px;
    align-items: flex-start;
  }

  .header-actions {
    width: 100%;
    justify-content: flex-start;
  }

  .section {
    padding: 16px;
  }

  .perf-header {
    padding: 10px 12px;
  }

  .perf-title {
    font-size: 12px;
  }
}
</style>
