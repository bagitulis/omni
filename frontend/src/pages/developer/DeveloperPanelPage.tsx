import { useSearchParams } from "react-router-dom";
import { Tabs, Typography } from "antd";
import OverviewTab from "./tabs/OverviewTab";

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
      children: <div style={{ padding: 24 }}>Coming soon</div>,
    },
    {
      key: "users",
      label: "Users",
      children: <div style={{ padding: 24 }}>Coming soon</div>,
    },
    {
      key: "system",
      label: "System",
      children: <div style={{ padding: 24 }}>Coming soon</div>,
    },
    {
      key: "settings",
      label: "Settings",
      children: <div style={{ padding: 24 }}>Coming soon</div>,
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
