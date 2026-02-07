import { ApexOptions } from "apexcharts";

// Shared chart configuration
export const getChartOptions = (
  categories: string[],
  colors: string[],
): ApexOptions => ({
  chart: {
    type: "bar",
    toolbar: { show: false },
    fontFamily: "system-ui, -apple-system, sans-serif",
  },
  colors,
  plotOptions: {
    bar: {
      horizontal: false,
      columnWidth: "55%",
      borderRadius: 3,
    },
  },
  dataLabels: { enabled: false },
  stroke: { show: true, width: 2, colors: ["transparent"] },
  xaxis: { categories },
  yaxis: { labels: { formatter: (val: number) => val.toLocaleString() } },
  fill: { opacity: 1 },
  tooltip: {
    y: { formatter: (val: number) => val.toLocaleString("id-ID") },
  },
  grid: { borderColor: "#f0f0f0" },
});
