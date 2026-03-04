<template>
  <div class="modal-overlay" @click="$emit('close')">
    <div class="modal-content" @click.stop>
      <div class="modal-header">
        <h3>Product Details</h3>
        <button
          class="close-btn"
          @click="$emit('close')"
          aria-label="Close modal"
          title="Close"
        >
          <i class="pi pi-times" aria-hidden="true"></i>
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
              <i
                v-if="product.tag_kit"
                class="pi pi-check"
                style="color: #27ae60"
              ></i>
              <i v-else class="pi pi-times" style="color: #e74c3c"></i>
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
        <button class="btn btn-secondary" @click="$emit('close')">Close</button>
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
.modal-overlay {
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background: rgba(0, 0, 0, 0.5);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 1000;
}

.modal-content {
  background: white;
  border-radius: 12px;
  box-shadow: 0 10px 40px rgba(0, 0, 0, 0.2);
  max-width: 600px;
  width: 90%;
  max-height: 90vh;
  overflow-y: auto;
}

.modal-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 20px;
  border-bottom: 1px solid #dee2e6;
}

.modal-header h3 {
  margin: 0;
  color: #2c3e50;
}

.close-btn {
  background: none;
  border: none;
  font-size: 1.5rem;
  cursor: pointer;
  color: #7f8c8d;
  padding: 0;
  width: 30px;
  height: 30px;
  display: flex;
  align-items: center;
  justify-content: center;
  border-radius: 4px;
  transition: all 0.2s ease;
}

.close-btn:hover {
  background: #f0f0f0;
  color: #2c3e50;
}

.modal-body {
  padding: 20px;
}

.product-header {
  display: flex;
  gap: 15px;
  margin-bottom: 20px;
  padding-bottom: 15px;
  border-bottom: 1px solid #f0f0f0;
}

.product-info h4 {
  margin: 0 0 5px 0;
  color: #2c3e50;
}

.product-info .shop-id {
  margin: 0;
  color: #7f8c8d;
  font-size: 0.9rem;
}

.details-section {
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.detail-item {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 10px 0;
  border-bottom: 1px solid #f0f0f0;
}

.detail-item:last-child {
  border-bottom: none;
}

.label {
  font-weight: 600;
  color: #2c3e50;
  min-width: 120px;
}

.value {
  color: #7f8c8d;
  text-align: right;
  flex: 1;
}

.value.code {
  font-family: "Courier New", monospace;
  background: #f0f0f0;
  padding: 4px 8px;
  border-radius: 4px;
  color: #3498db;
}

.status-badge {
  padding: 6px 12px;
  border-radius: 20px;
  font-size: 0.85rem;
  font-weight: 500;
}

.status-badge.normal {
  background: #d4edda;
  color: #155724;
}

.status-badge.inactive {
  background: #f8d7da;
  color: #721c24;
}

.modal-footer {
  padding: 15px 20px;
  border-top: 1px solid #dee2e6;
  display: flex;
  justify-content: flex-end;
  gap: 10px;
}

.btn {
  padding: 8px 16px;
  border: none;
  border-radius: 6px;
  cursor: pointer;
  font-weight: 500;
  transition: all 0.3s ease;
}

.btn-secondary {
  background: #95a5a6;
  color: white;
}

.btn-secondary:hover {
  background: #7f8c8d;
}
</style>
