<template>
  <div class="tiktok-ads-row">
    <div class="cell creative-id">{{ data.creative_id }}</div>
    <div class="cell creative-name">{{ data.creative_name }}</div>
    <div class="cell product-name">{{ data.product_name }}</div>
    <div class="cell cost">{{ formatCurrency(data.cost) }}</div>
    <div class="cell revenue">{{ formatCurrency(data.revenue) }}</div>
    <div class="cell roi" :class="getRoiClass(data.roi)">
      {{ formatRoi(data.roi) }}
    </div>
    <div class="cell conversions">{{ data.conversions }}</div>
    <div class="cell ctr">{{ formatPercent(data.ctr) }}</div>
  </div>
</template>

<script setup lang="ts">
interface TiktokAdsData {
  creative_id: string;
  creative_name: string;
  product_id: string;
  product_name: string;
  cost: number;
  revenue: number;
  roi: number;
  conversions: number;
  ctr: number;
  impressions: number;
  clicks: number;
}

interface Props {
  data: TiktokAdsData;
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

const formatRoi = (value: number) => {
  return `${value.toFixed(2)}x`;
};

const formatPercent = (value: number) => {
  return `${value.toFixed(2)}%`;
};

const getRoiClass = (roi: number) => {
  if (roi >= 5) return "excellent";
  if (roi >= 3) return "good";
  if (roi >= 1) return "ok";
  return "poor";
};
</script>

<style scoped>
.tiktok-ads-row {
  display: grid;
  grid-template-columns: 120px 200px 180px 100px 100px 80px 80px 70px;
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

.creative-name,
.product-name {
  font-weight: 500;
}

.cost,
.revenue,
.conversions {
  text-align: right;
}

.roi {
  font-weight: 600;
  text-align: center;
  padding: 4px 8px;
  border-radius: 4px;
}

.roi.excellent {
  background: #d4edda;
  color: #155724;
}

.roi.good {
  background: #d1ecf1;
  color: #0c5460;
}

.roi.ok {
  background: #fff3cd;
  color: #856404;
}

.roi.poor {
  background: #f8d7da;
  color: #721c24;
}
</style>
