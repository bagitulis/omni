http://localhost:5173/order-manager
<template>
  <Teleport to="body">
    <div v-if="show" class="modal-overlay" @click.self="closeIfNotProcessing">
      <div class="modal-content wholesale-modal">
        <!-- Header -->
        <div class="modal-header">
          <h3>🗑️ Hapus Wholesale - Shopee</h3>
          <button class="close-btn" @click="closeIfNotProcessing">✕</button>
        </div>

        <!-- Body -->
        <div class="modal-body">
          <!-- Selected Items Info -->
          <div class="info-banner">
            <span class="info-icon">ℹ️</span>
            <div class="info-text">
              <strong>{{ skus.length }} SKU</strong> dipilih
              <span v-if="uniqueItemCount > 0">
                → <strong>{{ uniqueItemCount }} produk unik</strong>
              </span>
              <p class="info-note">
                Wholesale dihapus per produk (item_id), bukan per SKU. SKU
                dengan item_id yang sama akan diproses 1x saja.
              </p>
            </div>
          </div>

          <!-- SKU List Preview -->
          <div class="sku-preview" v-if="skus.length <= 10">
            <strong>SKU yang dipilih:</strong>
            <ul>
              <li v-for="sku in skus" :key="sku">{{ sku }}</li>
            </ul>
          </div>
          <div class="sku-preview" v-else>
            <strong>SKU yang dipilih:</strong>
            <ul>
              <li v-for="sku in skus.slice(0, 5)" :key="sku">{{ sku }}</li>
              <li class="more-items">... dan {{ skus.length - 5 }} lainnya</li>
            </ul>
          </div>

          <!-- Processing Status -->
          <div v-if="processing" class="processing-status">
            <div class="spinner"></div>
            <span>Memproses penghapusan wholesale...</span>
          </div>

          <!-- Result Section -->
          <div v-if="result" class="result-section" :class="resultClass">
            <div class="result-header">
              {{ result.success ? "✅" : "⚠️" }} {{ result.message }}
            </div>
            <div class="result-details">
              <p>Total SKU: {{ result.data.totalSkus }}</p>
              <p>Item Unik: {{ result.data.uniqueItems }}</p>
              <p>Berhasil: {{ result.data.processed }}</p>
              <p v-if="result.data.failed > 0" class="failed">
                Gagal: {{ result.data.failed }}
              </p>
              <p v-if="result.data.skipped.length > 0" class="skipped">
                SKU tidak ditemukan: {{ result.data.skipped.join(", ") }}
              </p>
            </div>
          </div>

          <!-- Error Message -->
          <div v-if="error" class="error-message">⚠️ {{ error }}</div>
        </div>

        <!-- Footer -->
        <div class="modal-footer">
          <button class="btn-cancel" @click="closeIfNotProcessing">
            {{ result ? "Tutup" : "Batal" }}
          </button>
          <button
            v-if="!result"
            class="btn-delete"
            @click="handleDelete"
            :disabled="skus.length === 0 || processing"
          >
            {{ processing ? "Memproses..." : "🗑️ Hapus Wholesale" }}
          </button>
        </div>
      </div>
    </div>
  </Teleport>
</template>

<script setup lang="ts">
import { ref, computed, watch } from "vue";
import wholesaleService, {
  BatchDeleteBySkusResult,
} from "../../../services/wholesaleService";

const props = defineProps<{
  show: boolean;
  skus: string[];
}>();

const emit = defineEmits<{
  close: [];
  completed: [result: BatchDeleteBySkusResult];
}>();

const processing = ref(false);
const result = ref<BatchDeleteBySkusResult | null>(null);
const error = ref("");
const uniqueItemCount = ref(0);

const resultClass = computed(() => ({
  success: result.value?.success === true,
  partial: result.value?.success === false && result.value?.data?.processed > 0,
  failed:
    result.value?.success === false && result.value?.data?.processed === 0,
}));

// Reset when modal opens
watch(
  () => props.show,
  (newVal) => {
    if (newVal) {
      result.value = null;
      error.value = "";
      uniqueItemCount.value = 0;
    }
  }
);

function closeIfNotProcessing() {
  if (!processing.value) {
    emit("close");
  }
}

async function handleDelete() {
  if (props.skus.length === 0) return;

  processing.value = true;
  error.value = "";
  result.value = null;

  try {
    const deleteResult = await wholesaleService.batchDeleteBySkus(props.skus);
    result.value = deleteResult;
    uniqueItemCount.value = deleteResult.data.uniqueItems;
    emit("completed", deleteResult);
  } catch (err: any) {
    error.value = err.message || "Gagal menghapus wholesale";
  } finally {
    processing.value = false;
  }
}
</script>

<style src="./WholesaleBatchDeleteModal.styles.css" scoped></style>
