import { useEffect, useState } from "react";
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
} from "@ant-design/icons";
import { useNavigate } from "react-router-dom";
import { STORAGE_KEYS } from "@/lib/constants";
import { TokenStatusDropdown } from "./TokenStatusDropdown";

const { Header: AntHeader } = Layout;
const { Text } = Typography;

interface HeaderProps {
  collapsed: boolean;
  onCollapse: () => void;
}

export default function Header({ collapsed, onCollapse }: HeaderProps) {
  const navigate = useNavigate();
  const [userName, setUserName] = useState<string>("");
  const {
    token: { colorBgContainer, colorBorderSecondary },
  } = theme.useToken();

  useEffect(() => {
    const storedName = localStorage.getItem(STORAGE_KEYS.USER_NAME);
    if (storedName) {
      setUserName(storedName);
    }
  }, []);

  const handleLogout = () => {
    localStorage.removeItem(STORAGE_KEYS.AUTH_TOKEN);
    localStorage.removeItem(STORAGE_KEYS.AUTH_USER);
    localStorage.removeItem(STORAGE_KEYS.TENANT_ID);
    localStorage.removeItem(STORAGE_KEYS.USER_ROLE);
    localStorage.removeItem(STORAGE_KEYS.USER_NAME);
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
          }}
          className="desktop-trigger"
        />
        <div className="mobile-logo" style={{ fontWeight: 600, fontSize: 16 }}>
          OMNI
        </div>
      </div>

      <div style={{ display: "flex", alignItems: "center", gap: 8 }}>
        {/* Token Status Dropdown */}
        <TokenStatusDropdown />

        {/* User Menu */}
        <Dropdown menu={{ items: userMenu }} placement="bottomRight">
          <Space style={{ cursor: "pointer" }}>
            <Avatar size="small" icon={<UserOutlined />} />
            <Text className="username-label">{userName || "User"}</Text>
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
