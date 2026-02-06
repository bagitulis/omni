import { useState } from "react";
import {
  Table,
  Switch,
  Button,
  Card,
  Space,
  Typography,
  message,
  theme,
  Popconfirm,
} from "antd";
import { DeleteOutlined, ReloadOutlined } from "@ant-design/icons";
import { StatCard } from "./StatCard";
import { MOCK_ROUTES, type RouteConfig } from "./routeData";

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
  const [routes, setRoutes] = useState<RouteConfig[]>(MOCK_ROUTES);
  const [selectedKeys, setSelectedKeys] = useState<string[]>([]);

  const handleToggle = (id: string) => {
    setRoutes(
      routes.map((r) => (r.id === id ? { ...r, enabled: !r.enabled } : r)),
    );
    message.success("Route status updated");
  };

  const handleClearCache = (id: string) => {
    message.success(`Cache cleared for ${id}`);
  };

  const handleDelete = (id: string) => {
    setRoutes(routes.filter((r) => r.id !== id));
    message.success("Route deleted");
  };

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
    {
      title: "Status",
      dataIndex: "enabled",
      key: "enabled",
      width: 70,
      render: (e: boolean, r: RouteConfig) => (
        <Switch checked={e} onChange={() => handleToggle(r.id)} size="small" />
      ),
    },
    {
      title: "Actions",
      key: "actions",
      width: 100,
      render: (_: unknown, r: RouteConfig) => (
        <Space size="small">
          <Button
            type="text"
            size="small"
            icon={<ReloadOutlined />}
            onClick={() => handleClearCache(r.id)}
            style={{ color: token.colorPrimary }}
          />
          <Popconfirm
            title="Delete?"
            description="Remove this route?"
            onConfirm={() => handleDelete(r.id)}
            okText="Yes"
            cancelText="No"
          >
            <Button type="text" size="small" icon={<DeleteOutlined />} danger />
          </Popconfirm>
        </Space>
      ),
    },
  ];

  return (
    <div>
      <Card
        title="Route & Cache Management"
        size="small"
        style={{ marginBottom: 16, borderRadius: token.borderRadius }}
        extra={
          <Space>
            {selectedKeys.length > 0 && (
              <Text type="secondary" style={{ fontSize: 12 }}>
                {selectedKeys.length} selected
              </Text>
            )}
            <Popconfirm
              title="Clear All Cache?"
              description="Clear cache for all routes?"
              onConfirm={() => message.success("Cache cleared")}
              okText="Yes"
              cancelText="No"
            >
              <Button type="primary" size="small" icon={<ReloadOutlined />}>
                Clear All
              </Button>
            </Popconfirm>
          </Space>
        }
      >
        <Table
          dataSource={routes}
          columns={columns}
          rowKey="id"
          size="small"
          pagination={{ pageSize: 10, showTotal: (t) => `${t} routes` }}
          rowSelection={{
            selectedRowKeys: selectedKeys,
            onChange: (k) => setSelectedKeys(k as string[]),
          }}
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
            value={routes.length}
            color={token.colorPrimary}
          />
          <StatCard
            label="Enabled"
            value={routes.filter((r) => r.enabled).length}
            color="#52c41a"
          />
          <StatCard
            label="With Cache"
            value={routes.filter((r) => r.cache_ttl > 0).length}
            color="#1890ff"
          />
          <StatCard
            label="Total TTL"
            value={routes.reduce((s, r) => s + r.cache_ttl, 0)}
            color="#faad14"
          />
        </div>
      </Card>
    </div>
  );
}
