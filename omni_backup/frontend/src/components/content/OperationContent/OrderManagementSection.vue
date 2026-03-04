<template>
  <div class="operation-card">
    <div class="card-header">
      <div class="header-title">
        <i class="pi pi-shopping-cart"></i>
        <h3>Order Management</h3>
      </div>
      <span class="card-badge">{{ getOrderActionCount() }} actions</span>
    </div>

    <div class="card-content">
      <button
        class="op-btn op-export"
        @click="$emit('export-orders', platform)"
      >
        <div class="btn-icon">📥</div>
        <div class="btn-text">
          <h4>Export Orders</h4>
          <p>Download orders to Google Sheets</p>
        </div>
        <i class="pi pi-chevron-right"></i>
      </button>

      <!-- Shopee specific order actions -->
      <template v-if="isShopee()">
        <button class="op-btn op-wallet" @click="$emit('show-wallet-modal')">
          <div class="btn-icon">💰</div>
          <div class="btn-text">
            <h4>Wallet Report</h4>
            <p>View and export wallet statements</p>
          </div>
          <i class="pi pi-chevron-right"></i>
        </button>

        <button
          class="op-btn op-shipping"
          @click="$emit('show-shipping-modal')"
        >
          <div class="btn-icon">🚚</div>
          <div class="btn-text">
            <h4>Shipping Fee Report</h4>
            <p>Track and export shipping fees</p>
          </div>
          <i class="pi pi-chevron-right"></i>
        </button>
      </template>
    </div>
  </div>
</template>

<script lang="ts">
import { defineComponent } from "vue";
import { useOperationPlatform } from "./composables/useOperationPlatform";

export default defineComponent({
  name: "OrderManagementSection",
  props: {
    platform: {
      type: String,
      default: "shopee",
    },
  },
  emits: ["export-orders", "show-wallet-modal", "show-shipping-modal"],
  setup(props) {
    const { isShopee, getOrderActionCount } = useOperationPlatform(
      props.platform
    );

    return {
      isShopee,
      getOrderActionCount,
    };
  },
});
</script>

<style scoped>
.operation-card {
  background: white;
  border-radius: 12px;
  overflow: visible;
  box-shadow: 0 4px 16px rgba(0, 0, 0, 0.08);
  transition: all 0.3s cubic-bezier(0.4, 0, 0.2, 1);
  display: flex;
  flex-direction: column;
  min-height: 0;
}

.operation-card:hover {
  transform: translateY(-8px);
  box-shadow: 0 12px 28px rgba(0, 0, 0, 0.12);
}

.card-header {
  padding: 16px;
  background: linear-gradient(135deg, #f8f9fa, #eef2f8);
  border-bottom: 2px solid #e8eef5;
  display: flex;
  justify-content: space-between;
  align-items: center;
  flex-shrink: 0;
  min-height: 0;
}

.header-title {
  display: flex;
  align-items: center;
  gap: 12px;
}

.header-title i {
  font-size: 1.3rem;
  color: #4a90e2;
}

.header-title h3 {
  font-size: 1.1rem;
  font-weight: 700;
  color: #2c3e50;
  margin: 0;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.card-badge {
  background: #e8f0ff;
  color: #1e40af;
  padding: 6px 14px;
  border-radius: 20px;
  font-size: 0.85rem;
  font-weight: 600;
}

.card-content {
  padding: 12px;
  display: flex;
  flex-direction: column;
  gap: 10px;
  flex: 1;
  min-height: 0;
  overflow-y: auto;
}

.op-btn {
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 14px;
  background: #f8f9fa;
  border: 2px solid #e8eef5;
  border-radius: 10px;
  cursor: pointer;
  transition: all 0.3s cubic-bezier(0.4, 0, 0.2, 1);
  text-align: left;
  font-family: inherit;
  flex-shrink: 0;
  min-height: fit-content;
}

.op-btn:hover {
  background: #f0f5ff;
  border-color: #4a90e2;
  transform: translateX(6px);
  box-shadow: 0 4px 12px rgba(74, 144, 226, 0.15);
}

.op-btn:active {
  transform: translateX(4px);
}

.op-btn.op-wallet:hover {
  border-color: #ffc107;
  background: #fffbf0;
}

.op-btn.op-shipping:hover {
  border-color: #17a2b8;
  background: #f0f8f9;
}

.op-btn.op-export:hover {
  border-color: #ff6b6b;
  background: #fff5f5;
}

.btn-icon {
  font-size: 1.5rem;
  min-width: 32px;
  text-align: center;
  flex-shrink: 0;
}

.btn-text {
  flex: 1;
  min-width: 0;
}

.btn-text h4 {
  font-size: 0.9rem;
  font-weight: 700;
  color: #2c3e50;
  margin: 0 0 2px 0;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.btn-text p {
  font-size: 0.8rem;
  color: #4b5563; /* Improved from #7a8fa6 for WCAG AA contrast */
  margin: 0;
  white-space: normal;
  overflow: hidden;
  text-overflow: ellipsis;
  display: -webkit-box;
  -webkit-line-clamp: 2;
  line-clamp: 2;
  -webkit-box-orient: vertical;
}

.op-btn i {
  font-size: 1.2rem;
  color: #6b7280; /* Improved from #bcc3d4 for better contrast */
  transition: all 0.3s ease;
  flex-shrink: 0;
  min-width: 1.2rem;
}

.op-btn:hover i {
  color: #4a90e2;
  transform: translateX(4px);
}

@media (max-width: 768px) {
  .operation-card {
    padding: 16px;
    gap: 16px;
  }

  .op-btn {
    padding: 14px;
    gap: 12px;
  }

  .btn-icon {
    font-size: 1.5rem;
    min-width: 35px;
  }

  .btn-text h4 {
    font-size: 0.9rem;
  }

  .btn-text p {
    font-size: 0.8rem;
  }
}
</style>
