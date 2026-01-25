<template>
  <div class="bg-white p-4 rounded-lg shadow-sm border border-gray-100">
    <div class="flex justify-between items-center mb-4">
      <h3 class="text-lg font-semibold text-gray-700">Performance Trend</h3>
      <!-- Simple Toggle for Metrics could go here -->
    </div>
    <div
      v-if="loading"
      class="h-80 flex items-center justify-center bg-gray-50 rounded"
    >
      <span class="text-gray-400">Loading chart...</span>
    </div>
    <div
      v-else-if="series.length === 0"
      class="h-80 flex items-center justify-center bg-gray-50 rounded"
    >
      <span class="text-gray-400">No data available for this period</span>
    </div>
    <div v-else>
      <apexchart
        width="100%"
        height="350"
        type="line"
        :options="chartOptions"
        :series="series"
      ></apexchart>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, defineProps } from "vue";
import type { TrendDataPoint } from "../../stores/analytics";

const props = defineProps<{
  data: TrendDataPoint[];
  loading: boolean;
}>();

const series = computed(() => {
  if (!props.data || props.data.length === 0) return [];

  return [
    {
      name: "GMV (Omzet)",
      type: "column",
      data: props.data.map((d) => d.gmv),
    },
    {
      name: "Ad Spend (Biaya)",
      type: "line",
      data: props.data.map((d) => d.spend),
    },
  ];
});

const chartOptions = computed(() => ({
  chart: {
    height: 350,
    type: "line",
    toolbar: { show: false },
    zoom: { enabled: false },
  },
  stroke: {
    width: [0, 3],
    curve: "smooth",
  },
  plotOptions: {
    bar: {
      columnWidth: "50%",
      borderRadius: 4,
    },
  },
  dataLabels: {
    enabled: false,
  },
  labels: props.data.map((d) => d.label || d.period_start), // X-axis labels
  xaxis: {
    type: "category", // or 'datetime' if we parse dates
    tooltip: {
      enabled: false,
    },
  },
  yaxis: [
    {
      title: {
        text: "GMV (IDR)",
      },
      labels: {
        formatter: (val: number) => {
          if (val >= 1000000) return (val / 1000000).toFixed(1) + "M";
          if (val >= 1000) return (val / 1000).toFixed(0) + "k";
          return val;
        },
      },
    },
    {
      opposite: true,
      title: {
        text: "Ad Spend (IDR)",
      },
      labels: {
        formatter: (val: number) => {
          if (val >= 1000000) return (val / 1000000).toFixed(1) + "M";
          if (val >= 1000) return (val / 1000).toFixed(0) + "k";
          return val;
        },
      },
    },
  ],
  colors: ["#3B82F6", "#EF4444"], // Blue for Sales, Red for Cost
  tooltip: {
    shared: true,
    intersect: false,
    y: {
      formatter: (y: number) => {
        if (typeof y !== "undefined") {
          return new Intl.NumberFormat("id-ID", {
            style: "currency",
            currency: "IDR",
            maximumFractionDigits: 0,
          }).format(y);
        }
        return y;
      },
    },
  },
  legend: {
    position: "top",
    horizontalAlign: "right",
  },
}));
</script>
