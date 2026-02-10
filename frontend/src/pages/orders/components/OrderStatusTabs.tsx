import { Tabs, Card } from "antd";

export const ORDER_TABS = [
  { key: "unpaid", label: "Unpaid" },
  { key: "unprocess", label: "To Ship" },
  { key: "processed", label: "Shipped" },
  { key: "locked", label: "Locked Today" },
  { key: "today", label: "Today's Orders" },
];

interface OrderStatusTabsProps {
  activeTab: string;
  onChange: (key: string) => void;
  totalCount: number;
}

export function OrderStatusTabs({
  activeTab,
  onChange,
  totalCount,
}: OrderStatusTabsProps) {
  const items = ORDER_TABS.map((tab) => {
    const tabCount = tab.key === activeTab ? totalCount : 0;
    return {
      key: tab.key,
      label: (
        <span>
          {tab.label}
          {tab.key === activeTab && tabCount > 0 && (
            <span style={{ marginLeft: 4, color: "#ee4d2d" }}>
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
