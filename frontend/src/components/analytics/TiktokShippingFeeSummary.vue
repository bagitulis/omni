<template>
  <div class="summary-cards">
    <div class="card card-total">
      <div class="card-icon">📦</div>
      <div class="card-content">
        <span class="card-value">{{ summary.totalOrders }}</span>
        <span class="card-label">Total Orders</span>
      </div>
    </div>
    <div class="card card-diff">
      <div class="card-icon">⚠️</div>
      <div class="card-content">
        <span class="card-value">{{ summary.ordersWithDifference }}</span>
        <span class="card-label">With Difference</span>
      </div>
    </div>
    <div class="card card-profit">
      <div class="card-icon">📈</div>
      <div class="card-content">
        <span class="card-value profit">{{
          formatPrice(summary.totalProfit)
        }}</span>
        <span class="card-label">Total Profit</span>
      </div>
    </div>
    <div class="card card-loss">
      <div class="card-icon">📉</div>
      <div class="card-content">
        <span class="card-value loss">{{
          formatPrice(summary.totalLoss)
        }}</span>
        <span class="card-label">Total Loss</span>
      </div>
    </div>
    <div class="card card-net">
      <div class="card-icon">💰</div>
      <div class="card-content">
        <span
          class="card-value"
          :class="summary.netImpact >= 0 ? 'profit' : 'loss'"
        >
          {{ formatPrice(summary.netImpact) }}
        </span>
        <span class="card-label">Net Impact</span>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
interface TiktokShippingFeeSummary {
  totalOrders: number;
  ordersWithDifference: number;
  totalProfit: number;
  totalLoss: number;
  netImpact: number;
}

defineProps<{ summary: TiktokShippingFeeSummary }>();

function formatPrice(price: number): string {
  const prefix = price >= 0 ? "+" : "";
  return (
    prefix +
    new Intl.NumberFormat("id-ID", {
      style: "currency",
      currency: "IDR",
      minimumFractionDigits: 0,
    }).format(price)
  );
}
</script>

<style scoped>
.summary-cards {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(160px, 1fr));
  gap: var(--space-md);
  margin-bottom: var(--space-lg);
}

.card {
  display: flex;
  align-items: center;
  gap: var(--space-sm);
  padding: var(--space-md);
  background: var(--color-surface);
  border: 1px solid var(--color-border);
  border-radius: var(--radius-lg);
}

.card-icon {
  font-size: 1.5rem;
}

.card-content {
  display: flex;
  flex-direction: column;
}

.card-value {
  font-size: 1.2rem;
  font-weight: 700;
  color: var(--color-text-primary);
}

.card-label {
  font-size: 0.75rem;
  color: var(--color-text-muted);
}

.card-value.profit {
  color: var(--color-success);
}

.card-value.loss {
  color: var(--color-error);
}

.card-total {
  border-left: 3px solid var(--color-primary);
}
.card-diff {
  border-left: 3px solid var(--color-warning);
}
.card-profit {
  border-left: 3px solid var(--color-success);
}
.card-loss {
  border-left: 3px solid var(--color-error);
}
.card-net {
  border-left: 3px solid var(--color-info);
}
</style>
