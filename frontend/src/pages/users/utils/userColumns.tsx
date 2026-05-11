import { Button, Popconfirm, Space, Tag, Tooltip } from "antd";
import {
  DeleteOutlined,
  EditOutlined,
  LockOutlined,
  UnlockOutlined,
} from "@ant-design/icons";
import type { TableColumnsType } from "antd";
import type { UserResponse } from "@/api/users";
import dayjs from "dayjs";

const ROLE_LEVEL: Record<string, number> = {
  developer: 4,
  owner: 3,
  admin: 2,
  user: 1,
};

const ROLE_COLOR: Record<string, string> = {
  developer: "purple",
  owner: "gold",
  admin: "blue",
  user: "green",
};

interface ColumnParams {
  onEdit: (user: UserResponse) => void;
  onDelete: (id: string) => void;
  onUnlock: (id: string) => void;
  actorRole: string;
  currentUserId: string;
}

function canActOn(actorRole: string, targetRole: string): boolean {
  return (ROLE_LEVEL[actorRole] ?? 0) > (ROLE_LEVEL[targetRole] ?? 0);
}

export function getUserColumns({
  onEdit,
  onDelete,
  onUnlock,
  actorRole,
  currentUserId,
}: ColumnParams): TableColumnsType<UserResponse> {
  return [
    {
      title: "Username",
      dataIndex: "username",
      key: "username",
      width: 160,
      sorter: (a, b) => a.username.localeCompare(b.username),
    },
    {
      title: "Email",
      dataIndex: "email",
      key: "email",
      width: 220,
      ellipsis: true,
    },
    {
      title: "Role",
      dataIndex: "role",
      key: "role",
      width: 110,
      render: (role: string) => (
        <Tag color={ROLE_COLOR[role] ?? "default"}>
          {role.charAt(0).toUpperCase() + role.slice(1)}
        </Tag>
      ),
    },
    {
      title: "Status",
      key: "status",
      width: 100,
      render: (_: unknown, record: UserResponse) => {
        const isLocked =
          record.account_locked_until &&
          dayjs(record.account_locked_until).isAfter(dayjs());
        return isLocked ? (
          <Tag icon={<LockOutlined />} color="error">
            Locked
          </Tag>
        ) : (
          <Tag color="success">Active</Tag>
        );
      },
    },
    {
      title: "Created",
      dataIndex: "created_at",
      key: "created_at",
      width: 140,
      render: (val: string) => dayjs(val).format("YYYY-MM-DD HH:mm"),
      sorter: (a, b) => dayjs(a.created_at).unix() - dayjs(b.created_at).unix(),
    },
    {
      title: "Actions",
      key: "actions",
      width: 150,
      fixed: "right",
      render: (_: unknown, record: UserResponse) => {
        const canAct = canActOn(actorRole, record.role);
        const isSelf = record.id === currentUserId;
        const isLocked =
          record.account_locked_until &&
          dayjs(record.account_locked_until).isAfter(dayjs());

        return (
          <Space size="small">
            {canAct && (
              <Tooltip title="Edit">
                <Button
                  type="text"
                  size="small"
                  icon={<EditOutlined />}
                  onClick={() => onEdit(record)}
                />
              </Tooltip>
            )}
            {canAct && !isSelf && (
              <Popconfirm
                title="Delete user"
                description={`Are you sure you want to delete "${record.username}"?`}
                onConfirm={() => onDelete(record.id)}
                okText="Delete"
                okButtonProps={{ danger: true }}
              >
                <Tooltip title="Delete">
                  <Button
                    type="text"
                    size="small"
                    danger
                    icon={<DeleteOutlined />}
                  />
                </Tooltip>
              </Popconfirm>
            )}
            {canAct && isLocked && (
              <Tooltip title="Unlock">
                <Button
                  type="text"
                  size="small"
                  icon={<UnlockOutlined />}
                  onClick={() => onUnlock(record.id)}
                />
              </Tooltip>
            )}
          </Space>
        );
      },
    },
  ];
}
