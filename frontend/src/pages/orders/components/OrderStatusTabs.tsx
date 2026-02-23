import { Card, Flex, theme } from "antd";
import {
  SendOutlined,
  CheckCircleOutlined,
  CarOutlined,
  TrophyOutlined,
  CloseCircleOutlined,
  LockOutlined,
  CalendarOutlined,
} from "@ant-design/icons";
import { ORDER_TABS } from "./OrderStatusTabs.constants";

const TAB_ICONS: Record<string, React.ReactNode> = {
  unprocess: <SendOutlined />,
  processed: <CheckCircleOutlined />,
  shipped: <CarOutlined />,
  completed: <TrophyOutlined />,
  cancelled: <CloseCircleOutlined />,
  locked: <LockOutlined />,
  today: <CalendarOutlined />,
};

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

  return (
    <Card
      variant="borderless"
      styles={{ body: { padding: "8px 12px" } }}
      style={{
        borderRadius: 3,
        boxShadow: "0 1px 3px rgba(0,0,0,0.06)",
      }}
    >
      <Flex
        gap={6}
        wrap={false}
        style={{
          overflowX: "auto",
          scrollbarWidth: "none",
          msOverflowStyle: "none",
        }}
      >
        {ORDER_TABS.map((tab) => {
          const isActive = activeTab === tab.key;
          const tabCount = isActive ? totalCount : 0;
          const isSpecial = tab.key === "locked" || tab.key === "today";

          return (
            <button
              key={tab.key}
              onClick={() => onChange(tab.key)}
              style={{
                display: "inline-flex",
                alignItems: "center",
                gap: 6,
                padding: "6px 14px",
                border: "none",
                borderRadius: 3,
                cursor: "pointer",
                fontSize: 13,
                fontWeight: isActive ? 600 : 500,
                whiteSpace: "nowrap",
                flexShrink: 0,
                transition: "all 0.2s ease",
                background: isActive
                  ? isSpecial
                    ? token.colorErrorBg
                    : token.colorPrimaryBg
                  : "transparent",
                color: isActive
                  ? isSpecial
                    ? token.colorError
                    : token.colorPrimary
                  : token.colorTextSecondary,
                boxShadow: isActive
                  ? isSpecial
                    ? `inset 0 0 0 1px ${token.colorErrorBorder}`
                    : `inset 0 0 0 1px ${token.colorPrimaryBorder}`
                  : "none",
              }}
              onMouseEnter={(e) => {
                if (!isActive) {
                  e.currentTarget.style.background = token.colorFillSecondary;
                  e.currentTarget.style.color = token.colorText;
                }
              }}
              onMouseLeave={(e) => {
                if (!isActive) {
                  e.currentTarget.style.background = "transparent";
                  e.currentTarget.style.color = token.colorTextSecondary;
                }
              }}
            >
              <span style={{ fontSize: 13, lineHeight: 1 }}>
                {TAB_ICONS[tab.key]}
              </span>
              <span>{tab.label}</span>
              {isActive && tabCount > 0 && (
                <span
                  style={{
                    fontSize: 11,
                    fontWeight: 700,
                    background: isSpecial
                      ? token.colorErrorBgHover
                      : token.colorPrimaryBgHover,
                    color: "inherit",
                    borderRadius: 3,
                    padding: "1px 6px",
                    minWidth: 20,
                    textAlign: "center",
                  }}
                >
                  {tabCount}
                </span>
              )}
            </button>
          );
        })}
      </Flex>
    </Card>
  );
}
