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
            <Icon :name="copied ? 'check' : 'copy'" size="sm" />
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
          <Icon name="dollar" size="sm" /><span>{{
            order.payment_method
          }}</span>
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
          <Icon name="truck" size="sm" /><span>Arrange Shipment</span>
        </button>
        <button
          v-if="showCancelButton"
          @click="cancelOrder"
          class="om-btn om-btn-secondary cancel-btn"
        >
          <Icon name="close" size="sm" /><span>Cancel Order</span>
        </button>
        <button @click="viewDetail" class="om-btn om-btn-secondary">
          <Icon name="eye" size="sm" /><span>View Details</span>
        </button>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed } from "vue";
import OrderProductItem from "./OrderProductItem.vue";
import Icon from "@/components/ui/Icon.vue";
interface OrderItem {
  sku: string;
  product_name: string;
  variation_name?: string;
  quantity: number; // Backend returns snake_case
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
const formatStatus = (s: string) => {
  if (!s) return "-";
  const upper = s.toUpperCase();
  const map: Record<string, string> = {
    UNPAID: "Unpaid",
    READY_TO_SHIP: "Ready to Ship",
    SHIPPED: "Shipped",
    COMPLETED: "Completed",
    CANCELLED: "Cancelled",
    IN_CANCEL: "Cancelling",
    TO_CONFIRM_RECEIVE: "To Confirm Receive",
    TO_RETURN: "To Return",
  };
  if (map[upper]) return map[upper];
  // Generic fallback: snake_case -> Title Case
  return s
    .replace(/_/g, " ")
    .toLowerCase()
    .replace(/\b\w/g, (l) => l.toUpperCase());
};
const getStatusClass = (s: string) => {
  if (!s) return "om-status-toship";
  const map: Record<string, string> = {
    UNPAID: "om-status-unpaid",
    READY_TO_SHIP: "om-status-toship",
    SHIPPED: "om-status-shipped",
    COMPLETED: "om-status-completed",
    CANCELLED: "om-status-cancelled",
    IN_CANCEL: "om-status-cancelled",
  };
  return map[s.toUpperCase()] || "om-status-toship";
};
const getBuyerInitial = (u: string) => (u ? u.charAt(0).toUpperCase() : "?");
const formatAmount = (amt: number, cur: string) => {
  if (!amt) return "-";
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
/* 
 * Styles are centralized in OrderManager.theme.css 
 * to keep the component under 300 lines.
 */
</style>
