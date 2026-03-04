<template>
  <div class="order-table-container">
    <!-- Table Header -->
    <div class="table-header">
      <span class="table-title">
        {{ getTableTitle }}
      </span>
      <span class="table-count">
        {{
          activeTab === "locked"
            ? `${filteredOrders.length} products`
            : `${filteredOrders.length} items`
        }}
      </span>
    </div>

    <!-- Locked Orders Table View -->
    <div v-if="activeTab === 'locked'" class="table-responsive">
      <table class="orders-table">
        <thead>
          <tr>
            <th style="width: 15%">SKU</th>
            <th style="width: 40%">Nama Produk</th>
            <th style="width: 30%">Variasi</th>
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

    <!-- Order Today View - With Tracking Info -->
    <div v-else-if="activeTab === 'today'" class="table-responsive">
      <table class="orders-table">
        <thead>
          <tr>
            <th style="width: 14%">No. Pesanan</th>
            <th style="width: 14%">No Resi</th>
            <th style="width: 12%">Ekspedisi</th>
            <th style="width: 12%">SKU Seller</th>
            <th style="width: 22%">Nama Produk</th>
            <th style="width: 16%">Nama Variasi</th>
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
              <span class="tracking-number" :title="item.trackingNo">
                {{ item.trackingNo || "-" }}
              </span>
            </td>
            <td>
              <span class="courier" :title="item.courier">
                {{ item.courier || "-" }}
              </span>
            </td>
            <td>
              <span class="sku">{{ item.sellerSku || "-" }}</span>
            </td>
            <td>
              <span class="product-name" :title="item.productName">
                {{ item.productName || "-" }}
              </span>
            </td>
            <td>
              <span class="variation" :title="item.variationName">
                {{ item.variationName || "-" }}
              </span>
            </td>
            <td style="text-align: center">
              <span class="qty-badge">{{ item.quantity || 1 }}</span>
            </td>
          </tr>
        </tbody>
      </table>
    </div>

    <!-- Regular Items View -->
    <div v-else class="table-responsive">
      <table class="orders-table">
        <thead>
          <tr>
            <th style="width: 10%">Platform</th>
            <th style="width: 18%">No. Pesanan</th>
            <th style="width: 15%">SKU</th>
            <th style="width: 25%">Nama Produk</th>
            <th style="width: 20%">Variasi</th>
            <th style="width: 8%; text-align: center">Qty</th>
          </tr>
        </thead>
        <tbody>
          <tr
            v-for="(item, index) in filteredOrders"
            :key="index"
            class="order-row"
          >
            <td>
              <span :class="['platform-badge', item.platform.toLowerCase()]">
                {{ item.platform }}
              </span>
            </td>
            <td>
              <span class="order-number">{{ item.order_no }}</span>
            </td>
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
                item.variation_name
              }}</span>
            </td>
            <td style="text-align: center">
              <span class="qty-badge">{{ item.qty }}</span>
            </td>
          </tr>
        </tbody>
      </table>
    </div>
  </div>
</template>

<script setup lang="ts">
interface Order {
  order_no: string;
  platform: string;
  sku: string;
  product_name: string;
  variation_name?: string;
  qty: number;
  // Order Today fields
  orderSn?: string;
  trackingNo?: string;
  courier?: string;
  sellerSku?: string;
  productName?: string;
  variationName?: string;
  quantity?: number;
  [key: string]: any;
}

import { computed } from "vue";

const props = defineProps<{
  filteredOrders: Order[];
  activeTab: string;
}>();

const getTableTitle = computed(() => {
  const titleMap: Record<string, string> = {
    locked: "Locked Today Orders",
    today: "Order Today - Siap Kirim",
    unpaid: "Unpaid Orders",
    unprocess: "Unprocess Orders",
    processed: "Processed Orders",
  };
  return (
    titleMap[props.activeTab] ||
    `${props.activeTab.charAt(0).toUpperCase() + props.activeTab.slice(1)} Orders`
  );
});
</script>

<style scoped>
@import "./OrderManager.styles.css";
@import "./OrderTable.styles.css";
</style>
