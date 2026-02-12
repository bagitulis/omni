import { Tabs, Card, theme } from "antd";

export const ALL_TABS = [
  { key: "unpaid", label: "Unpaid" },
  { key: "unprocess", label: "To Ship" },
  { key: "processed", label: "Shipped" },
  { key: "locked", label: "Locked Today" },
  { key: "today", label: "Today's Orders" },
];

export const SHOPEE_TABS = [
  { key: "UNPAID", label: "Unpaid" },
  { key: "READY_TO_SHIP", label: "To Ship" },
  { key: "PROCESSED", label: "Processed" },
  { key: "SHIPPED", label: "Shipped" },
  { key: "COMPLETED", label: "Completed" },
  { key: "IN_CANCEL", label: "In Cancel" },
  { key: "CANCELLED", label: "Cancelled" },
];

export const LAZADA_TABS = [
  { key: "unpaid", label: "Unpaid" },
  { key: "topack", label: "To Pack" },
  { key: "toship", label: "To Ship" },
  { key: "shipped", label: "Shipped" },
  { key: "delivered", label: "Delivered" },
  { key: "failed", label: "Failed" },
  { key: "returned", label: "Returned" },
];

export const TIKTOK_TABS = [
  { key: "AWAITING_SHIPMENT", label: "To Ship" },
  { key: "AWAITING_COLLECTION", label: "To Collect" },
  { key: "IN_TRANSIT", label: "In Transit" },
  { key: "DELIVERED", label: "Delivered" },
  { key: "COMPLETED", label: "Completed" },
  { key: "CANCELLED", label: "Cancelled" },
];

export const ORDER_TABS = ALL_TABS;

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
  platform = "all",
}: OrderStatusTabsProps) {
  const { token } = theme.useToken();
  const getTabs = () => {
    switch (platform) {
      case "shopee":
        return SHOPEE_TABS;
      case "lazada":
        return LAZADA_TABS;
      case "tiktok":
        return TIKTOK_TABS;
      default:
        return ALL_TABS;
    }
  };

  const items = getTabs().map((tab) => {
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
