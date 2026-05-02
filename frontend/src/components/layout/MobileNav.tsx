import { useNavigate, useLocation } from "react-router-dom";
import { theme } from "antd";
import {
  DashboardOutlined,
  ShoppingOutlined,
  SkinOutlined,
  InboxOutlined,
  CodeOutlined,
} from "@ant-design/icons";

export default function MobileNav() {
  const navigate = useNavigate();
  const location = useLocation();
  const {
    token: { colorBgContainer, colorBorder, colorPrimary, colorTextSecondary },
  } = theme.useToken();

  const navItems = [
    { key: "/", icon: <DashboardOutlined />, label: "Home" },
    { key: "/order-manager", icon: <ShoppingOutlined />, label: "Orders" },
    { key: "/products", icon: <SkinOutlined />, label: "Products" },
    { key: "/inventory", icon: <InboxOutlined />, label: "Inventory" },
    {
      key: "/script-monitor?tab=current",
      icon: <CodeOutlined />,
      label: "Scripts",
    },
  ];

  return (
    <div
      className="mobile-nav"
      style={{
        position: "fixed",
        bottom: 0,
        left: 0,
        right: 0,
        height: 56,
        background: colorBgContainer,
        borderTop: `1px solid ${colorBorder}`,
        display: "flex",
        justifyContent: "space-around",
        alignItems: "center",
        zIndex: 1000,
        paddingBottom: "env(safe-area-inset-bottom)",
      }}
    >
      {navItems.map((item) => {
        const isScriptMonitor = item.key.startsWith("/script-monitor");
        const isProducts = item.key === "/products";
        const isActive = isScriptMonitor
          ? location.pathname === "/script-monitor"
          : isProducts
            ? location.pathname.startsWith("/products")
            : location.pathname === item.key;
        return (
          <button
            type="button"
            key={item.key}
            onClick={() => navigate(item.key)}
            style={{
              display: "flex",
              flexDirection: "column",
              alignItems: "center",
              justifyContent: "center",
              flex: 1,
              height: "100%",
              cursor: "pointer",
              color: isActive ? colorPrimary : colorTextSecondary,
              transition:
                "color var(--motion-mid) var(--ease-standard), transform var(--motion-mid) var(--ease-standard)",
              transform: isActive ? "scale(1.1)" : "scale(1)",
              border: "none",
              background: "transparent",
              padding: 0,
            }}
          >
            <div style={{ fontSize: 20, marginBottom: 2 }}>{item.icon}</div>
            <div style={{ fontSize: 10, fontWeight: isActive ? 500 : 400 }}>
              {item.label}
            </div>
          </button>
        );
      })}
      <style>{`
        @media (min-width: 768px) {
          .mobile-nav {
            display: none !important;
          }
        }
      `}</style>
    </div>
  );
}
