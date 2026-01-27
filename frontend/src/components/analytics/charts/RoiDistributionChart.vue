<template>
  <div class="roi-distribution-chart">
    <apexchart
      v-if="series.length > 0"
      type="donut"
      height="300"
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
  data?: {
    excellent: number; // ROI >= 5
    good: number; // ROI 3-5
    ok: number; // ROI 1-3
    poor: number; // ROI < 1
  };
}

const props = withDefaults(defineProps<Props>(), {
  data: () => ({ excellent: 0, good: 0, ok: 0, poor: 0 }),
});

const series = computed(() => {
  const { excellent, good, ok, poor } = props.data;
  const total = excellent + good + ok + poor;
  if (total === 0) return [];

  return [excellent, good, ok, poor];
});

const chartOptions = computed(() => ({
  chart: {
    type: "donut",
  },
  labels: ["Excellent (≥5x)", "Good (3-5x)", "OK (1-3x)", "Poor (<1x)"],
  colors: ["#10B981", "#3B82F6", "#F59E0B", "#EF4444"],
  legend: {
    position: "bottom",
  },
  plotOptions: {
    pie: {
      donut: {
        size: "65%",
        labels: {
          show: true,
          name: {
            show: true,
          },
          value: {
            show: true,
            formatter: (val: string) => `${val} products`,
          },
          total: {
            show: true,
            label: "Total Products",
            formatter: (w: any) => {
              const total = w.globals.seriesTotals.reduce(
                (a: number, b: number) => a + b,
                0,
              );
              return String(total);
            },
          },
        },
      },
    },
  },
  dataLabels: {
    enabled: true,
    formatter: (val: number) => {
      return val.toFixed(1) + "%";
    },
  },
}));
</script>

<style scoped>
.roi-distribution-chart {
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
  height: 300px;
  color: #6b7280;
}
</style>
