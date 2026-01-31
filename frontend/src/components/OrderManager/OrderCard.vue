<template>
  <div class="order-card om-card">
    <div class="card-header">
      <div class="header-left">
        <span
          :class="[
            'om-platform-badge',
            `om-platform-${order.platform.toLowerCase()}`,
          ]"
        >
          {{ formatPlatform(order.platform) }}
        </span>
        <div class="order-number">
          <span class="order-label">Order No.</span>
          <span class="order-value">{{ order.order_no }}</span>
          <button
            @click="copyOrderNumber"
            class="copy-btn"
            title="Copy Order No."
          >
            <i :class="copied ? 'pi pi-check' : 'pi pi-copy'"></i>
          </button>
        </div>
      </div>
      <span :class="['om-status-badge', getStatusClass(order.status)]">
        {{ formatStatus(order.status) }}
      </span>
    </div>
    <div class="buyer-info">
      <div class="buyer-avatar">
        {{ getBuyerInitial(order.buyer_username) }}
      </div>
      <span class="buyer-name">{{ order.buyer_username }}</span>
    </div>
    <div class="products-list">
      <OrderProductItem
        v-for="(item, i) in order.items"
        :key="i"
        :product="item"
        :platform="order.platform"
      />
    </div>
    <div class="card-footer">
      <div class="footer-info">
        <div v-if="order.payment_method" class="payment-method">
          <i class="pi pi-credit-card"></i
          ><span>{{ order.payment_method }}</span>
        </div>
        <div class="total-amount">
          <span class="total-label">Total:</span>
          <span class="total-value">{{
            formatAmount(order.total_amount, order.currency)
          }}</span>
        </div>
      </div>
      <div class="footer-actions">
        <button
          v-if="showShipButton"
          @click="shipOrder"
          class="om-btn om-btn-primary"
        >
          <i class="pi pi-truck"></i><span>Arrange Shipment</span>
        </button>
        <button
          v-if="showCancelButton"
          @click="cancelOrder"
          class="om-btn om-btn-secondary cancel-btn"
        >
          <i class="pi pi-times"></i><span>Cancel Order</span>
        </button>
        <button @click="viewDetail" class="om-btn om-btn-secondary">
          <i class="pi pi-eye"></i><span>View Details</span>
        </button>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed } from "vue";
import OrderProductItem from "./OrderProductItem.vue";
interface OrderItem {
  sku: string;
  product_name: string;
  variation_name?: string;
  qty: number;
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
  items: OrderItem[];
}
interface Props {
  order: Order;
  activeTab: string;
}
const props = defineProps<Props>();
const emit = defineEmits<{
  "ship-order": [Order];
  "cancel-order": [Order];
  "view-detail": [Order];
  "copy-order-number": [string];
}>();
const copied = ref(false);
const showShipButton = computed(
  () =>
    ["unprocess", "unpaid"].includes(props.activeTab) &&
    props.order.status !== "CANCELLED",
);
const showCancelButton = computed(
  () =>
    ["unprocess", "unpaid"].includes(props.activeTab) &&
    props.order.status !== "CANCELLED",
);
const formatPlatform = (p: string) =>
  ({ shopee: "Shopee", lazada: "Lazada", tiktok: "TikTok" })[p.toLowerCase()] ||
  p;
const formatStatus = (s: string) =>
  ({
    UNPAID: "Unpaid",
    READY_TO_SHIP: "To Ship",
    SHIPPED: "Shipped",
    COMPLETED: "Completed",
    CANCELLED: "Cancelled",
    IN_CANCEL: "Cancelling",
  })[s.toUpperCase()] || s;
const getStatusClass = (s: string) =>
  ({
    UNPAID: "om-status-unpaid",
    READY_TO_SHIP: "om-status-toship",
    SHIPPED: "om-status-shipped",
    COMPLETED: "om-status-completed",
    CANCELLED: "om-status-cancelled",
  })[s.toUpperCase()] || "om-status-toship";
const getBuyerInitial = (u: string) => (u ? u.charAt(0).toUpperCase() : "?");
const formatAmount = (amt: number, cur: string) => {
  const locales: Record<string, string> = {
    IDR: "id-ID",
    MYR: "ms-MY",
    SGD: "en-SG",
    PHP: "en-PH",
    THB: "th-TH",
    VND: "vi-VN",
  };
  return new Intl.NumberFormat(locales[cur] || "id-ID", {
    style: "currency",
    currency: cur || "IDR",
    minimumFractionDigits: 0,
  }).format(amt);
};
const copyOrderNumber = async () => {
  try {
    await navigator.clipboard.writeText(props.order.order_no);
    copied.value = true;
    emit("copy-order-number", props.order.order_no);
    setTimeout(() => {
      copied.value = false;
    }, 2000);
  } catch (e) {
    console.error(e);
  }
};
const shipOrder = () => emit("ship-order", props.order);
const cancelOrder = () => emit("cancel-order", props.order);
const viewDetail = () => emit("view-detail", props.order);
</script>

<style scoped>
@import "./OrderManager.theme.css";
.order-card {
  padding: var(--om-spacing-md);
  animation: om-slideUp var(--om-transition-normal) ease-out;
}
.card-header {
  display: flex;
  justify-content: space-between;
  align-items: flex-start;
  margin-bottom: var(--om-spacing-md);
  padding-bottom: var(--om-spacing-sm);
  border-bottom: 1px solid var(--om-border);
}
.header-left {
  display: flex;
  align-items: center;
  gap: var(--om-spacing-md);
  flex-wrap: wrap;
}
.order-number {
  display: flex;
  align-items: center;
  gap: var(--om-spacing-xs);
}
.order-label {
  font-size: var(--om-font-xs);
  color: var(--om-text-secondary);
}
.order-value {
  font-size: var(--om-font-sm);
  font-weight: 600;
  color: var(--om-text-primary);
  font-family: monospace;
}
.copy-btn {
  background: none;
  border: none;
  padding: 4px;
  cursor: pointer;
  color: var(--om-text-secondary);
  display: flex;
  transition: color var(--om-transition-fast);
}
.copy-btn:hover {
  color: var(--om-primary);
}
.buyer-info {
  display: flex;
  align-items: center;
  gap: var(--om-spacing-sm);
  margin-bottom: var(--om-spacing-md);
}
.buyer-avatar {
  width: 32px;
  height: 32px;
  border-radius: 50%;
  background: var(--om-primary-light);
  color: var(--om-primary);
  display: flex;
  align-items: center;
  justify-content: center;
  font-weight: 600;
  font-size: var(--om-font-sm);
}
.buyer-name {
  font-size: var(--om-font-sm);
  color: var(--om-text-primary);
}
.products-list {
  margin-bottom: var(--om-spacing-md);
}
.card-footer {
  display: flex;
  justify-content: space-between;
  padding-top: var(--om-spacing-sm);
  border-top: 1px solid var(--om-border);
  flex-wrap: wrap;
  gap: var(--om-spacing-md);
}
.footer-info {
  display: flex;
  align-items: center;
  gap: var(--om-spacing-lg);
}
.payment-method {
  display: flex;
  align-items: center;
  gap: var(--om-spacing-xs);
  font-size: var(--om-font-sm);
  color: var(--om-text-secondary);
}
.total-amount {
  display: flex;
  align-items: center;
  gap: var(--om-spacing-xs);
}
.total-label {
  font-size: var(--om-font-sm);
  color: var(--om-text-secondary);
}
.total-value {
  font-size: var(--om-font-lg);
  font-weight: 700;
  color: var(--om-primary);
}
.footer-actions {
  display: flex;
  gap: var(--om-spacing-sm);
}
.cancel-btn {
  color: var(--om-status-cancelled);
  border-color: var(--om-status-cancelled);
}
.cancel-btn:hover:not(:disabled) {
  background: #ffebee;
}
@media (max-width: 768px) {
  .card-header {
    flex-direction: column;
    gap: var(--om-spacing-sm);
  }
  .card-footer {
    flex-direction: column;
    align-items: stretch;
  }
  .footer-info {
    justify-content: space-between;
  }
  .footer-actions {
    flex-direction: column;
  }
  .footer-actions .om-btn {
    width: 100%;
  }
}
</style>
