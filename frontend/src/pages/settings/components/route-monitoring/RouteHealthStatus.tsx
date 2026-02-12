import {
  Card,
  Col,
  Progress,
  Row,
  Space,
  Statistic,
  Typography,
  theme,
} from "antd";
import {
  CheckCircleOutlined,
  CloseCircleOutlined,
  DashboardOutlined,
  WarningOutlined,
} from "@ant-design/icons";
import type { RouteConfig } from "@/types/routeConfig";

const { Text } = Typography;

interface RouteHealthStatusProps {
  routes: RouteConfig[];
}

function getHealthStatus(healthy_percent: number): "good" | "fair" | "poor" {
  if (healthy_percent >= 80) {
    return "good";
  }
  if (healthy_percent >= 50) {
    return "fair";
  }
  return "poor";
}

export function RouteHealthStatus({ routes }: RouteHealthStatusProps) {
  const { token } = theme.useToken();

  const healthy_count = routes.filter((route) => route.enabled).length;
  const warning_count = routes.filter(
    (route) => route.rate_limit_enabled,
  ).length;
  const error_count = routes.filter((route) => !route.enabled).length;
  const total_count = routes.length;
  const healthy_percent =
    total_count === 0 ? 0 : Math.round((healthy_count / total_count) * 100);
  const health_status = getHealthStatus(healthy_percent);

  const progress_status =
    health_status === "good"
      ? "success"
      : health_status === "fair"
        ? "normal"
        : "exception";

  return (
    <Card
      size="small"
      title={
        <Space>
          <DashboardOutlined style={{ color: token.colorPrimary }} />
          <span>Route Health Status</span>
        </Space>
      }
      style={{ borderRadius: token.borderRadius }}
    >
      <Space direction="vertical" size={12} style={{ width: "100%" }}>
        <Progress
          percent={healthy_percent}
          status={progress_status}
          strokeWidth={10}
        />
        <Text type="secondary" style={{ fontSize: token.fontSizeSM }}>
          Health level: {health_status.toUpperCase()} ({healthy_count} of{" "}
          {total_count} routes healthy)
        </Text>
        <Row gutter={[12, 12]}>
          <Col xs={24} sm={8}>
            <Card size="small" style={{ borderRadius: token.borderRadiusSM }}>
              <Statistic
                title="Healthy"
                value={healthy_count}
                prefix={
                  <CheckCircleOutlined style={{ color: token.colorSuccess }} />
                }
              />
            </Card>
          </Col>
          <Col xs={24} sm={8}>
            <Card size="small" style={{ borderRadius: token.borderRadiusSM }}>
              <Statistic
                title="Warning"
                value={warning_count}
                prefix={
                  <WarningOutlined style={{ color: token.colorWarning }} />
                }
              />
            </Card>
          </Col>
          <Col xs={24} sm={8}>
            <Card size="small" style={{ borderRadius: token.borderRadiusSM }}>
              <Statistic
                title="Error"
                value={error_count}
                prefix={
                  <CloseCircleOutlined style={{ color: token.colorError }} />
                }
              />
            </Card>
          </Col>
        </Row>
      </Space>
    </Card>
  );
}
