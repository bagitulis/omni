import { Card, Col, Row, Typography, theme } from "antd";
import {
  ShopOutlined,
  VideoCameraOutlined,
  FundOutlined,
  ExperimentOutlined,
  CalculatorOutlined,
  AppstoreOutlined,
  ArrowRightOutlined,
} from "@ant-design/icons";
import { useNavigate } from "react-router-dom";

const { Title, Text } = Typography;
const { useToken } = theme;

export const QuickActions = () => {
  const navigate = useNavigate();
  const { token } = useToken();

  const cards = [
    {
      title: "Shopee Analytics",
      description: "Sales performance and traffic insights for Shopee stores",
      icon: (
        <ShopOutlined style={{ fontSize: 24, color: token.colorPrimary }} />
      ),
      path: "/analytics/shopee",
      bgColor: token.colorFillQuaternary,
    },
    {
      title: "Shopee Ads",
      description: "Shopee advertising ROI, CTR, CPC, and ROAS tracking",
      icon: (
        <FundOutlined style={{ fontSize: 24, color: token.colorPrimary }} />
      ),
      path: "/analytics/shopee-ads",
      bgColor: token.colorFillQuaternary,
    },
    {
      title: "TikTok Analytics",
      description: "Video engagement and livestream metrics for TikTok Shop",
      icon: (
        <VideoCameraOutlined style={{ fontSize: 24, color: token.colorText }} />
      ),
      path: "/analytics/tiktok",
      bgColor: token.colorFillSecondary,
    },
    {
      title: "TikTok Ads",
      description: "TikTok creative performance, ROI, and engagement metrics",
      icon: <FundOutlined style={{ fontSize: 24, color: token.colorText }} />,
      path: "/analytics/tiktok-ads",
      bgColor: token.colorFillSecondary,
    },
    {
      title: "ML Dashboard",
      description: "AI-driven predictions and inventory optimization models",
      icon: (
        <ExperimentOutlined
          style={{ fontSize: 24, color: token.colorPrimary }}
        />
      ),
      path: "/analytics/ml",
      bgColor: token.colorFillSecondary,
    },
    {
      title: "Budget Simulator",
      description: "Simulate ad budget allocation and predict campaign ROI",
      icon: (
        <CalculatorOutlined style={{ fontSize: 24, color: token.colorInfo }} />
      ),
      path: "/analytics/budget-simulator",
      bgColor: token.colorFillSecondary,
    },
    {
      title: "Product Classification",
      description: "AI-powered product categorization and performance grouping",
      icon: (
        <AppstoreOutlined style={{ fontSize: 24, color: token.colorSuccess }} />
      ),
      path: "/analytics/product-classification",
      bgColor: token.colorFillSecondary,
    },

  ];

  return (
    <div style={{ marginBottom: 24 }}>
      <Title level={4} style={{ marginTop: 0 }}>
        Quick Actions
      </Title>
      <Row gutter={[16, 16]}>
        {cards.map((card) => (
          <Col xs={24} sm={12} md={12} lg={6} key={card.title}>
            <Card
              hoverable
              bordered={false}
              style={{
                height: "100%",
                cursor: "pointer",
                borderRadius: token.borderRadiusLG,
              }}
              styles={{ body: { padding: 20, height: "100%" } }}
              onClick={() => navigate(card.path)}
            >
              <div
                style={{
                  display: "flex",
                  flexDirection: "column",
                  height: "100%",
                  gap: 16,
                }}
              >
                <div
                  style={{
                    width: 48,
                    height: 48,
                    borderRadius: token.borderRadius,
                    background: card.bgColor,
                    display: "flex",
                    alignItems: "center",
                    justifyContent: "center",
                  }}
                >
                  {card.icon}
                </div>

                <div style={{ flex: 1 }}>
                  <Title
                    level={5}
                    style={{
                      marginTop: 0,
                      marginBottom: 8,
                      fontSize: 16,
                    }}
                  >
                    {card.title}
                  </Title>
                  <Text
                    type="secondary"
                    style={{
                      fontSize: token.fontSize,
                      lineHeight: 1.5,
                    }}
                  >
                    {card.description}
                  </Text>
                </div>

                <div
                  style={{
                    display: "flex",
                    alignItems: "center",
                    color: token.colorPrimary,
                  }}
                >
                  <span
                    style={{
                      fontSize: token.fontSize,
                      fontWeight: 500,
                      marginRight: 4,
                    }}
                  >
                    View Dashboard
                  </span>
                  <ArrowRightOutlined style={{ fontSize: 12 }} />
                </div>
              </div>
            </Card>
          </Col>
        ))}
      </Row>
    </div>
  );
};
