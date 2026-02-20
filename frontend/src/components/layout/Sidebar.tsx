import React, { useMemo, useState } from "react";
import { Layout, Menu, theme } from "antd";
import { useLocation, useNavigate } from "react-router-dom";
import {
  DashboardOutlined,
  ShoppingOutlined,
  SkinOutlined,
  InboxOutlined,
  BarChartOutlined,
  SettingOutlined,
  NodeIndexOutlined,
  CodeOutlined,
} from "@ant-design/icons";

const { Sider } = Layout;

interface SidebarProps {
  collapsed: boolean;
  onCollapse: (collapsed: boolean) => void;
}

function Sidebar({ collapsed, onCollapse }: SidebarProps) {
  const navigate = useNavigate();
  const location = useLocation();
  const {
    token: {
      colorBgContainer,
      colorPrimary,
      colorTextLightSolid,
      colorBorder,
    },
  } = theme.useToken();

  const menuItems = useMemo(
    () => [
      { key: "/", icon: <DashboardOutlined />, label: "Dashboard" },
      { key: "/order-manager", icon: <ShoppingOutlined />, label: "Orders" },
      {
        key: "/products",
        icon: <SkinOutlined />,
        label: "Products",
        children: [
          { key: "/products", label: "All Products" },
          { key: "/products/add", label: "Add Product" },
          { key: "/products/sync-history", label: "Sync History" },
        ],
      },
      { key: "/inventory", icon: <InboxOutlined />, label: "Inventory" },
      {
        key: "/route-mapping",
        icon: <NodeIndexOutlined />,
        label: "Route Mapping",
      },
      { key: "/analytics", icon: <BarChartOutlined />, label: "Analytics" },
      {
        key: "/script-monitor",
        icon: <CodeOutlined />,
        label: "Script Monitor",
      },
      { key: "/settings", icon: <SettingOutlined />, label: "Settings" },
    ],
    [],
  );

  const selectedKey = useMemo(() => {
    // Products sub-routes
    if (location.pathname.startsWith("/products")) {
      if (location.pathname === "/products/sync-history")
        return "/products/sync-history";
      if (location.pathname === "/products/add") return "/products/add";
      // All other /products paths (including /products/:id/edit) → highlight "All Products"
      return "/products";
    }
    // Script monitor: always highlight the single menu item regardless of ?tab=
    if (location.pathname.startsWith("/script-monitor")) {
      return "/script-monitor";
    }
    return location.pathname;
  }, [location.pathname]);

  const [openKeys, setOpenKeys] = useState<string[]>(() => {
    const keys: string[] = [];
    if (location.pathname.startsWith("/products")) keys.push("/products");
    return keys;
  });

  return (
    <Sider
      trigger={null}
      collapsible
      collapsed={collapsed}
      breakpoint="md"
      onBreakpoint={onCollapse}
      width={220} // Standard width
      collapsedWidth={80} // Standard collapsed width
      style={{
        background: colorBgContainer,
        height: "100vh",
        position: "fixed",
        top: 0,
        left: 0,
        borderRight: `1px solid ${colorBorder}`,
        zIndex: 100,
        overflowY: "auto",
        overflowX: "hidden",
        boxShadow: "2px 0 8px rgba(0, 0, 0, 0.05)", // Subtle depth
      }}
      // Class used for media query targeting in style tag below
      className="main-sidebar"
    >
      <div
        style={{
          height: 48, // Standard header height
          margin: 16,
          background: collapsed ? colorPrimary : "rgba(3, 105, 161, 0.1)", // Light brand bg
          borderRadius: 3, // Sharp corners
          display: "flex",
          alignItems: "center",
          justifyContent: "center",
          color: collapsed ? "#fff" : colorPrimary,
          fontWeight: 800,
          fontSize: collapsed ? 20 : 24, // Impactful brand text
          letterSpacing: collapsed ? 0 : -0.5,
          overflow: "hidden",
          whiteSpace: "nowrap",
          transition: "all 0.3s cubic-bezier(0.2, 0, 0, 1)",
          boxShadow: collapsed ? "0 2px 4px rgba(0,0,0,0.1)" : "none",
        }}
      >
        {collapsed ? "O" : "OMNI"}
      </div>
      <Menu
        mode="inline"
        selectedKeys={[selectedKey]}
        openKeys={openKeys}
        onOpenChange={(keys) => setOpenKeys(keys as string[])}
        onClick={({ key }) => navigate(key)}
        items={menuItems}
        style={{ borderRight: 0 }}
        theme="light"
      />
      <style>{`
        .ant-menu-item {
          transition: background-color var(--motion-mid) var(--ease-standard),
                      color var(--motion-mid) var(--ease-standard) !important;
        }
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

export default React.memo(Sidebar);
