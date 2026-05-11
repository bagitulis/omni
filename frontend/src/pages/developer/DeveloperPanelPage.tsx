import { useCallback, useEffect, useState } from "react";
import {
  App,
  Alert,
  Button,
  Card,
  Col,
  Row,
  Space,
  Spin,
  Statistic,
  Table,
  Tag,
  Typography,
} from "antd";
import {
  CrownOutlined,
  ReloadOutlined,
  SafetyCertificateOutlined,
  TeamOutlined,
  UserOutlined,
} from "@ant-design/icons";
import apiClient from "@/api/client";
import { developerApi, type TenantOverview } from "@/api/developer";
import { useAuthStore } from "@/stores/authStore";
import type { ColumnsType } from "antd/es/table";

const { Title, Text } = Typography;

export default function DeveloperPanelPage() {
  const { message } = App.useApp();
  const accessToken = useAuthStore((state) => state.accessToken);
  const setAuth = useAuthStore((state) => state.setAuth);
  const user = useAuthStore((state) => state.user);

  const [tenants, setTenants] = useState<TenantOverview[]>([]);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const fetchOverview = useCallback(async () => {
    setLoading(true);
    setError(null);
    try {
      const response = await developerApi.getOverview();
      if (response.success && response.data) {
        setTenants(response.data.tenants);
      } else {
        setError(response.error || "Failed to load tenant overview");
      }
    } catch (err) {
      const msg = err instanceof Error ? err.message : "Failed to load tenant overview";
      setError(msg);
    } finally {
      setLoading(false);
    }
  }, []);

  useEffect(() => {
    if (accessToken) {
      void fetchOverview();
    }
  }, [accessToken, fetchOverview]);

  const handleManageUsers = async (tenantId: string) => {
    try {
      const res = await apiClient.post<{
        token: string;
        tenant_id: string;
      }>("/auth/switch-tenant", { tenant_id: tenantId });
      if (res.success && res.data && user) {
        setAuth({
          access_token: res.data.token,
          user,
          tenant_id: res.data.tenant_id,
        });
        message.success(`Switched to tenant: ${tenantId}`);
        window.location.href = "/users";
      } else {
        throw new Error(res.error || "Failed to switch tenant");
      }
    } catch (err) {
      const msg = err instanceof Error ? err.message : "Failed to switch tenant";
      message.error(msg);
    }
  };

  const totals = tenants.reduce(
    (acc, t) => ({
      users: acc.users + (t.user_count || 0),
      owners: acc.owners + (t.owners || 0),
      admins: acc.admins + (t.admins || 0),
      members: acc.members + (t.users || 0),
    }),
    { users: 0, owners: 0, admins: 0, members: 0 },
  );

  const columns: ColumnsType<TenantOverview> = [
    {
      title: "Tenant",
      dataIndex: "shop_name",
      key: "shop_name",
      render: (shop: string, record) => (
        <Space direction="vertical" size={0}>
          <Text strong>{shop}</Text>
          <Text type="secondary" style={{ fontSize: 12 }}>
            {record.id}
          </Text>
        </Space>
      ),
    },
    {
      title: "Total Users",
      dataIndex: "user_count",
      key: "user_count",
      align: "right",
      render: (n: number) => <Text strong>{n}</Text>,
    },
    {
      title: "Owners",
      dataIndex: "owners",
      key: "owners",
      align: "right",
      render: (n: number) => <Tag color="gold">{n}</Tag>,
    },
    {
      title: "Admins",
      dataIndex: "admins",
      key: "admins",
      align: "right",
      render: (n: number) => <Tag color="blue">{n}</Tag>,
    },
    {
      title: "Users",
      dataIndex: "users",
      key: "users",
      align: "right",
      render: (n: number) => <Tag color="default">{n}</Tag>,
    },
    {
      title: "Status",
      dataIndex: "error",
      key: "error",
      render: (err?: string) =>
        err ? <Tag color="red">{err}</Tag> : <Tag color="green">Healthy</Tag>,
    },
    {
      title: "Actions",
      key: "actions",
      fixed: "right",
      render: (_: unknown, record) => (
        <Space>
          <Button
            type="primary"
            size="small"
            onClick={() => handleManageUsers(record.id)}
            disabled={!!record.error}
          >
            Manage Users
          </Button>
        </Space>
      ),
    },
  ];

  return (
    <div style={{ padding: 24 }}>
      <div
        style={{
          display: "flex",
          justifyContent: "space-between",
          alignItems: "center",
          marginBottom: 16,
        }}
      >
        <div>
          <Title level={3} style={{ margin: 0 }}>
            Developer Panel
          </Title>
          <Text type="secondary">
            Cross-tenant overview — visible to developer role only
          </Text>
        </div>
        <Button icon={<ReloadOutlined />} onClick={() => void fetchOverview()}>
          Refresh
        </Button>
      </div>

      {user && user.role !== "developer" && (
        <Alert
          type="warning"
          showIcon
          style={{ marginBottom: 16 }}
          message="You are not a developer"
          description="This page is intended for the developer role. Backend access is enforced server-side."
        />
      )}

      {error && (
        <Alert
          type="error"
          showIcon
          style={{ marginBottom: 16 }}
          message="Failed to load overview"
          description={error}
        />
      )}

      <Row gutter={16} style={{ marginBottom: 16 }}>
        <Col xs={12} md={6}>
          <Card>
            <Statistic
              title="Tenants"
              value={tenants.length}
              prefix={<TeamOutlined />}
            />
          </Card>
        </Col>
        <Col xs={12} md={6}>
          <Card>
            <Statistic
              title="Owners"
              value={totals.owners}
              prefix={<CrownOutlined />}
            />
          </Card>
        </Col>
        <Col xs={12} md={6}>
          <Card>
            <Statistic
              title="Admins"
              value={totals.admins}
              prefix={<SafetyCertificateOutlined />}
            />
          </Card>
        </Col>
        <Col xs={12} md={6}>
          <Card>
            <Statistic
              title="Members"
              value={totals.members}
              prefix={<UserOutlined />}
            />
          </Card>
        </Col>
      </Row>

      <Spin spinning={loading}>
        <Card>
          <Table<TenantOverview>
            rowKey="id"
            dataSource={tenants}
            columns={columns}
            pagination={false}
            scroll={{ x: 900 }}
            locale={{ emptyText: loading ? "Loading..." : "No tenants found" }}
          />
        </Card>
      </Spin>
    </div>
  );
}
