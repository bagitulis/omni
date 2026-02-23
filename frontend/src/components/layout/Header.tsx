import React, { useMemo, useCallback, useEffect, useState } from "react";
import {
  Layout,
  Button,
  Avatar,
  Dropdown,
  Select,
  Space,
  theme,
  Typography,
  message,
} from "antd";
import {
  MenuFoldOutlined,
  MenuUnfoldOutlined,
  UserOutlined,
  LogoutOutlined,
  SettingOutlined,
  SunOutlined,
  MoonOutlined,
  SwapOutlined,
} from "@ant-design/icons";
import { useNavigate } from "react-router-dom";
import { useAuthStore } from "@/stores/authStore";
import { TokenStatusDropdown } from "./TokenStatusDropdown";
import { useTheme } from "@/contexts/ThemeContext.hooks";
import apiClient from "@/api/client";

const { Header: AntHeader } = Layout;
const { Text } = Typography;

interface HeaderProps {
  collapsed: boolean;
  onCollapse: () => void;
}

interface TenantOption {
  id: string;
  shop_name: string;
}

function Header({ collapsed, onCollapse }: HeaderProps) {
  const navigate = useNavigate();
  const { user, logout, tenantId, setAuth } = useAuthStore();
  const {
    token: { colorBgContainer, colorBorderSecondary },
  } = theme.useToken();
  const { isDark, toggle } = useTheme();

  // Tenant switcher state (developer only)
  const [tenants, setTenants] = useState<TenantOption[]>([]);
  const [switching, setSwitching] = useState(false);
  const isDeveloper = user?.role === "developer";

  useEffect(() => {
    if (!isDeveloper) return;
    apiClient
      .get<{ tenants: TenantOption[] }>("/auth/tenants")
      .then((res) => {
        if (res.success && res.data) {
          const list =
            (res.data as unknown as { tenants: TenantOption[] }).tenants || [];
          setTenants(list);
        }
      })
      .catch(() => {
        /* ignore */
      });
  }, [isDeveloper]);

  const handleSwitchTenant = useCallback(
    async (newTenantId: string) => {
      if (!newTenantId || newTenantId === tenantId) return;
      setSwitching(true);
      try {
        const res = await apiClient.post<{
          token: string;
          tenant_id: string;
        }>("/auth/switch-tenant", { tenant_id: newTenantId });
        if (res.success && res.data) {
          const data = res.data as unknown as {
            token: string;
            tenant_id: string;
          };
          if (data.token && user) {
            setAuth({
              access_token: data.token,
              user,
              tenant_id: data.tenant_id,
            });
          }
          message.success(`Switched to ${newTenantId}`);
          window.location.reload();
        }
      } catch {
        message.error("Failed to switch tenant");
      } finally {
        setSwitching(false);
      }
    },
    [tenantId, user, setAuth],
  );

  const handleLogout = useCallback(async () => {
    await logout();
    navigate("/login");
  }, [logout, navigate]);

  const userMenu = useMemo(
    () => [
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
    ],
    [navigate, handleLogout],
  );

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
          aria-label={collapsed ? "Expand sidebar" : "Collapse sidebar"}
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
        {/* Tenant Switcher (Developer Only) */}
        {isDeveloper && tenants.length > 0 && (
          <Select
            value={tenantId || undefined}
            onChange={handleSwitchTenant}
            loading={switching}
            size="small"
            style={{ minWidth: 160 }}
            suffixIcon={<SwapOutlined />}
            options={tenants.map((t) => ({
              value: t.id,
              label: t.shop_name || t.id,
            }))}
            className="tenant-switcher"
          />
        )}

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
          .tenant-switcher {
            display: none !important;
          }
        }
      `}</style>
    </AntHeader>
  );
}

export default React.memo(Header);
