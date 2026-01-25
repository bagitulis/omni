<template>
  <div v-if="loading" class="loading-spinner">
    <i class="pi pi-spin pi-spinner" aria-hidden="true"></i>
    <p>Loading products...</p>
  </div>

  <div v-else-if="products.length === 0" class="empty-state">
    <i class="pi pi-inbox" aria-hidden="true"></i>
    <p>No products in database</p>
    <small>Products will appear here after fetching from API</small>
  </div>

  <div v-else class="products-table-wrapper">
    <div class="products-table">
      <table>
        <thead>
          <tr>
            <th v-for="field in columnFields" :key="field" v-show="visibleColumnsWithDefaults[field]">
              {{ columnLabels[field] }}
            </th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="product in (products as Array<Record<string, any>>)" :key="getProductKey(product as Record<string, any>)">
            <td v-for="field in columnFields" :key="field" v-show="visibleColumnsWithDefaults[field]">
              <slot :name="`cell-${field}`" :value="(product as Record<string, any>)[field]" :product="product">
                {{ formatCell(field, (product as Record<string, any>)[field]) }}
              </slot>
            </td>
          </tr>
        </tbody>
      </table>
    </div>

    <!-- Pagination -->
    <div class="pagination">
      <button
        @click="$emit('prev-page')"
        :disabled="currentPage === 1"
        class="btn-pagination"
        aria-label="Previous page"
        title="Previous page"
      >
        <i class="pi pi-chevron-left" aria-hidden="true"></i>
      </button>
      <span class="page-info">
        Page {{ currentPage }} of {{ totalPages }}
      </span>
      <button
        @click="$emit('next-page')"
        :disabled="currentPage === totalPages"
        class="btn-pagination"
        aria-label="Next page"
        title="Next page"
      >
        <i class="pi pi-chevron-right" aria-hidden="true"></i>
      </button>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed } from "vue";
import { formatNumber, formatDateTime } from "@/utils/productManagerUtils";

const props = defineProps({
  products: {
    type: Array,
    required: true,
  },
  columnFields: {
    type: Array as () => string[],
    required: true,
  },
  columnLabels: {
    type: Object,
    required: true,
  },
  visibleColumns: {
    type: Object,
    required: true,
  },
  currentPage: {
    type: Number,
    required: true,
  },
  totalPages: {
    type: Number,
    required: true,
  },
  loading: {
    type: Boolean,
    default: false,
  },
  keyFields: {
    type: Array as () => string[],
    default: () => ["item_id", "itemId"],
  },
});

defineEmits(["next-page", "prev-page"]);

// Ensure all columns visible by default if visibleColumns not set or incomplete
const visibleColumnsWithDefaults = computed(() => {
  const result: Record<string, boolean> = {};
  for (const field of props.columnFields) {
    // Only hide if explicitly set to false
    result[field] = props.visibleColumns && props.visibleColumns.hasOwnProperty(field) 
      ? props.visibleColumns[field] === true 
      : true; // Default: show
  }
  return result;
});

const formatCell = (field: string, value: any): string => {
  // Handle price fields
  if (["price", "currentPrice", "originalPrice"].includes(field)) {
    return value ? "Rp " + formatNumber(value) : "-";
  }
  // Handle date fields
  if (["updated_at", "updated"].includes(field)) {
    return formatDateTime(value);
  }
  // Handle stock/quantity fields
  if (["quantity", "sellerStock", "shopeeStock"].includes(field)) {
    return value || 0;
  }
  // Handle status badges
  if (field === "status") {
    return value || "-";
  }
  // Default: return formatted string
  return value ? String(value) : "-";
};

const getProductKey = (product: Record<string, any>): string => {
  // Try to find a unique key combination
  for (const keyField of props.keyFields) {
    if (product[keyField]) {
      // Try to find a secondary key
      const secondaryKeys = ["sku_id", "skuId", "modelId"];
      for (const secondKey of secondaryKeys) {
        if (product[secondKey]) {
          return `${product[keyField]}-${product[secondKey]}`;
        }
      }
      return String(product[keyField]);
    }
  }
  return Math.random().toString();
};
</script>

<style scoped>
.loading-spinner {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  padding: 40px 20px;
  color: #6b7280;
}

.loading-spinner i {
  font-size: 2rem;
  margin-bottom: 10px;
  color: #1976d2;
}

.empty-state {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  padding: 60px 20px;
  text-align: center;
  color: #6b7280;
}

.empty-state i {
  font-size: 3rem;
  margin-bottom: 15px;
  color: #ddd;
}

.empty-state p {
  margin: 0 0 5px 0;
  font-size: 1rem;
  color: #555;
}

.empty-state small {
  color: #6b7280;
  font-size: 0.9rem;
}

.products-table-wrapper {
  background: white;
  border-radius: 6px;
  box-shadow: 0 2px 4px rgba(0, 0, 0, 0.08);
  overflow: hidden;
}

.products-table {
  overflow-x: auto;
}

table {
  width: 100%;
  border-collapse: collapse;
  font-size: 0.9rem;
}

thead {
  background: #f5f5f5;
  border-bottom: 2px solid #e0e0e0;
}

th {
  padding: 12px 15px;
  text-align: left;
  font-weight: 600;
  color: #2c3e50;
  white-space: nowrap;
}

tbody tr {
  border-bottom: 1px solid #f0f0f0;
  transition: background-color 0.2s ease;
}

tbody tr:hover {
  background-color: #fafafa;
}

td {
  padding: 12px 15px;
  color: #555;
  word-break: break-word;
}

.pagination {
  display: flex;
  justify-content: center;
  align-items: center;
  gap: 15px;
  padding: 20px;
  background: #f9f9f9;
  border-top: 1px solid #e0e0e0;
}

.btn-pagination {
  padding: 8px 12px;
  border: 1px solid #ddd;
  background: white;
  border-radius: 4px;
  cursor: pointer;
  display: flex;
  align-items: center;
  justify-content: center;
  transition: all 0.3s ease;
}

.btn-pagination:hover:not(:disabled) {
  background: #f0f0f0;
  border-color: #6b7280;
}

.btn-pagination:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

.page-info {
  color: #666;
  font-size: 0.9rem;
  min-width: 150px;
  text-align: center;
}
</style>
