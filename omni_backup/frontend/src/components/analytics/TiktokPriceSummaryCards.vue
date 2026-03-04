<template>
  <div class="summary-cards">
    <div class="summary-card total">
      <div class="card-icon">📦</div>
      <div class="card-content">
        <div class="card-value">{{ summary.totalSku }}</div>
        <div class="card-label">Total SKU</div>
      </div>
    </div>
    <div class="summary-card transactions">
      <div class="card-icon">🛒</div>
      <div class="card-content">
        <div class="card-value">{{ summary.totalTransactions }}</div>
        <div class="card-label">Transactions</div>
      </div>
    </div>
    <div class="summary-card ok">
      <div class="card-icon">✅</div>
      <div class="card-content">
        <div class="card-value">{{ summary.skuOk }}</div>
        <div class="card-label">Price OK</div>
      </div>
    </div>
    <div class="summary-card warning">
      <div class="card-icon">⚠️</div>
      <div class="card-content">
        <div class="card-value">{{ summary.skuWithPriceDiff }}</div>
        <div class="card-label">Price Diff</div>
      </div>
    </div>
    <div class="summary-card info">
      <div class="card-icon">❓</div>
      <div class="card-content">
        <div class="card-value">{{ summary.skuNoInventory }}</div>
        <div class="card-label">No Inventory</div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
interface Summary {
  totalSku: number;
  totalTransactions: number;
  skuOk: number;
  skuWithPriceDiff: number;
  skuNoInventory: number;
}

defineProps<{ summary: Summary }>();
</script>

<style scoped>
.summary-cards {
  display: grid;
  grid-template-columns: repeat(5, 1fr);
  gap: 16px;
  margin-bottom: 20px;
}

.summary-card {
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 16px 20px;
  background: white;
  border-radius: 8px;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.08);
}

.summary-card .card-icon {
  font-size: 1.8rem;
}

.summary-card .card-value {
  font-size: 1.8rem;
  font-weight: 700;
  color: #2c3e50;
}

.summary-card .card-label {
  font-size: 0.9rem;
  color: #6b7280;
}

.summary-card.ok {
  border-left: 3px solid var(--color-success);
}

.summary-card.warning {
  border-left: 3px solid var(--color-warning);
}

.summary-card.info {
  border-left: 3px solid var(--color-info);
}

@media (max-width: 1200px) {
  .summary-cards {
    grid-template-columns: repeat(3, 1fr);
  }
}

@media (max-width: 768px) {
  .summary-cards {
    grid-template-columns: repeat(2, 1fr);
  }
}
</style>
