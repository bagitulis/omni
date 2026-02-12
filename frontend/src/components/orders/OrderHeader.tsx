import { Row, Col, Card, Typography, Flex, Skeleton, theme } from "antd";
import { ShoppingOutlined } from "@ant-design/icons";
import { PlatformCounts } from "@/types/order";

const { Title, Text } = Typography;

interface OrderHeaderProps {
  activeTab: string;
  platformCounts?: PlatformCounts;
  totalCount: number;
  loading?: boolean;
}

const PLATFORMS = ["shopee", "tiktok", "lazada"];

export function OrderHeader({
  activeTab,
  platformCounts,
  totalCount,
  loading,
}: OrderHeaderProps) {
  const { token } = theme.useToken();
  const platformConfig: Record<
    string,
    { color: string; bgColor: string; label: string }
  > = {
    shopee: {
      color: token.colorPrimary,
      bgColor: token.colorFillQuaternary,
      label: "Shopee",
    },
    tiktok: {
      color: token.colorText,
      bgColor: token.colorFillQuaternary,
      label: "TikTok",
    },
    lazada: {
      color: token.colorInfo,
      bgColor: token.colorFillSecondary,
      label: "Lazada",
    },
  };

  const getTabLabel = () => {
    switch (activeTab) {
      case "unpaid":
        return "Unpaid Orders";
      case "unprocess":
        return "To Process";
      case "processed":
        return "Processed";
      case "locked":
        return "Locked Products";
      case "today":
        return "Today's Orders";
      default:
        return "Orders";
    }
  };

  return (
    <Flex vertical gap={16}>
      {/* Page Title */}
      <Flex justify="space-between" align="center">
        <div>
          <Flex align="center" gap={8}>
            <ShoppingOutlined
              style={{ fontSize: 20, color: token.colorPrimary }}
            />
            <Title level={3} style={{ margin: 0 }}>
              Order Manager
            </Title>
          </Flex>
          <Text type="secondary" style={{ fontSize: 12, marginLeft: 28 }}>
            Manage and monitor all orders from various platforms
          </Text>
        </div>
        <div>
          <Text strong style={{ fontSize: 14 }}>
            {getTabLabel()}: {loading ? "..." : totalCount}
          </Text>
        </div>
      </Flex>

      {/* Platform Stats Cards */}
      <Row gutter={[12, 12]}>
        {PLATFORMS.map((platform) => {
          const config = platformConfig[platform];
          const count = platformCounts?.[platform] ?? 0;

          return (
            <Col xs={8} key={platform}>
              <Card
                data-testid={`order-header-platform-${platform}`}
                size="small"
                style={{
                  borderRadius: 4,
                  border: `1px solid ${config.color}20`,
                  background: config.bgColor,
                }}
                styles={{ body: { padding: "12px 16px" } }}
              >
                {loading ? (
                  <Skeleton.Button active size="small" style={{ width: 60 }} />
                ) : (
                  <Flex align="center" justify="space-between">
                    <div>
                      <Text
                        strong
                        style={{
                          color: config.color,
                          fontSize: 11,
                          textTransform: "uppercase",
                          letterSpacing: "0.5px",
                        }}
                      >
                        {config.label}
                      </Text>
                      <div
                        data-testid={`order-header-platform-count-${platform}`}
                        style={{
                          fontSize: 20,
                          fontWeight: 700,
                          color: config.color,
                          lineHeight: 1.2,
                        }}
                      >
                        {count}
                      </div>
                      <Text type="secondary" style={{ fontSize: 10 }}>
                        {activeTab === "locked" ? "Products" : "Orders"}
                      </Text>
                    </div>
                    <div
                      style={{
                        width: 36,
                        height: 36,
                        borderRadius: 4,
                        background: config.color,
                        display: "flex",
                        alignItems: "center",
                        justifyContent: "center",
                      }}
                    >
                      <ShoppingOutlined
                        style={{ color: "white", fontSize: 16 }}
                      />
                    </div>
                  </Flex>
                )}
              </Card>
            </Col>
          );
        })}
      </Row>
    </Flex>
  );
}
