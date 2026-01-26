<template>
  <div class="results-section">
    <div class="section-header">
      <h2>📋 SKU Price Analysis</h2>
      <span class="result-count">{{ skuGroups.length }} items</span>
    </div>
    <div class="table-container">
      <table class="results-table">
        <thead>
          <tr>
            <th class="col-status">Status</th>
            <th class="col-sku">SKU</th>
            <th class="col-name">Item Name</th>
            <th class="col-price">Marketplace Price</th>
            <th class="col-price">Inventory Price</th>
            <th class="col-price">Actual Income</th>
            <th class="col-price">Expected Income</th>
          </tr>
        </thead>
        <tbody>
          <tr
            v-for="sku in skuGroups"
            :key="sku.sku"
            :class="getRowClass(sku.status)"
          >
            <td>
              <span :class="['status-badge', sku.status.toLowerCase()]">
                {{ getStatusLabel(sku.status) }}
              </span>
            </td>
            <td class="cell-sku">{{ sku.model_sku || sku.sku }}</td>
            <td class="cell-name">
              {{ sku.item_name }}
              <span v-if="sku.model_name" class="variant-name">
                - {{ sku.model_name }}
              </span>
            </td>
            <td class="cell-price">
              <span v-for="(price, idx) in sku.unique_unit_prices" :key="idx">
                {{ formatPrice(price)
                }}<span v-if="idx < sku.unique_unit_prices.length - 1">, </span>
              </span>
            </td>
            <td class="cell-price">{{ formatPrice(sku.inventory_price) }}</td>
            <td class="cell-price">
              <span
                v-for="(income, idx) in getTopTwoSmallestIncomes(
                  sku.unique_actual_incomes,
                )"
                :key="idx"
              >
                {{ formatPrice(income)
                }}<span
                  v-if="
                    idx <
                    getTopTwoSmallestIncomes(sku.unique_actual_incomes).length -
                      1
                  "
                  >,
                </span>
              </span>
            </td>
            <td class="cell-price expected">
              {{ formatPrice(sku.expected_income) }}
            </td>
          </tr>
        </tbody>
      </table>
    </div>
  </div>
</template>

<script setup lang="ts">
interface SkuGroup {
  sku: string;
  model_sku: string;
  item_name: string;
  model_name: string;
  status: string;
  inventory_price: number | null;
  unique_unit_prices: number[];
  unique_actual_incomes: number[];
  expected_income: number | null;
  total_transactions: number;
}

defineProps<{ skuGroups: SkuGroup[] }>();

function getStatusLabel(status: string): string {
  switch (status) {
    case "OK":
      return "✓ OK";
    case "PRICE_DIFF":
      return "⚠ Diff";
    case "NO_INVENTORY":
      return "? N/A";
    default:
      return status;
  }
}

function getRowClass(status: string): string {
  switch (status) {
    case "OK":
      return "row-ok";
    case "PRICE_DIFF":
      return "row-warning";
    case "NO_INVENTORY":
      return "row-info";
    default:
      return "";
  }
}

function formatPrice(price: number | null): string {
  if (price === null || price === undefined) return "-";
  return new Intl.NumberFormat("id-ID", {
    style: "currency",
    currency: "IDR",
    minimumFractionDigits: 0,
  }).format(price);
}

function getTopTwoSmallestIncomes(incomes: number[]): number[] {
  if (!incomes || incomes.length === 0) return [];
  // Sort ascending and take first 2
  const sorted = [...incomes].sort((a, b) => a - b);
  return sorted.slice(0, 2);
}
</script>

<style scoped>
.results-section {
  background: white;
  border-radius: 8px;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.08);
  overflow: hidden;
}

.section-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 16px 20px;
  border-bottom: 2px solid #e0e0e0;
  background: #f5f5f5;
}

.section-header h2 {
  margin: 0;
  font-size: 1.1rem;
  color: #2c3e50;
}

.result-count {
  font-size: 0.9rem;
  color: #6b7280;
}

.table-container {
  overflow-x: auto;
}

.results-table {
  width: 100%;
  border-collapse: collapse;
  font-size: 0.9rem;
}

.results-table th {
  padding: 12px 15px;
  text-align: left;
  background: #f5f5f5;
  color: #2c3e50;
  font-weight: 600;
  position: sticky;
  top: 0;
  font-size: 0.85rem;
  text-transform: uppercase;
}

.results-table td {
  padding: 12px 15px;
  border-bottom: 1px solid #f0f0f0;
}

.results-table tbody tr:hover {
  background: rgba(99, 102, 241, 0.05);
}

.row-ok {
  background: rgba(34, 197, 94, 0.03);
}
.row-warning {
  background: rgba(245, 158, 11, 0.05);
}
.row-info {
  background: rgba(59, 130, 246, 0.03);
}

.status-badge {
  padding: 6px 10px;
  border-radius: 4px;
  font-size: 0.8rem;
  font-weight: 500;
}

.status-badge.ok {
  background: #dcfce7;
  color: #166534;
}

.status-badge.price_diff {
  background: #fef3c7;
  color: #92400e;
}

.status-badge.no_inventory {
  background: #dbeafe;
  color: #1e40af;
}

.cell-sku {
  font-family: "Consolas", monospace;
  font-size: 0.9rem;
  color: #1976d2;
  text-align: left;
  font-weight: 600;
}

.cell-name {
  color: var(--color-text-secondary);
  text-align: left;
  word-wrap: break-word;
  white-space: normal;
  line-height: 1.4;
  min-width: 200px;
  max-width: 300px;
}

.variant-name {
  color: var(--color-text-muted);
  font-style: italic;
}

.cell-price {
  font-family: "Consolas", monospace;
  text-align: left;
  color: var(--color-text-secondary);
}

.cell-price.expected {
  color: var(--color-success);
  font-weight: 600;
}
</style>
