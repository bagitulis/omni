<template>
  <div v-if="hasError" class="error-boundary">
    <div class="error-content">
      <h2>⚠️ An Error Occurred</h2>
      <p>{{ errorMessage }}</p>
      <div class="error-actions">
        <button @click="reload" class="btn-primary">🔄 Reload</button>
        <button @click="clearCache" class="btn-secondary">
          🗑️ Clear Cache
        </button>
      </div>
      <details v-if="errorDetails" class="error-details">
        <summary>Error Details (for debugging)</summary>
        <pre>{{ errorDetails }}</pre>
      </details>
    </div>
  </div>
  <slot v-else></slot>
</template>

<script setup lang="ts">
import { ref, onErrorCaptured } from "vue";
import cacheService from "@/services/cacheService";

const hasError = ref(false);
const errorMessage = ref("");
const errorDetails = ref("");

onErrorCaptured((err: any, _instance: any, info: string) => {
  console.error("❌ Error captured by boundary:", err);
  console.error("Component info:", info);

  hasError.value = true;
  errorMessage.value = err?.message || "An unknown error occurred";
  errorDetails.value = `${err?.stack || err}\n\nInfo: ${info}`;

  // Prevent error from propagating
  return false;
});

const reload = () => {
  hasError.value = false;
  errorMessage.value = "";
  errorDetails.value = "";
  window.location.reload();
};

const clearCache = () => {
  // Use centralized cache service
  cacheService.invalidateAll();

  alert("✅ Cache cleared successfully. Page will reload.");
  window.location.reload();
};
</script>

<style scoped>
.error-boundary {
  min-height: 400px;
  display: flex;
  align-items: center;
  justify-content: center;
  padding: 40px 20px;
  background: linear-gradient(135deg, #fef5e7 0%, #fadbd8 100%);
}

.error-content {
  background: white;
  border-radius: 12px;
  padding: 40px;
  max-width: 600px;
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.15);
  text-align: center;
}

.error-content h2 {
  color: #e74c3c;
  margin-bottom: 16px;
  font-size: 24px;
}

.error-content p {
  color: #555;
  margin-bottom: 24px;
  font-size: 16px;
  line-height: 1.6;
}

.error-actions {
  display: flex;
  gap: 12px;
  justify-content: center;
  margin-bottom: 20px;
}

.btn-primary,
.btn-secondary {
  padding: 12px 24px;
  border: none;
  border-radius: 8px;
  font-size: 14px;
  font-weight: 600;
  cursor: pointer;
  transition: all 0.3s ease;
}

.btn-primary {
  background: #3498db;
  color: white;
}

.btn-primary:hover {
  background: #2980b9;
  transform: translateY(-2px);
  box-shadow: 0 4px 8px rgba(52, 152, 219, 0.3);
}

.btn-secondary {
  background: #ecf0f1;
  color: #555;
}

.btn-secondary:hover {
  background: #bdc3c7;
  transform: translateY(-2px);
}

.error-details {
  margin-top: 20px;
  text-align: left;
  background: #f8f9fa;
  border-radius: 6px;
  padding: 12px;
}

.error-details summary {
  cursor: pointer;
  font-weight: 600;
  color: #666;
  padding: 8px;
}

.error-details summary:hover {
  color: #333;
}

.error-details pre {
  margin-top: 12px;
  padding: 12px;
  background: #2c3e50;
  color: #ecf0f1;
  border-radius: 4px;
  overflow-x: auto;
  font-size: 12px;
  line-height: 1.4;
  max-height: 300px;
  overflow-y: auto;
}
</style>
