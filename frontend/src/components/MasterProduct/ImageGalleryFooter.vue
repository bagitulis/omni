<template>
  <div class="modal-footer">
    <div class="footer-left">
      <div class="selection-info">
        Terpilih: {{ selectedCount }} / {{ maxImages }}
      </div>
      <div class="pagination" v-if="meta.pages > 1">
        <button
          class="page-btn"
          :disabled="meta.page === 1"
          @click="$emit('page-change', meta.page - 1)"
        >
          ‹
        </button>
        <span class="page-info">{{ meta.page }} / {{ meta.pages }}</span>
        <button
          class="page-btn"
          :disabled="meta.page === meta.pages"
          @click="$emit('page-change', meta.page + 1)"
        >
          ›
        </button>
      </div>
    </div>

    <div class="footer-actions">
      <button class="btn-cancel" @click="$emit('cancel')">Batal</button>
      <button
        class="btn-confirm"
        @click="$emit('confirm')"
        :disabled="selectedCount === 0"
      >
        Pilih ({{ selectedCount }})
      </button>
    </div>
  </div>
</template>

<script setup lang="ts">
defineProps<{
  selectedCount: number;
  maxImages: number;
  meta: {
    page: number;
    pages: number;
    total: number;
    page_size: number;
  };
}>();

defineEmits<{
  (e: "page-change", page: number): void;
  (e: "cancel"): void;
  (e: "confirm"): void;
}>();
</script>

<style scoped>
.modal-footer {
  padding: 1rem 1.25rem;
  border-top: 1px solid #e5e7eb;
  display: flex;
  justify-content: space-between;
  align-items: center;
  background: white;
}

.footer-left {
  display: flex;
  gap: 1.5rem;
  align-items: center;
}

.selection-info {
  font-size: 0.875rem;
  font-weight: 600;
  color: #374151;
}

.pagination {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  font-size: 0.875rem;
}

.page-btn {
  padding: 0.25rem 0.5rem;
  border: 1px solid #e5e7eb;
  background: white;
  border-radius: 0.25rem;
  cursor: pointer;
}

.page-btn:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

.footer-actions {
  display: flex;
  gap: 0.75rem;
}

.btn-cancel {
  padding: 0.625rem 1.25rem;
  background: white;
  border: 1px solid #d1d5db;
  color: #374151;
  border-radius: 0.5rem;
  font-weight: 500;
  cursor: pointer;
}

.btn-confirm {
  padding: 0.625rem 1.5rem;
  background: #ff6b2c;
  color: white;
  border: none;
  border-radius: 0.5rem;
  font-weight: 600;
  cursor: pointer;
}

.btn-confirm:disabled {
  opacity: 0.6;
  cursor: not-allowed;
}
</style>
