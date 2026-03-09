import { Card, Table, Tag, theme } from "antd";
import type { TableProps } from "antd";
import type { RouteConfig } from "@/types/routeConfig";

interface RoutePerformanceTableProps {
  routes: RouteConfig[];
}

interface PerformanceRow {
  key: number;
  route_path: string;
  route_method: string;
  timeout: number;
  cache_ttl: number;
  max_concurrent: number;
  rate_limit_max: number;
}

export function RoutePerformanceTable({ routes }: RoutePerformanceTableProps) {
  const { token } = theme.useToken();

  const data_source: PerformanceRow[] = routes.map((route) => ({
    key: route.id,
    route_path: route.route_path,
    route_method: route.route_method ?? "-",
    timeout: route.timeout,
    cache_ttl: route.cache_ttl ?? 0,
    max_concurrent: route.max_concurrent ?? 0,
    rate_limit_max: route.rate_limit_max ?? 0,
  }));

  const method_colors: Record<string, string> = {
    GET: token.colorSuccess,
    POST: token.colorPrimary,
    PUT: token.colorWarning,
    DELETE: token.colorError,
  };

  const columns: TableProps<PerformanceRow>["columns"] = [
    {
      title: "Route",
      dataIndex: "route_path",
      key: "route_path",
      width: 280,
      render: (value: string) => <Tag bordered={false}>{value}</Tag>,
    },
    {
      title: "Method",
      dataIndex: "route_method",
      key: "route_method",
      width: 100,
      render: (value: string) => (
        <Tag color={method_colors[value] || token.colorTextSecondary}>
          {value}
        </Tag>
      ),
    },
    {
      title: "Timeout",
      dataIndex: "timeout",
      key: "timeout",
      width: 110,
      sorter: (a, b) => a.timeout - b.timeout,
      render: (value: number) => `${value} ms`,
    },
    {
      title: "Cache TTL",
      dataIndex: "cache_ttl",
      key: "cache_ttl",
      width: 100,
      sorter: (a, b) => a.cache_ttl - b.cache_ttl,
      render: (value: number) => `${value}s`,
    },
    {
      title: "Max Concurrent",
      dataIndex: "max_concurrent",
      key: "max_concurrent",
      width: 130,
      sorter: (a, b) => a.max_concurrent - b.max_concurrent,
    },
    {
      title: "Rate Limit Max",
      dataIndex: "rate_limit_max",
      key: "rate_limit_max",
      width: 130,
      sorter: (a, b) => a.rate_limit_max - b.rate_limit_max,
    },
  ];

  return (
    <Card
      size="small"
      title="Performance Metrics"
      style={{ borderRadius: token.borderRadius }}
    >
      <Table<PerformanceRow>
        rowKey="key"
        size="small"
        columns={columns}
        dataSource={data_source}
        pagination={{ pageSize: 8, showSizeChanger: true }}
        scroll={{ x: 860 }}
      />
    </Card>
  );
}
