<template>
  <div class="modal-overlay" @click="$emit('close')">
    <div class="modal-content" @click.stop>
      <div class="modal-header">
        <h3>Product Details</h3>
        <button
          class="modal-close-btn"
          @click="$emit('close')"
          aria-label="Close modal"
          title="Close"
        >
          <svg
            width="20"
            height="20"
            fill="none"
            stroke="currentColor"
            viewBox="0 0 24 24"
          >
            <path
              stroke-linecap="round"
              stroke-linejoin="round"
              stroke-width="2"
              d="M6 18L18 6M6 6l12 12"
            ></path>
          </svg>
        </button>
      </div>

      <div class="modal-body">
        <div class="product-header">
          <div class="product-info">
            <h4>Item ID: {{ product.item_id }}</h4>
            <p class="shop-id">Shop ID: {{ product.shop_id }}</p>
          </div>
        </div>

        <div class="details-section">
          <div class="detail-item">
            <span class="label">ID:</span>
            <span class="value">{{ product.id }}</span>
          </div>
          <div class="detail-item">
            <span class="label">Item ID:</span>
            <span class="value code">{{ product.item_id }}</span>
          </div>
          <div class="detail-item">
            <span class="label">Shop ID:</span>
            <span class="value">{{ product.shop_id }}</span>
          </div>
          <div class="detail-item">
            <span class="label">Status:</span>
            <span
              class="status-badge"
              :class="product.item_status?.toLowerCase() || 'normal'"
            >
              {{ product.item_status || "NORMAL" }}
            </span>
          </div>
          <div class="detail-item">
            <span class="label">Tag Kit:</span>
            <span class="value">
              <span v-if="product.tag_kit" class="tag-yes">Yes</span>
              <span v-else class="tag-no">No</span>
            </span>
          </div>
          <div class="detail-item">
            <span class="label">Update Time:</span>
            <span class="value">{{ formatDate(product.update_time_str) }}</span>
          </div>
          <div class="detail-item">
            <span class="label">Created At:</span>
            <span class="value">{{ formatDate(product.created_at) }}</span>
          </div>
          <div class="detail-item">
            <span class="label">Updated At:</span>
            <span class="value">{{ formatDate(product.updated_at) }}</span>
          </div>
        </div>
      </div>

      <div class="modal-footer">
        <button class="modal-btn modal-btn-secondary" @click="$emit('close')">
          Close
        </button>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { formatDate } from "@/utils/helpers";

interface Product {
  [key: string]: any;
}

interface Props {
  product: Product;
}

defineProps<Props>();
defineEmits<{
  close: [];
}>();
</script>

<style scoped>
/* ProductDetailModal - Unified Tokopedia Style */
.modal-overlay {
  position: fixed;
  inset: 0;
  background: rgba(0, 0, 0, 0.5);
  backdrop-filter: blur(4px);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 9999;
  padding: 16px;
  animation: fadeIn 200ms ease-out;
}

.modal-content {
  background: #ffffff;
  border-radius: 12px;
  box-shadow: 0 20px 60px rgba(0, 0, 0, 0.15);
  max-width: 500px;
  width: 100%;
  max-height: 90vh;
  overflow: hidden;
  display: flex;
  flex-direction: column;
  animation: slideUp 250ms cubic-bezier(0.4, 0, 0.2, 1);
}

.modal-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 16px 20px;
  border-bottom: 1px solid #e0e0e0;
  background: #ffffff;
  flex-shrink: 0;
}

.modal-header h3 {
  margin: 0;
  font-size: 1.125rem;
  font-weight: 700;
  color: #212121;
}

.modal-close-btn {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 32px;
  height: 32px;
  background: transparent;
  border: none;
  border-radius: 6px;
  color: #6c727c;
  cursor: pointer;
  transition: all 150ms ease;
  padding: 0;
}

.modal-close-btn:hover {
  background: #f3f4f5;
  color: #212121;
}

.modal-close-btn:focus-visible {
  outline: 2px solid #03ac0e;
  outline-offset: 2px;
}

.modal-body {
  padding: 20px;
  overflow-y: auto;
  flex: 1;
  background: #ffffff;
}

.product-header {
  display: flex;
  gap: 12px;
  margin-bottom: 20px;
  padding-bottom: 16px;
  border-bottom: 1px solid #e0e0e0;
}

.product-info h4 {
  margin: 0 0 4px 0;
  font-size: 1rem;
  font-weight: 700;
  color: #212121;
}

.product-info .shop-id {
  margin: 0;
  color: #6c727c;
  font-size: 0.875rem;
}

.details-section {
  display: flex;
  flex-direction: column;
  gap: 0;
}

.detail-item {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 12px 0;
  border-bottom: 1px solid #f3f4f5;
}

.detail-item:last-child {
  border-bottom: none;
}

.label {
  font-weight: 600;
  color: #212121;
  font-size: 0.875rem;
}

.value {
  color: #6c727c;
  text-align: right;
  font-size: 0.875rem;
}

.value.code {
  font-family: monospace;
  background: #f3f4f5;
  padding: 4px 8px;
  border-radius: 4px;
  color: #03ac0e;
  font-weight: 600;
}

.status-badge {
  padding: 4px 12px;
  border-radius: 16px;
  font-size: 0.75rem;
  font-weight: 700;
  text-transform: uppercase;
}

.status-badge.normal {
  background: #e5f9e6;
  color: #03ac0e;
}

.status-badge.inactive {
  background: #ffebee;
  color: #d32f2f;
}

.tag-yes {
  color: #03ac0e;
  font-weight: 600;
}

.tag-no {
  color: #9fa6b0;
}

.modal-footer {
  padding: 16px 20px;
  border-top: 1px solid #e0e0e0;
  display: flex;
  justify-content: flex-end;
  gap: 12px;
  background: #ffffff;
  flex-shrink: 0;
}

.modal-btn {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  gap: 8px;
  padding: 10px 20px;
  font-size: 0.875rem;
  font-weight: 600;
  border-radius: 8px;
  cursor: pointer;
  transition: all 150ms ease;
  border: 1px solid #e0e0e0;
}

.modal-btn-secondary {
  background: #ffffff;
  color: #212121;
  border-color: #e0e0e0;
}

.modal-btn-secondary:hover {
  background: #f5f5f5;
  border-color: #bdbdbd;
}

.modal-btn:focus-visible {
  outline: 2px solid #03ac0e;
  outline-offset: 2px;
}

/* Animations */
@keyframes fadeIn {
  from {
    opacity: 0;
  }
  to {
    opacity: 1;
  }
}

@keyframes slideUp {
  from {
    opacity: 0;
    transform: translateY(20px);
  }
  to {
    opacity: 1;
    transform: translateY(0);
  }
}

/* Responsive */
@media (max-width: 640px) {
  .modal-overlay {
    padding: 0;
  }

  .modal-content {
    max-width: 100%;
    max-height: 100%;
    height: 100%;
    border-radius: 0;
  }

  .modal-header {
    padding: 12px 16px;
  }

  .modal-body {
    padding: 16px;
  }

  .modal-footer {
    padding: 12px 16px;
  }
}

/* Accessibility */
@media (prefers-reduced-motion: reduce) {
  .modal-overlay,
  .modal-content {
    animation: none;
  }
}
</style>
