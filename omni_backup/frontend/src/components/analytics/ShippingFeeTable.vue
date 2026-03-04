<template>
  <div class="results-section">
    <div class="section-header">
      <h2>🚚 Shipping Fee Analysis</h2>
      <span class="result-count">{{ orders.length }} orders with difference</span>
    </div>
    <div class="table-container">
      <table class="results-table">
        <thead>
          <tr>
            <th class="col-date">Order Date</th>
            <th class="col-order">Order SN</th>
            <th class="col-price">Buyer Paid</th>
            <th class="col-price">Actual Fee</th>
            <th class="col-price">Rebate</th>
            <th class="col-price">Difference</th>
            <th class="col-method">Payment Method</th>
          </tr>
        </thead>
        <tbody>
          <tr
            v-for="order in orders"
            :key="order.orderSn"
            :class="getDifferenceClass(order.difference)"
          >
            <td class="cell-date">{{ order.orderDate || "-" }}</td>
            <td class="cell-order">{{ order.orderSn }}</td>
            <td class="cell-price">{{ formatPrice(order.buyerPaid) }}</td>
            <td class="cell-price">{{ formatPrice(order.actualFee) }}</td>
            <td class="cell-price">{{ formatPrice(order.shopeeRebate) }}</td>
            <td class="cell-price" :class="getDifferenceValueClass(order.difference)">
              {{ formatDifference(order.difference) }}
            </td>
            <td class="cell-method">{{ order.paymentMethod || "-" }}</td>
          </tr>
        </tbody>
      </table>
    </div>
  </div>
</template>

<script setup lang="ts">
interface ShippingFeeOrder {
  orderSn: string;
  orderDate: string | null;
  buyerPaid: number;
  actualFee: number;
  shopeeRebate: number;
  difference: number;
  buyerName: string | null;
  paymentMethod: string | null;
}

defineProps<{ orders: ShippingFeeOrder[] }>();

function getDifferenceClass(diff: number): string {
  return diff > 0 ? "row-profit" : "row-loss";
}

function getDifferenceValueClass(diff: number): string {
  return diff > 0 ? "value-profit" : "value-loss";
}

function formatPrice(price: number | null): string {
  if (price === null || price === undefined) return "-";
  return new Intl.NumberFormat("id-ID", {
    style: "currency",
    currency: "IDR",
    minimumFractionDigits: 0,
  }).format(price);
}

function formatDifference(diff: number): string {
  const prefix = diff > 0 ? "+" : "";
  return prefix + formatPrice(diff);
}
</script>

<style scoped>
.results-section {
  background: var(--color-surface);
  border-radius: var(--radius-lg);
  border: 1px solid var(--color-border);
  overflow: hidden;
}

.section-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: var(--space-md) var(--space-lg);
  border-bottom: 1px solid var(--color-border);
  background: var(--color-surface-alt);
}

.section-header h2 {
  margin: 0;
  font-size: 1rem;
  color: var(--color-text-primary);
}

.result-count {
  font-size: 0.85rem;
  color: var(--color-text-muted);
}

.table-container {
  overflow-x: auto;
}

.results-table {
  width: 100%;
  border-collapse: collapse;
  font-size: 0.85rem;
}

.results-table th,
.results-table td {
  padding: var(--space-sm) var(--space-md);
  text-align: left;
  border-bottom: 1px solid var(--color-border);
}

.results-table th {
  background: var(--color-surface-alt);
  font-weight: 600;
  color: var(--color-text-secondary);
  white-space: nowrap;
}

.col-date { width: 100px; }
.col-order { width: 180px; }
.col-price { width: 110px; text-align: right; }
.col-method { width: 140px; }

.cell-date { color: var(--color-text-muted); }
.cell-order { font-family: monospace; font-size: 0.8rem; }
.cell-price { text-align: right; }
.cell-method { color: var(--color-text-secondary); }

.row-profit {
  background: var(--color-success-bg, rgba(34, 197, 94, 0.05));
}

.row-loss {
  background: var(--color-error-bg, rgba(239, 68, 68, 0.05));
}

.value-profit {
  color: var(--color-success, #22c55e);
  font-weight: 600;
}

.value-loss {
  color: var(--color-error, #ef4444);
  font-weight: 600;
}
</style>
