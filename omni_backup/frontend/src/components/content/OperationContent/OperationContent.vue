<template>
  <!-- Enhanced Operation Content with improved UI/UX -->
  <div class="operation-content">
    <!-- Header -->
    <OperationHeader
      :platform="platform"
      :status="status"
      :loading="loading"
      @refresh-status="$emit('refresh-status')"
    />

    <!-- Operation sections with improved layout -->
    <div class="operation-sections">
      <!-- Order Management Section -->
      <OrderManagementSection
        :platform="platform"
        @export-orders="$emit('export-orders', $event)"
        @show-wallet-modal="$emit('show-wallet-modal')"
        @show-shipping-modal="$emit('show-shipping-modal')"
      />

      <!-- Authentication Section -->
      <AuthenticationSection
        :platform="platform"
        @handle-operation="$emit('handle-operation', $event, platform)"
      />
    </div>

    <!-- Loading overlay -->
    <div v-if="loading" class="loading-overlay">
      <div class="spinner"></div>
      <p>Processing operation...</p>
    </div>
  </div>
</template>

<script lang="ts">
import { defineComponent } from "vue";
import OperationHeader from "./OperationHeader.vue";
import OrderManagementSection from "./OrderManagementSection.vue";
import AuthenticationSection from "./AuthenticationSection.vue";

export default defineComponent({
  name: "OperationContent",
  components: {
    OperationHeader,
    OrderManagementSection,
    AuthenticationSection,
  },
  props: {
    platform: {
      type: String,
      default: "shopee",
    },
    status: {
      type: Object,
      default: null,
    },
    loading: {
      type: Boolean,
      default: false,
    },
  },
  emits: [
    "refresh-status",
    "show-token-modal",
    "export-orders",
    "handle-operation",
    "update-price",
    "show-wallet-modal",
    "show-shipping-modal",
  ],
});
</script>

<style scoped>
* {
  margin: 0;
  padding: 0;
  box-sizing: border-box;
}

.operation-content {
  height: 100%;
  display: flex;
  flex-direction: column;
  background: linear-gradient(135deg, #f5f7fa 0%, #c3cfe2 100%);
  position: relative;
  overflow: hidden;
}

/* ==================== OPERATION SECTIONS ==================== */
.operation-sections {
  flex: 1;
  overflow-y: auto;
  padding: 20px;
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(320px, 1fr));
  gap: 16px;
  background: linear-gradient(135deg, #f5f7fa 0%, #c3cfe2 100%);
}

.operation-sections::-webkit-scrollbar {
  width: 8px;
}

.operation-sections::-webkit-scrollbar-track {
  background: transparent;
}

.operation-sections::-webkit-scrollbar-thumb {
  background: rgba(0, 0, 0, 0.2);
  border-radius: 4px;
}

.operation-sections::-webkit-scrollbar-thumb:hover {
  background: rgba(0, 0, 0, 0.3);
}

/* ==================== LOADING OVERLAY ==================== */
.loading-overlay {
  position: absolute;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background: rgba(255, 255, 255, 0.95);
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  z-index: 1000;
  backdrop-filter: blur(4px);
}

.spinner {
  width: 50px;
  height: 50px;
  border: 4px solid #e8eef5;
  border-top-color: #4a90e2;
  border-radius: 50%;
  animation: spin 0.8s linear infinite;
  margin-bottom: 16px;
}

@keyframes spin {
  to {
    transform: rotate(360deg);
  }
}

.loading-overlay p {
  font-size: 0.95rem;
  color: #4b5563; /* Improved from #7a8fa6 for WCAG AA 4.5:1 contrast */
  font-weight: 500;
}

/* ==================== RESPONSIVE DESIGN ==================== */
@media (max-width: 768px) {
  .operation-sections {
    grid-template-columns: 1fr;
    padding: 16px;
    gap: 16px;
  }
}
</style>
