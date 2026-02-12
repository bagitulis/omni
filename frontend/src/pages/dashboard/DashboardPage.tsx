import { useState } from "react";
import {
  Row,
  Col,
  Typography,
  Card,
  Table,
  Tag,
  Button,
  Flex,
  Menu,
  theme,
} from "antd";
import {
  ShoppingOutlined,
  AlertOutlined,
  CarOutlined,
  CheckCircleOutlined,
  RightOutlined,
  DashboardOutlined,
  FileTextOutlined,
  SettingOutlined,
} from "@ant-design/icons";
import { useNavigate } from "react-router-dom";
import { TaskCard } from "@/components/ui/TaskCard";
import { useDashboard } from "@/hooks/useDashboard";
import { WalletWidget } from "./components/WalletWidget";
import { ShippingWidget } from "./components/ShippingWidget";
import { PlatformHealthWidget } from "./components/PlatformHealthWidget";
import { QuickActions } from "./components/QuickActions";
import { DashboardModals } from "./components/DashboardModals";
import { DashboardActionBar } from "./components/DashboardActionBar";

const { Title, Text } = Typography;

export function DashboardPage() {
  const { token } = theme.useToken();
  const { data, isLoading } = useDashboard();
  const navigate = useNavigate();
  const [activeTab, setActiveTab] = useState("overview");

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
        <span
          style={{
            color: token.colorPrimary,
            fontWeight: 500,
            cursor: "pointer",
          }}
        >
          {text}
        </span>
      ),
    },
    {
      title: "Platform",
      dataIndex: "platform",
      key: "platform",
      render: (platform: string) => (
        <Tag
          color={getPlatformColor(platform)}
          style={{ textTransform: "capitalize" }}
        >
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

  const renderOverview = () => (
    <Flex vertical gap={24}>
      <Flex justify="space-between" align="center">
        <Title level={2} style={{ margin: 0 }}>
          Dashboard
        </Title>
        <Text type="secondary" style={{ fontSize: 14 }}>
          Overview of your operations
        </Text>
      </Flex>

      {/* Operational Task Counters */}
      <Row gutter={[16, 16]}>
        <Col xs={24} sm={12} md={6}>
          <TaskCard
            title="Orders Pending"
            value={data?.orders_pending ?? 0}
            loading={isLoading}
            icon={
              <ShoppingOutlined
                style={{ marginRight: 8, color: token.colorPrimary }}
              />
            }
            onClick={() => navigate("/order-manager?type=unpaid")}
          />
        </Col>
        <Col xs={24} sm={12} md={6}>
          <TaskCard
            title="Total Orders"
            subtitle="Last 30 Days"
            value={data?.analytics?.total_orders ?? 0}
            loading={isLoading}
            icon={
              <AlertOutlined
                style={{ marginRight: 8, color: token.colorWarning }}
              />
            }
            onClick={() => navigate("/order-manager")}
          />
        </Col>
        <Col xs={24} sm={12} md={6}>
          <TaskCard
            title="Ready to Ship"
            value={data?.ready_to_ship ?? 0}
            loading={isLoading}
            icon={
              <CarOutlined style={{ marginRight: 8, color: token.colorInfo }} />
            }
            onClick={() => navigate("/order-manager?type=unprocess")}
          />
        </Col>
        <Col xs={24} sm={12} md={6}>
          <TaskCard
            title="Total Sales"
            subtitle="Last 30 Days"
            value={
              data?.analytics?.total_sales
                ? `Rp ${Math.floor(data.analytics.total_sales / 1_000_000)}M`
                : 0
            }
            loading={isLoading}
            valueStyle={{ color: token.colorSuccess }}
            icon={
              <CheckCircleOutlined
                style={{ marginRight: 8, color: token.colorSuccess }}
              />
            }
          />
        </Col>
      </Row>

      {/* Widgets Row */}
      <Row gutter={[16, 16]}>
        <Col xs={24} md={16}>
          <WalletWidget />
        </Col>
        <Col xs={24} md={8}>
          <QuickActions />
        </Col>
      </Row>

      <Row gutter={[16, 16]}>
        <Col xs={24} md={12}>
          <ShippingWidget />
        </Col>
        <Col xs={24} md={12}>
          <PlatformHealthWidget />
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
  );

  const handleMenuClick = ({ key }: { key: string }) => {
    // Navigation map for features that have dedicated pages
    const navigationMap: Record<string, string> = {
      "product-management": "/product-manager",
      "order-management": "/order-manager",
      settings: "/settings",
    };

    if (navigationMap[key]) {
      navigate(navigationMap[key]);
    } else {
      setActiveTab(key);
    }
  };

  const renderContent = () => {
    switch (activeTab) {
      case "overview":
        return renderOverview();
      default:
        return null;
    }
  };

  const menuItems = [
    { key: "overview", icon: <DashboardOutlined />, label: "Overview" },
    {
      key: "product-management",
      icon: <ShoppingOutlined />,
      label: "Product Management",
    },
    {
      key: "order-management",
      icon: <FileTextOutlined />,
      label: "Order Management",
    },
    { key: "settings", icon: <SettingOutlined />, label: "Settings" },
  ];

  return (
    <div
      style={{
        height: "calc(100vh - 96px)",
        display: "flex",
        flexDirection: "column",
      }}
    >
      {/* Top Action Bar */}
      <div
        style={{
          padding: "12px 24px",
          background: token.colorBgContainer,
          borderBottom: `1px solid ${token.colorBorder}`,
        }}
      >
        <DashboardActionBar />
      </div>

      <div style={{ display: "flex", flex: 1, overflow: "hidden" }}>
        {/* Left Sidebar */}
        <div
          style={{
            width: 200,
            background: token.colorBgContainer,
            borderRight: `1px solid ${token.colorBorder}`,
          }}
        >
          <Menu
            mode="inline"
            selectedKeys={[activeTab]}
            items={menuItems}
            onClick={handleMenuClick}
            style={{ borderRight: 0, height: "100%" }}
          />
        </div>

        {/* Right Content Area */}
        <div
          style={{
            flex: 1,
            overflowY: "auto",
            padding: 24,
            background: token.colorBgLayout,
          }}
        >
          {renderContent()}
        </div>
      </div>

      <DashboardModals />
    </div>
  );
}

export default DashboardPage;
