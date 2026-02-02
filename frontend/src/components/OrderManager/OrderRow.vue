<template>
  <div class="order-row">
    <!-- Buyer Header Row (gray background #f5f5f5) -->
    <div class="buyer-header">
      <div class="buyer-left">
        <div class="buyer-avatar">
          {{ getBuyerInitial(order.buyer_username) }}
        </div>
        <span class="buyer-name">{{ order.buyer_username }}</span>
        <button class="chat-icon-btn" title="Chat with buyer">
          <i class="pi pi-envelope"></i>
        </button>
      </div>
      <div class="buyer-right">
        <span class="order-label">No. Pesanan</span>
        <span class="order-number">{{ order.order_no }}</span>
        <button
          @click="copyOrderNumber"
          class="copy-btn"
          title="Copy Order Number"
        >
          <i :class="copied ? 'pi pi-check' : 'pi pi-copy'"></i>
        </button>
      </div>
    </div>

    <!-- Order Content Grid - 6 columns -->
    <div class="order-content">
      <!-- Product Column (3fr) -->
      <OrderProductCell :items="order.items" :platform="order.platform" />

      <!-- Amount Column (1.2fr) -->
      <div class="amount-col">
        <div class="amount-value">
          {{ formatAmount(order.total_amount, order.currency) }}
        </div>
        <div v-if="order.payment_method" class="payment-method">
          {{ order.payment_method }}
        </div>
      </div>

      <!-- Status Column (1fr) -->
      <div class="status-col">
        <span :class="['status-badge', getStatusClass(order.status)]">
          {{ formatStatus(order.status) }}
        </span>
      </div>

      <!-- Countdown Column (1.5fr) -->
      <div class="countdown-col">
        <div v-if="order.countdown" class="countdown-text">
          {{ order.countdown }}
        </div>
        <div v-else class="countdown-text">-</div>
      </div>

      <!-- Shipping Column (1.2fr) -->
      <div class="shipping-col">
        <div v-if="order.shipping_carrier" class="shipping-info">
          <div class="carrier-name">{{ order.shipping_carrier }}</div>
          <div v-if="order.shipping_type" class="shipping-type">
            {{ order.shipping_type }}
          </div>
        </div>
        <div v-else class="shipping-placeholder">-</div>
      </div>

      <!-- Action Column (1fr) -->
      <OrderActionCell
        :active-tab="activeTab"
        :order-status="order.status"
        @ship-order="shipOrder"
        @cancel-order="cancelOrder"
        @view-detail="viewDetail"
      />
    </div>

    <!-- Buyer Message (optional, yellow background) -->
    <div v-if="order.buyer_message" class="buyer-message">
      <i class="pi pi-bell"></i>
      <span class="message-text">{{ order.buyer_message }}</span>
      <a href="#" class="message-link">Buka</a>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref } from "vue";
import OrderProductCell, { type OrderItem } from "./OrderProductCell.vue";
import OrderActionCell from "./OrderActionCell.vue";

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
  countdown?: string;
  buyer_message?: string;
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

const getBuyerInitial = (username: string) =>
  username ? username.charAt(0).toUpperCase() : "?";

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

/* 
 * Styles are now centralized in OrderManager.theme.css 
 * to reduce file size and maintain consistency.
 * 
 * This keeps the component under 300 lines (currently ~180 lines)
 * satisfying the style guide requirements.
 */
</style>
