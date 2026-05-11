import { useCallback, useEffect, useMemo, useState } from "react";
import { App, Button, Input, Table, Typography } from "antd";
import { PlusOutlined, SearchOutlined } from "@ant-design/icons";
import { useAuthStore } from "@/stores/authStore";
import { usePermission } from "@/hooks/usePermission";
import { usersApi, type UserResponse } from "@/api/users";
import { getUserColumns } from "./utils/userColumns";
import { CreateUserModal } from "./components/CreateUserModal";
import { EditUserModal } from "./components/EditUserModal";

const { Title } = Typography;

export default function UserManagementPage() {
  const { message } = App.useApp();
  const user = useAuthStore((state) => state.user);
  const accessToken = useAuthStore((state) => state.accessToken);
  const { userRole, hasPermission } = usePermission();
  const canCreate = hasPermission("users.create");

  const [users, setUsers] = useState<UserResponse[]>([]);
  const [loading, setLoading] = useState(false);
  const [pagination, setPagination] = useState({ page: 1, limit: 20, total: 0 });
  const [search, setSearch] = useState("");

  // Modal state
  const [createOpen, setCreateOpen] = useState(false);
  const [editOpen, setEditOpen] = useState(false);
  const [editUser, setEditUser] = useState<UserResponse | null>(null);

  const fetchUsers = useCallback(async (page = 1, limit = 20) => {
    setLoading(true);
    try {
      const response = await usersApi.list({ page, limit });
      if (response.success && response.data) {
        setUsers(response.data.users);
        setPagination({
          page: response.data.page,
          limit: response.data.limit,
          total: response.data.total,
        });
      } else {
        message.error(response.error ?? "Failed to load users");
      }
    } catch (err) {
      // 401 retry may have failed — don't show error if user was redirected
      if (useAuthStore.getState().isAuthenticated) {
        message.error(err instanceof Error ? err.message : "Failed to load users");
      }
    } finally {
      setLoading(false);
    }
  }, [message]);

  useEffect(() => {
    // Wait for access token to be available before fetching
    // Prevents race: page mounts before initializeAuth completes refresh
    if (accessToken) {
      fetchUsers();
    }
  }, [fetchUsers, accessToken]);

  const handleEdit = (record: UserResponse) => {
    setEditUser(record);
    setEditOpen(true);
  };

  const handleDelete = async (id: string) => {
    const response = await usersApi.delete(id);
    if (response.success) {
      message.success("User deleted");
      fetchUsers(pagination.page, pagination.limit);
    } else {
      message.error(response.error ?? "Failed to delete user");
    }
  };

  const handleUnlock = async (id: string) => {
    const response = await usersApi.unlock(id);
    if (response.success) {
      message.success("User unlocked");
      fetchUsers(pagination.page, pagination.limit);
    } else {
      message.error(response.error ?? "Failed to unlock user");
    }
  };

  const columns = useMemo(
    () =>
      getUserColumns({
        onEdit: handleEdit,
        onDelete: handleDelete,
        onUnlock: handleUnlock,
        actorRole: userRole,
        currentUserId: user?.id ?? "",
      }),
    // eslint-disable-next-line react-hooks/exhaustive-deps
    [userRole, user?.id, pagination.page],
  );

  const filteredUsers = useMemo(() => {
    if (!search.trim()) return users;
    const q = search.toLowerCase();
    return users.filter(
      (u) =>
        u.username.toLowerCase().includes(q) ||
        u.email.toLowerCase().includes(q),
    );
  }, [users, search]);

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
        <Title level={4} style={{ margin: 0 }}>
          User Management
        </Title>
        {canCreate && (
          <Button
            type="primary"
            icon={<PlusOutlined />}
            onClick={() => setCreateOpen(true)}
          >
            Add User
          </Button>
        )}
      </div>

      <Input
        placeholder="Search by username or email"
        prefix={<SearchOutlined />}
        value={search}
        onChange={(e) => setSearch(e.target.value)}
        style={{ marginBottom: 16, maxWidth: 320 }}
        allowClear
      />

      <Table<UserResponse>
        columns={columns}
        dataSource={filteredUsers}
        rowKey="id"
        loading={loading}
        size="small"
        scroll={{ x: 900 }}
        pagination={{
          current: pagination.page,
          pageSize: pagination.limit,
          total: pagination.total,
          showSizeChanger: true,
          showTotal: (total) => `Total ${total} users`,
          onChange: (page, pageSize) => fetchUsers(page, pageSize),
        }}
      />

      <CreateUserModal
        open={createOpen}
        onClose={() => setCreateOpen(false)}
        onSuccess={() => fetchUsers(pagination.page, pagination.limit)}
        actorRole={userRole}
      />

      <EditUserModal
        open={editOpen}
        user={editUser}
        onClose={() => {
          setEditOpen(false);
          setEditUser(null);
        }}
        onSuccess={() => fetchUsers(pagination.page, pagination.limit)}
        actorRole={userRole}
      />
    </div>
  );
}
