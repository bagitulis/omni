import { useMemo } from "react";
import ReactApexChart from "react-apexcharts";
import { Spin, theme } from "antd";
import type { ApexOptions } from "apexcharts";
import type { FC } from "react";

interface TrendDataPoint {
  label: string;
  spend: number;
  gmv: number;
}

interface AdsTrendChartProps {
  data: TrendDataPoint[];
  loading?: boolean;
}

export const AdsTrendChart: FC<AdsTrendChartProps> = ({ data, loading }) => {
  const { token } = theme.useToken();

  const series = useMemo(() => {
    if (!data || data.length === 0) return [];
    return [
      {
        name: "GMV (Omzet)",
        type: "column",
        data: data.map((d) => d.gmv),
      },
      {
        name: "Ad Spend (Biaya)",
        type: "line",
        data: data.map((d) => d.spend),
      },
    ];
  }, [data]);

  const options: ApexOptions = useMemo(
    () => ({
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
      labels: data.map((d) => d.label),
      xaxis: {
        type: "category",
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
              if (val >= 1000000) return `${(val / 1000000).toFixed(1)}M`;
              if (val >= 1000) return `${(val / 1000).toFixed(0)}k`;
              return String(val);
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
              if (val >= 1000000) return `${(val / 1000000).toFixed(1)}M`;
              if (val >= 1000) return `${(val / 1000).toFixed(0)}k`;
              return String(val);
            },
          },
        },
      ],
      colors: [token.colorPrimary, token.colorError],
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
            return "";
          },
        },
      },
      legend: {
        position: "top",
        horizontalAlign: "right",
      },
    }),
    [data, token.colorPrimary, token.colorError],
  );

  if (loading) {
    return (
      <div
        style={{
          height: 350,
          display: "flex",
          justifyContent: "center",
          alignItems: "center",
          background: token.colorFillAlter,
          borderRadius: token.borderRadius,
        }}
      >
        <Spin />
      </div>
    );
  }

  if (!data || data.length === 0) {
    return (
      <div
        style={{
          height: 350,
          display: "flex",
          justifyContent: "center",
          alignItems: "center",
          background: token.colorFillAlter,
          borderRadius: token.borderRadius,
        }}
      >
        <span style={{ color: token.colorTextSecondary }}>
          No data available for this period
        </span>
      </div>
    );
  }

  return (
    <ReactApexChart
      options={options}
      series={series}
      type="line"
      height={350}
    />
  );
};

export default AdsTrendChart;
