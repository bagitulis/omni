<template>
  <div class="tab-content">
    <!-- Info Banner -->
    <div class="info-banner">
      <span class="info-icon">ℹ️</span>
      <div class="info-text">
        <strong>{{ items.length }} SKU</strong> dipilih - Shopee Only
        <p class="info-note">
          Reset MPQ ke 1, lalu set 3 tier harga grosir. Formula: (Harga - Admin)
          + (Admin / MinQty)
        </p>
      </div>
    </div>

    <!-- Preview Table -->
    <div class="preview-table-container" v-if="items.length > 0">
      <table class="preview-table" aria-label="Wholesale pricing preview">
        <thead>
          <tr>
            <th scope="col">SKU</th>
            <th scope="col">Harga</th>
            <th scope="col">
              T1 ({{ settings.minOrder1 }}-{{ settings.maxOrder1 }})
            </th>
            <th scope="col">T2 ({{ tier2Min }}-{{ tier2Max }})</th>
            <th scope="col">
              T3 ({{ tier3Min }}-{{ settings.maxOrderTier3 }})
            </th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="item in previewItems" :key="item.sku">
            <td class="sku-cell">{{ item.sku }}</td>
            <td class="price-cell">{{ formatPrice(item.price) }}</td>
            <td>{{ formatPrice(item.tiers[0]?.unitPrice) }}</td>
            <td>{{ formatPrice(item.tiers[1]?.unitPrice) }}</td>
            <td>{{ formatPrice(item.tiers[2]?.unitPrice) }}</td>
          </tr>
        </tbody>
      </table>
      <p v-if="items.length > 5" class="more-items-note">
        ... dan {{ items.length - 5 }} lainnya
      </p>
    </div>

    <div v-else class="no-items">Tidak ada item Shopee yang dipilih</div>

    <!-- Update Button -->
    <div class="action-buttons" v-if="items.length > 0">
      <button
        class="btn-update"
        @click="handleUpdate"
        :disabled="processing || localProcessing"
      >
        <span v-if="localProcessing" class="btn-loading">
          <span class="spinner"></span>
          Updating...
        </span>
        <span v-else>📦 Update Wholesale</span>
      </button>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, ref } from "vue";
import wholesaleService, {
  WholesaleSettings,
} from "../../../../services/wholesaleService";

interface UpdateItem {
  sku: string;
  price: number;
  platform: string;
}

const props = defineProps<{
  items: UpdateItem[];
  settings: WholesaleSettings;
  processing: boolean;
}>();

const emit = defineEmits<{
  update: [result: any];
}>();

const localProcessing = ref(false);

// Computed tier ranges
const tier2Min = computed(() => props.settings.maxOrder1 + 1);
const tier2Max = computed(() => tier2Min.value + 1);
const tier3Min = computed(() => tier2Max.value + 1);

// Preview items (first 5)
const previewItems = computed(() => {
  return props.items.slice(0, 5).map((item) => ({
    ...item,
    tiers: wholesaleService.calculateTiersLocal(item.price, props.settings),
  }));
});

function formatPrice(price: number | undefined): string {
  if (price === undefined) return "-";
  return new Intl.NumberFormat("id-ID").format(price);
}

async function handleUpdate() {
  if (props.items.length === 0) return;

  localProcessing.value = true;

  try {
    // Use the new endpoint with MPQ reset
    const result = await wholesaleService.batchWholesaleWithReset(props.items);
    emit("update", result);
  } catch (err: any) {
    emit("update", {
      success: false,
      message: err.message || "Gagal update wholesale",
      data: { processed: 0, failed: props.items.length },
    });
  } finally {
    localProcessing.value = false;
  }
}
</script>

<style scoped>
.tab-content {
  min-height: 200px;
}

.info-banner {
  display: flex;
  gap: 12px;
  background: #fff3e0;
  padding: 12px 16px;
  border-radius: 8px;
  margin-bottom: 16px;
}

.info-icon {
  font-size: 24px;
}
.info-text strong {
  color: #e65100;
}
.info-note {
  font-size: 12px;
  color: #666;
  margin: 4px 0 0 0;
}

.preview-table-container {
  overflow-x: auto;
  margin-bottom: 8px;
}

.preview-table {
  width: 100%;
  border-collapse: collapse;
  font-size: 13px;
}

.preview-table th,
.preview-table td {
  padding: 8px 12px;
  border: 1px solid #e0e0e0;
  text-align: left;
}

.preview-table th {
  background: #f5f5f5;
  font-weight: 600;
}
.sku-cell {
  font-family: monospace;
}
.price-cell {
  font-weight: 600;
  color: #1565c0;
}

.more-items-note {
  font-size: 12px;
  color: #666;
  font-style: italic;
}

.no-items {
  text-align: center;
  padding: 40px;
  color: #6b7280;
}

.action-buttons {
  display: flex;
  justify-content: flex-end;
  margin-top: 16px;
}

.btn-update {
  padding: 10px 20px;
  background: #ee4d2d;
  color: white;
  border: none;
  border-radius: 4px;
  cursor: pointer;
  font-weight: 500;
  min-width: 180px;
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 8px;
}

.btn-update:hover:not(:disabled) {
  background: #d73211;
}
.btn-update:disabled {
  background: #ccc;
  cursor: not-allowed;
}

.btn-loading {
  display: flex;
  align-items: center;
  gap: 8px;
}

.spinner {
  width: 14px;
  height: 14px;
  border: 2px solid rgba(255, 255, 255, 0.3);
  border-top-color: white;
  border-radius: 50%;
  animation: spin 0.8s linear infinite;
}

@keyframes spin {
  to {
    transform: rotate(360deg);
  }
}
</style>
