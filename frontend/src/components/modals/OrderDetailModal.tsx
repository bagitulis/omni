import {
  Drawer,
  Descriptions,
  Table,
  Tag,
  Typography,
  Button,
  Space,
  Divider,
} from "antd";
import type { ColumnsType } from "antd/es/table";
import { OrderDetail, OrderItem } from "@/types/order";

const { Text } = Typography;

interface OrderDetailModalProps {
  open: boolean;
  onClose: () => void;
  order: OrderDetail | null;
}

export function OrderDetailModal({
  open,
  onClose,
  order,
}: OrderDetailModalProps) {
  if (!order) return null;

  const columns: ColumnsType<OrderItem> = [
    {
      title: "Product",
      dataIndex: "item_name",
      key: "item_name",
      render: (text, record) => (
        <div style={{ display: "flex", flexDirection: "column" }}>
          <Text strong>{text}</Text>
          <Text type="secondary" style={{ fontSize: 11 }}>
            SKU: {record.item_sku}
          </Text>
        </div>
      ),
    },
    {
      title: "Price",
      dataIndex: "price",
      key: "price",
      width: 100,
      render: (val) => val.toLocaleString(),
    },
    {
      title: "Qty",
      dataIndex: "quantity",
      key: "quantity",
      width: 80,
    },
    {
      title: "Total",
      dataIndex: "total",
      key: "total",
      width: 100,
      render: (val) => val.toLocaleString(),
    },
  ];

  const getStatusColor = (status: string) => {
    switch (status.toLowerCase()) {
      case "pending":
        return "warning";
      case "paid":
        return "blue";
      case "ready_to_ship":
        return "cyan";
      case "shipped":
        return "purple";
      case "completed":
        return "success";
      case "cancelled":
        return "error";
      default:
        return "default";
    }
  };

  return (
    <Drawer
      title={
        <Space>
          <span>Order Details</span>
          <Text code>#{order.order_sn}</Text>
        </Space>
      }
      placement="right"
      width={640}
      onClose={onClose}
      open={open}
      extra={<Button onClick={onClose}>Close</Button>}
    >
      <Descriptions title="Order Information" bordered size="small" column={2}>
        <Descriptions.Item label="Order SN">{order.order_sn}</Descriptions.Item>
        <Descriptions.Item label="Status">
          <Tag color={getStatusColor(order.order_status)}>
            {order.order_status}
          </Tag>
        </Descriptions.Item>
        <Descriptions.Item label="Platform">
          <Tag>{order.platform}</Tag>
        </Descriptions.Item>
        <Descriptions.Item label="Created At">
          {new Date(order.created_at).toLocaleString()}
        </Descriptions.Item>
        <Descriptions.Item label="Buyer">
          {order.buyer_username}
        </Descriptions.Item>
        <Descriptions.Item label="Total Amount">
          <Text strong style={{ fontSize: 16 }}>
            {order.total_amount.toLocaleString()}
          </Text>
        </Descriptions.Item>
      </Descriptions>

      <Divider />

      <Typography.Title level={5}>Order Items</Typography.Title>
      <Table<OrderItem>
        columns={columns}
        dataSource={order.items}
        rowKey="item_id"
        pagination={false}
        size="small"
        bordered
      />

      {order.shipping_address && (
        <>
          <Divider />
          <Descriptions
            title="Shipping Address"
            bordered
            size="small"
            column={1}
          >
            <Descriptions.Item label="Receiver">
              {order.shipping_address.receiver_name} (
              {order.shipping_address.receiver_phone})
            </Descriptions.Item>
            <Descriptions.Item label="Address">
              {order.shipping_address.address}
            </Descriptions.Item>
            <Descriptions.Item label="Location">
              {order.shipping_address.city}, {order.shipping_address.state}{" "}
              {order.shipping_address.postal_code},{" "}
              {order.shipping_address.country}
            </Descriptions.Item>
          </Descriptions>
        </>
      )}
    </Drawer>
  );
}
