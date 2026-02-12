import { Card, Col, Row, Space, Tag, Typography, theme } from "antd";
import { ClockCircleOutlined, WarningOutlined } from "@ant-design/icons";
import type { RouteConfig } from "@/types/routeConfig";

const { Text } = Typography;

interface RouteSlowRoutesSectionProps {
  routes: RouteConfig[];
}

export function RouteSlowRoutesSection({
  routes,
}: RouteSlowRoutesSectionProps) {
  const { token } = theme.useToken();

  const slow_routes = routes.filter((route) => route.timeout > 30000);

  return (
    <Card
      size="small"
      title={
        <Space>
          <WarningOutlined style={{ color: token.colorWarning }} />
          <span>Slow Routes ({slow_routes.length})</span>
        </Space>
      }
      style={{ borderRadius: token.borderRadius }}
    >
      {slow_routes.length === 0 ? (
        <Text type="secondary">No routes exceed 30000 ms timeout.</Text>
      ) : (
        <Row gutter={[12, 12]}>
          {slow_routes.map((route) => (
            <Col xs={24} md={12} key={route.id}>
              <Card size="small" style={{ borderRadius: token.borderRadiusSM }}>
                <Space direction="vertical" size={8} style={{ width: "100%" }}>
                  <Text code>{route.route_path}</Text>
                  <Space size={8} wrap>
                    <Tag color={token.colorPrimary}>{route.route_method}</Tag>
                    <Tag color={token.colorWarning}>
                      <ClockCircleOutlined /> {route.timeout} ms
                    </Tag>
                  </Space>
                </Space>
              </Card>
            </Col>
          ))}
        </Row>
      )}
    </Card>
  );
}
