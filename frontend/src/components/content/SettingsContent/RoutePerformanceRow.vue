<template>
  <tr 
    :class="['route-row', route.status]"
  >
    <td class="route-name">
      <span class="route-badge">{{ route.routeName }}</span>
    </td>
    <td class="status-cell">
      <span :class="['status-badge', route.status]">
        {{ statusEmoji[route.status] }} {{ route.status.toUpperCase() }}
      </span>
    </td>
    <td :class="['response-time', route.responseTimeStatus]">
      <span class="time-value">{{ route.responseTimeFormatted }}</span>
    </td>
    <td class="response-time">
      <span class="time-value">{{ route.lastTime }}ms</span>
    </td>
    <td class="request-count">
      <span class="count-badge">{{ route.requestCount }}</span>
    </td>
    <td class="success-rate">
      <span :class="['rate-badge', getRateClass(route.successRate)]">
        {{ route.successRate }}%
      </span>
    </td>
    <td class="error-count">
      <span :class="['error-badge', route.errorCount > 0 ? 'has-errors' : 'no-errors']">
        {{ route.errorCount }}
      </span>
    </td>
  </tr>
</template>

<script setup lang="ts">
/**
 * Route Performance Table Row
 * Renders single route performance row
 * Single Responsibility: Render one route row only
 */

interface PerformanceRoute {
  routeName: string;
  status: 'healthy' | 'warning' | 'error';
  averageTime: number;
  lastTime: number;
  requestCount: number;
  successRate: number;
  errorCount: number;
  responseTimeStatus: 'fast' | 'normal' | 'slow' | 'very-slow';
  responseTimeFormatted: string;
}

defineProps<{
  route: PerformanceRoute;
}>();

const statusEmoji: Record<string, string> = {
  'healthy': '✅',
  'warning': '⚠️',
  'error': '❌',
};

const getRateClass = (rate: number): string => {
  if (rate >= 95) return 'excellent';
  if (rate >= 80) return 'good';
  return 'poor';
};
</script>

<style scoped lang="css">
/* Route Performance Row - Clean CSS from scratch */

.route-row {
  border-left: 4px solid transparent;
  transition: background-color 0.2s ease;
}

.route-row:hover {
  background-color: #f0f5ff;
  border-left-color: #0066cc;
}

.route-row.healthy {
  border-left-color: #27ae60;
}

.route-row.healthy:hover {
  background-color: #f0fdf4;
}

.route-row.warning {
  border-left-color: #f39c12;
}

.route-row.warning:hover {
  background-color: #fffbf0;
}

.route-row.error {
  border-left-color: #e74c3c;
}

.route-row.error:hover {
  background-color: #fff5f5;
}

/* Route Name */
.route-name {
  font-weight: 600;
  color: #333;
}

.route-badge {
  display: inline-block;
  padding: 4px 8px;
  background: #e8f0ff;
  color: #0066cc;
  border-radius: 4px;
  font-size: 12px;
  font-weight: 700;
  transition: background-color 0.2s ease;
}

.route-row:hover .route-badge {
  background: #d0e1ff;
}

/* Status Cell */
.status-cell {
  text-align: left;
}

.status-badge {
  display: inline-block;
  padding: 4px 8px;
  border-radius: 4px;
  font-size: 11px;
  font-weight: 700;
  white-space: nowrap;
  transition: all 0.2s ease;
}

.status-badge.healthy {
  background: #d4edda;
  color: #155724;
  border: 1px solid #c3e6cb;
}

.status-badge.warning {
  background: #fff3cd;
  color: #856404;
  border: 1px solid #ffeaa7;
}

.status-badge.error {
  background: #f8d7da;
  color: #721c24;
  border: 1px solid #f5c6cb;
}

/* Response Time */
.response-time {
  text-align: left;
  font-weight: 500;
}

.time-value {
  display: inline-block;
  padding: 4px 8px;
  border-radius: 4px;
  font-size: 12px;
  font-weight: 600;
  font-variant-numeric: tabular-nums;
}

.response-time.fast .time-value {
  background: #d4edda;
  color: #155724;
}

.response-time.normal .time-value {
  background: #e8f4f8;
  color: #0c5460;
}

.response-time.slow .time-value {
  background: #fff3cd;
  color: #856404;
}

.response-time.very-slow .time-value {
  background: #f8d7da;
  color: #721c24;
}

/* Request Count */
.request-count {
  text-align: left;
}

.count-badge {
  display: inline-block;
  min-width: 24px;
  padding: 4px 8px;
  background: #e8f0ff;
  color: #0066cc;
  border-radius: 4px;
  font-weight: 700;
  font-size: 12px;
}

/* Success Rate */
.success-rate {
  text-align: left;
}

.rate-badge {
  display: inline-block;
  min-width: 40px;
  padding: 4px 8px;
  border-radius: 4px;
  font-weight: 700;
  font-size: 12px;
  font-variant-numeric: tabular-nums;
  transition: all 0.2s ease;
}

.rate-badge.excellent {
  background: #d4edda;
  color: #155724;
  border: 1px solid #c3e6cb;
}

.rate-badge.good {
  background: #fff3cd;
  color: #856404;
  border: 1px solid #ffeaa7;
}

.rate-badge.poor {
  background: #f8d7da;
  color: #721c24;
  border: 1px solid #f5c6cb;
}

/* Error Count */
.error-count {
  text-align: left;
}

.error-badge {
  display: inline-block;
  min-width: 24px;
  padding: 4px 8px;
  border-radius: 4px;
  font-weight: 700;
  font-size: 12px;
  font-variant-numeric: tabular-nums;
}

.error-badge.has-errors {
  background: #f8d7da;
  color: #721c24;
  border: 1px solid #f5c6cb;
}

.error-badge.no-errors {
  background: #d4edda;
  color: #155724;
  border: 1px solid #c3e6cb;
}

/* Responsive */
@media (max-width: 1024px) {
  .route-badge {
    padding: 3px 6px;
    font-size: 11px;
  }
  
  .status-badge,
  .time-value,
  .count-badge,
  .rate-badge,
  .error-badge {
    padding: 3px 6px;
    font-size: 11px;
  }
}

@media (max-width: 768px) {
  .route-badge {
    padding: 2px 4px;
    font-size: 10px;
  }
  
  .status-badge,
  .time-value,
  .count-badge,
  .rate-badge,
  .error-badge {
    padding: 2px 4px;
    font-size: 10px;
  }
}
</style>
