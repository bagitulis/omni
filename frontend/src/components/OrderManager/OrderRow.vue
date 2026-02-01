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
      <div class="product-col">
        <div v-for="(item, i) in order.items" :key="i" class="product-item">
          <div class="product-image">
            <img
              v-if="getProductImage(item)"
              :src="getProductImage(item)"
              :alt="item.product_name"
              @error="(e) => onImageError(e)"
              loading="lazy"
            />
            <div v-else class="image-placeholder">
              <i class="pi pi-image"></i>
            </div>
          </div>
          <div class="product-info">
            <div class="product-name">{{ item.product_name }}</div>
            <span class="product-qty">x{{ item.qty }}</span>
            <div v-if="item.variation_name" class="product-variant">
              Variasi: {{ item.variation_name }}
            </div>
          </div>
        </div>
      </div>

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
      <div class="action-col">
        <div class="action-buttons">
          <button
            v-if="showShipButton"
            @click="shipOrder"
            class="btn-action primary"
            title="Arrange Shipment"
          >
            Atur Pengiriman
          </button>
          <button
            v-else-if="showResponseButton"
            @click="shipOrder"
            class="btn-action primary"
            title="Respond"
          >
            Respon
          </button>
          <button
            v-if="showCancelButton"
            @click="cancelOrder"
            class="btn-action secondary"
            title="Cancel Order"
          >
            <i class="pi pi-times"></i>
          </button>
          <button
            @click="viewDetail"
            class="btn-action secondary"
            title="View Details"
          >
            <i class="pi pi-eye"></i>
          </button>
        </div>
      </div>
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
import { ref, computed } from "vue";

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
const imageErrors = ref<Set<string>>(new Set());

const showShipButton = computed(
  () => props.activeTab === "unprocess" && props.order.status !== "CANCELLED",
);

const showResponseButton = computed(
  () => props.activeTab === "unpaid" && props.order.status !== "CANCELLED",
);

const showCancelButton = computed(
  () =>
    ["unprocess", "unpaid"].includes(props.activeTab) &&
    props.order.status !== "CANCELLED",
);

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

const getProductImage = (item: OrderItem) => {
  const key = item.sku;
  if (imageErrors.value.has(key) || !item.product_image) {
    return null;
  }
  const url = item.product_image;
  if (props.order.platform === "shopee" && !url.includes("_tn")) {
    return url.replace(/\.(jpg|jpeg|png)$/i, "_tn.$1");
  }
  return url;
};

const onImageError = (e: Event) => {
  const img = e.target as HTMLImageElement;
  const parent = img.closest(".product-image");
  if (parent) {
    const item = parent.parentElement?.querySelector(".product-name-text");
    if (item) imageErrors.value.add(item.textContent || "");
  }
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

.order-row {
  background: var(--om-bg-primary);
  border: 1px solid var(--om-border);
  border-radius: var(--om-radius-md);
  overflow: hidden;
  margin-bottom: var(--om-spacing-md);
  animation: om-slideUp var(--om-transition-normal) ease-out;
  transition: box-shadow var(--om-transition-normal);
}

.order-row:hover {
  box-shadow: var(--om-shadow-md);
}

/* ============ BUYER HEADER ============ */
.buyer-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: var(--om-spacing-md);
  background: #f5f5f5;
  border-bottom: 1px solid var(--om-border);
  gap: var(--om-spacing-md);
  min-height: 56px;
}

.buyer-left {
  display: flex;
  align-items: center;
  gap: var(--om-spacing-sm);
  flex: 1;
  min-width: 0;
}

.buyer-avatar {
  width: 40px;
  height: 40px;
  border-radius: 50%;
  background: var(--om-primary-light);
  color: var(--om-primary);
  display: flex;
  align-items: center;
  justify-content: center;
  font-weight: 600;
  font-size: var(--om-font-sm);
  flex-shrink: 0;
}

.buyer-name {
  font-size: var(--om-font-sm);
  font-weight: 500;
  color: var(--om-text-primary);
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.chat-icon-btn {
  background: none;
  border: none;
  padding: 4px;
  cursor: pointer;
  color: var(--om-text-secondary);
  display: flex;
  align-items: center;
  justify-content: center;
  transition: color var(--om-transition-fast);
  font-size: 1rem;
  flex-shrink: 0;
}

.chat-icon-btn:hover {
  color: var(--om-primary);
}

.buyer-right {
  display: flex;
  align-items: center;
  gap: var(--om-spacing-xs);
  flex-shrink: 0;
}

.order-label {
  font-size: var(--om-font-xs);
  color: var(--om-text-secondary);
  white-space: nowrap;
}

.order-number {
  font-size: var(--om-font-sm);
  font-weight: 600;
  color: var(--om-text-primary);
  font-family: monospace;
  white-space: nowrap;
}

.copy-btn {
  background: none;
  border: none;
  padding: 4px;
  cursor: pointer;
  color: var(--om-text-secondary);
  display: flex;
  align-items: center;
  justify-content: center;
  transition: color var(--om-transition-fast);
  font-size: 0.875rem;
  flex-shrink: 0;
}

.copy-btn:hover {
  color: var(--om-primary);
}

/* ============ ORDER CONTENT GRID ============ */
.order-content {
  display: grid;
  grid-template-columns: 3fr 1.2fr 1fr 1.5fr 1.2fr 1fr;
  gap: var(--om-spacing-md);
  padding: var(--om-spacing-md);
  align-items: flex-start;
}

/* Product Column */
.product-col {
  display: flex;
  flex-direction: column;
  gap: var(--om-spacing-md);
}

.product-item {
  display: flex;
  gap: var(--om-spacing-sm);
  align-items: flex-start;
}

.product-image {
  width: 60px;
  height: 60px;
  border-radius: var(--om-radius-sm);
  overflow: hidden;
  background: var(--om-bg-secondary);
  flex-shrink: 0;
  display: flex;
  align-items: center;
  justify-content: center;
}

.product-image img {
  width: 100%;
  height: 100%;
  object-fit: cover;
}

.image-placeholder {
  width: 100%;
  height: 100%;
  display: flex;
  align-items: center;
  justify-content: center;
  color: var(--om-text-disabled);
  font-size: 1.5rem;
}

.product-info {
  display: flex;
  flex-direction: column;
  gap: 2px;
  flex: 1;
  min-width: 0;
}

.product-name {
  font-size: var(--om-font-sm);
  font-weight: 500;
  color: var(--om-text-primary);
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
  line-height: 1.4;
}

.product-qty {
  font-size: var(--om-font-xs);
  color: var(--om-text-secondary);
  font-weight: 500;
}

.product-variant {
  font-size: var(--om-font-xs);
  color: var(--om-text-secondary);
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

/* Amount Column */
.amount-col {
  display: flex;
  flex-direction: column;
  gap: 2px;
  justify-content: flex-start;
}

.amount-value {
  font-size: var(--om-font-base);
  font-weight: 600;
  color: var(--om-primary);
}

.payment-method {
  font-size: var(--om-font-xs);
  color: var(--om-text-secondary);
}

/* Status Column */
.status-col {
  display: flex;
  align-items: flex-start;
}

.status-badge {
  display: inline-flex;
  align-items: center;
  padding: var(--om-spacing-xs) var(--om-spacing-sm);
  font-size: var(--om-font-xs);
  font-weight: 600;
  border-radius: var(--om-radius-full);
}

/* Countdown Column */
.countdown-col {
  display: flex;
  align-items: flex-start;
}

.countdown-text {
  font-size: var(--om-font-sm);
  color: var(--om-text-primary);
  line-height: 1.4;
}

/* Shipping Column */
.shipping-col {
  display: flex;
  flex-direction: column;
  gap: 2px;
}

.shipping-info {
  display: flex;
  flex-direction: column;
  gap: 2px;
}

.carrier-name {
  font-size: var(--om-font-sm);
  font-weight: 500;
  color: var(--om-text-primary);
}

.shipping-type {
  font-size: var(--om-font-xs);
  color: var(--om-text-secondary);
}

.shipping-placeholder {
  font-size: var(--om-font-sm);
  color: var(--om-text-disabled);
}

/* Action Column */
.action-col {
  display: flex;
  align-items: flex-start;
}

.action-buttons {
  display: flex;
  gap: var(--om-spacing-xs);
  flex-wrap: wrap;
  width: 100%;
}

.btn-action {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  gap: var(--om-spacing-xs);
  padding: 6px 12px;
  font-size: var(--om-font-xs);
  font-weight: 500;
  border-radius: var(--om-radius-sm);
  cursor: pointer;
  transition: all var(--om-transition-fast);
  border: 1px solid transparent;
  white-space: nowrap;
  flex: 1;
  min-width: 60px;
}

.btn-action.primary {
  background: transparent;
  color: #ee4d2d;
  border: 1px solid #ee4d2d;
}

.btn-action.primary:hover:not(:disabled) {
  background: #fff0ed;
  color: #d73211;
  border-color: #d73211;
}

.btn-action.secondary {
  background: transparent;
  color: var(--om-text-secondary);
  border: 1px solid var(--om-border);
}

.btn-action.secondary:hover:not(:disabled) {
  background: var(--om-bg-secondary);
  color: var(--om-text-primary);
}

.btn-action:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

.btn-action i {
  font-size: 1rem;
}

/* ============ BUYER MESSAGE ============ */
.buyer-message {
  display: flex;
  align-items: center;
  gap: var(--om-spacing-sm);
  padding: var(--om-spacing-sm) var(--om-spacing-md);
  background: #fffbf0;
  border-top: 1px solid #fff1c0;
  font-size: var(--om-font-sm);
  color: var(--om-text-primary);
}

.buyer-message i {
  font-size: 0.875rem;
  flex-shrink: 0;
  color: #ff9800;
}

.message-text {
  flex: 1;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.message-link {
  color: var(--om-primary);
  text-decoration: none;
  font-weight: 500;
  cursor: pointer;
  white-space: nowrap;
  flex-shrink: 0;
}

.message-link:hover {
  text-decoration: underline;
}

/* ============ RESPONSIVE DESIGN ============ */
@media (max-width: 1200px) {
  .order-content {
    grid-template-columns: 3fr 1fr 0.8fr 1fr 1fr 1.2fr;
    gap: var(--om-spacing-sm);
    padding: var(--om-spacing-sm);
  }

  .btn-action {
    padding: 6px 8px;
    font-size: var(--om-font-xs);
  }

  .product-name {
    font-size: var(--om-font-xs);
  }

  .amount-value {
    font-size: var(--om-font-sm);
  }
}

@media (max-width: 768px) {
  .order-row {
    border-radius: var(--om-radius-sm);
  }

  .buyer-header {
    flex-direction: column;
    align-items: flex-start;
    gap: var(--om-spacing-sm);
    padding: var(--om-spacing-sm);
  }

  .buyer-left {
    width: 100%;
  }

  .buyer-right {
    width: 100%;
    justify-content: flex-start;
  }

  .order-content {
    display: flex;
    flex-direction: column;
    gap: var(--om-spacing-md);
    padding: var(--om-spacing-sm);
    border-bottom: none;
  }

  .product-col,
  .amount-col,
  .status-col,
  .countdown-col,
  .shipping-col,
  .action-col {
    border: 1px solid var(--om-border);
    border-radius: var(--om-radius-sm);
    padding: var(--om-spacing-sm);
    background: var(--om-bg-secondary);
  }

  .product-item {
    gap: var(--om-spacing-xs);
  }

  .product-image {
    width: 50px;
    height: 50px;
  }

  .action-buttons {
    flex-direction: column;
  }

  .btn-action {
    width: 100%;
    justify-content: flex-start;
  }

  .buyer-message {
    padding: var(--om-spacing-xs) var(--om-spacing-sm);
    font-size: var(--om-font-xs);
  }
}
</style>
