import { useState } from "react";
import { Layout, theme } from "antd";
import { Outlet } from "react-router-dom";
import Sidebar from "./Sidebar";
import Header from "./Header";
import MobileNav from "./MobileNav";

const { Content } = Layout;

export default function AppLayout() {
  const [collapsed, setCollapsed] = useState(false);
  const {
    token: { colorBgLayout },
  } = theme.useToken();

  // Calculate sidebar width based on collapsed state
  const sidebarWidth = collapsed ? 80 : 220;

  return (
    <Layout style={{ minHeight: "100vh" }}>
      <Sidebar collapsed={collapsed} onCollapse={setCollapsed} />
      <Layout
        className="site-layout"
        style={{
          marginLeft: sidebarWidth, // Always apply margin for fixed sidebar
          background: colorBgLayout,
          transition: "margin-left 0.2s",
        }}
      >
        <Header
          collapsed={collapsed}
          onCollapse={() => setCollapsed(!collapsed)}
        />
        <Content
          style={{
            padding: 24,
            minHeight: 280,
            overflow: "initial",
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
