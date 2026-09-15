import React, { useMemo, useState } from "react";
import { Layout, Menu, theme } from "antd";
import { useTheme } from "@/contexts/ThemeContext.hooks";
import { useLocation, useNavigate } from "react-router-dom";
import {
  DashboardOutlined,
  ShoppingOutlined,
  SkinOutlined,
  InboxOutlined,
  BellOutlined,
  SettingOutlined,
  CodeOutlined,
  FileTextOutlined,
  TeamOutlined,
  ExperimentOutlined,
  ApiOutlined,
} from "@ant-design/icons";
import { usePermission } from "@/hooks/usePermission";

const { Sider } = Layout;

interface SidebarProps {
  collapsed: boolean;
  onCollapse: (collapsed: boolean) => void;
}

function Sidebar({ collapsed, onCollapse }: SidebarProps) {
  const navigate = useNavigate();
  const location = useLocation();
  const { isDark } = useTheme();
  const {
    token: {
      colorBgContainer,
      colorPrimary,
      colorTextLightSolid,
      colorBorder,
    },
  } = theme.useToken();

  const { canManageUsers, isDeveloper } = usePermission();

  const menuItems = useMemo(() => {
    const items = [
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
      { key: "/notifications", icon: <BellOutlined />, label: "Notifications" },
      {
        key: "/report",
        icon: <FileTextOutlined />,
        label: "Report",
        children: [
          { key: "/report/shopee", label: "Shopee Report" },
          { key: "/report/tiktok", label: "TikTok Report" },
        ],
      },
      {
        key: "/script-monitor",
        icon: <CodeOutlined />,
        label: "Script Monitor",
      },
      {
        key: "/extensions",
        icon: <ApiOutlined />,
        label: "Extensions",
        children: [
          { key: "/extensions", label: "Installed Extensions" },
          { key: "/extensions/shopee", label: "Shopee Scraper" },
          { key: "/extensions/results", label: "Results" },
        ],
      },
      { key: "/settings", icon: <SettingOutlined />, label: "Settings" },
    ];

    if (canManageUsers) {
      items.push({
        key: "/users",
        icon: <TeamOutlined />,
        label: "User Management",
      });
    }

    if (isDeveloper) {
      items.push({
        key: "/developer",
        icon: <ExperimentOutlined />,
        label: "Developer Panel",
      });
    }

    return items;
  }, [canManageUsers, isDeveloper]);

  const selectedKey = useMemo(() => {
    // Products sub-routes
    if (location.pathname.startsWith("/products")) {
      if (location.pathname === "/products/sync-history")
        return "/products/sync-history";
      if (location.pathname === "/products/add") return "/products/add";
      // All other /products paths (including /products/:id/edit) → highlight "All Products"
      return "/products";
    }
    // Report sub-routes
    if (location.pathname.startsWith("/report")) {
      return location.pathname;
    }
    // Script monitor: always highlight the single menu item regardless of ?tab=
    if (location.pathname.startsWith("/script-monitor")) {
      return "/script-monitor";
    }
    // Extensions sub-routes: each child is its own key, and the exact path is
    // the selected key so the submenu shows which page is open.
    if (location.pathname.startsWith("/extensions")) {
      return location.pathname;
    }
    return location.pathname;
  }, [location.pathname]);

  const [openKeys, setOpenKeys] = useState<string[]>(() => {
    const keys: string[] = [];
    if (location.pathname.startsWith("/products")) keys.push("/products");
    if (location.pathname.startsWith("/report")) {
      keys.push("/report");
    }
    if (location.pathname.startsWith("/extensions")) {
      keys.push("/extensions");
    }
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
          background: collapsed ? colorPrimary : (isDark ? "rgba(14, 165, 233, 0.15)" : "rgba(3, 105, 161, 0.1)"),
          borderRadius: 3, // Sharp corners
          display: "flex",
          alignItems: "center",
          justifyContent: "center",
          color: collapsed ? colorTextLightSolid : colorPrimary,
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
