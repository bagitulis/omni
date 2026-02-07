import { Table, Card, Typography, theme } from "antd";
import { StatCard } from "./StatCard";
import { API_ROUTES } from "./routeData";

const { Text } = Typography;
const { useToken } = theme;

const METHOD_COLORS: Record<string, string> = {
  GET: "#52c41a",
  POST: "#1890ff",
  PUT: "#faad14",
  DELETE: "#f5222d",
};

export default function RouteManagementTab() {
  const { token } = useToken();

  const columns = [
    {
      title: "Route Path",
      dataIndex: "route_path",
      key: "route_path",
      render: (t: string) => (
        <Text code style={{ fontSize: 12 }}>
          {t}
        </Text>
      ),
    },
    {
      title: "Method",
      dataIndex: "method",
      key: "method",
      width: 80,
      render: (m: string) => (
        <span
          style={{
            color: METHOD_COLORS[m] || "#666",
            fontWeight: 600,
            fontSize: 12,
          }}
        >
          {m}
        </span>
      ),
    },
    {
      title: "Description",
      dataIndex: "description",
      key: "description",
      render: (t: string) => (
        <Text type="secondary" style={{ fontSize: 12 }}>
          {t}
        </Text>
      ),
    },
    {
      title: "TTL (s)",
      dataIndex: "cache_ttl",
      key: "cache_ttl",
      width: 80,
      render: (t: number) => (
        <Text style={{ fontSize: 12 }}>
          {t > 0 ? t : <span style={{ color: "#999" }}>Off</span>}
        </Text>
      ),
    },
  ];

  return (
    <div>
      <Card
        title="API Route Reference"
        size="small"
        style={{ marginBottom: 16, borderRadius: token.borderRadius }}
      >
        <Text
          type="secondary"
          style={{ fontSize: 12, marginBottom: 16, display: "block" }}
        >
          Reference view of available API routes. Route configuration is managed
          server-side.
        </Text>
        <Table
          dataSource={API_ROUTES}
          columns={columns}
          rowKey="id"
          size="small"
          pagination={{ pageSize: 10, showTotal: (t) => `${t} routes` }}
          style={{ fontSize: 12 }}
        />
      </Card>

      <Card
        title="Cache Configuration"
        size="small"
        style={{ borderRadius: token.borderRadius }}
      >
        <div
          style={{ display: "grid", gridTemplateColumns: "1fr 1fr", gap: 24 }}
        >
          <StatCard
            label="Total Routes"
            value={API_ROUTES.length}
            color={token.colorPrimary}
          />
          <StatCard
            label="With Cache"
            value={API_ROUTES.filter((r) => r.cache_ttl > 0).length}
            color="#1890ff"
          />
        </div>
      </Card>
    </div>
  );
}
