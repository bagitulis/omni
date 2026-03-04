<template>
  <div class="perf-table-wrapper">
    <table class="perf-table">
      <thead>
        <tr>
          <th class="col-route">Route</th>
          <th class="col-status">Status</th>
          <th class="col-avg">Avg</th>
          <th class="col-last">Last</th>
          <th class="col-req">Req</th>
          <th class="col-rate">Rate</th>
          <th class="col-err">Err</th>
        </tr>
      </thead>
      <tbody>
        <tr v-if="performanceData.length === 0" class="empty-row">
          <td colspan="7" class="empty-message">
            <span class="empty-icon">📭</span>
            <span>No data</span>
          </td>
        </tr>
        <RoutePerformanceRow 
          v-for="route in performanceData" 
          :key="route.routeName"
          :route="route"
        />
      </tbody>
    </table>
  </div>
</template>

<script setup lang="ts">
import RoutePerformanceRow from './RoutePerformanceRow.vue';

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
  performanceData: PerformanceRoute[];
}>();
</script>

<style scoped lang="css">
.perf-table-wrapper {
  width: 100%;
  overflow-x: auto;
}

.perf-table {
  width: 100%;
  border-collapse: collapse;
  font-size: 11px;
  font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif;
}

.perf-table thead {
  background: #f8f9fa;
  border-bottom: 1px solid #0066cc;
}

.perf-table th {
  padding: 7px 5px;
  text-align: left;
  font-weight: 700;
  color: #333;
  font-size: 9px;
  text-transform: uppercase;
  letter-spacing: 0.2px;
  white-space: nowrap;
  border: none;
}

/* Column widths - compact */
.col-route { width: 25%; min-width: 100px; }
.col-status { width: 12%; min-width: 60px; }
.col-avg { width: 10%; min-width: 55px; }
.col-last { width: 10%; min-width: 55px; }
.col-req { width: 8%; min-width: 45px; }
.col-rate { width: 10%; min-width: 55px; }
.col-err { width: 8%; min-width: 45px; }

.perf-table tbody tr {
  border-bottom: 1px solid #e8e8e8;
  transition: background-color 0.15s ease;
}

.perf-table tbody tr:hover {
  background-color: #fafbfc;
}

.perf-table td {
  padding: 7px 5px;
  vertical-align: middle;
  text-align: left;
  border: none;
}

.empty-row {
  background: #ffffff !important;
}

.empty-row td {
  text-align: center;
  padding: 20px 5px !important;
}

.empty-message {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 4px;
  color: #6b7280;
  font-size: 11px;
}

.empty-icon {
  font-size: 20px;
}

@media (max-width: 1024px) {
  .perf-table th,
  .perf-table td {
    padding: 6px 4px;
    font-size: 10px;
  }
  
  .col-route { width: 24%; min-width: 90px; }
  .col-status { width: 12%; min-width: 55px; }
  .col-avg { width: 10%; min-width: 50px; }
  .col-last { width: 10%; min-width: 50px; }
  .col-req { width: 8%; min-width: 40px; }
  .col-rate { width: 10%; min-width: 50px; }
  .col-err { width: 8%; min-width: 40px; }
}

@media (max-width: 768px) {
  .perf-table th,
  .perf-table td {
    padding: 5px 3px;
    font-size: 9px;
  }
  
  .col-route { width: 22%; min-width: 75px; }
  .col-status { width: 12%; min-width: 45px; }
  .col-avg { width: 10%; min-width: 40px; }
  .col-last { width: 10%; min-width: 40px; }
  .col-req { width: 8%; min-width: 35px; }
  .col-rate { width: 10%; min-width: 40px; }
  .col-err { width: 8%; min-width: 35px; }
}
</style>
