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
  Grid,
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
  const screens = Grid.useBreakpoint();
  const isMobile = !screens.md;

  const getPlatformColor = (platform: string) => {
    switch (platform) {
      case "shopee":
        return "#ee4d2d";
      case "lazada":
        return "#0f136d";
      case "tiktok":
        return "#000000";
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
            fontSize: 12,
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
          style={{ textTransform: "capitalize", fontSize: 10, borderRadius: 2 }}
        >
          {platform}
        </Tag>
      ),
    },
    {
      title: "Status",
      dataIndex: "status",
      key: "status",
      render: (status: string) => (
        <Tag color="processing" bordered={false} style={{ fontSize: 10, borderRadius: 2 }}>
          {status}
        </Tag>
      ),
    },
    {
      title: "Amount",
      dataIndex: "total_amount",
      key: "total_amount",
      render: (amount: number) => (
        <Text strong style={{ fontSize: 12 }}>
          Rp {(amount || 0).toLocaleString("id-ID")}
        </Text>
      ),
    },
  ];

  const renderOverview = () => (
    <Flex vertical gap={24}>
      <Flex
        justify={isMobile ? "start" : "space-between"}
        align={isMobile ? "start" : "center"}
        vertical={isMobile}
        gap={isMobile ? 4 : 0}
      >
        <div>
          <Title level={isMobile ? 3 : 2} style={{ margin: 0, fontSize: isMobile ? 20 : 24, fontWeight: 600 }}>
            Dashboard
          </Title>
          <Text type="secondary" style={{ fontSize: 12, marginTop: 4, display: "block" }}>
            Overview of your operations
          </Text>
        </div>
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

      {/* Ready to Ship Orders Section */}
      <Card
        title={<Text strong style={{ fontSize: 14 }}>Ready to Ship</Text>}
        extra={
          <Button type="link" onClick={() => navigate("/order-manager")} style={{ fontSize: 12 }}>
            View All <RightOutlined style={{ fontSize: 10 }} />
          </Button>
        }
        style={{ borderRadius: 3, border: "1px solid #f0f0f0" }}
        bodyStyle={{ padding: 0 }}
      >
        <Table
          dataSource={data?.recent_orders}
          columns={columns}
          rowKey="order_sn"
          pagination={false}
          loading={isLoading}
          size="middle"
          rowClassName={() => 'dashboard-table-row'}
        />
        <style>{`
          .dashboard-table-row td {
            font-size: 12px;
          }
          .ant-table-thead > tr > th {
            font-size: 12px;
            font-weight: 500;
            background: #f8fafc;
          }
        `}</style>
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
    { key: "overview", icon: <DashboardOutlined style={{ fontSize: 14 }}/>, label: "Overview", style: { fontSize: 13, fontWeight: 500 } },
    {
      key: "product-management",
      icon: <ShoppingOutlined style={{ fontSize: 14 }}/>,
      label: "Product Management",
      style: { fontSize: 13, fontWeight: 500 }
    },
    {
      key: "order-management",
      icon: <FileTextOutlined style={{ fontSize: 14 }}/>,
      label: "Order Management",
      style: { fontSize: 13, fontWeight: 500 }
    },
    { key: "settings", icon: <SettingOutlined style={{ fontSize: 14 }}/>, label: "Settings", style: { fontSize: 13, fontWeight: 500 } },
  ];

  return (
    <div
      style={{
        height: "calc(100vh - 96px)",
        display: "flex",
        flexDirection: "column",
        fontFamily: `system-ui, -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, sans-serif`
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
        {!isMobile && (
          <div
            style={{
              width: 220,
              background: token.colorBgContainer,
              borderRight: `1px solid ${token.colorBorder}`,
              paddingTop: 8,
            }}
          >
            <Menu
              mode="inline"
              selectedKeys={[activeTab]}
              items={menuItems}
              onClick={handleMenuClick}
              style={{ borderRight: 0, height: "100%", background: "transparent" }}
            />
          </div>
        )}

        {/* Right Content Area */}
        <div
          style={{
            flex: 1,
            overflowY: "auto",
            padding: isMobile ? 16 : 24,
            background: "#f8fafc", // A more modern, light slate background 
          }}
        >
          {/* Main content max width for ultra-wide screens to maintain readibility optionally */}
          <div style={{ maxWidth: 1440, margin: "0 auto" }}>
            {renderContent()}
          </div>
        </div>
      </div>

      <DashboardModals />
    </div>
  );
}

export default DashboardPage;
