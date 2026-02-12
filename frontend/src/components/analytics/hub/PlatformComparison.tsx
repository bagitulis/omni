import { Card, Col, Row, Statistic, Button, Typography, theme } from "antd";
import { VideoCameraOutlined, ShopOutlined } from "@ant-design/icons";
import { useNavigate } from "react-router-dom";
import type { UnifiedSummary } from "@/types/analytics";

const { Title } = Typography;
const { useToken } = theme;

interface Props {
  summary?: UnifiedSummary;
}

export const PlatformComparison = ({ summary }: Props) => {
  const navigate = useNavigate();
  const { token } = useToken();

  const formatCurrency = (value: number) => {
    return new Intl.NumberFormat("id-ID", {
      style: "currency",
      currency: "IDR",
      minimumFractionDigits: 0,
      maximumFractionDigits: 0,
    }).format(value);
  };

  return (
    <div style={{ marginBottom: 24 }}>
      <Title level={4}>Platform Comparison</Title>
      <Row gutter={[16, 16]}>
        {/* TikTok */}
        <Col xs={24} md={12}>
          <Card
            title={
              <div style={{ display: "flex", alignItems: "center" }}>
                <VideoCameraOutlined style={{ marginRight: 8, fontSize: 20 }} />
                TikTok Shop
              </div>
            }
            extra={
              <Button type="link" onClick={() => navigate("/analytics/tiktok")}>
                Details
              </Button>
            }
            style={{ height: "100%", borderTop: "3px solid #000000" }}
          >
            <Row gutter={16}>
              <Col span={8}>
                <Statistic
                  title="Revenue"
                  value={summary?.tiktok.total_revenue}
                  formatter={(val) => formatCurrency(Number(val))}
                  valueStyle={{ fontSize: 16 }}
                />
              </Col>
              <Col span={8}>
                <Statistic
                  title="Cost"
                  value={summary?.tiktok.total_cost}
                  formatter={(val) => formatCurrency(Number(val))}
                  valueStyle={{ fontSize: 16 }}
                />
              </Col>
              <Col span={8}>
                <Statistic
                  title="ROAS"
                  value={summary?.tiktok.avg_roas}
                  precision={2}
                  suffix="x"
                  valueStyle={{
                    fontSize: 16,
                    color:
                      (summary?.tiktok.avg_roas || 0) > 0
                        ? token.colorSuccess
                        : token.colorError,
                  }}
                />
              </Col>
            </Row>
          </Card>
        </Col>

        {/* Shopee */}
        <Col xs={24} md={12}>
          <Card
            title={
              <div style={{ display: "flex", alignItems: "center" }}>
                <ShopOutlined
                  style={{ marginRight: 8, fontSize: 20, color: "#ee4d2d" }}
                />
                Shopee
              </div>
            }
            extra={
              <Button type="link" onClick={() => navigate("/analytics/shopee")}>
                Details
              </Button>
            }
            style={{ height: "100%", borderTop: "3px solid #ee4d2d" }}
          >
            <Row gutter={16}>
              <Col span={8}>
                <Statistic
                  title="Revenue"
                  value={summary?.shopee.total_revenue}
                  formatter={(val) => formatCurrency(Number(val))}
                  valueStyle={{ fontSize: 16 }}
                />
              </Col>
              <Col span={8}>
                <Statistic
                  title="Cost"
                  value={summary?.shopee.total_cost}
                  formatter={(val) => formatCurrency(Number(val))}
                  valueStyle={{ fontSize: 16 }}
                />
              </Col>
              <Col span={8}>
                <Statistic
                  title="ROAS"
                  value={summary?.shopee.avg_roas}
                  precision={2}
                  suffix="x"
                  valueStyle={{
                    fontSize: 16,
                    color:
                      (summary?.shopee.avg_roas || 0) > 0
                        ? token.colorSuccess
                        : token.colorError,
                  }}
                />
              </Col>
            </Row>
          </Card>
        </Col>
      </Row>
    </div>
  );
};
