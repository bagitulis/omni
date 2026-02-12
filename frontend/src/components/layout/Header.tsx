import {
  Layout,
  Button,
  Avatar,
  Dropdown,
  Space,
  theme,
  Typography,
} from "antd";
import {
  MenuFoldOutlined,
  MenuUnfoldOutlined,
  UserOutlined,
  LogoutOutlined,
  SettingOutlined,
  SunOutlined,
  MoonOutlined,
} from "@ant-design/icons";
import { useNavigate } from "react-router-dom";
import { useAuthStore } from "@/stores/authStore";
import { TokenStatusDropdown } from "./TokenStatusDropdown";
import { useTheme } from "@/contexts/ThemeContext";

const { Header: AntHeader } = Layout;
const { Text } = Typography;

interface HeaderProps {
  collapsed: boolean;
  onCollapse: () => void;
}

export default function Header({ collapsed, onCollapse }: HeaderProps) {
  const navigate = useNavigate();
  const { user, logout } = useAuthStore();
  const {
    token: { colorBgContainer, colorBorderSecondary },
  } = theme.useToken();
  const { isDark, toggle } = useTheme();

  const handleLogout = async () => {
    await logout();
    navigate("/login");
  };

  const userMenu = [
    {
      key: "profile",
      label: "Profile",
      icon: <UserOutlined />,
    },
    {
      key: "settings",
      label: "Settings",
      icon: <SettingOutlined />,
      onClick: () => navigate("/settings"),
    },
    {
      type: "divider" as const,
    },
    {
      key: "logout",
      label: "Logout",
      icon: <LogoutOutlined />,
      danger: true,
      onClick: handleLogout,
    },
  ];

  return (
    <AntHeader
      style={{
        padding: "0 16px",
        background: colorBgContainer,
        borderBottom: `1px solid ${colorBorderSecondary}`,
        height: 48,
        display: "flex",
        alignItems: "center",
        justifyContent: "space-between",
        position: "sticky",
        top: 0,
        zIndex: 9,
        lineHeight: "48px",
      }}
    >
      <div style={{ display: "flex", alignItems: "center" }}>
        <Button
          type="text"
          icon={collapsed ? <MenuUnfoldOutlined /> : <MenuFoldOutlined />}
          onClick={onCollapse}
          style={{
            fontSize: "16px",
            width: 48,
            height: 48,
            display: "none", // Hidden by default, shown via CSS for desktop
            transition:
              "color var(--motion-mid) var(--ease-standard), background-color var(--motion-mid) var(--ease-standard)",
          }}
          className="desktop-trigger"
        />
        <div className="mobile-logo" style={{ fontWeight: 600, fontSize: 16 }}>
          OMNI
        </div>
      </div>

      <div style={{ display: "flex", alignItems: "center", gap: 8 }}>
        <Button
          type="text"
          icon={isDark ? <SunOutlined /> : <MoonOutlined />}
          onClick={toggle}
          aria-label={isDark ? "Switch to light mode" : "Switch to dark mode"}
          style={{ fontSize: 16 }}
        />

        {/* Token Status Dropdown */}
        <TokenStatusDropdown />

        {/* User Menu */}
        <Dropdown menu={{ items: userMenu }} placement="bottomRight">
          <Space style={{ cursor: "pointer" }}>
            <Avatar size="small" icon={<UserOutlined />} />
            <Text className="username-label">{user?.username || "User"}</Text>
          </Space>
        </Dropdown>
      </div>

      <style>{`
        @media (min-width: 768px) {
          .desktop-trigger {
            display: inline-flex !important;
          }
          .mobile-logo {
            display: none !important;
          }
          .username-label {
            display: inline-block !important;
          }
        }
        @media (max-width: 767px) {
          .desktop-trigger {
            display: none !important;
          }
          .mobile-logo {
            display: block !important;
          }
          .username-label {
            display: none !important;
          }
        }
      `}</style>
    </AntHeader>
  );
}
