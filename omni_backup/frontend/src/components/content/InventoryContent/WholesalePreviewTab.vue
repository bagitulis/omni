<template>
  <div class="tab-content">
    <div class="info-banner">
      <span class="info-icon">ℹ️</span>
      <div class="info-text">
        <strong>{{ itemCount }} SKU</strong> dipilih
        <p class="info-note">
          Wholesale akan dihitung dengan rumus: (Harga - Admin) + (Admin /
          MinQty)
        </p>
      </div>
    </div>

    <div class="preview-table-container">
      <table class="preview-table" aria-label="Wholesale pricing preview">
        <thead>
          <tr>
            <th scope="col">SKU</th>
            <th scope="col">Harga</th>
            <th scope="col">T1 ({{ minOrder1 }}-{{ maxOrder1 }})</th>
            <th scope="col">T2 ({{ tier2Min }}-{{ tier2Max }})</th>
            <th scope="col">T3 ({{ tier3Min }}-{{ maxOrderTier3 }})</th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="item in previewItems" :key="item.sku">
            <td class="sku-cell">{{ item.sku }}</td>
            <td class="price-cell">{{ formatPrice(item.price) }}</td>
            <td>{{ formatPrice(item.tiers[0]?.unitPrice) }}</td>
            <td>{{ formatPrice(item.tiers[1]?.unitPrice) }}</td>
            <td>{{ formatPrice(item.tiers[2]?.unitPrice) }}</td>
          </tr>
        </tbody>
      </table>
    </div>
    <p v-if="itemCount > 5" class="more-items-note">
      ... dan {{ itemCount - 5 }} lainnya
    </p>
  </div>
</template>

<script setup lang="ts">
import type { PreviewItem } from "./composables/useWholesaleUpdate";

defineProps<{
  itemCount: number;
  previewItems: PreviewItem[];
  minOrder1: number;
  maxOrder1: number;
  tier2Min: number;
  tier2Max: number;
  tier3Min: number;
  maxOrderTier3: number;
  formatPrice: (price: number | undefined) => string;
}>();
</script>

<style src="./WholesaleUpdateModal.styles.css" scoped></style>
