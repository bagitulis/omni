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
      colorFillSecondary,
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
        children: [
          { key: "/script-monitor?tab=current", label: "Current Job" },
          { key: "/script-monitor?tab=queue", label: "Queue" },
          { key: "/script-monitor?tab=history", label: "History" },
          {
            key: "/script-monitor?tab=auto-functions",
            label: "Auto-Functions",
          },
        ],
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
    // Script monitor special handling (existing)
    if (location.pathname === "/script-monitor") {
      return location.search
        ? `${location.pathname}${location.search}`
        : "/script-monitor";
    }
    return location.pathname;
  }, [location.pathname, location.search]);

  const [openKeys, setOpenKeys] = useState<string[]>(() => {
    const keys: string[] = [];
    if (location.pathname.startsWith("/products")) keys.push("/products");
    if (location.pathname.startsWith("/script-monitor"))
      keys.push("/script-monitor");
    return keys;
  });

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
        borderRight: `1px solid ${colorBorder}`,
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
          background: colorFillSecondary,
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
