import { Button, Descriptions, Empty, Image, Space, Table, Tag, Typography } from "antd";
import type { ColumnsType } from "antd/es/table";
import type { Booking, BookingItem } from "@/types/booking";

type RecipientAddress = Record<string, unknown>;

export function maskPhone(phone: string): string {
  if (!phone) return "N/A";
  if (phone.length <= 7) return phone;
  return `${phone.slice(0, 3)}${"*".repeat(phone.length - 7)}${phone.slice(-4)}`;
}

export function formatAddress(addressJson: string): string[] {
  if (!addressJson) return [];

  try {
    const parsed = JSON.parse(addressJson) as RecipientAddress;
    const fields = [
      "full_address",
      "address",
      "street",
      "district",
      "city",
      "state",
      "province",
      "zipcode",
      "postal_code",
      "country",
    ];

    return fields
      .map((field) => parsed[field])
      .filter(
        (value): value is string =>
          typeof value === "string" && value.trim().length > 0,
      )
      .map((value) => value.trim());
  } catch {
    return [];
  }
}

export function DetailTag({ value, color }: { value: string; color: string }) {
  return (
    <Tag color={color} style={{ borderRadius: 3, fontSize: 12 }}>
      {value || "N/A"}
    </Tag>
  );
}

export function ParentOrderLink({ booking }: { booking: Booking }) {
  if (booking.has_parent_order && booking.order_sn) {
    return (
      <Button
        type="link"
        size="small"
        style={{ padding: 0, fontSize: 12 }}
        onClick={() => console.log("View parent order:", booking.order_sn)}
      >
        View Order
      </Button>
    );
  }

  if (booking.order_sn) {
    return (
      <Space direction="vertical" size={2}>
        <Typography.Text style={{ fontSize: 12 }}>
          Order SN: {booking.order_sn}
        </Typography.Text>
        <Typography.Text type="secondary" style={{ fontSize: 12 }}>
          Parent order not synced yet
        </Typography.Text>
      </Space>
    );
  }

  return (
    <Typography.Text type="secondary" style={{ fontSize: 12 }}>
      No matched order yet
    </Typography.Text>
  );
}

export function RecipientSummary({ booking }: { booking: Booking }) {
  const addressLines = formatAddress(booking.recipient_address_json);

  return (
    <Descriptions column={1} size="small" bordered>
      <Descriptions.Item label="Recipient Name">
        {booking.recipient_name || "N/A"}
      </Descriptions.Item>
      <Descriptions.Item label="Phone">
        {maskPhone(booking.recipient_phone)}
      </Descriptions.Item>
      <Descriptions.Item label="Address">
        {addressLines.length > 0 ? (
          <Space direction="vertical" size={2}>
            {addressLines.map((line) => (
              <Typography.Text key={line} style={{ fontSize: 12 }}>
                {line}
              </Typography.Text>
            ))}
          </Space>
        ) : (
          "N/A"
        )}
      </Descriptions.Item>
    </Descriptions>
  );
}

export function ItemsTable({ items }: { items: BookingItem[] }) {
  const columns: ColumnsType<BookingItem> = [
    {
      title: "Image",
      dataIndex: "image_url",
      key: "image_url",
      width: 64,
      render: (imageUrl: string, item) =>
        imageUrl ? (
          <Image
            src={imageUrl}
            alt={item.item_name || "Booking item"}
            width={48}
            height={48}
            style={{ objectFit: "cover", borderRadius: 3 }}
            preview={false}
          />
        ) : (
          <Empty
            image={Empty.PRESENTED_IMAGE_SIMPLE}
            description={false}
            style={{ margin: 0 }}
          />
        ),
    },
    {
      title: "Item Name",
      dataIndex: "item_name",
      key: "item_name",
      ellipsis: true,
      render: (value: string) => value || "N/A",
    },
    {
      title: "Model/SKU",
      key: "model_sku",
      width: 130,
      render: (_, item) => (
        <Space direction="vertical" size={2}>
          <Typography.Text style={{ fontSize: 12 }}>
            {item.model_name || "N/A"}
          </Typography.Text>
          <Typography.Text type="secondary" style={{ fontSize: 12 }}>
            {item.model_sku || item.item_sku || item.sku || "N/A"}
          </Typography.Text>
        </Space>
      ),
    },
    { title: "Qty", dataIndex: "quantity", key: "quantity", width: 56 },
    {
      title: "Weight",
      dataIndex: "weight",
      key: "weight",
      width: 80,
      render: (weight: number) => (weight ? `${weight} kg` : "N/A"),
    },
  ];

  if (items.length === 0) {
    return <Empty description="No items" />;
  }

  return (
    <Table<BookingItem>
      columns={columns}
      dataSource={items}
      rowKey={(item) => `${item.booking_sn}-${item.item_id}-${item.model_id}`}
      pagination={false}
      size="small"
      scroll={{ x: 480 }}
    />
  );
}
