import { Table, Badge, Button } from "antd";
import type { ProductPlatform } from "../types";

interface ProductSyncTabProps {
  platforms: ProductPlatform[];
  onSync: (platform: string) => void;
  loading: boolean;
}

export const ProductSyncTab = ({
  platforms,
  onSync,
  loading,
}: ProductSyncTabProps) => {
  const columns = [
    { title: "Platform", dataIndex: "platform", key: "platform" },
    {
      title: "Status",
      dataIndex: "status",
      key: "status",
      render: (status: string) => {
        const color: "success" | "warning" | "error" =
          status === "synced"
            ? "success"
            : status === "pending"
              ? "warning"
              : "error";
        return <Badge status={color} text={status.toUpperCase()} />;
      },
    },
    { title: "Last Sync", dataIndex: "last_sync", key: "last_sync" },
    {
      title: "Action",
      key: "action",
      render: (_: unknown, record: ProductPlatform) => (
        <Button
          size="small"
          onClick={() => onSync(record.platform)}
          loading={loading}
        >
          Sync Now
        </Button>
      ),
    },
  ];
  return <Table dataSource={platforms} columns={columns} pagination={false} />;
};
