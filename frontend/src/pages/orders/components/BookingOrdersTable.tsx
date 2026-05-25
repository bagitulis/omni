import { useCallback, useEffect, useState } from "react";
import {
  Alert,
  Button,
  Card,
  Empty,
  Flex,
  Spin,
  Table,
  Tag,
  Typography,
  theme,
} from "antd";
import type { ColumnsType } from "antd/es/table";
import { getBookingOrders } from "@/api/orders";
import type { Booking } from "@/types/booking";
import {
  formatBookingStatus,
  formatBookingTime,
  formatMatchStatus,
  getBookingStatusColor,
} from "../utils/bookingTransforms";

interface BookingOrdersTableProps {
  platform: string;
  onViewDetail: (booking: Booking) => void;
  refreshTrigger?: number;
}

export function BookingOrdersTable({
  platform,
  onViewDetail,
  refreshTrigger = 0,
}: BookingOrdersTableProps) {
  const [bookings, setBookings] = useState<Booking[]>([]);
  const [loading, setLoading] = useState(true);
  const [total, setTotal] = useState(0);
  const [page, setPage] = useState(1);
  const [pageSize, setPageSize] = useState(20);
  const [error, setError] = useState<string | null>(null);
  const { token } = theme.useToken();

  const fetchBookings = useCallback(async () => {
    // If platform is lazada/tiktok, show unavailable message
    if (platform && platform !== "all" && platform !== "shopee") {
      setBookings([]);
      setTotal(0);
      setLoading(false);
      return;
    }

    setLoading(true);
    setError(null);
    try {
      const response = await getBookingOrders({
        page,
        page_size: pageSize,
        platform: platform === "all" ? undefined : platform,
      });
      setBookings(response.data || []);
      setTotal(response.pagination?.total ?? response.count ?? 0);
    } catch (err) {
      console.warn("Failed to fetch booking orders:", err);
      setError(err instanceof Error ? err.message : "Failed to load booking orders");
      setBookings([]);
    } finally {
      setLoading(false);
    }
  }, [page, pageSize, platform]);

  useEffect(() => {
    if (refreshTrigger >= 0) {
      void fetchBookings();
    }
  }, [fetchBookings, refreshTrigger]);

  // Check if platform should show unavailable message
  const isPlatformUnavailable =
    platform && platform !== "all" && platform !== "shopee";

  if (isPlatformUnavailable) {
    return (
      <Card
        style={{ borderRadius: 3, boxShadow: token.boxShadow }}
      >
        <Empty description="Booking orders are only available for Shopee in this version." />
      </Card>
    );
  }

  const columns: ColumnsType<Booking> = [
    {
      title: "Booking SN",
      dataIndex: "booking_sn",
      key: "booking_sn",
      width: 170,
      fixed: "left",
      render: (v: string) => (
        <Typography.Text
          copyable
          style={{ fontSize: 12, fontFamily: "monospace" }}
        >
          {v}
        </Typography.Text>
      ),
    },
    {
      title: "Order SN",
      dataIndex: "has_parent_order",
      key: "order_sn",
      width: 170,
      render: (_: boolean, record: Booking) => {
        if (record.has_parent_order && record.order_sn) {
          return (
            <Button
              type="link"
              size="small"
              style={{ padding: 0, fontSize: 12, fontFamily: "monospace" }}
              onClick={() => window.open(`/orders/${record.order_sn}`, "_blank")}
            >
              {record.order_sn}
            </Button>
          );
        }
        return (
          <Typography.Text
            type="secondary"
            style={{ fontSize: 12 }}
          >
            Not matched yet
          </Typography.Text>
        );
      },
    },
    {
      title: "Booking Status",
      dataIndex: "booking_status",
      key: "booking_status",
      width: 130,
      render: (v: string) => (
        <Tag color={getBookingStatusColor(v)} style={{ borderRadius: 3, fontSize: 12 }}>
          {formatBookingStatus(v)}
        </Tag>
      ),
    },
    {
      title: "Match Status",
      dataIndex: "match_status",
      key: "match_status",
      width: 120,
      render: (v: string) => {
        if (!v) return <Typography.Text type="secondary">-</Typography.Text>;
        return (
          <Typography.Text style={{ fontSize: 12 }}>
            {formatMatchStatus(v)}
          </Typography.Text>
        );
      },
    },
    {
      title: "Recipient",
      dataIndex: "recipient_name",
      key: "recipient_name",
      width: 140,
      ellipsis: true,
      render: (v: string) => v || "-",
    },
    {
      title: "Items",
      key: "items",
      width: 120,
      render: (_: unknown, record: Booking) => (
        <Typography.Text style={{ fontSize: 12 }} ellipsis>
          {record.item_count} items
        </Typography.Text>
      ),
    },
    {
      title: "Courier",
      dataIndex: "shipping_carrier",
      key: "shipping_carrier",
      width: 120,
      ellipsis: true,
      render: (v: string) => v || "-",
    },
    {
      title: "Created",
      dataIndex: "create_time",
      key: "create_time",
      width: 160,
      render: (v: number) => (
        <Typography.Text style={{ fontSize: 12 }}>
          {formatBookingTime(v)}
        </Typography.Text>
      ),
    },
    {
      title: "Updated",
      dataIndex: "update_time",
      key: "update_time",
      width: 160,
      render: (v: number) => (
        <Typography.Text style={{ fontSize: 12 }}>
          {formatBookingTime(v)}
        </Typography.Text>
      ),
    },
    {
      title: "Actions",
      key: "actions",
      width: 90,
      fixed: "right",
      render: (_: unknown, record: Booking) => (
        <Button
          type="link"
          size="small"
          style={{ fontSize: 12 }}
          onClick={() => onViewDetail(record)}
        >
          Details
        </Button>
      ),
    },
  ];

  if (loading) {
    return (
      <Card style={{ borderRadius: 3, boxShadow: token.boxShadow }}>
        <Flex justify="center" align="center" style={{ padding: 48 }}>
          <Spin tip="Loading booking orders..." />
        </Flex>
      </Card>
    );
  }

  if (error) {
    return (
      <Card style={{ borderRadius: 3, boxShadow: token.boxShadow }}>
        <Alert
          type="error"
          message="Unable to sync booking orders"
          description={error}
          showIcon
          action={<Button size="small" onClick={fetchBookings}>Retry</Button>}
        />
      </Card>
    );
  }

  if (bookings.length === 0) {
    return (
      <Card style={{ borderRadius: 3, boxShadow: token.boxShadow }}>
        <Empty description="No booking orders found. Click Refresh to sync latest Shopee bookings.">
          <Button type="primary" onClick={fetchBookings}>Refresh</Button>
        </Empty>
      </Card>
    );
  }

  return (
    <Card
      style={{ borderRadius: 3, boxShadow: token.boxShadow, overflow: "hidden" }}
      styles={{
        header: {
          background: token.colorBgLayout,
          borderBottom: `1px solid ${token.colorBorderSecondary}`,
          padding: "12px 16px",
          minHeight: "auto",
        },
        body: { padding: 0 },
      }}
      title={
        <span style={{ fontSize: 14, fontWeight: 600 }}>
          Booking Orders
          {total > 0 && (
            <span
              style={{
                marginLeft: 8,
                fontSize: 12,
                fontWeight: 400,
                color: token.colorTextSecondary,
              }}
            >
              ({total} items)
            </span>
          )}
        </span>
      }
    >
      <Table<Booking>
        columns={columns}
        dataSource={bookings}
        rowKey="booking_sn"
        size="small"
        pagination={{
          current: page,
          pageSize,
          total,
          onChange: (p, ps) => {
            setPage(p);
            setPageSize(ps);
          },
          showSizeChanger: true,
          pageSizeOptions: ["10", "20", "50"],
          size: "small",
        }}
        scroll={{ x: 1440 }}
        style={{ fontSize: 12 }}
      />
    </Card>
  );
}
