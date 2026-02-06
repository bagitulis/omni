import { useMemo, useState } from "react";
import { Table, Avatar, Button, Typography, Tag, Dropdown, Space, Flex } from "antd";
import type { ColumnsType } from "antd/es/table";
import { 
  UserOutlined, 
  MessageOutlined, 
  CopyOutlined, 
  CheckOutlined, 
  MoreOutlined,
  ShoppingOutlined 
} from "@ant-design/icons";
import { Order } from "@/types/order";
import { StatusPipeline } from "../ui/StatusPipeline";

const { Text } = Typography;

interface OrderItem {
  sku: string;
  product_name: string;
  variation_name?: string;
  qty: number;
  price: number;
  product_image?: string;
}

interface GroupedOrder {
  key: string;
  order_no: string;
  order_sn?: string;
  buyer_username: string;
  platform: string;
  status: string;
  total_amount: number;
  currency: string;
  payment_method?: string;
  shipping_carrier?: string;
  ship_by_date?: number;
  items: OrderItem[];
}

interface OrderTableProps {
  orders: Order[];
  loading: boolean;
  pagination: {
    current: number;
    pageSize: number;
    total: number;
    onChange: (page: number, pageSize: number) => void;
  };
  selectedRowKeys: React.Key[];
  onSelectionChange: (selectedRowKeys: React.Key[]) => void;
  onShip: (order: any) => void;
  onPrint: (order: any) => void;
  onViewDetail?: (order: any) => void;
}

// ============ HELPER FUNCTIONS ============

function formatAmount(amt: number, cur: string): string {
  return new Intl.NumberFormat("id-ID", {
    style: "currency",
    currency: cur || "IDR",
    minimumFractionDigits: 0,
  }).format(amt);
}

function getCountdown(shipByDate?: number): string {
  if (!shipByDate) return "-";
  const now = Date.now();
  const deadline = shipByDate * 1000;
  const diff = deadline - now;
  if (diff <= 0) return "Overdue";

  const hours = Math.floor(diff / (1000 * 60 * 60));
  const minutes = Math.floor((diff % (1000 * 60 * 60)) / (1000 * 60));

  if (hours >= 24) {
    const days = Math.floor(hours / 24);
    return `${days}d ${hours % 24}h`;
  }
  return `${hours}h ${minutes}m`;
}

function getCountdownColor(shipByDate?: number): string {
  if (!shipByDate) return "#666";
  const now = Date.now();
  const deadline = shipByDate * 1000;
  const diff = deadline - now;
  if (diff <= 0) return "#f5222d";
  if (diff <= 24 * 60 * 60 * 1000) return "#fa8c16";
  return "#666";
}

// ============ SUB COMPONENTS ============

function CopyButton({ text }: { text: string }) {
  const [copied, setCopied] = useState(false);

  const handleCopy = () => {
    navigator.clipboard.writeText(text);
    setCopied(true);
    setTimeout(() => setCopied(false), 2000);
  };

  return (
    <Button
      type="text"
      size="small"
      icon={copied ? <CheckOutlined style={{ color: "#52c41a" }} /> : <CopyOutlined style={{ color: "#999" }} />}
      onClick={handleCopy}
      style={{ padding: 0, height: "auto" }}
    />
  );
}

// Product Cell with Buyer Header
function ProductCellWithBuyer({ order }: { order: GroupedOrder }) {
  return (
    <Flex vertical gap={8}>
      {/* Buyer Header */}
      <Flex 
        justify="space-between" 
        align="center"
        style={{
          backgroundColor: "#f5f5f5",
          padding: "8px 12px",
          borderRadius: 4,
          marginBottom: 4,
        }}
      >
        <Space size="small">
          <Avatar size={24} icon={<UserOutlined />} style={{ backgroundColor: "#0369a1" }} />
          <Text strong style={{ fontSize: 13 }}>{order.buyer_username}</Text>
          <Button type="text" size="small" icon={<MessageOutlined style={{ color: "#999", fontSize: 12 }} />} style={{ padding: 0 }} />
        </Space>
        <Space size={4}>
          <Text type="secondary" style={{ fontSize: 11 }}>ID:</Text>
          <Text style={{ fontSize: 12, fontFamily: "monospace" }}>{order.order_no}</Text>
          <CopyButton text={order.order_no} />
        </Space>
      </Flex>

      {/* Products */}
      <Flex vertical gap={10}>
        {order.items.map((item, idx) => (
          <Flex key={`${item.sku || "item"}-${idx}`} gap={10} align="flex-start">
            <Avatar
              shape="square"
              size={52}
              src={item.product_image}
              icon={<ShoppingOutlined />}
              style={{ 
                flexShrink: 0, 
                border: "1px solid #e0e0e0",
                backgroundColor: "#fafafa"
              }}
            />
            <Flex vertical gap={2} style={{ flex: 1, minWidth: 0 }}>
              <Flex justify="space-between" align="flex-start" gap={8}>
                <Text
                  style={{
                    fontSize: 13,
                    fontWeight: 500,
                    lineHeight: 1.3,
                    display: "-webkit-box",
                    WebkitLineClamp: 2,
                    WebkitBoxOrient: "vertical",
                    overflow: "hidden",
                  }}
                  title={item.product_name}
                >
                  {item.product_name}
                </Text>
                <Tag color="default" style={{ margin: 0, flexShrink: 0, fontSize: 11 }}>
                  x{item.qty}
                </Tag>
              </Flex>
              {item.variation_name && (
                <Text type="secondary" style={{ fontSize: 11 }}>
                  Var: {item.variation_name}
                </Text>
              )}
              {item.sku && (
                <Text type="secondary" style={{ fontSize: 10, fontFamily: "monospace" }}>
                  SKU: {item.sku}
                </Text>
              )}
            </Flex>
          </Flex>
        ))}
      </Flex>
    </Flex>
  );
}

// ============ MAIN COMPONENT ============

export function OrderTable({
  orders,
  loading,
  pagination,
  selectedRowKeys,
  onSelectionChange,
  onShip,
  onPrint,
  onViewDetail,
}: OrderTableProps) {
  // Group orders by order_no
  const groupedOrders = useMemo(() => {
    const orderMap = new Map<string, GroupedOrder>();

    orders.forEach((item) => {
      const key = item.order_no;

      if (!orderMap.has(key)) {
        orderMap.set(key, {
          key,
          ...item,
          items: [],
          status: item.status || item.order_status,
        });
      }

      const order = orderMap.get(key)!;
      order.items.push({
        sku: item.sku,
        product_name: item.product_name,
        variation_name: item.variation_name,
        qty: item.qty,
        price: item.price,
        product_image: item.product_image,
      });
    });

    return Array.from(orderMap.values());
  }, [orders]);

  // ============ TABLE COLUMNS ============
  const columns: ColumnsType<GroupedOrder> = [
    {
      title: "Product",
      key: "product",
      width: 320,
      render: (_, record) => <ProductCellWithBuyer order={record} />,
    },
    {
      title: "Amount Paid",
      key: "amount",
      width: 120,
      render: (_, record) => (
        <Flex vertical gap={2}>
          <Text strong style={{ color: "#0369a1", fontSize: 14 }}>
            {formatAmount(record.total_amount, record.currency)}
          </Text>
          <Text type="secondary" style={{ fontSize: 11 }}>
            {record.payment_method || "Online Payment"}
          </Text>
        </Flex>
      ),
    },
    {
      title: "Status",
      key: "status",
      width: 130,
      render: (_, record) => <StatusPipeline status={record.status} />,
    },
    {
      title: "Countdown",
      key: "countdown",
      width: 110,
      render: (_, record) => (
        <Flex vertical gap={2}>
          <Text style={{ color: getCountdownColor(record.ship_by_date), fontWeight: 500, fontSize: 13 }}>
            {getCountdown(record.ship_by_date)}
          </Text>
          {record.ship_by_date && (
            <Text type="secondary" style={{ fontSize: 10 }}>
              Ship by: {new Date(record.ship_by_date * 1000).toLocaleDateString()}
            </Text>
          )}
        </Flex>
      ),
    },
    {
      title: "Shipping",
      key: "shipping",
      width: 120,
      render: (_, record) => (
        <Text style={{ fontWeight: 500, fontSize: 12 }}>{record.shipping_carrier || "-"}</Text>
      ),
    },
    {
      title: "Action",
      key: "action",
      width: 90,
      fixed: "right",
      render: (_, record) => (
        <Flex vertical gap={6}>
          <Button
            type="primary"
            size="small"
            block
            disabled={record.status !== "READY_TO_SHIP"}
            onClick={() => onShip(record)}
            style={{ 
              backgroundColor: record.status === "READY_TO_SHIP" ? "#0369a1" : undefined,
              fontSize: 12
            }}
          >
            Ship
          </Button>
          <Flex justify="space-between" align="center">
            <Button
              type="link"
              size="small"
              style={{ padding: 0, fontSize: 11, height: "auto" }}
              onClick={() => onViewDetail?.(record)}
            >
              Details
            </Button>
            <Dropdown
              menu={{
                items: [
                  { key: "view", label: "View Details", onClick: () => onViewDetail?.(record) },
                  { key: "print", label: "Print Label", onClick: () => onPrint(record) },
                ],
              }}
              trigger={["click"]}
            >
              <Button type="text" size="small" icon={<MoreOutlined />} style={{ padding: 0 }} />
            </Dropdown>
          </Flex>
        </Flex>
      ),
    },
  ];

  return (
    <Table<GroupedOrder>
      columns={columns}
      dataSource={groupedOrders}
      loading={loading}
      rowKey="key"
      size="middle"
      pagination={{
        current: pagination.current,
        pageSize: pagination.pageSize,
        total: pagination.total,
        onChange: pagination.onChange,
        showSizeChanger: true,
        showTotal: (total) => `Total ${total} orders`,
        size: "small",
      }}
      rowSelection={{
        selectedRowKeys,
        onChange: onSelectionChange,
      }}
      scroll={{ x: 900 }}
      style={{ backgroundColor: "#fff" }}
    />
  );
}
