<template>
  <div>
    <!-- Loading State -->
    <div v-if="loading" class="loading-spinner">
      <div class="spinner"></div>
      <p>Memuat produk...</p>
    </div>

    <!-- Empty State -->
    <div v-else-if="isEmpty" class="empty-state">
      <div class="empty-icon">📦</div>
      <h3>Belum ada produk</h3>
      <p>Mulai dengan mengimpor produk dari Shopee atau tambah produk baru</p>
      <div class="empty-actions">
        <router-link to="/master-products/import" class="btn btn-secondary">
          Import dari Shopee
        </router-link>
        <router-link to="/master-products/add" class="btn btn-primary">
          Tambah Produk
        </router-link>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
defineProps<{
  loading?: boolean;
  isEmpty?: boolean;
}>();
</script>

<style scoped>
/* Loading State */
.loading-spinner {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  padding: 3rem;
  color: var(--text-secondary, #6b7280);
}

.spinner {
  width: 40px;
  height: 40px;
  border: 3px solid var(--border, #e5e7eb);
  border-top-color: var(--primary, #3b82f6);
  border-radius: 50%;
  animation: spin 1s linear infinite;
  margin-bottom: 1rem;
}

@keyframes spin {
  to {
    transform: rotate(360deg);
  }
}

/* Empty State */
.empty-state {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  padding: 4rem 2rem;
  text-align: center;
  background: var(--surface, white);
  border-radius: 0.5rem;
  border: 1px dashed var(--border, #e5e7eb);
}

.empty-icon {
  font-size: 3rem;
  margin-bottom: 1rem;
}

.empty-state h3 {
  font-size: 1.25rem;
  font-weight: 600;
  color: var(--text-primary, #1a1a1a);
  margin: 0 0 0.5rem 0;
}

.empty-state p {
  color: var(--text-secondary, #6b7280);
  margin: 0 0 1.5rem 0;
}

.empty-actions {
  display: flex;
  gap: 0.75rem;
}

.btn {
  padding: 0.5rem 1rem;
  border-radius: 0.375rem;
  font-weight: 500;
  text-decoration: none;
  cursor: pointer;
  transition: all 0.2s;
  font-size: 0.875rem;
}

.btn-primary {
  background: var(--primary, #3b82f6);
  color: white;
  border: none;
}

.btn-primary:hover {
  background: var(--primary-hover, #2563eb);
}

.btn-secondary {
  background: var(--surface, white);
  color: var(--text-primary, #1a1a1a);
  border: 1px solid var(--border, #e5e7eb);
}

.btn-secondary:hover {
  background: var(--surface-hover, #f3f4f6);
}
</style>
