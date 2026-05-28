import { useEffect, useState } from "react";
import {
  Alert,
  Descriptions,
  Drawer,
  Empty,
  Flex,
  Grid,
  Space,
  Spin,
  Typography,
  theme,
} from "antd";
import { getBookingOrderDetail } from "@/api/orders";
import type { Booking, BookingItem } from "@/types/booking";
import {
  formatBookingStatus,
  formatBookingTime,
  formatMatchStatus,
  getBookingStatusColor,
  getMatchStatusColor,
} from "../utils/bookingTransforms";
import {
  DetailTag,
  ItemsTable,
  ParentOrderLink,
  RecipientSummary,
} from "./BookingDetailDrawerSections";

interface BookingDetailDrawerProps {
  bookingSn: string | null;
  open: boolean;
  onClose: () => void;
  platform: string;
}

export function BookingDetailDrawer({
  bookingSn,
  open,
  onClose,
  platform,
}: BookingDetailDrawerProps) {
  const screens = Grid.useBreakpoint();
  const { token } = theme.useToken();
  const [booking, setBooking] = useState<Booking | null>(null);
  const [items, setItems] = useState<BookingItem[]>([]);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState(false);

  useEffect(() => {
    if (!open || !bookingSn) {
      setBooking(null);
      setItems([]);
      setError(false);
      return;
    }

    let ignore = false;
    setLoading(true);
    setError(false);

    getBookingOrderDetail(bookingSn)
      .then((response) => {
        if (ignore) return;
        setBooking(response.data.booking);
        setItems(response.data.items || []);
      })
      .catch(() => {
        if (ignore) return;
        setBooking(null);
        setItems([]);
        setError(true);
      })
      .finally(() => {
        if (!ignore) setLoading(false);
      });

    return () => {
      ignore = true;
    };
  }, [bookingSn, open]);

  const title = booking ? (
    <Space>
      <Typography.Text strong>{`Booking ${booking.booking_sn}`}</Typography.Text>
      <DetailTag
        color={getBookingStatusColor(booking.booking_status)}
        value={formatBookingStatus(booking.booking_status)}
      />
    </Space>
  ) : (
    <Typography.Text strong>{bookingSn ? `Booking ${bookingSn}` : "Booking Detail"}</Typography.Text>
  );

  return (
    <Drawer
      title={title}
      open={open}
      onClose={onClose}
      width={screens.md ? 520 : "100vw"}
      styles={{ body: { padding: 16, background: token.colorBgLayout } }}
      afterOpenChange={(visible) => {
        if (visible) window.setTimeout(() => document.querySelector<HTMLButtonElement>(".ant-drawer-close")?.focus(), 0);
      }}
    >
      {loading ? (
        <Flex justify="center" align="center" style={{ minHeight: 240 }}>
          <Spin />
        </Flex>
      ) : error ? (
        <Alert type="error" message="Unable to load booking detail" showIcon />
      ) : booking ? (
        <Space direction="vertical" size={16} style={{ width: "100%" }}>
          <Typography.Text type="secondary" style={{ fontSize: 12 }}>
            Platform: {platform}
          </Typography.Text>

          <section>
            <Typography.Title level={5} style={{ marginTop: 0, marginBottom: 8 }}>
              Booking Metadata
            </Typography.Title>
            <Descriptions column={1} size="small" bordered>
              <Descriptions.Item label="Booking SN">{booking.booking_sn}</Descriptions.Item>
              <Descriptions.Item label="Order SN">{booking.order_sn || "N/A"}</Descriptions.Item>
              <Descriptions.Item label="Booking Status">
                <DetailTag
                  color={getBookingStatusColor(booking.booking_status)}
                  value={formatBookingStatus(booking.booking_status)}
                />
              </Descriptions.Item>
              <Descriptions.Item label="Match Status">
                <DetailTag
                  color={getMatchStatusColor(booking.match_status)}
                  value={formatMatchStatus(booking.match_status)}
                />
              </Descriptions.Item>
              <Descriptions.Item label="Shipping Carrier">
                {booking.shipping_carrier || "N/A"}
              </Descriptions.Item>
              <Descriptions.Item label="Fulfillment Flag">
                {booking.fulfillment_flag || "N/A"}
              </Descriptions.Item>
              <Descriptions.Item label="Create Time">
                {formatBookingTime(booking.create_time) || "N/A"}
              </Descriptions.Item>
              <Descriptions.Item label="Update Time">
                {formatBookingTime(booking.update_time) || "N/A"}
              </Descriptions.Item>
            </Descriptions>
          </section>

          <section>
            <Typography.Title level={5} style={{ marginTop: 0, marginBottom: 8 }}>
              Parent Order Link
            </Typography.Title>
            <ParentOrderLink booking={booking} />
          </section>

          <section>
            <Typography.Title level={5} style={{ marginTop: 0, marginBottom: 8 }}>
              Recipient Information
            </Typography.Title>
            <RecipientSummary booking={booking} />
          </section>

          <section>
            <Typography.Title level={5} style={{ marginTop: 0, marginBottom: 8 }}>
              Items List
            </Typography.Title>
            <ItemsTable items={items} />
          </section>
        </Space>
      ) : (
        <Empty description="No booking selected" />
      )}
    </Drawer>
  );
}
