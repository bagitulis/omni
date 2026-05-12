import { useCallback, useEffect, useRef, useState } from "react";
import { App, Button, Input, Space, Table, Tag, Typography } from "antd";
import { SearchOutlined, LockOutlined, StopOutlined } from "@ant-design/icons";
import type { ColumnsType } from "antd/es/table";
import { developerApi, type CrossTenantUser } from "@/api/developer";
import { ResetPasswordModal } from "../components/ResetPasswordModal";
import { BulkActionModal } from "../components/BulkActionModal";

const { Text } = Typography;

type BulkAction = "reset-passwords" | "disable";

export default function UsersTab() {
  const { message } = App.useApp();
  const [query, setQuery] = useState("");
  const [users, setUsers] = useState<CrossTenantUser[]>([]);
  const [loading, setLoading] = useState(false);
  const [searched, setSearched] = useState(false);
  const [selectedRowKeys, setSelectedRowKeys] = useState<React.Key[]>([]);
  const debounceRef = useRef<ReturnType<typeof setTimeout> | null>(null);

  // Modal state
  const [resetUser, setResetUser] = useState<CrossTenantUser | null>(null);
  const [bulkAction, setBulkAction] = useState<BulkAction | null>(null);

  const searchUsers = useCallback(
    async (q: string) => {
      if (q.length < 2) {
        setUsers([]);
        setSearched(false);
        return;
      }
      setLoading(true);
      setSearched(true);
      try {
        const response = await developerApi.searchUsers(q);
        if (response.success && response.data) {
          setUsers(response.data);
        } else {
          message.error(response.error ?? "Search failed");
          setUsers([]);
        }
      } catch {
        message.error("Search failed");
        setUsers([]);
      } finally {
        setLoading(false);
      }
    },
    [message],
  );

  const handleInputChange = (value: string) => {
    setQuery(value);
    if (debounceRef.current) clearTimeout(debounceRef.current);
    debounceRef.current = setTimeout(() => {
      searchUsers(value.trim());
    }, 300);
  };

  useEffect(() => {
    return () => {
      if (debounceRef.current) clearTimeout(debounceRef.current);
    };
  }, []);

  const handleResetSuccess = () => {
    if (query.length >= 2) searchUsers(query.trim());
  };

  const handleBulkSuccess = () => {
    setSelectedRowKeys([]);
    if (query.length >= 2) searchUsers(query.trim());
  };

  const columns: ColumnsType<CrossTenantUser> = [
    {
      title: "Username",
      dataIndex: "username",
      width: 150,
    },
    {
      title: "Email",
      dataIndex: "email",
      width: 200,
    },
    {
      title: "Role",
      dataIndex: "role",
      width: 100,
      render: (role: string) => {
        const colorMap: Record<string, string> = {
          developer: "purple",
          owner: "gold",
          admin: "blue",
          user: "default",
        };
        return <Tag color={colorMap[role] || "default"}>{role}</Tag>;
      },
    },
    {
      title: "Tenant",
      dataIndex: "tenant_name",
      width: 150,
    },
    {
      title: "Status",
      dataIndex: "status",
      width: 100,
      render: (status: string) => (
        <Tag color={status === "active" ? "success" : "error"}>{status}</Tag>
      ),
    },
    {
      title: "Actions",
      key: "actions",
      width: 140,
      render: (_, record) => (
        <Button
          size="small"
          icon={<LockOutlined />}
          onClick={() => setResetUser(record)}
        >
          Reset Password
        </Button>
      ),
    },
  ];

  const hasSelection = selectedRowKeys.length > 0;

  return (
    <div>
      {/* Search */}
      <Input
        prefix={<SearchOutlined />}
        placeholder="Search users by name or email (min 2 characters)"
        value={query}
        onChange={(e) => handleInputChange(e.target.value)}
        allowClear
        style={{ maxWidth: 400, marginBottom: 16 }}
      />

      {/* Bulk action bar */}
      {hasSelection && (
        <div
          style={{
            marginBottom: 16,
            padding: "8px 12px",
            background: "#e0f2fe",
            borderRadius: 3,
            display: "flex",
            alignItems: "center",
            gap: 12,
          }}
        >
          <Text style={{ fontSize: 12 }}>
            {selectedRowKeys.length} user{selectedRowKeys.length > 1 ? "s" : ""} selected
          </Text>
          <Space size="small">
            <Button
              size="small"
              icon={<LockOutlined />}
              onClick={() => setBulkAction("reset-passwords")}
            >
              Reset All Passwords
            </Button>
            <Button
              size="small"
              danger
              icon={<StopOutlined />}
              onClick={() => setBulkAction("disable")}
            >
              Disable All
            </Button>
          </Space>
        </div>
      )}

      {/* Empty state */}
      {!searched && !loading && (
        <div style={{ textAlign: "center", padding: "48px 0" }}>
          <Text type="secondary">
            Enter a search query to find users across all tenants.
          </Text>
        </div>
      )}

      {/* Results table */}
      {searched && (
        <Table<CrossTenantUser>
          columns={columns}
          dataSource={users}
          rowKey="id"
          loading={loading}
          size="small"
          pagination={{ pageSize: 20, showSizeChanger: false }}
          rowSelection={{
            selectedRowKeys,
            onChange: (keys) => setSelectedRowKeys(keys),
          }}
          locale={{ emptyText: "No users found" }}
        />
      )}

      {/* Modals */}
      <ResetPasswordModal
        open={!!resetUser}
        user={resetUser}
        onClose={() => setResetUser(null)}
        onSuccess={handleResetSuccess}
      />

      {bulkAction && (
        <BulkActionModal
          open={!!bulkAction}
          action={bulkAction}
          userIds={selectedRowKeys as string[]}
          onClose={() => setBulkAction(null)}
          onSuccess={handleBulkSuccess}
        />
      )}
    </div>
  );
}
