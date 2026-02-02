<template>
  <div v-if="loading" class="loading-container">
    <div class="spinner"></div>
    <p>Memuat preview...</p>
  </div>

  <div v-else-if="error" class="error-container">
    <span class="error-icon">!</span>
    <p>{{ error }}</p>
    <button @click="$emit('retry')" class="btn-retry">Coba Lagi</button>
  </div>
</template>

<script setup lang="ts">
defineProps<{
  loading: boolean;
  error: string;
}>();

defineEmits<{
  (e: "retry"): void;
}>();
</script>

<style scoped>
.loading-container,
.error-container {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  padding: 4rem 2rem;
  gap: 1rem;
  color: #6b7280;
}

.spinner {
  width: 48px;
  height: 48px;
  border: 4px solid #f3f4f6;
  border-top-color: #ff6b2c;
  border-radius: 50%;
  animation: spin 0.8s linear infinite;
}

@keyframes spin {
  to {
    transform: rotate(360deg);
  }
}

.error-icon {
  width: 48px;
  height: 48px;
  background: #fee2e2;
  color: #ef4444;
  border-radius: 50%;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 1.5rem;
  font-weight: 700;
}

.btn-retry {
  padding: 0.75rem 1.5rem;
  background: #ff6b2c;
  color: white;
  border: none;
  border-radius: 0.5rem;
  cursor: pointer;
  font-weight: 600;
}
</style>
