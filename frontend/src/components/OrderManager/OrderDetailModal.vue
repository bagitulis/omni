<template>
  <Teleport to="body">
    <div v-if="visible" class="o" @click.self="close">
      <div
        class="om"
        role="dialog"
        aria-modal="true"
        aria-labelledby="modal-title"
      >
        <div class="h">
          <h3 id="modal-title">Order Details</h3>
          <button @click="close" aria-label="Close">
            <Icon name="close" size="md" />
          </button>
        </div>

        <div class="b" v-if="order">
          <!-- Order Header Info -->
          <div class="info-group">
            <div class="flex-row">
              <span class="label">Order #</span>
              <span class="value font-mono">{{ order.order_no }}</span>
            </div>
            <div class="badges">
              <span class="badge" :class="order.platform">{{
                order.platform
              }}</span>
              <span class="badge status">{{ order.status }}</span>
            </div>
          </div>

          <!-- Buyer Info -->
          <div class="info-group">
            <div class="flex-row align-center">
              <Icon name="user" size="sm" class="text-secondary" />
              <span class="value ml-2">{{ order.buyer_username }}</span>
            </div>
            <div v-if="order.buyer_message" class="message">
              "{{ order.buyer_message }}"
            </div>
          </div>

          <!-- Products List -->
          <div class="products-list">
            <div
              v-for="(item, idx) in order.items"
              :key="idx"
              class="product-item"
            >
              <img
                v-if="item.product_image"
                :src="item.product_image"
                alt="Product"
                class="product-thumb"
              />
              <div class="product-details">
                <div class="product-name">{{ item.product_name }}</div>
                <div class="product-meta">
                  <span v-if="item.variation_name" class="variation">{{
                    item.variation_name
                  }}</span>
                  <span class="sku">SKU: {{ item.sku }}</span>
                </div>
                <div class="product-price">
                  x{{ item.quantity }}
                  <span v-if="item.price" class="price"
                    >{{ order.currency }} {{ item.price }}</span
                  >
                </div>
              </div>
            </div>
          </div>

          <!-- Shipping & Payment -->
          <div class="info-grid">
            <div class="info-card">
              <div class="card-header">
                <Icon name="truck" size="sm" />
                <span>Shipping</span>
              </div>
              <div class="card-body">
                <div>{{ order.shipping_carrier || "Not allocated" }}</div>
                <div v-if="order.shipping_type" class="text-sm text-secondary">
                  {{ order.shipping_type }}
                </div>
              </div>
            </div>
            <div class="info-card">
              <div class="card-header">
                <Icon name="shopping-bag" size="sm" />
                <span>Payment</span>
              </div>
              <div class="card-body">
                <div>{{ order.payment_method || "Unspecified" }}</div>
                <div class="total-amount">
                  {{ order.currency }} {{ order.total_amount }}
                </div>
              </div>
            </div>
          </div>
        </div>

        <div class="footer">
          <button @click="close" class="om-btn om-btn-secondary">Close</button>
        </div>
      </div>
    </div>
  </Teleport>
</template>

<script setup lang="ts">
import Icon from "@/components/ui/Icon.vue";

interface OrderItem {
  sku: string;
  product_name: string;
  variation_name?: string;
  quantity: number;
  price?: number;
  product_image?: string;
}

interface Order {
  order_no: string;
  platform: string;
  status: string;
  buyer_username: string;
  total_amount: number;
  currency: string;
  payment_method?: string;
  shipping_carrier?: string;
  shipping_type?: string;
  buyer_message?: string;
  items: OrderItem[];
}

defineProps<{ visible: boolean; order: Order | null }>();
const emit = defineEmits<{ close: [] }>();
const close = () => emit("close");
</script>

<style scoped>
@import "./OrderManager.theme.css";

.o {
  position: fixed;
  inset: 0;
  background: rgba(0, 0, 0, 0.5);
  z-index: 1000;
  display: flex;
  align-items: center;
  justify-content: center;
  animation: om-fadeIn var(--om-transition-fast);
}
.om {
  background: var(--om-bg-primary);
  border-radius: var(--om-radius-lg);
  box-shadow: var(--om-shadow-lg);
  width: 100%;
  max-width: 600px;
  max-height: 90vh;
  overflow: hidden;
  display: flex;
  flex-direction: column;
  animation: om-slideUp var(--om-transition-normal);
}
.h {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: var(--om-spacing-md) var(--om-spacing-lg);
  border-bottom: 1px solid var(--om-border);
}
.h h3 {
  font-size: var(--om-font-lg);
  font-weight: 600;
  margin: 0;
}
.h button {
  background: none;
  border: none;
  cursor: pointer;
  color: var(--om-text-secondary);
}
.b {
  padding: var(--om-spacing-lg);
  overflow-y: auto;
  flex: 1;
}
.footer {
  padding: var(--om-spacing-md) var(--om-spacing-lg);
  border-top: 1px solid var(--om-border);
  display: flex;
  justify-content: flex-end;
}

/* Internal Layout */
.info-group {
  margin-bottom: var(--om-spacing-md);
}
.flex-row {
  display: flex;
  align-items: center;
  justify-content: space-between;
}
.label {
  color: var(--om-text-secondary);
  font-size: var(--om-font-sm);
}
.value {
  font-weight: 500;
}
.font-mono {
  font-family: monospace;
}
.text-secondary {
  color: var(--om-text-secondary);
}
.ml-2 {
  margin-left: 0.5rem;
}
.badges {
  display: flex;
  gap: 0.5rem;
  margin-top: 0.25rem;
}
.badge {
  padding: 2px 8px;
  border-radius: 4px;
  font-size: 0.75rem;
  font-weight: 600;
  text-transform: uppercase;
  background: var(--om-bg-secondary);
}
.badge.shopee {
  background: #ee4d2d;
  color: white;
}
.badge.tiktok {
  background: #000;
  color: white;
}
.badge.lazada {
  background: #0f146d;
  color: white;
}
.message {
  background: var(--om-bg-secondary);
  padding: 0.5rem;
  border-radius: 4px;
  font-style: italic;
  margin-top: 0.5rem;
  font-size: 0.9em;
}

.products-list {
  border: 1px solid var(--om-border);
  border-radius: 6px;
  overflow: hidden;
  margin: 1rem 0;
}
.product-item {
  display: flex;
  gap: 1rem;
  padding: 0.75rem;
  border-bottom: 1px solid var(--om-border);
}
.product-item:last-child {
  border-bottom: none;
}
.product-thumb {
  width: 48px;
  height: 48px;
  object-fit: cover;
  border-radius: 4px;
  background: #eee;
}
.product-details {
  flex: 1;
}
.product-name {
  font-weight: 500;
  font-size: 0.95rem;
  line-height: 1.2;
}
.product-meta {
  font-size: 0.8rem;
  color: var(--om-text-secondary);
  margin-top: 2px;
}
.sku {
  margin-left: 0.5rem;
  font-family: monospace;
}
.product-price {
  margin-top: 4px;
  font-size: 0.9rem;
  font-weight: 600;
}

.info-grid {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 1rem;
}
.info-card {
  background: var(--om-bg-secondary);
  padding: 0.75rem;
  border-radius: 6px;
}
.card-header {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  font-weight: 600;
  margin-bottom: 0.5rem;
  font-size: 0.9rem;
}
.card-body {
  font-size: 0.9rem;
}
.total-amount {
  font-size: 1.1rem;
  font-weight: 700;
  color: var(--om-primary);
  margin-top: 2px;
}
</style>
