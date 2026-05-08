import { Drawer, Table, Tag, Typography, theme, Empty, Badge, Spin } from "antd";
import { CheckCircleOutlined, CloseCircleOutlined } from "@ant-design/icons";
import type { FC } from "react";
import type {
  BulkOperationMetadata,
  FailedItemDetail,
} from "@/types/notificationMetadata";

const { Text, Title } = Typography;

interface SyncResultsDrawerProps {
  open: boolean;
  onClose: () => void;
  results: BulkOperationMetadata | null;
}

export const SyncResultsDrawer: FC<SyncResultsDrawerProps> = ({
  open,
  onClose,
  results,
}) => {
  const { token } = theme.useToken();

  if (!results) {
    return (
      <Drawer title="Sync Results" placement="right" width={480} open={open} onClose={onClose}>
        <Spin style={{ display: "block", margin: "60px auto" }} />
      </Drawer>
    );
  }

  const platformEntries = Object.entries(results.platforms || {});
  const failedItems = results.failed_items || [];

  const platformColumns = [
    {
      title: "Platform",
      dataIndex: "platform",
      key: "platform",
      render: (v: string) => (
        <Text strong style={{ textTransform: "capitalize" as const }}>
          {v}
        </Text>
      ),
    },
    {
      title: "Success",
      dataIndex: "succeeded",
      key: "succeeded",
      render: (v: number) => <Tag color="success">{v} ✓</Tag>,
    },
    {
      title: "Failed",
      dataIndex: "failed",
      key: "failed",
      render: (v: number) =>
        v > 0 ? <Tag color="error">{v} ✗</Tag> : <Tag>{v}</Tag>,
    },
  ];

  const platformData = platformEntries.map(([name, stats]) => ({
    key: name,
    platform: name,
    succeeded: stats.succeeded,
    failed: stats.failed,
  }));

  const failedColumns = [
    { title: "SKU", dataIndex: "sku", key: "sku", width: 120 },
    {
      title: "Platform",
      dataIndex: "platform",
      key: "platform",
      width: 100,
      render: (v: string) => <Tag>{v}</Tag>,
    },
    { title: "Error", dataIndex: "error", key: "error", ellipsis: true },
  ];

  const operationLabel =
    results.operation_type === "stock_sync" ? "Stock Sync" : "Price Sync";

  return (
    <Drawer
      title={`${operationLabel} Results`}
      placement="right"
      width={480}
      open={open}
      onClose={onClose}
      styles={{ body: { padding: "16px 24px" } }}
    >
      {/* Summary */}
      <div style={{ display: "flex", gap: 12, marginBottom: 20 }}>
        <Badge
          count={results.succeeded}
          showZero
          style={{ backgroundColor: token.colorSuccess }}
        >
          <Tag
            icon={<CheckCircleOutlined />}
            color="success"
            style={{ fontSize: 14, padding: "4px 12px" }}
          >
            Succeeded
          </Tag>
        </Badge>
        <Badge
          count={results.failed}
          showZero
          style={{
            backgroundColor:
              results.failed > 0
                ? token.colorError
                : token.colorTextQuaternary,
          }}
        >
          <Tag
            icon={<CloseCircleOutlined />}
            color={results.failed > 0 ? "error" : "default"}
            style={{ fontSize: 14, padding: "4px 12px" }}
          >
            Failed
          </Tag>
        </Badge>
      </div>

      {/* Per-Platform Table */}
      {platformData.length > 0 && (
        <>
          <Title level={5} style={{ marginBottom: 8 }}>
            Per Platform
          </Title>
          <Table
            columns={platformColumns}
            dataSource={platformData}
            pagination={false}
            size="small"
            style={{ marginBottom: 20 }}
          />
        </>
      )}

      {/* Failed Items */}
      {failedItems.length > 0 && (
        <>
          <Title level={5} style={{ marginBottom: 8, color: token.colorError }}>
            Failed Items ({failedItems.length})
          </Title>
          <Table<FailedItemDetail & { key: number }>
            columns={failedColumns}
            dataSource={failedItems.map((item, i) => ({ ...item, key: i }))}
            pagination={failedItems.length > 10 ? { pageSize: 10 } : false}
            size="small"
            scroll={{ y: 300 }}
          />
        </>
      )}

      {/* All success */}
      {failedItems.length === 0 && results.succeeded > 0 && (
        <Empty
          image={Empty.PRESENTED_IMAGE_SIMPLE}
          description="All items synced successfully!"
          style={{ marginTop: 40 }}
        />
      )}
    </Drawer>
  );
};
