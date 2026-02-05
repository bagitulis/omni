import { Layout, Menu, theme } from "antd";
import { useLocation, useNavigate } from "react-router-dom";
import {
  DashboardOutlined,
  ShoppingOutlined,
  SkinOutlined,
  InboxOutlined,
  BarChartOutlined,
  SettingOutlined,
} from "@ant-design/icons";

const { Sider } = Layout;

interface SidebarProps {
  collapsed: boolean;
  onCollapse: (collapsed: boolean) => void;
}

export default function Sidebar({ collapsed, onCollapse }: SidebarProps) {
  const navigate = useNavigate();
  const location = useLocation();
  const {
    token: { colorBgContainer, colorPrimary, colorTextLightSolid },
  } = theme.useToken();

  const menuItems = [
    { key: "/", icon: <DashboardOutlined />, label: "Dashboard" },
    { key: "/order-manager", icon: <ShoppingOutlined />, label: "Orders" },
    { key: "/master-products", icon: <SkinOutlined />, label: "Products" },
    { key: "/inventory", icon: <InboxOutlined />, label: "Inventory" },
    { key: "/analytics", icon: <BarChartOutlined />, label: "Analytics" },
    { key: "/settings", icon: <SettingOutlined />, label: "Settings" },
  ];

  return (
    <Sider
      trigger={null}
      collapsible
      collapsed={collapsed}
      breakpoint="md"
      onBreakpoint={onCollapse}
      width={220}
      collapsedWidth={80}
      style={{
        background: colorBgContainer,
        height: "100vh",
        position: "fixed",
        top: 0,
        left: 0,
        borderRight: "1px solid #e2e8f0",
        zIndex: 100,
        overflow: "auto",
      }}
      // Class used for media query targeting in style tag below
      className="main-sidebar"
    >
      <div
        style={{
          height: 48,
          margin: 16,
          background: "rgba(3, 105, 161, 0.1)", // sky-700 with opacity
          borderRadius: 3,
          display: "flex",
          alignItems: "center",
          justifyContent: "center",
          color: colorPrimary,
          fontWeight: "bold",
          overflow: "hidden",
          whiteSpace: "nowrap",
        }}
      >
        {collapsed ? "O" : "OMNI"}
      </div>
      <Menu
        mode="inline"
        selectedKeys={[location.pathname]}
        onClick={({ key }) => navigate(key)}
        items={menuItems}
        style={{ borderRight: 0 }}
        theme="light"
      />
      <style>{`
        .ant-menu-item-selected {
          background-color: ${colorPrimary} !important;
          color: ${colorTextLightSolid} !important;
        }
        .ant-layout-sider-zero-width-trigger {
          display: none; 
        }
        @media (max-width: 768px) {
          .main-sidebar {
            display: none !important;
          }
        }
      `}</style>
    </Sider>
  );
}
