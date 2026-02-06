<template>
  <div class="revenue-area-chart">
    <apexchart
      v-if="series.length > 0"
      type="area"
      height="350"
      :options="chartOptions"
      :series="series"
    />
    <div v-else class="empty-state">
      <p>No data available</p>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed } from "vue";
import VueApexCharts from "vue3-apexcharts";

const apexchart = VueApexCharts;

interface Props {
  data?: Array<{
    date: string;
    revenue: number;
    cost: number;
  }>;
}

const props = withDefaults(defineProps<Props>(), {
  data: () => [],
});

const series = computed(() => {
  if (!props.data || props.data.length === 0) return [];

  return [
    {
      name: "Revenue",
      data: props.data.map((d) => ({
        x: new Date(d.date).getTime(),
        y: d.revenue,
      })),
    },
    {
      name: "Cost",
      data: props.data.map((d) => ({
        x: new Date(d.date).getTime(),
        y: d.cost,
      })),
    },
  ];
});

const chartOptions = computed(() => ({
  chart: {
    type: "area",
    height: 350,
    zoom: {
      enabled: true,
    },
    toolbar: {
      show: true,
    },
  },
  dataLabels: {
    enabled: false,
  },
  stroke: {
    curve: "smooth",
    width: 2,
  },
  fill: {
    type: "gradient",
    gradient: {
      opacityFrom: 0.6,
      opacityTo: 0.1,
    },
  },
  colors: ["#10B981", "#F59E0B"],
  xaxis: {
    type: "datetime",
    labels: {
      format: "dd MMM",
    },
  },
  yaxis: {
    labels: {
      formatter: (value: number) => {
        return new Intl.NumberFormat("id-ID", {
          style: "currency",
          currency: "IDR",
          minimumFractionDigits: 0,
          maximumFractionDigits: 0,
          notation: "compact",
        }).format(value);
      },
    },
  },
  tooltip: {
    x: {
      format: "dd MMM yyyy",
    },
    y: {
      formatter: (value: number) => {
        return new Intl.NumberFormat("id-ID", {
          style: "currency",
          currency: "IDR",
          minimumFractionDigits: 0,
          maximumFractionDigits: 0,
        }).format(value);
      },
    },
  },
  legend: {
    position: "top",
    horizontalAlign: "left",
  },
}));
</script>

<style scoped>
.revenue-area-chart {
  width: 100%;
  background: white;
  border-radius: 8px;
  padding: 1rem;
  box-shadow: 0 1px 3px 0 rgba(0, 0, 0, 0.1);
}

.empty-state {
  display: flex;
  align-items: center;
  justify-content: center;
  height: 350px;
  color: #6b7280;
}
</style>
