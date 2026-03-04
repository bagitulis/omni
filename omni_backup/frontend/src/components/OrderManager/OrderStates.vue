<template>
  <div>
    <!-- Loading state -->
    <div v-if="loading" class="loading-state">
      <i class="pi pi-spin pi-spinner" aria-hidden="true"></i>
      <p>Loading orders from all platforms...</p>
    </div>

    <!-- Error state -->
    <div v-if="error && !loading" class="error-state">
      <i class="pi pi-exclamation-triangle" aria-hidden="true"></i>
      <p>{{ error }}</p>
      <button @click="emit('close-error')" class="btn-close">Close</button>
    </div>

    <!-- Empty state -->
    <div v-if="!loading && !error && isNoData" class="empty-state">
      <i class="pi pi-inbox" aria-hidden="true"></i>
      <p>
        {{ searchActive ? "Tidak ada hasil yang cocok" : "Tidak ada order" }}
      </p>
    </div>
  </div>
</template>

<script setup lang="ts">
defineProps<{
  loading: boolean;
  error: string | null;
  isNoData: boolean;
  searchActive: boolean;
}>();

const emit = defineEmits<{
  "close-error": [];
}>();
</script>

<style scoped>
@import "./OrderManager.styles.css";

/* States */
.loading-state,
.error-state,
.empty-state {
  background: white;
  border-radius: 12px;
  padding: 80px 20px;
  text-align: center;
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.08);
}

.loading-state i,
.error-state i,
.empty-state i {
  font-size: 64px;
  margin-bottom: 20px;
  display: block;
}

.loading-state i {
  color: #3498db;
  animation: spin 1s linear infinite;
}

.error-state i {
  color: #e74c3c;
}

.error-state {
  color: #e74c3c;
}

.empty-state i {
  color: #bdc3c7;
}

.empty-state {
  color: #7f8c8d;
}

.error-state p,
.empty-state p {
  margin: 15px 0;
  font-size: 18px;
  font-weight: 500;
}

.btn-close {
  padding: 10px 20px;
  background: #e74c3c;
  color: white;
  border: none;
  border-radius: 6px;
  cursor: pointer;
  margin-top: 15px;
  font-weight: 600;
  transition: all 0.3s;
}

.btn-close:hover {
  background: #c0392b;
}

/* Animations */
@keyframes spin {
  from {
    transform: rotate(0deg);
  }
  to {
    transform: rotate(360deg);
  }
}

/* Responsive */
@media (max-width: 768px) {
  .loading-state,
  .error-state,
  .empty-state {
    padding: 60px 15px;
  }

  .loading-state i,
  .error-state i,
  .empty-state i {
    font-size: 48px;
    margin-bottom: 15px;
  }

  .error-state p,
  .empty-state p {
    font-size: 16px;
  }

  .btn-close {
    padding: 8px 16px;
    font-size: 13px;
  }
}
</style>
