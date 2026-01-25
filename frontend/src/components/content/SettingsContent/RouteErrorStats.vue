<template>
  <div class="error-stats-grid">
    <div class="stat-card">
      <div class="stat-icon">📊</div>
      <div class="stat-details">
        <div class="stat-value">{{ errorStats.totalRequests }}</div>
        <div class="stat-label">Total Requests</div>
      </div>
    </div>
    <div class="stat-card">
      <div class="stat-icon">❌</div>
      <div class="stat-details">
        <div
          class="stat-value"
          :style="{ color: getErrorColor(errorStats.totalErrors) }"
        >
          {{ errorStats.totalErrors }}
        </div>
        <div class="stat-label">Total Errors</div>
      </div>
    </div>
    <div class="stat-card">
      <div class="stat-icon">📈</div>
      <div class="stat-details">
        <div
          class="stat-value"
          :style="{ color: getRateColor(errorStats.errorRate) }"
        >
          {{ errorStats.errorRate }}
        </div>
        <div class="stat-label">Error Rate</div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
/**
 * Route Error Statistics Grid
 * Displays error metrics and statistics
 * Single Responsibility: Show error stats only
 */

interface ErrorStats {
  totalErrors: number;
  totalRequests: number;
  errorRate: string;
}

defineProps<{
  errorStats: ErrorStats;
}>();

const getErrorColor = (errors: number): string => {
  return errors > 0 ? "#e74c3c" : "#27ae60";
};

const getRateColor = (rate: string): string => {
  const rateNum = Number(rate.toString().replace("%", ""));
  return rateNum > 10 ? "#e74c3c" : "#27ae60";
};
</script>

<style scoped lang="css">
@import "./ScriptMonitor.styles.css";

.error-stats-grid {
  display: grid;
  grid-template-columns: repeat(3, 1fr);
  gap: 12px;
}

.stat-card {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 12px;
  padding: 20px;
  background: linear-gradient(135deg, #f8f9fa 0%, #ffffff 100%);
  border-radius: 12px;
  border: 1px solid #e0e0e0;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.03);
  transition: all 0.3s ease;
  text-align: center;
}

.stat-card:hover {
  border-color: #0066cc;
  box-shadow: 0 4px 12px rgba(0, 102, 204, 0.15);
  transform: translateY(-2px);
}

.stat-icon {
  font-size: 24px;
  height: 32px;
  display: flex;
  align-items: center;
  justify-content: center;
}

.stat-details {
  flex: 1;
  display: flex;
  flex-direction: column;
  gap: 6px;
  width: 100%;
}

.stat-value {
  font-size: 20px;
  font-weight: 700;
  color: #1a1a1a;
  font-variant-numeric: tabular-nums;
  line-height: 1.2;
}

.stat-label {
  font-size: 12px;
  color: #6b7280; /* Improved from #888 for WCAG AA contrast */
  font-weight: 500;
  text-transform: uppercase;
  letter-spacing: 0.3px;
}

@media (max-width: 768px) {
  .error-stats-grid {
    grid-template-columns: 1fr;
  }
}
</style>
