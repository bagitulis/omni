import { Card, Col, Row, Space, Statistic, theme } from "antd";
import {
  CloseCircleOutlined,
  HourglassOutlined,
  RedoOutlined,
  StopOutlined,
} from "@ant-design/icons";
import type { RouteConfig } from "@/types/routeConfig";

interface RouteErrorStatsProps {
  routes: RouteConfig[];
}

export function RouteErrorStats({ routes }: RouteErrorStatsProps) {
  const { token } = theme.useToken();

  const disabled_count = routes.filter((route) => !route.enabled).length;
  const rate_limited_count = routes.filter(
    (route) => route.rate_limit_enabled,
  ).length;
  const high_timeout_count = routes.filter(
    (route) => route.timeout > 30000,
  ).length;
  const retries_enabled_count = routes.filter(
    (route) => route.retry_enabled,
  ).length;

  return (
    <Card
      size="small"
      title="Error Statistics"
      style={{ borderRadius: token.borderRadius }}
    >
      <Row gutter={[12, 12]}>
        <Col xs={24} sm={12} lg={6}>
          <Card size="small" style={{ borderRadius: token.borderRadiusSM }}>
            <Space direction="vertical" size={0}>
              <Statistic
                title="Disabled Routes"
                value={disabled_count}
                valueStyle={{ color: token.colorError }}
                prefix={<CloseCircleOutlined />}
              />
            </Space>
          </Card>
        </Col>
        <Col xs={24} sm={12} lg={6}>
          <Card size="small" style={{ borderRadius: token.borderRadiusSM }}>
            <Statistic
              title="Rate Limited"
              value={rate_limited_count}
              valueStyle={{ color: token.colorWarning }}
              prefix={<StopOutlined />}
            />
          </Card>
        </Col>
        <Col xs={24} sm={12} lg={6}>
          <Card size="small" style={{ borderRadius: token.borderRadiusSM }}>
            <Statistic
              title="High Timeout"
              value={high_timeout_count}
              valueStyle={{ color: token.colorError }}
              prefix={<HourglassOutlined />}
            />
          </Card>
        </Col>
        <Col xs={24} sm={12} lg={6}>
          <Card size="small" style={{ borderRadius: token.borderRadiusSM }}>
            <Statistic
              title="Retries Enabled"
              value={retries_enabled_count}
              valueStyle={{ color: token.colorPrimary }}
              prefix={<RedoOutlined />}
            />
          </Card>
        </Col>
      </Row>
    </Card>
  );
}
