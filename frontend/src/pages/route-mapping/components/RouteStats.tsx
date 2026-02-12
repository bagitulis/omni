import {
  ApiOutlined,
  AppstoreOutlined,
  CheckCircleOutlined,
  DisconnectOutlined,
} from "@ant-design/icons";
import { Card, Col, Row, Statistic, theme } from "antd";
import type { RouteData } from "@/types/routeMapping";

interface RouteStatsProps {
  data: RouteData | undefined;
  categoryStats: {
    connected: number;
    frontend_only: number;
    backend_only: number;
    unused: number;
  };
}

export function RouteStats({ data, categoryStats }: RouteStatsProps) {
  const { token } = theme.useToken();
  return (
    <Row gutter={[16, 16]}>
      <Col xs={24} sm={12} md={4} style={{ flex: 1 }}>
        <Card bordered={false} bodyStyle={{ padding: "16px" }}>
          <Statistic
            title="Total Routes"
            value={data?.total_routes || 0}
            prefix={<ApiOutlined style={{ color: token.colorPrimary }} />}
            valueStyle={{ fontSize: "1.25rem", fontWeight: 600 }}
          />
        </Card>
      </Col>
      <Col xs={24} sm={12} md={4} style={{ flex: 1 }}>
        <Card bordered={false} bodyStyle={{ padding: "16px" }}>
          <Statistic
            title="Connection Rate"
            value={data?.connection_rate || "0%"}
            prefix={
              <CheckCircleOutlined style={{ color: token.colorSuccess }} />
            }
            valueStyle={{ fontSize: "1.25rem", fontWeight: 600 }}
          />
        </Card>
      </Col>
      <Col xs={24} sm={12} md={4} style={{ flex: 1 }}>
        <Card bordered={false} bodyStyle={{ padding: "16px" }}>
          <Statistic
            title="Frontend Only"
            value={categoryStats.frontend_only}
            prefix={<DisconnectOutlined style={{ color: token.colorError }} />}
            valueStyle={{ fontSize: "1.25rem", fontWeight: 600 }}
          />
        </Card>
      </Col>
      <Col xs={24} sm={12} md={4} style={{ flex: 1 }}>
        <Card bordered={false} bodyStyle={{ padding: "16px" }}>
          <Statistic
            title="Backend Only"
            value={categoryStats.backend_only}
            prefix={<ApiOutlined style={{ color: token.colorWarning }} />}
            valueStyle={{ fontSize: "1.25rem", fontWeight: 600 }}
          />
        </Card>
      </Col>
      <Col xs={24} sm={12} md={4} style={{ flex: 1 }}>
        <Card bordered={false} bodyStyle={{ padding: "16px" }}>
          <Statistic
            title="Components"
            value={data?.total_components || 0}
            prefix={<AppstoreOutlined style={{ color: token.colorInfo }} />}
            valueStyle={{ fontSize: "1.25rem", fontWeight: 600 }}
          />
        </Card>
      </Col>
    </Row>
  );
}
