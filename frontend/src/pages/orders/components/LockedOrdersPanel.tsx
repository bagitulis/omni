import { useState, useEffect, useCallback } from "react";
import {
  Button,
  Card,
  Empty,
  Flex,
  Space,
  Spin,
  Table,
  Tag,
  Typography,
  message,
} from "antd";
import { LockOutlined, SyncOutlined } from "@ant-design/icons";
import type { ColumnsType } from "antd/es/table";
import {
  syncLockedToday,
  getLockedOrders,
  type LockedOrderItem,
} from "@/api/lockedOrders";

/**
 * LockedOrdersPanel — displays locked (pending) orders aggregated by SKU.
 * Shows which SKUs have orders that haven't been shipped yet.
 */
export function LockedOrdersPanel() {
  const [items, setItems] = useState<LockedOrderItem[]>([]);
  const [loading, setLoading] = useState(false);
  const [syncing, setSyncing] = useState(false);
  const [loaded, setLoaded] = useState(false);

  const handleFetch = useCallback(async () => {
    setLoading(true);
    try {
      const data = await getLockedOrders();
      setItems(data);
      setLoaded(true);
    } catch {
      message.error("Failed to fetch locked orders");
    } finally {
      setLoading(false);
    }
  }, []);

  // Auto-fetch locked orders on mount
  useEffect(() => {
    void handleFetch();
  }, [handleFetch]);

  const handleSync = async () => {
    setSyncing(true);
    try {
      const data = await syncLockedToday(7);
      setItems(data);
      setLoaded(true);
      message.success(`Synced ${data.length} locked orders`);
    } catch {
      message.error("Failed to sync locked orders");
    } finally {
      setSyncing(false);
    }
  };

  const columns: ColumnsType<LockedOrderItem> = [
    {
      title: "SKU",
      dataIndex: "sku",
      key: "sku",
      width: 180,
      render: (sku: string) => (
        <Typography.Text code style={{ fontSize: 12 }}>
          {sku}
        </Typography.Text>
      ),
    },
    {
      title: "Product",
      dataIndex: "product_name",
      key: "product_name",
      ellipsis: true,
    },
    {
      title: "Variation",
      dataIndex: "variation_name",
      key: "variation_name",
      width: 150,
      render: (v: string) => v || "-",
    },
    {
      title: "Locked Qty",
      dataIndex: "qty",
      key: "qty",
      width: 100,
      align: "center",
      render: (qty: number) => (
        <Tag color="error" style={{ borderRadius: 3, fontWeight: 600 }}>
          {qty}
        </Tag>
      ),
    },
  ];

  const totalLocked = items.reduce((sum, item) => sum + item.qty, 0);

  return (
    <Card
      size="small"
      title={
        <Flex align="center" gap={8}>
          <LockOutlined />
          <span>Locked Orders (Pending Shipment)</span>
          {loaded && (
            <Tag color="error" style={{ borderRadius: 3 }}>
              {items.length} SKUs · {totalLocked} pcs locked
            </Tag>
          )}
        </Flex>
      }
      extra={
        <Space>
          <Button
            size="small"
            onClick={handleFetch}
            loading={loading}
            disabled={syncing}
          >
            Load
          </Button>
          <Button
            size="small"
            type="primary"
            icon={<SyncOutlined spin={syncing} />}
            onClick={handleSync}
            loading={syncing}
            disabled={loading}
          >
            Sync & Refresh
          </Button>
        </Space>
      }
      style={{ borderRadius: 3 }}
    >
      {loading && !loaded ? (
        <Flex justify="center" align="center" style={{ padding: 48 }}>
          <Spin />
        </Flex>
      ) : !loaded ? (
        <Empty
          description="Click Load or Sync to view locked orders"
          image={Empty.PRESENTED_IMAGE_SIMPLE}
        />
      ) : (
        <Table<LockedOrderItem>
          columns={columns}
          dataSource={items}
          rowKey={(record) => `${record.sku}-${record.product_name}`}
          size="small"
          pagination={{ pageSize: 20, size: "small" }}
          scroll={{ y: 400 }}
          style={{ fontSize: 12 }}
        />
      )}
    </Card>
  );
}
