<template>
  <div class="tab-content">
    <!-- Platform Support Info -->
    <div class="info-banner warning">
      <span class="info-icon">🗑️</span>
      <div class="info-text">
        <strong>Delete Wholesale & Reset MPQ</strong>
        <p class="info-note">
          Shopee: Hapus wholesale tiers + reset MPQ ke 1 | TikTok: Reset
          minimum_order_quantity ke 1
        </p>
      </div>
    </div>

    <!-- Preview Section -->
    <div
      class="preview-section"
      v-if="shopeeItems.length + tiktokItems.length > 0"
    >
      <!-- Shopee Preview -->
      <div v-if="shopeeItems.length > 0" class="platform-preview">
        <h4 class="platform-title">
          <span class="platform-badge shopee">Shopee</span>
          {{ shopeeItems.length }} item
        </h4>
        <ul class="item-list">
          <li v-for="item in shopeePreview" :key="item.sku" class="sku-item">
            {{ item.sku }}
          </li>
        </ul>
        <p v-if="shopeeItems.length > 5" class="more-note">
          ... +{{ shopeeItems.length - 5 }} lainnya
        </p>
      </div>

      <!-- TikTok Preview -->
      <div v-if="tiktokItems.length > 0" class="platform-preview">
        <h4 class="platform-title">
          <span class="platform-badge tiktok">TikTok</span>
          {{ tiktokItems.length }} item
        </h4>
        <ul class="item-list">
          <li v-for="item in tiktokPreview" :key="item.sku" class="sku-item">
            {{ item.sku }}
          </li>
        </ul>
        <p v-if="tiktokItems.length > 5" class="more-note">
          ... +{{ tiktokItems.length - 5 }} lainnya
        </p>
      </div>
    </div>

    <div v-else class="no-items">Tidak ada item Shopee/TikTok yang dipilih</div>

    <!-- Action Buttons -->
    <div
      class="action-buttons"
      v-if="shopeeItems.length + tiktokItems.length > 0"
    >
      <button
        class="btn-delete"
        @click="handleDelete"
        :disabled="processing || localProcessing"
      >
        <span v-if="localProcessing" class="btn-loading">
          <span class="spinner"></span>
          Deleting...
        </span>
        <span v-else>🗑️ Delete Wholesale & Reset MPQ</span>
      </button>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, ref } from "vue";
import wholesaleService from "../../../../services/wholesaleService";

interface UpdateItem {
  sku: string;
  price: number;
  platform: string;
}

const props = defineProps<{
  shopeeItems: UpdateItem[];
  tiktokItems: UpdateItem[];
  processing: boolean;
}>();

const emit = defineEmits<{
  update: [result: any];
}>();

const localProcessing = ref(false);

const shopeePreview = computed(() => props.shopeeItems.slice(0, 5));
const tiktokPreview = computed(() => props.tiktokItems.slice(0, 5));

async function handleDelete() {
  if (props.shopeeItems.length === 0 && props.tiktokItems.length === 0) return;

  localProcessing.value = true;
  const results: any[] = [];

  try {
    // Delete Shopee wholesale & reset MPQ to 1
    if (props.shopeeItems.length > 0) {
      const shopeeSkus = props.shopeeItems.map((item) => item.sku);
      const shopeeResult = await wholesaleService.batchDeleteBySkus(shopeeSkus);
      results.push({
        platform: "shopee",
        processed: shopeeResult.data?.processed || 0,
        failed: shopeeResult.data?.failed || 0,
        action: "delete wholesale + reset MPQ",
      });
    }

    // Reset TikTok MPQ to 1 (no wholesale concept in TikTok)
    if (props.tiktokItems.length > 0) {
      const tiktokItems = props.tiktokItems.map((item) => ({
        sku: item.sku,
        price: item.price, // Keep original price when resetting
      }));
      const tiktokResult = await wholesaleService.batchTiktokMpq(
        tiktokItems,
        1 // Reset to MPQ = 1
      );
      results.push({
        platform: "tiktok",
        processed: tiktokResult.data?.processed || 0,
        failed: tiktokResult.data?.failed || 0,
        action: "reset MPQ to 1",
      });
    }

    const anyFailed = results.some((r) => r.failed > 0);
    emit("update", {
      success: !anyFailed,
      results,
      message: anyFailed ? "Beberapa item gagal" : "Berhasil hapus wholesale",
    });
  } catch (err: any) {
    emit("update", {
      success: false,
      message: err.message || "Gagal delete wholesale",
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
  background: #ffebee;
  padding: 12px 16px;
  border-radius: 8px;
  margin-bottom: 16px;
  border-left: 4px solid #e53935;
}

.info-icon {
  font-size: 24px;
}
.info-text strong {
  color: #c62828;
}
.info-note {
  font-size: 12px;
  color: #666;
  margin: 4px 0 0;
}

.platform-preview {
  margin-bottom: 12px;
  padding: 12px;
  background: #fafafa;
  border-radius: 8px;
}
.platform-title {
  font-size: 14px;
  margin-bottom: 8px;
  display: flex;
  align-items: center;
  gap: 8px;
}

.platform-badge {
  padding: 2px 8px;
  border-radius: 4px;
  font-size: 11px;
  color: white;
}
.platform-badge.shopee {
  background: #ee4d2d;
}
.platform-badge.tiktok {
  background: #000;
}

.item-list {
  margin: 0;
  padding-left: 20px;
  font-size: 12px;
  font-family: monospace;
}
.sku-item {
  padding: 2px 0;
}
.more-note {
  font-size: 11px;
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

.btn-delete {
  padding: 10px 20px;
  background: #e53935;
  color: white;
  border: none;
  border-radius: 4px;
  cursor: pointer;
  font-weight: 500;
  min-width: 220px;
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 8px;
}

.btn-delete:hover:not(:disabled) {
  background: #c62828;
}
.btn-delete:disabled {
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
