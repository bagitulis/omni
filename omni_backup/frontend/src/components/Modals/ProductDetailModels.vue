<template>
  <div v-if="models.length > 0" class="models-section">
    <h4 class="section-title">
      <i class="pi pi-tag"></i>
      Product Variants ({{ models.length }})
    </h4>
    <div class="models-table">
      <table class="data-table" aria-label="Product variants">
        <thead>
          <tr>
            <th scope="col">Model ID</th>
            <th scope="col">Model SKU</th>
            <th scope="col">Model Name</th>
            <th scope="col">Price</th>
            <th scope="col">Stock</th>
            <th scope="col">Status</th>
            <th scope="col">Tier Index</th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="model in models" :key="model.id" class="model-row">
            <td class="model-id">{{ model.model_id }}</td>
            <td class="sku">{{ model.model_sku || "-" }}</td>
            <td class="model-name">{{ model.model_name || "-" }}</td>
            <td class="price">Rp {{ formatPrice(model.current_price) }}</td>
            <td class="stock" :class="{ low: (model.seller_stock || 0) < 5 }">
              {{ model.seller_stock || 0 }}
            </td>
            <td>
              <span
                class="status-badge"
                :class="model.model_status?.toLowerCase() || 'normal'"
              >
                {{ model.model_status || "NORMAL" }}
              </span>
            </td>
            <td class="tier-index">{{ model.tier_index || "-" }}</td>
          </tr>
        </tbody>
      </table>
    </div>
  </div>
  <div v-else class="empty-message">No variants found for this product</div>
</template>

<script setup lang="ts">
import { useProductFormatters } from "@/composables/useProductFormatters";

interface Props {
  models: any[];
}

defineProps<Props>();

const { formatPrice } = useProductFormatters();
</script>

<style scoped>
.models-section {
  margin-bottom: 30px;
}

.section-title {
  display: flex;
  align-items: center;
  gap: 10px;
  color: #2c3e50;
  font-size: 1.1rem;
  margin-bottom: 15px;
  padding-bottom: 10px;
  border-bottom: 2px solid #3498db;
}

.models-table {
  overflow-x: auto;
  margin-bottom: 15px;
}

.data-table {
  width: 100%;
  border-collapse: collapse;
  font-size: 0.9rem;
}

.data-table thead {
  background: #f8f9fa;
  border-bottom: 2px solid #ecf0f1;
}

.data-table th {
  padding: 10px;
  text-align: left;
  font-weight: 600;
  color: #2c3e50;
}

.data-table td {
  padding: 10px;
  border-bottom: 1px solid #ecf0f1;
  color: #34495e;
}

.model-row:hover {
  background: #f8f9fa;
}

.model-id {
  color: #9b59b6;
  font-weight: 600;
}

.sku {
  font-family: "Monaco", "Menlo", monospace;
  background: #f8f9fa;
  padding: 4px;
  border-radius: 3px;
  font-size: 0.85rem;
}

.price {
  color: #27ae60;
  font-weight: 600;
}

.stock {
  font-weight: 600;
}

.stock.low {
  color: #e74c3c;
  background: #fadbd8;
  padding: 4px;
  border-radius: 3px;
}

.status-badge {
  padding: 4px 8px;
  border-radius: 4px;
  font-size: 0.85rem;
  font-weight: 500;
}

.status-badge.normal {
  background: #d4edda;
  color: #155724;
}

.status-badge.inactive {
  background: #e2e3e5;
  color: #383d41;
}

.tier-index {
  font-family: "Monaco", "Menlo", monospace;
  font-size: 0.9rem;
}

.empty-message {
  text-align: center;
  padding: 30px;
  color: #95a5a6;
  background: #f8f9fa;
  border-radius: 6px;
}

@media (max-width: 768px) {
  .data-table {
    font-size: 0.8rem;
  }

  .data-table th,
  .data-table td {
    padding: 6px;
  }
}
</style>
