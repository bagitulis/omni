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

  return (
    <Layout style={{ minHeight: "100vh" }}>
      <Sidebar collapsed={collapsed} onCollapse={setCollapsed} />
      <Layout
        className="site-layout"
        style={{
          marginLeft: 0, // Default for mobile (sidebar hidden)
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
            margin: "24px 24px 80px 24px", // Extra bottom margin for mobile nav
            minHeight: 280,
            overflow: "initial",
          }}
        >
          <Outlet />
        </Content>
        <MobileNav />
      </Layout>
      <style>{`
        @media (min-width: 768px) {
          .site-layout {
            margin-left: ${collapsed ? 80 : 220}px !important;
          }
          main.ant-layout-content {
            margin-bottom: 24px !important; /* Reset bottom margin on desktop */
          }
        }
      `}</style>
    </Layout>
  );
}
