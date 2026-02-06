import { useSearchParams } from "react-router-dom";
import { Tabs, Typography, theme } from "antd";
import {
  SettingOutlined,
  ApiOutlined,
  CloudOutlined,
  UserOutlined,
  DeploymentUnitOutlined,
} from "@ant-design/icons";
import GeneralTab from "./tabs/GeneralTab";
import PlatformsTab from "./tabs/PlatformsTab";
import WebhooksTab from "./tabs/WebhooksTab";
import AccountTab from "./tabs/AccountTab";
import RouteManagementTab from "./tabs/RouteManagementTab";

const { Title, Text } = Typography;
const { useToken } = theme;

export default function SettingsPage() {
  const { token } = useToken();
  const [searchParams, setSearchParams] = useSearchParams();
  const activeTab = searchParams.get("tab") || "general";

  const handleTabChange = (key: string) => {
    setSearchParams({ tab: key });
  };

  const tabItems = [
    {
      key: "general",
      label: (
        <span>
          <SettingOutlined />
          General
        </span>
      ),
      children: <GeneralTab />,
    },
    {
      key: "platforms",
      label: (
        <span>
          <ApiOutlined />
          Platforms
        </span>
      ),
      children: <PlatformsTab />,
    },
    {
      key: "webhooks",
      label: (
        <span>
          <CloudOutlined />
          Webhooks
        </span>
      ),
      children: <WebhooksTab />,
    },
    {
      key: "routes",
      label: (
        <span>
          <DeploymentUnitOutlined />
          Routes
        </span>
      ),
      children: <RouteManagementTab />,
    },
    {
      key: "account",
      label: (
        <span>
          <UserOutlined />
          Account
        </span>
      ),
      children: <AccountTab />,
    },
  ];

  return (
    <div style={{ padding: 24 }}>
      <div style={{ marginBottom: 24 }}>
        <Title level={2} style={{ margin: 0, fontSize: 20 }}>
          Settings
        </Title>
        <Text type="secondary" style={{ fontSize: 12 }}>
          Manage application settings and platform integrations
        </Text>
      </div>

      <Tabs
        activeKey={activeTab}
        onChange={handleTabChange}
        items={tabItems}
        style={{
          background: token.colorBgContainer,
          borderRadius: token.borderRadius,
          padding: 16,
        }}
      />
    </div>
  );
}
