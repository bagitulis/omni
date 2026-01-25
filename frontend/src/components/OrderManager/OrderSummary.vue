<template>
  <div class="order-summary">
    <div class="summary-stats">
      <div class="stat-item">
        <span class="stat-label">Total Items:</span>
        <span class="stat-value">{{ filteredOrders.length }}</span>
      </div>
      <div class="stat-item">
        <span class="stat-label">Total Orders:</span>
        <span class="stat-value">{{ uniqueOrders.length }}</span>
      </div>
      <div class="stat-item">
        <span class="stat-label">Platforms:</span>
        <span class="stat-value">{{ uniquePlatforms.length }}</span>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
interface Order {
  order_no: string;
  platform: string;
  [key: string]: any;
}

defineProps<{
  filteredOrders: Order[];
  uniqueOrders: string[];
  uniquePlatforms: string[];
}>();
</script>

<style scoped>
@import './OrderManager.styles.css';

/* Summary */
.order-summary {
  padding: 20px 25px;
  background: #f8f9fa;
  border-top: 1px solid #e2e8f0;
}

.summary-stats {
  display: flex;
  gap: 30px;
  flex-wrap: wrap;
}

.stat-item {
  display: flex;
  align-items: center;
  gap: 12px;
}

.stat-label {
  color: #7f8c8d;
  font-weight: 600;
  font-size: 14px;
}

.stat-value {
  color: #2c3e50;
  font-weight: 700;
  font-size: 20px;
}

/* Responsive */
@media (max-width: 768px) {
  .order-summary {
    padding: 15px 20px;
  }

  .summary-stats {
    gap: 15px;
    flex-direction: column;
  }

  .stat-item {
    justify-content: space-between;
  }

  .stat-label {
    font-size: 13px;
  }

  .stat-value {
    font-size: 18px;
  }
}
</style>
