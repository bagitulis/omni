import { Row, Col, Typography, Card, Table, Tag, Button, Flex } from "antd";
import {
  ShoppingOutlined,
  AlertOutlined,
  CarOutlined,
  CheckCircleOutlined,
  RightOutlined,
} from "@ant-design/icons";
import { useNavigate } from "react-router-dom";
import { TaskCard } from "@/components/ui/TaskCard";
import { useDashboard } from "@/hooks/useDashboard";

const { Title, Text } = Typography;

export function DashboardPage() {
  const { data, isLoading } = useDashboard();
  const navigate = useNavigate();

  const getPlatformColor = (platform: string) => {
    switch (platform) {
      case "shopee":
        return "orange";
      case "lazada":
        return "blue";
      case "tiktok":
        return "black";
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
        <span style={{ color: "#0369a1", fontWeight: 500, cursor: "pointer" }}>{text}</span>
      ),
    },
    {
      title: "Platform",
      dataIndex: "platform",
      key: "platform",
      render: (platform: string) => (
        <Tag color={getPlatformColor(platform)} style={{ textTransform: "capitalize" }}>
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
       dataIndex: "total_amount",
       key: "total_amount",
       render: (amount: number) => `Rp ${(amount || 0).toLocaleString("id-ID")}`,
     },
  ];

  return (
    <div style={{ padding: 24 }}>
      <Flex vertical gap={24}>
        <Flex justify="space-between" align="center">
          <Title level={2} style={{ margin: 0 }}>
            Dashboard
          </Title>
          <Text type="secondary" style={{ fontSize: 14 }}>Overview of your operations</Text>
        </Flex>

        {/* Operational Task Counters */}
        <Row gutter={[16, 16]}>
           <Col xs={24} sm={12} md={6}>
             <TaskCard
               title="Orders Pending"
               value={data?.orders_pending ?? 0}
               loading={isLoading}
               icon={<ShoppingOutlined style={{ marginRight: 8, color: "#0369a1" }} />}
               onClick={() => navigate("/order-manager?type=unpaid")}
             />
           </Col>
           <Col xs={24} sm={12} md={6}>
             <TaskCard
               title="Total Orders"
               value={data?.analytics?.total_orders ?? 0}
               loading={isLoading}
               icon={<AlertOutlined style={{ marginRight: 8, color: "#d97706" }} />}
               onClick={() => navigate("/order-manager")}
             />
           </Col>
           <Col xs={24} sm={12} md={6}>
             <TaskCard
               title="Ready to Ship"
               value={data?.ready_to_ship ?? 0}
               loading={isLoading}
               icon={<CarOutlined style={{ marginRight: 8, color: "#2563eb" }} />}
               onClick={() => navigate("/order-manager?type=unprocess")}
             />
           </Col>
           <Col xs={24} sm={12} md={6}>
             <TaskCard
               title="Total Sales"
               value={
                 data?.analytics?.total_sales
                   ? `Rp ${Math.floor(data.analytics.total_sales / 1_000_000)}M`
                   : 0
               }
               loading={isLoading}
               valueStyle={{ color: "#16a34a" }}
               icon={<CheckCircleOutlined style={{ marginRight: 8, color: "#16a34a" }} />}
             />
           </Col>
        </Row>

        {/* Recent Orders Section */}
        <Card
          title="Recent Orders"
          extra={
            <Button type="link" href="/order-manager">
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
      </Flex>
    </div>
  );
}
