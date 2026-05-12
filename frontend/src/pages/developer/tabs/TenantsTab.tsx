import { useCallback, useEffect, useState } from "react";
import { App, Button, Card, Space, Table, Tag, Typography } from "antd";
import { PlusOutlined, ReloadOutlined } from "@ant-design/icons";
import type { ColumnsType } from "antd/es/table";
import { developerApi, type TenantDetail } from "@/api/developer";
import { TypeToConfirmModal } from "@/components/modals/TypeToConfirmModal";
import { CreateTenantModal } from "../components/CreateTenantModal";

const { Title, Text } = Typography;

export default function TenantsTab() {
  const { message } = App.useApp();

  const [tenants, setTenants] = useState<TenantDetail[]>([]);
  const [loading, setLoading] = useState(false);
  const [createModalOpen, setCreateModalOpen] = useState(false);
  const [deactivateTarget, setDeactivateTarget] = useState<TenantDetail | null>(null);
  const [deactivateLoading, setDeactivateLoading] = useState(false);

  const fetchTenants = useCallback(async () => {
    setLoading(true);
    try {
      const response = await developerApi.getTenants();
      if (response.success && response.data) {
        setTenants(response.data);
      } else {
        message.error(response.error || "Failed to load tenants");
      }
    } catch (err) {
      const msg = err instanceof Error ? err.message : "Failed to load tenants";
      message.error(msg);
    } finally {
      setLoading(false);
    }
  }, [message]);

  useEffect(() => {
    void fetchTenants();
  }, [fetchTenants]);

  const handleDeactivate = async () => {
    if (!deactivateTarget) return;
    setDeactivateLoading(true);
    try {
      const response = await developerApi.deactivateTenant(deactivateTarget.id);
      if (response.success) {
        message.success(`Tenant "${deactivateTarget.name}" deactivated`);
        setDeactivateTarget(null);
        void fetchTenants();
      } else {
        message.error(response.error || "Failed to deactivate tenant");
      }
    } catch (err) {
      const msg = err instanceof Error ? err.message : "Failed to deactivate tenant";
      message.error(msg);
    } finally {
      setDeactivateLoading(false);
    }
  };

  const handleCreateSuccess = () => {
    setCreateModalOpen(false);
    void fetchTenants();
  };

  const columns: ColumnsType<TenantDetail> = [
    {
      title: "Name",
      dataIndex: "name",
      key: "name",
      render: (name: string) => <Text strong>{name}</Text>,
    },
    {
      title: "Status",
      dataIndex: "is_active",
      key: "is_active",
      width: 100,
      render: (active: boolean) =>
        active ? <Tag color="green">Active</Tag> : <Tag color="red">Inactive</Tag>,
    },
    {
      title: "Users",
      dataIndex: "user_count",
      key: "user_count",
      width: 80,
      align: "right",
      render: (count: number) => count,
    },
    {
      title: "Created",
      dataIndex: "created_at",
      key: "created_at",
      width: 180,
      render: (date: string) => new Date(date).toLocaleDateString(),
    },
    {
      title: "Actions",
      key: "actions",
      width: 120,
      render: (_: unknown, record) => (
        <Button
          size="small"
          danger
          disabled={!record.is_active}
          onClick={() => setDeactivateTarget(record)}
        >
          Deactivate
        </Button>
      ),
    },
  ];

  return (
    <>
      <div
        style={{
          display: "flex",
          justifyContent: "space-between",
          alignItems: "center",
          marginBottom: 16,
        }}
      >
        <div>
          <Title level={4} style={{ margin: 0 }}>
            Tenant Management
          </Title>
          <Text type="secondary">Create and manage tenants across the system</Text>
        </div>
        <Space>
          <Button icon={<ReloadOutlined />} onClick={() => void fetchTenants()}>
            Refresh
          </Button>
          <Button
            type="primary"
            icon={<PlusOutlined />}
            onClick={() => setCreateModalOpen(true)}
          >
            Create Tenant
          </Button>
        </Space>
      </div>

      <Card>
        <Table<TenantDetail>
          rowKey="id"
          dataSource={tenants}
          columns={columns}
          loading={loading}
          pagination={false}
          scroll={{ x: 600 }}
          locale={{ emptyText: loading ? "Loading..." : "No tenants found" }}
        />
      </Card>

      <CreateTenantModal
        open={createModalOpen}
        onClose={() => setCreateModalOpen(false)}
        onSuccess={handleCreateSuccess}
      />

      <TypeToConfirmModal
        open={!!deactivateTarget}
        title="Deactivate Tenant"
        description={`This will deactivate tenant '${deactivateTarget?.name ?? ""}' and prevent all users from logging in. Data will be preserved.`}
        confirmText={deactivateTarget?.name ?? ""}
        onConfirm={handleDeactivate}
        onCancel={() => setDeactivateTarget(null)}
        danger
        loading={deactivateLoading}
      />
    </>
  );
}
