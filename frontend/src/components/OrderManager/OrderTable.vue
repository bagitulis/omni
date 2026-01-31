<template>
  <div class="order-table-container">
    <div class="table-header">
      <span class="table-title">{{ getTableTitle }}</span>
      <span class="table-count">{{
        activeTab === "locked"
          ? `${filteredOrders.length} products`
          : `${filteredOrders.length} orders`
      }}</span>
    </div>
    <!-- Locked Orders Table -->
    <div v-if="activeTab === 'locked'" class="table-responsive">
      <table class="orders-table">
        <thead>
          <tr>
            <th style="width: 15%">SKU</th>
            <th style="width: 40%">Product Name</th>
            <th style="width: 30%">Variation</th>
            <th style="width: 15%; text-align: center">Qty</th>
          </tr>
        </thead>
        <tbody>
          <tr
            v-for="(item, index) in filteredOrders"
            :key="index"
            class="order-row"
          >
            <td>
              <span class="sku">{{ item.sku }}</span>
            </td>
            <td>
              <span class="product-name" :title="item.product_name">{{
                item.product_name
              }}</span>
            </td>
            <td>
              <span class="variation" :title="item.variation_name">{{
                item.variation_name || "-"
              }}</span>
            </td>
            <td style="text-align: center">
              <span class="qty-badge">{{ item.qty }}</span>
            </td>
          </tr>
        </tbody>
      </table>
    </div>
    <!-- Today's Orders Table -->
    <div v-else-if="activeTab === 'today'" class="table-responsive">
      <table class="orders-table">
        <thead>
          <tr>
            <th style="width: 14%">Order No.</th>
            <th style="width: 14%">Tracking No.</th>
            <th style="width: 12%">Courier</th>
            <th style="width: 12%">Seller SKU</th>
            <th style="width: 22%">Product Name</th>
            <th style="width: 16%">Variation</th>
            <th style="width: 6%; text-align: center">Qty</th>
          </tr>
        </thead>
        <tbody>
          <tr
            v-for="(item, index) in filteredOrders"
            :key="index"
            class="order-row"
          >
            <td>
              <span class="order-number">{{ item.orderSn }}</span>
            </td>
            <td>
              <span class="tracking-number" :title="item.trackingNo">{{
                item.trackingNo || "-"
              }}</span>
            </td>
            <td>
              <span class="courier" :title="item.courier">{{
                item.courier || "-"
              }}</span>
            </td>
            <td>
              <span class="sku">{{ item.sellerSku || "-" }}</span>
            </td>
            <td>
              <span class="product-name" :title="item.productName">{{
                item.productName || "-"
              }}</span>
            </td>
            <td>
              <span class="variation" :title="item.variationName">{{
                item.variationName || "-"
              }}</span>
            </td>
            <td style="text-align: center">
              <span class="qty-badge">{{ item.quantity || 1 }}</span>
            </td>
          </tr>
        </tbody>
      </table>
    </div>
    <!-- Card View for Regular Orders -->
    <div v-else class="orders-grid">
      <OrderCard
        v-for="order in groupedOrders"
        :key="order.order_no"
        :order="order"
        :activeTab="activeTab"
        @ship-order="$emit('ship-order', $event)"
        @cancel-order="$emit('cancel-order', $event)"
        @view-detail="$emit('view-detail', $event)"
        @copy-order-number="$emit('copy-order-number', $event)"
      />
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed } from "vue";
import OrderCard from "./OrderCard.vue";

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
  sku: string;
  product_name: string;
  variation_name?: string;
  qty: number;
  status?: string;
  buyer_username?: string;
  total_amount?: number;
  currency?: string;
  payment_method?: string;
  items?: OrderItem[];
  orderSn?: string;
  trackingNo?: string;
  courier?: string;
  sellerSku?: string;
  productName?: string;
  variationName?: string;
  quantity?: number;
  [key: string]: any;
}
interface GroupedOrder {
  order_no: string;
  platform: string;
  status: string;
  buyer_username: string;
  total_amount: number;
  currency: string;
  payment_method?: string;
  items: OrderItem[];
}

const props = defineProps<{ filteredOrders: Order[]; activeTab: string }>();
defineEmits<{
  "ship-order": [order: GroupedOrder];
  "cancel-order": [order: GroupedOrder];
  "view-detail": [order: GroupedOrder];
  "copy-order-number": [orderNo: string];
}>();

const groupedOrders = computed<GroupedOrder[]>(() => {
  if (["locked", "today"].includes(props.activeTab)) return [];
  const orderMap = new Map<string, GroupedOrder>();
  for (const item of props.filteredOrders) {
    if (!orderMap.has(item.order_no)) {
      orderMap.set(item.order_no, {
        order_no: item.order_no,
        platform: item.platform,
        status: item.status || "READY_TO_SHIP",
        buyer_username: item.buyer_username || "Unknown Buyer",
        total_amount: item.total_amount || 0,
        currency: item.currency || "IDR",
        payment_method: item.payment_method,
        items: [],
      });
    }
    const order = orderMap.get(item.order_no)!;
    order.items.push({
      sku: item.sku,
      product_name: item.product_name,
      variation_name: item.variation_name,
      qty: item.qty,
      price: item.price,
      product_image: item.product_image,
    });
    if (item.price && item.qty) order.total_amount += item.price * item.qty;
  }
  return Array.from(orderMap.values());
});

const getTableTitle = computed(() => {
  const titleMap: Record<string, string> = {
    locked: "Locked Today",
    today: "Today's Shipped Orders",
    unpaid: "Unpaid Orders",
    unprocess: "To Ship",
    processed: "Shipped Orders",
  };
  return (
    titleMap[props.activeTab] ||
    `${props.activeTab.charAt(0).toUpperCase() + props.activeTab.slice(1)} Orders`
  );
});
</script>

<style scoped>
@import "./OrderManager.theme.css";
@import "./OrderTable.styles.css";
.orders-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(400px, 1fr));
  gap: var(--om-spacing-md);
  padding: var(--om-spacing-md) 0;
}
@media (max-width: 768px) {
  .orders-grid {
    grid-template-columns: 1fr;
  }
}
</style>
