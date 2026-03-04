<template>
  <div class="health-wrapper">
    <RouteHealthStatus :healthData="healthData" />
    <RouteErrorStats :errorStats="errorStats" />
  </div>
</template>

<script setup lang="ts">
/**
 * Route Health Card Container
 * Orchestrates health status and error statistics
 * Single Responsibility: Delegate to sub-components
 */

import RouteHealthStatus from './RouteHealthStatus.vue';
import RouteErrorStats from './RouteErrorStats.vue';

interface HealthData {
  total: number;
  healthy: number;
  warning: number;
  error: number;
  healthyPercent: number;
  status: 'good' | 'fair' | 'poor';
}

interface ErrorStats {
  totalErrors: number;
  totalRequests: number;
  errorRate: string;
}

defineProps<{
  healthData: HealthData;
  errorStats: ErrorStats;
}>();
</script>

<style scoped lang="css">
@import './ScriptMonitor.styles.css';

.health-wrapper {
  display: grid;
  grid-template-columns: 1.2fr 1fr;
  gap: 24px;
  margin-bottom: 24px;
}

@media (max-width: 1024px) {
  .health-wrapper {
    grid-template-columns: 1fr;
  }
}

@media (max-width: 768px) {
  .health-wrapper {
    grid-template-columns: 1fr;
    gap: 16px;
    margin-bottom: 16px;
  }
}
</style>

