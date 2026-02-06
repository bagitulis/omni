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
          <Icon name="chat" size="sm" />
        </button>
      </div>
      <div class="buyer-right">
        <span class="order-label">Order Number</span>
        <span class="order-number">{{ order.order_no }}</span>
        <!-- Copy Button moved closer for better UX -->
        <button
          @click="copyOrderNumber"
          class="copy-btn"
          title="Copy Order Number"
        >
          <Icon :name="copied ? 'check' : 'copy'" size="sm" />
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
        <div :class="['countdown-text', countdownClass]">
          {{ countdownText }}
        </div>
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
        :order="order"
        @ship-order="shipOrder"
        @cancel-order="cancelOrder"
        @view-detail="viewDetail"
      />
    </div>

    <!-- Buyer Message (optional, yellow background) -->
    <div v-if="order.buyer_message" class="buyer-message">
      <Icon name="bell" size="sm" />
      <span class="message-text">{{ order.buyer_message }}</span>
      <a href="#" class="message-link">View</a>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, onUnmounted } from "vue";
import OrderProductCell, { type OrderItem } from "./OrderProductCell.vue";
import OrderActionCell from "./OrderActionCell.vue";
import Icon from "@/components/ui/Icon.vue";

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
  ship_by_date?: number; // Unix timestamp for shipping deadline
  countdown?: string; // Legacy: pre-calculated countdown string
  buyer_message?: string;
  items: OrderItem[];
  [key: string]: any;
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
const now = ref(Date.now());
let countdownInterval: ReturnType<typeof setInterval> | null = null;

// Calculate countdown from ship_by_date timestamp
const countdownText = computed(() => {
  // If pre-calculated countdown exists, use it
  if (props.order.countdown) return props.order.countdown;

  // If no ship_by_date, show dash
  if (!props.order.ship_by_date) return "-";

  const deadline = props.order.ship_by_date * 1000; // Convert to ms
  const diff = deadline - now.value;

  // If already passed
  if (diff <= 0) return "Overdue";

  // Calculate time components
  const hours = Math.floor(diff / (1000 * 60 * 60));
  const minutes = Math.floor((diff % (1000 * 60 * 60)) / (1000 * 60));

  if (hours >= 24) {
    const days = Math.floor(hours / 24);
    const remainingHours = hours % 24;
    return `${days}d ${remainingHours}h`;
  }

  return `${hours}h ${minutes}m`;
});

// Countdown urgency class for styling
const countdownClass = computed(() => {
  if (!props.order.ship_by_date) return "";

  const deadline = props.order.ship_by_date * 1000;
  const diff = deadline - now.value;

  if (diff <= 0) return "countdown-overdue";
  if (diff <= 2 * 60 * 60 * 1000) return "countdown-urgent"; // < 2 hours
  if (diff <= 6 * 60 * 60 * 1000) return "countdown-warning"; // < 6 hours
  return "";
});

// Start countdown timer on mount
onMounted(() => {
  if (props.order.ship_by_date) {
    countdownInterval = setInterval(() => {
      now.value = Date.now();
    }, 60000); // Update every minute
  }
});

// Cleanup on unmount
onUnmounted(() => {
  if (countdownInterval) {
    clearInterval(countdownInterval);
  }
});

const getBuyerInitial = (username: string) =>
  username ? username.charAt(0).toUpperCase() : "?";

const formatStatus = (s: string) => {
  if (!s) return "-";

  const upper = s.toUpperCase();
  const map: Record<string, string> = {
    UNPAID: "Unpaid",
    READY_TO_SHIP: "Ready to Ship",
    toship: "To Ship", // Handle "toship" explicitly if needed, though generic handles it
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
  if (!s) return "om-status-toship"; // Default fallback
  const map: Record<string, string> = {
    UNPAID: "om-status-unpaid",
    READY_TO_SHIP: "om-status-toship",
    PROCESSED: "om-status-shipped", // Added mapping
    SHIPPED: "om-status-shipped",
    COMPLETED: "om-status-completed",
    CANCELLED: "om-status-cancelled",
    IN_CANCEL: "om-status-cancelled",
    TO_CONFIRM_RECEIVE: "om-status-shipped", // Treat as shipped
    TO_RETURN: "om-status-cancelled", // Treat as cancelled type
  };
  return map[s.toUpperCase()] || "om-status-toship";
};

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
@import "./styles/variables.css";
@import "./styles/components.css";
@import "./styles/layout.css";
</style>
