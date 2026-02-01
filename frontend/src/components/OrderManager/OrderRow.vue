<template>
  <div class="order-row">
    <!-- Buyer Info Header Row -->
    <div class="buyer-header">
      <div class="buyer-info-left">
        <div class="buyer-avatar">
          {{ getBuyerInitial(order.buyer_username) }}
        </div>
        <div class="buyer-details">
          <div class="buyer-name">{{ order.buyer_username }}</div>
          <div class="buyer-actions">
            <button class="chat-btn" title="Chat with buyer">
              <i class="pi pi-comments"></i>
            </button>
          </div>
        </div>
      </div>
      <div class="order-number-section">
        <span class="order-number-label">Order No.</span>
        <span class="order-number-value">{{ order.order_no }}</span>
        <button
          @click="copyOrderNumber"
          class="copy-btn"
          title="Copy Order No."
        >
          <i :class="copied ? 'pi pi-check' : 'pi pi-copy'"></i>
        </button>
      </div>
    </div>

    <!-- Products Grid Header (visual alignment guide) -->
    <div class="products-grid-container">
      <!-- Product Column (3fr) -->
      <div class="grid-column product-column">
        <div v-for="(item, i) in order.items" :key="i" class="product-cell">
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
            <div class="product-name-text">{{ item.product_name }}</div>
            <div v-if="item.variation_name" class="product-variant">
              {{ item.variation_name }}
            </div>
            <div class="product-qty-text">x{{ item.qty }}</div>
          </div>
        </div>
      </div>

      <!-- Amount Paid Column (1.2fr) -->
      <div class="grid-column amount-column">
        <div v-for="(item, i) in order.items" :key="i" class="amount-cell">
          <div v-if="i === 0" class="amount-value">
            {{ formatAmount(order.total_amount, order.currency) }}
          </div>
          <div v-if="i === 0 && order.payment_method" class="payment-method">
            {{ order.payment_method }}
          </div>
        </div>
      </div>

      <!-- Status Column (1fr) -->
      <div class="grid-column status-column">
        <div v-for="(item, i) in order.items" :key="i" class="status-cell">
          <div
            v-if="i === 0"
            :class="['status-badge', getStatusClass(order.status)]"
          >
            {{ formatStatus(order.status) }}
          </div>
        </div>
      </div>

      <!-- Countdown Column (1.2fr) -->
      <div class="grid-column countdown-column">
        <div v-for="(item, i) in order.items" :key="i" class="countdown-cell">
          <div v-if="i === 0 && order.countdown" class="countdown-text">
            {{ order.countdown }}
          </div>
          <div v-else-if="i === 0" class="countdown-text">-</div>
        </div>
      </div>

      <!-- Shipping Column (1.2fr) -->
      <div class="grid-column shipping-column">
        <div v-for="(item, i) in order.items" :key="i" class="shipping-cell">
          <div v-if="i === 0" class="shipping-info">
            <div v-if="order.shipping_carrier" class="carrier-name">
              {{ order.shipping_carrier }}
            </div>
            <div v-if="order.shipping_type" class="shipping-type">
              {{ order.shipping_type }}
            </div>
            <div v-if="!order.shipping_carrier" class="shipping-placeholder">
              -
            </div>
          </div>
        </div>
      </div>

      <!-- Action Column (1.5fr) -->
      <div class="grid-column action-column">
        <div v-for="(item, i) in order.items" :key="i" class="action-cell">
          <div v-if="i === 0" class="action-buttons">
            <button
              v-if="showShipButton"
              @click="shipOrder"
              class="om-btn om-btn-primary"
              title="Arrange Shipment"
            >
              <i class="pi pi-truck"></i>
              <span>Arrange</span>
            </button>
            <button
              v-if="showCancelButton"
              @click="cancelOrder"
              class="om-btn om-btn-secondary cancel-btn"
              title="Cancel Order"
            >
              <i class="pi pi-times"></i>
            </button>
            <button
              @click="viewDetail"
              class="om-btn om-btn-secondary"
              title="View Details"
            >
              <i class="pi pi-eye"></i>
              <span>Details</span>
            </button>
          </div>
        </div>
      </div>
    </div>

    <!-- Buyer Message Section (if exists) -->
    <div v-if="order.buyer_message" class="buyer-message">
      <i class="pi pi-bell"></i>
      <span class="message-text">Buyer message</span>
      <a href="#" class="message-link">Open</a>
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
  () =>
    ["unprocess", "unpaid"].includes(props.activeTab) &&
    props.order.status !== "CANCELLED",
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

/* Buyer Header */
.buyer-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: var(--om-spacing-md);
  background: var(--om-bg-secondary);
  border-bottom: 1px solid var(--om-border);
  gap: var(--om-spacing-md);
}

.buyer-info-left {
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
  font-size: var(--om-font-base);
  flex-shrink: 0;
}

.buyer-details {
  display: flex;
  align-items: center;
  gap: var(--om-spacing-md);
  min-width: 0;
}

.buyer-name {
  font-size: var(--om-font-sm);
  font-weight: 500;
  color: var(--om-text-primary);
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.buyer-actions {
  display: flex;
  gap: var(--om-spacing-xs);
  flex-shrink: 0;
}

.chat-btn {
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
}

.chat-btn:hover {
  color: var(--om-primary);
}

.order-number-section {
  display: flex;
  align-items: center;
  gap: var(--om-spacing-xs);
  flex-shrink: 0;
}

.order-number-label {
  font-size: var(--om-font-xs);
  color: var(--om-text-secondary);
  white-space: nowrap;
}

.order-number-value {
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
  align-items: center;
  justify-content: center;
  transition: color var(--om-transition-fast);
  font-size: 0.875rem;
  flex-shrink: 0;
}

.copy-btn:hover {
  color: var(--om-primary);
}

/* Products Grid Container */
.products-grid-container {
  display: grid;
  grid-template-columns: 3fr 1.2fr 1fr 1.2fr 1.2fr 1.5fr;
  gap: var(--om-spacing-md);
  padding: var(--om-spacing-md);
  border-bottom: 1px solid var(--om-border);
  align-items: start;
}

.grid-column {
  display: flex;
  flex-direction: column;
  gap: var(--om-spacing-sm);
}

/* Product Column */
.product-column {
  gap: var(--om-spacing-md);
}

.product-cell {
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

.product-name-text {
  font-size: var(--om-font-sm);
  font-weight: 500;
  color: var(--om-text-primary);
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.product-variant {
  font-size: var(--om-font-xs);
  color: var(--om-text-secondary);
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.product-qty-text {
  font-size: var(--om-font-xs);
  color: var(--om-text-secondary);
  font-weight: 500;
}

/* Amount Column */
.amount-column {
  justify-content: flex-start;
}

.amount-cell {
  display: flex;
  flex-direction: column;
  gap: 2px;
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
.status-column {
  justify-content: flex-start;
}

.status-cell {
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
.countdown-column {
  justify-content: flex-start;
}

.countdown-cell {
  display: flex;
  align-items: flex-start;
}

.countdown-text {
  font-size: var(--om-font-sm);
  color: var(--om-text-primary);
  font-weight: 500;
}

/* Shipping Column */
.shipping-column {
  justify-content: flex-start;
}

.shipping-cell {
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
.action-column {
  justify-content: flex-start;
}

.action-cell {
  display: flex;
  align-items: flex-start;
}

.action-buttons {
  display: flex;
  gap: var(--om-spacing-xs);
  flex-wrap: wrap;
}

.om-btn {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  gap: var(--om-spacing-xs);
  padding: var(--om-spacing-xs) var(--om-spacing-sm);
  font-size: var(--om-font-xs);
  font-weight: 500;
  border-radius: var(--om-radius-sm);
  cursor: pointer;
  transition: all var(--om-transition-fast);
  border: none;
  white-space: nowrap;
}

.om-btn-primary {
  background: var(--om-primary);
  color: white;
}

.om-btn-primary:hover:not(:disabled) {
  background: var(--om-primary-hover);
}

.om-btn-secondary {
  background: var(--om-bg-secondary);
  color: var(--om-text-primary);
  border: 1px solid var(--om-border);
}

.om-btn-secondary:hover:not(:disabled) {
  background: var(--om-bg-hover);
}

.om-btn:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

.cancel-btn {
  color: var(--om-status-cancelled);
  border-color: var(--om-status-cancelled);
}

.cancel-btn:hover:not(:disabled) {
  background: #ffebee;
}

.om-btn span {
  display: inline;
}

/* Buyer Message Section */
.buyer-message {
  display: flex;
  align-items: center;
  gap: var(--om-spacing-sm);
  padding: var(--om-spacing-sm) var(--om-spacing-md);
  background: #fff9c4;
  border-top: 1px solid #fff59d;
  font-size: var(--om-font-sm);
  color: var(--om-text-primary);
}

.buyer-message i {
  font-size: 0.875rem;
  flex-shrink: 0;
}

.message-text {
  flex: 1;
}

.message-link {
  color: var(--om-primary);
  text-decoration: none;
  font-weight: 500;
  cursor: pointer;
  white-space: nowrap;
}

.message-link:hover {
  text-decoration: underline;
}

/* Responsive Design */
@media (max-width: 1200px) {
  .products-grid-container {
    grid-template-columns: 3fr 1fr 0.8fr 1fr 1fr 1.2fr;
    gap: var(--om-spacing-sm);
    padding: var(--om-spacing-sm);
  }

  .om-btn span {
    display: none;
  }

  .product-name-text {
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

  .buyer-info-left {
    width: 100%;
  }

  .order-number-section {
    width: 100%;
    justify-content: flex-start;
  }

  .products-grid-container {
    display: flex;
    flex-direction: column;
    gap: var(--om-spacing-md);
    padding: var(--om-spacing-sm);
    border-bottom: 1px solid var(--om-border);
  }

  .grid-column {
    border: 1px solid var(--om-border);
    border-radius: var(--om-radius-sm);
    padding: var(--om-spacing-sm);
  }

  .product-cell {
    gap: var(--om-spacing-xs);
  }

  .product-image {
    width: 50px;
    height: 50px;
  }

  .action-buttons {
    width: 100%;
    flex-direction: column;
  }

  .om-btn {
    width: 100%;
    justify-content: flex-start;
  }

  .om-btn span {
    display: inline;
  }

  .buyer-message {
    padding: var(--om-spacing-xs) var(--om-spacing-sm);
    font-size: var(--om-font-xs);
  }
}
</style>
