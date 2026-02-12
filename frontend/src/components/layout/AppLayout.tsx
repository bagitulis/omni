import { useState, useCallback } from "react";
import { Layout, theme } from "antd";
import { Outlet } from "react-router-dom";
import Sidebar from "./Sidebar";
import Header from "./Header";
import MobileNav from "./MobileNav";

const { Content } = Layout;

export default function AppLayout() {
  const [collapsed, setCollapsed] = useState(false);
  const {
    token: { colorBgLayout, colorPrimary, colorBgContainer },
  } = theme.useToken();

  // Calculate sidebar width based on collapsed state
  const sidebarWidth = collapsed ? 80 : 220;

  const handleToggle = useCallback(() => {
    setCollapsed((prev) => !prev);
  }, []);

  return (
    <Layout style={{ minHeight: "100vh" }}>
      {/* Skip to content link for keyboard navigation accessibility */}
      <a
        href="#main-content"
        className="skip-to-content"
        style={{
          position: "absolute",
          left: "-9999px",
          top: "auto",
          width: "1px",
          height: "1px",
          overflow: "hidden",
          zIndex: 9999,
        }}
        onFocus={(e) => {
          e.currentTarget.style.position = "fixed";
          e.currentTarget.style.top = "8px";
          e.currentTarget.style.left = "8px";
          e.currentTarget.style.width = "auto";
          e.currentTarget.style.height = "auto";
          e.currentTarget.style.overflow = "visible";
          e.currentTarget.style.padding = "8px 16px";
          e.currentTarget.style.background = colorPrimary;
          e.currentTarget.style.color = colorBgContainer;
          e.currentTarget.style.borderRadius = "3px";
          e.currentTarget.style.textDecoration = "none";
          e.currentTarget.style.fontWeight = "600";
        }}
        onBlur={(e) => {
          e.currentTarget.style.position = "absolute";
          e.currentTarget.style.left = "-9999px";
          e.currentTarget.style.width = "1px";
          e.currentTarget.style.height = "1px";
          e.currentTarget.style.overflow = "hidden";
        }}
      >
        Skip to main content
      </a>
      <Sidebar collapsed={collapsed} onCollapse={setCollapsed} />
      <Layout
        className="site-layout"
        style={{
          marginLeft: sidebarWidth, // Always apply margin for fixed sidebar
          background: colorBgLayout,
          transition: "margin-left var(--motion-mid)",
        }}
      >
        <Header collapsed={collapsed} onCollapse={handleToggle} />
        <Content
          id="main-content"
          style={{
            padding: 24,
            minHeight: 0,
            overflow: "auto",
            flex: 1,
          }}
        >
          <Outlet />
        </Content>
        <MobileNav />
      </Layout>
      <style>{`
        @media (max-width: 768px) {
          .site-layout {
            margin-left: 0 !important;
          }
          main.ant-layout-content {
            padding: 16px !important;
            margin-bottom: 56px !important; /* Space for mobile nav */
          }
        }
      `}</style>
    </Layout>
  );
}
