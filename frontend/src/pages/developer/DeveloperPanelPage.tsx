import { useSearchParams } from "react-router-dom";
import { Tabs, Typography } from "antd";
import OverviewTab from "./tabs/OverviewTab";
import SystemTab from "./tabs/SystemTab";
import TenantsTab from "./tabs/TenantsTab";
import UsersTab from "./tabs/UsersTab";
import SettingsTab from "./tabs/SettingsTab";

const { Title, Text } = Typography;

export default function DeveloperPanelPage() {
  const [searchParams, setSearchParams] = useSearchParams();
  const activeTab = searchParams.get("tab") || "overview";

  const handleTabChange = (key: string) => {
    setSearchParams({ tab: key });
  };

  const tabItems = [
    {
      key: "overview",
      label: "Overview",
      children: <OverviewTab />,
    },
    {
      key: "tenants",
      label: "Tenants",
      children: <TenantsTab />,
    },
    {
      key: "users",
      label: "Users",
      children: <UsersTab />,
    },
    {
      key: "system",
      label: "System",
      children: <SystemTab />,
    },
    {
      key: "settings",
      label: "Settings",
      children: <SettingsTab />,
    },
  ];

  return (
    <div style={{ padding: 24 }}>
      <div style={{ marginBottom: 16 }}>
        <Title level={3} style={{ margin: 0 }}>
          Developer Panel
        </Title>
        <Text type="secondary">
          Cross-tenant management — visible to developer role only
        </Text>
      </div>

      <Tabs
        activeKey={activeTab}
        onChange={handleTabChange}
        items={tabItems}
      />
    </div>
  );
}
