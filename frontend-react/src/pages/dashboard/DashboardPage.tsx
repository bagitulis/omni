import { Row, Col, Typography, Card, Table, Tag, Button } from "antd";
import {
  ShoppingOutlined,
  AlertOutlined,
  CarOutlined,
  CheckCircleOutlined,
  RightOutlined,
} from "@ant-design/icons";
import { TaskCard } from "@/components/ui/TaskCard";
import { useDashboard } from "@/hooks/useDashboard";

const { Title } = Typography;

export function DashboardPage() {
  const { data, isLoading } = useDashboard();
  const metrics = data?.metrics;

  const getPlatformColor = (platform: string) => {
    switch (platform) {
      case "shopee":
        return "orange";
      case "lazada":
        return "blue";
      case "tiktok":
        return "black";
      case "tokopedia":
        return "green";
      default:
        return "default";
    }
  };

  const columns = [
    {
      title: "Order SN",
      dataIndex: "order_sn",
      key: "order_sn",
      render: (text: string) => (
        <a className="text-sky-700 font-medium">{text}</a>
      ),
    },
    {
      title: "Platform",
      dataIndex: "platform",
      key: "platform",
      render: (platform: string) => (
        <Tag color={getPlatformColor(platform)} className="capitalize">
          {platform}
        </Tag>
      ),
    },
    {
      title: "Status",
      dataIndex: "status",
      key: "status",
      render: (status: string) => <Tag bordered={false}>{status}</Tag>,
    },
    {
      title: "Amount",
      dataIndex: "amount",
      key: "amount",
      render: (amount: number) => `Rp ${amount.toLocaleString("id-ID")}`,
    },
  ];

  return (
    <div className="p-6 space-y-6">
      <div className="flex justify-between items-center">
        <Title level={2} style={{ margin: 0 }}>
          Dashboard
        </Title>
        <div className="text-gray-500 text-sm">Overview of your operations</div>
      </div>

      {/* Operational Task Counters */}
      <Row gutter={[16, 16]}>
        <Col xs={24} sm={12} md={6}>
          <TaskCard
            title="Orders Pending"
            value={metrics?.orders_pending ?? 0}
            loading={isLoading}
            icon={<ShoppingOutlined className="mr-2 text-sky-700" />}
            onClick={() => console.log("Go to pending orders")}
          />
        </Col>
        <Col xs={24} sm={12} md={6}>
          <TaskCard
            title="Stock Warning"
            value={metrics?.stock_warning ?? 0}
            loading={isLoading}
            valueStyle={{ color: "#d97706" }} // Warning color
            icon={<AlertOutlined className="mr-2 text-amber-600" />}
            onClick={() => console.log("Go to low stock")}
          />
        </Col>
        <Col xs={24} sm={12} md={6}>
          <TaskCard
            title="Ready to Ship"
            value={metrics?.ready_to_ship ?? 0}
            loading={isLoading}
            icon={<CarOutlined className="mr-2 text-blue-600" />}
            onClick={() => console.log("Go to ready to ship")}
          />
        </Col>
        <Col xs={24} sm={12} md={6}>
          <TaskCard
            title="Platform Status"
            value={metrics?.platform_status ?? 0}
            loading={isLoading}
            suffix="%"
            valueStyle={{ color: "#16a34a" }} // Success color
            icon={<CheckCircleOutlined className="mr-2 text-green-600" />}
          />
        </Col>
      </Row>

      {/* Recent Orders Section */}
      <Card
        title="Recent Orders"
        className="shadow-sm"
        extra={
          <Button type="link" href="/orders">
            View All <RightOutlined />
          </Button>
        }
      >
        <Table
          dataSource={data?.recent_orders}
          columns={columns}
          rowKey="order_sn"
          pagination={false}
          loading={isLoading}
          size="small"
        />
      </Card>
    </div>
  );
}
