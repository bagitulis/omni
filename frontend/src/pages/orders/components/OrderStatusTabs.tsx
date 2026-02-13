import { Tabs, Card, theme } from "antd";
import { ORDER_TABS } from "./OrderStatusTabs.constants";

interface OrderStatusTabsProps {
  activeTab: string;
  onChange: (key: string) => void;
  totalCount: number;
  platform?: string;
}

export function OrderStatusTabs({
  activeTab,
  onChange,
  totalCount,
  platform: _platform = "all",
}: OrderStatusTabsProps) {
  const { token } = theme.useToken();

  const items = ORDER_TABS.map((tab) => {
    const tabCount = tab.key === activeTab ? totalCount : 0;
    return {
      key: tab.key,
      label: (
        <span>
          {tab.label}
          {tab.key === activeTab && tabCount > 0 && (
            <span style={{ marginLeft: 4, color: token.colorPrimary }}>
              ({tabCount})
            </span>
          )}
        </span>
      ),
    };
  });

  return (
    <Card
      variant="borderless"
      styles={{ body: { padding: "0 16px" } }}
      style={{ borderRadius: 4 }}
    >
      <Tabs
        activeKey={activeTab}
        onChange={onChange}
        items={items}
        tabBarStyle={{ margin: 0 }}
      />
    </Card>
  );
}
