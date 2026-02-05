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
          <!-- Order Info Card -->
          <div class="card order-info-card">
            <div class="order-no font-mono">#{{ order.order_no }}</div>
            <div class="badges-row">
              <span class="badge" :class="order.platform.toLowerCase()">{{
                order.platform
              }}</span>
              <span class="badge status">{{ order.status }}</span>
            </div>
            <div class="buyer-row">
              <Icon name="user" size="sm" class="text-secondary" />
              <span class="buyer-name">{{ order.buyer_username }}</span>
            </div>
            <div v-if="order.buyer_message" class="message">
              "{{ order.buyer_message }}"
            </div>
          </div>

          <!-- Product Details Card -->
          <div class="card product-card">
            <div class="section-header">Product Details</div>
            <div class="products-list">
              <div
                v-for="(item, idx) in order.items"
                :key="idx"
                class="product-item"
              >
                <div class="product-thumb-wrapper">
                  <img
                    v-if="item.product_image"
                    :src="item.product_image"
                    alt="Product"
                    class="product-thumb"
                  />
                  <div v-else class="product-thumb-placeholder"></div>
                </div>
                <div class="product-details">
                  <div class="product-name">{{ item.product_name }}</div>
                  <div class="product-meta">
                    <span v-if="item.variation_name" class="variation"
                      >Variant: {{ item.variation_name }}</span
                    >
                    <span class="sku">SKU: {{ item.sku }}</span>
                  </div>
                  <div class="product-price-row">
                    <span class="quantity">{{ item.quantity }} x</span>
                    <span v-if="item.price" class="price"
                      >{{ order.currency }} {{ item.price }}</span
                    >
                  </div>
                </div>
              </div>
            </div>
          </div>

          <!-- Shipping & Payment Grid -->
          <div class="info-grid">
            <!-- Shipping Card -->
            <div class="card shipping-card">
              <div class="card-header">
                <Icon name="truck" size="sm" />
                <span>Shipping</span>
              </div>
              <div class="card-body">
                <div class="info-value font-medium">
                  {{ order.shipping_carrier || "Not allocated" }}
                </div>
                <div v-if="order.shipping_type" class="text-sm text-secondary">
                  {{ order.shipping_type }}
                </div>
                <div
                  v-if="order.ship_by_date"
                  class="deadline-section"
                  :class="countdownClass"
                >
                  <div class="deadline-row">
                    <span class="text-xs text-secondary">Ship by:</span>
                    <span class="font-medium text-sm">{{
                      formatDeadline(order.ship_by_date)
                    }}</span>
                  </div>
                  <div class="deadline-timer">
                    <Icon name="clock" size="xs" />
                    <span>{{ countdownText }}</span>
                  </div>
                </div>
              </div>
            </div>

            <!-- Payment Card -->
            <div class="card payment-card">
              <div class="card-header">
                <Icon name="shopping-bag" size="sm" />
                <span>Payment</span>
              </div>
              <div class="card-body">
                <div class="info-value font-medium">
                  {{ order.payment_method || "Unspecified" }}
                </div>
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
import { ref, computed, watch, onMounted, onUnmounted } from "vue";
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
  ship_by_date?: number;
  countdown?: string;
}

const props = defineProps<{ visible: boolean; order: Order | null }>();
const emit = defineEmits<{ close: [] }>();
const close = () => emit("close");

// ESC key to close modal
const handleKeydown = (e: KeyboardEvent) => {
  if (e.key === "Escape" && props.visible) {
    close();
  }
};

// Body scroll lock
watch(
  () => props.visible,
  (visible) => {
    document.body.style.overflow = visible ? "hidden" : "";
  },
);

const now = ref(Date.now());
let countdownInterval: ReturnType<typeof setInterval> | null = null;

const formatDeadline = (ts: number) => {
  return new Date(ts * 1000).toLocaleDateString(undefined, {
    month: "short",
    day: "numeric",
    hour: "2-digit",
    minute: "2-digit",
  });
};

const countdownText = computed(() => {
  if (!props.order?.ship_by_date) return "";
  if (props.order.countdown) return props.order.countdown;

  const deadline = props.order.ship_by_date * 1000;
  const diff = deadline - now.value;

  if (diff <= 0) return "Overdue";

  const hours = Math.floor(diff / (1000 * 60 * 60));
  const minutes = Math.floor((diff % (1000 * 60 * 60)) / (1000 * 60));

  if (hours >= 24) {
    const days = Math.floor(hours / 24);
    const remainingHours = hours % 24;
    return `${days}d ${remainingHours}h`;
  }

  return `${hours}h ${minutes}m`;
});

const countdownClass = computed(() => {
  if (!props.order?.ship_by_date) return "";

  const deadline = props.order.ship_by_date * 1000;
  const diff = deadline - now.value;

  if (diff <= 0) return "text-overdue";
  if (diff <= 2 * 60 * 60 * 1000) return "text-urgent";
  if (diff <= 6 * 60 * 60 * 1000) return "text-warning";
  return "";
});

onMounted(() => {
  document.addEventListener("keydown", handleKeydown);
  countdownInterval = setInterval(() => {
    now.value = Date.now();
  }, 60000);
});

onUnmounted(() => {
  document.removeEventListener("keydown", handleKeydown);
  document.body.style.overflow = "";
  if (countdownInterval) clearInterval(countdownInterval);
});
</script>

<style scoped>
@import "./OrderDetailModal.styles.css";
</style>
