import {
  Table,
  Button,
  Space,
  Typography,
  Tooltip,
  Dropdown,
  MenuProps,
} from "antd";
import {
  PrinterOutlined,
  SendOutlined,
  MoreOutlined,
  CopyOutlined,
  EyeOutlined,
} from "@ant-design/icons";
import { Order } from "@/types/order";
import { PlatformBadge } from "../ui/PlatformBadge";
import { StatusPipeline } from "../ui/StatusPipeline";
import { TableRowSelection } from "antd/es/table/interface";

const { Text } = Typography;

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
  onShip: (order: Order) => void;
  onPrint: (order: Order) => void;
}

export function OrderTable({
  orders,
  loading,
  pagination,
  selectedRowKeys,
  onSelectionChange,
  onShip,
  onPrint,
}: OrderTableProps) {
  const rowSelection: TableRowSelection<Order> = {
    selectedRowKeys,
    onChange: onSelectionChange,
  };

  const columns = [
    {
      title: "Order No.",
      dataIndex: "order_sn",
      key: "order_sn",
      width: 180,
      render: (text: string) => (
        <Space>
          <Text strong style={{ color: "#0369a1" }}>
            {text}
          </Text>
          <Tooltip title="Copy Order ID">
            <Button
              type="text"
              size="small"
              icon={<CopyOutlined className="text-xs" />}
              onClick={(e) => {
                e.stopPropagation();
                navigator.clipboard.writeText(text);
              }}
            />
          </Tooltip>
        </Space>
      ),
    },
    {
      title: "Platform",
      dataIndex: "platform",
      key: "platform",
      width: 100,
      render: (platform: string) => <PlatformBadge platform={platform} />,
    },
    {
      title: "Customer",
      dataIndex: "buyer_username",
      key: "buyer_username",
      render: (text: string) => <Text>{text}</Text>,
    },
    {
      title: "Total",
      dataIndex: "total_amount",
      key: "total_amount",
      width: 120,
      render: (amount: number) => (
        <Text strong>
          {new Intl.NumberFormat("id-ID", {
            style: "currency",
            currency: "IDR",
            minimumFractionDigits: 0,
          }).format(amount)}
        </Text>
      ),
    },
    {
      title: "Status",
      dataIndex: "order_status",
      key: "status",
      width: 150,
      render: (status: string) => <StatusPipeline status={status} />,
    },
    {
      title: "Actions",
      key: "actions",
      width: 120,
      fixed: "right" as const,
      render: (_: any, record: Order) => {
        const menuItems: MenuProps["items"] = [
          {
            key: "view",
            label: "View Details",
            icon: <EyeOutlined />,
          },
          {
            key: "print",
            label: "Print Label",
            icon: <PrinterOutlined />,
            onClick: () => onPrint(record),
          },
          {
            key: "ship",
            label: "Ship Order",
            icon: <SendOutlined />,
            onClick: () => onShip(record),
          },
        ];

        return (
          <Space size="small">
            <Tooltip title="Ship Order">
              <Button
                size="small"
                type="primary"
                icon={<SendOutlined />}
                onClick={() => onShip(record)}
                disabled={record.order_status !== "READY_TO_SHIP"}
              />
            </Tooltip>
            <Dropdown menu={{ items: menuItems }} trigger={["click"]}>
              <Button size="small" icon={<MoreOutlined />} />
            </Dropdown>
          </Space>
        );
      },
    },
  ];

  return (
    <Table
      rowKey="order_sn"
      columns={columns}
      dataSource={orders}
      loading={loading}
      rowSelection={rowSelection}
      pagination={{
        ...pagination,
        showSizeChanger: true,
        showTotal: (total) => `Total ${total} orders`,
        size: "small",
      }}
      scroll={{ x: 1000 }}
      size="small" // Data-dense
      className="border border-slate-200 rounded-sm"
    />
  );
}
