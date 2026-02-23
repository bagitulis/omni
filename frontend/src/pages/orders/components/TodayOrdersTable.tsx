import { useState, useEffect } from "react";
import { Card, Table, Tag, Typography, Flex, Spin, Empty } from "antd";
import type { ColumnsType } from "antd/es/table";
import apiClient from "@/api/client";

interface TodayOrderItem {
  order_sn: string;
  tracking_no?: string;
  courier?: string;
  seller_sku: string;
  product_name: string;
  variation_name?: string;
  quantity: number;
  platform?: string;
}

interface TodayOrdersData {
  items?: TodayOrderItem[];
  count?: number;
}

/**
 * TodayOrdersTable — displays today's shipped orders in a flat table.
 * Different from the generic OrderTable which groups by order_no.
 */
const platformColors: Record<string, string> = {
  shopee: "orange",
  lazada: "blue",
  tiktok: "cyan",
};

export function TodayOrdersTable() {
  const [items, setItems] = useState<TodayOrderItem[]>([]);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    const fetchToday = async () => {
      setLoading(true);
      try {
        const response =
          await apiClient.get<TodayOrdersData>("/orders/today");
        // Backend returns items at top level (not inside data wrapper)
        const raw = response as unknown as Record<string, unknown>;
        const data = (response.data?.items ?? (raw.items as TodayOrderItem[]) ?? []);
        setItems(data);
      } catch {
        setItems([]);
      } finally {
        setLoading(false);
      }
    };
    fetchToday();
  }, []);

  const columns: ColumnsType<TodayOrderItem> = [
    {
      title: "Order No.",
      dataIndex: "order_sn",
      key: "order_sn",
      width: 160,
      render: (v: string) => (
        <Typography.Text copyable style={{ fontSize: 12 }}>
          {v}
        </Typography.Text>
      ),
    },
    {
      title: "Tracking No.",
      dataIndex: "tracking_no",
      key: "tracking_no",
      width: 150,
      render: (v: string) => v || "-",
    },
    {
      title: "Courier",
      dataIndex: "courier",
      key: "courier",
      width: 120,
      render: (v: string) => v || "-",
    },
    {
      title: "Seller SKU",
      dataIndex: "seller_sku",
      key: "seller_sku",
      width: 140,
      render: (v: string) => (
        <Typography.Text code style={{ fontSize: 11 }}>
          {v || "-"}
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
      width: 130,
      render: (v: string) => v || "-",
    },
    {
      title: "Qty",
      dataIndex: "quantity",
      key: "quantity",
      width: 70,
      align: "center",
      render: (qty: number) => (
        <Tag color="blue" style={{ borderRadius: 3 }}>
          {qty || 1}
        </Tag>
      ),
    },
    {
      title: "Platform",
      dataIndex: "platform",
      key: "platform",
      width: 100,
      render: (v: string) => (
        <Tag
          color={platformColors[v?.toLowerCase()] ?? "default"}
          style={{ borderRadius: 3 }}
        >
          {v || "-"}
        </Tag>
      ),
    },
  ];

  if (loading) {
    return (
      <Card style={{ borderRadius: 3 }}>
        <Flex justify="center" align="center" style={{ padding: 48 }}>
          <Spin />
        </Flex>
      </Card>
    );
  }

  if (items.length === 0) {
    return (
      <Card style={{ borderRadius: 3 }}>
        <Empty description="No orders today" />
      </Card>
    );
  }

  return (
    <Card
      style={{ borderRadius: 3 }}
      title={`Today's Orders (${items.length})`}
    >
      <Table<TodayOrderItem>
        columns={columns}
        dataSource={items}
        rowKey={(r) => `${r.order_sn}-${r.seller_sku}`}
        size="small"
        pagination={{ pageSize: 20, size: "small" }}
        scroll={{ x: 900 }}
        style={{ fontSize: 12 }}
      />
    </Card>
  );
}
