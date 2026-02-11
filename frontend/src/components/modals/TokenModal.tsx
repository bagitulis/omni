import React from "react";
import { Modal, Button, Table, Tag, Space, Alert, Spin } from "antd";
import { ReloadOutlined } from "@ant-design/icons";
import {
  useAllTokenStatus,
  useRefreshToken,
  useRefreshAllTokens,
} from "@/hooks/useTokens";
import type { PlatformTokenStatus } from "@/api/tokens";
import type { ColumnsType } from "antd/es/table";

interface TokenModalProps {
  open: boolean;
  onClose: () => void;
}

export const TokenModal: React.FC<TokenModalProps> = ({ open, onClose }) => {
  const { data: tokenStatuses, isLoading, error } = useAllTokenStatus();
  const { mutate: refreshToken, isPending: isRefreshing } = useRefreshToken();
  const { mutate: refreshAll, isPending: isRefreshingAll } =
    useRefreshAllTokens();

  // Convert Record<string, PlatformTokenStatus> to array for Table
  const tableData = React.useMemo(() => {
    if (!tokenStatuses) return [];
    return Object.values(tokenStatuses);
  }, [tokenStatuses]);

  const getStatusTag = (status: PlatformTokenStatus) => {
    if (status.isExpired) {
      return <Tag color="error">Expired</Tag>;
    }
    if (status.needsRefresh) {
      return <Tag color="warning">Needs Refresh</Tag>;
    }
    if (status.isValid) {
      return <Tag color="success">Active</Tag>;
    }
    return <Tag color="default">Unknown</Tag>;
  };

  const formatDate = (dateString?: string) => {
    if (!dateString) return "—";
    try {
      const date = new Date(dateString);
      return date.toLocaleString("en-US", {
        year: "numeric",
        month: "short",
        day: "numeric",
        hour: "2-digit",
        minute: "2-digit",
      });
    } catch {
      return dateString;
    }
  };

  const columns: ColumnsType<PlatformTokenStatus> = [
    {
      title: "Platform",
      dataIndex: "platform",
      key: "platform",
      width: 120,
      render: (platform: string) =>
        platform.charAt(0).toUpperCase() + platform.slice(1),
    },
    {
      title: "Status",
      key: "status",
      width: 140,
      render: (_, record) => getStatusTag(record),
    },
    {
      title: "Expires At",
      dataIndex: "expiresAt",
      key: "expiresAt",
      width: 180,
      render: formatDate,
    },
    {
      title: "Action",
      key: "action",
      width: 100,
      align: "center",
      render: (_, record) => (
        <Button
          size="small"
          icon={<ReloadOutlined />}
          onClick={() => refreshToken(record.platform)}
          disabled={isRefreshing || isRefreshingAll}
        >
          Refresh
        </Button>
      ),
    },
  ];

  const renderContent = () => {
    if (isLoading) {
      return (
        <div style={{ textAlign: "center", padding: "48px 0" }}>
          <Spin size="large" />
          <div style={{ marginTop: "16px", color: "#666" }}>
            Loading token status...
          </div>
        </div>
      );
    }

    if (error) {
      return (
        <Alert
          message="Failed to Load Token Status"
          description={
            error instanceof Error ? error.message : "Unknown error occurred"
          }
          type="error"
          showIcon
        />
      );
    }

    if (!tableData || tableData.length === 0) {
      return (
        <Alert
          message="No Tokens Found"
          description="No platform tokens are configured yet."
          type="info"
          showIcon
        />
      );
    }

    return (
      <Table<PlatformTokenStatus>
        columns={columns}
        dataSource={tableData}
        rowKey="platform"
        pagination={false}
        size="small"
        style={{ marginTop: "8px" }}
      />
    );
  };

  return (
    <Modal
      title="Token Management"
      open={open}
      onCancel={onClose}
      destroyOnHidden
      width={700}
      footer={[
        <Button
          key="refresh-all"
          type="primary"
          icon={<ReloadOutlined />}
          onClick={() => refreshAll(undefined)}
          disabled={isRefreshingAll || isRefreshing}
          loading={isRefreshingAll}
        >
          Refresh All Tokens
        </Button>,
        <Button key="close" onClick={onClose}>
          Close
        </Button>,
      ]}
    >
      <Space direction="vertical" size="middle" style={{ width: "100%" }}>
        <Alert
          message="Manage API tokens for all connected platforms"
          description="View token status and refresh tokens as needed. Tokens are automatically refreshed when they expire."
          type="info"
          showIcon
        />
        {renderContent()}
      </Space>
    </Modal>
  );
};

export default TokenModal;
