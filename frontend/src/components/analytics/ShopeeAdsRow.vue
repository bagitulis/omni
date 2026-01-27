<template>
  <div class="shopee-ads-row">
    <div class="cell product-id">{{ data.product_id }}</div>
    <div class="cell product-name">{{ data.product_name }}</div>
    <div class="cell cost">{{ formatCurrency(data.cost) }}</div>
    <div class="cell revenue">{{ formatCurrency(data.revenue) }}</div>
    <div class="cell roas" :class="getRoasClass(data.roas)">
      {{ formatRoas(data.roas) }}
    </div>
    <div class="cell conversions">{{ data.conversions }}</div>
    <div class="cell ctr">{{ formatPercent(data.ctr) }}</div>
    <div class="cell units">{{ data.units_sold }}</div>
  </div>
</template>

<script setup lang="ts">
interface ShopeeAdsData {
  product_id: string;
  product_name: string;
  cost: number;
  revenue: number;
  roas: number;
  conversions: number;
  ctr: number;
  units_sold: number;
  impressions: number;
  clicks: number;
}

interface Props {
  data: ShopeeAdsData;
}

defineProps<Props>();

const formatCurrency = (value: number) => {
  return new Intl.NumberFormat("id-ID", {
    style: "currency",
    currency: "IDR",
    minimumFractionDigits: 0,
    maximumFractionDigits: 0,
    notation: "compact",
  }).format(value);
};

const formatRoas = (value: number) => {
  return `${value.toFixed(2)}x`;
};

const formatPercent = (value: number) => {
  return `${value.toFixed(2)}%`;
};

const getRoasClass = (roas: number) => {
  if (roas >= 5) return "excellent";
  if (roas >= 3) return "good";
  if (roas >= 1) return "ok";
  return "poor";
};
</script>

<style scoped>
.shopee-ads-row {
  display: grid;
  grid-template-columns: 100px 200px 100px 100px 80px 80px 70px 70px;
  gap: 8px;
  padding: 12px 16px;
  align-items: center;
  font-size: 14px;
}

.cell {
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.product-name {
  font-weight: 500;
}

.cost,
.revenue,
.conversions,
.units {
  text-align: right;
}

.roas {
  font-weight: 600;
  text-align: center;
  padding: 4px 8px;
  border-radius: 4px;
}

.roas.excellent {
  background: #d4edda;
  color: #155724;
}

.roas.good {
  background: #d1ecf1;
  color: #0c5460;
}

.roas.ok {
  background: #fff3cd;
  color: #856404;
}

.roas.poor {
  background: #f8d7da;
  color: #721c24;
}
</style>
