import { Card, Col, Row, Statistic } from "antd";
import {
  ApiOutlined,
  CheckCircleOutlined,
  DisconnectOutlined,
  AppstoreOutlined,
} from "@ant-design/icons";
import { RouteData } from "@/types/routeMapping";

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
  return (
    <Row gutter={[16, 16]}>
      <Col xs={24} sm={12} md={4} style={{ flex: 1 }}>
        <Card bordered={false} bodyStyle={{ padding: "16px" }}>
          <Statistic
            title="Total Routes"
            value={data?.total_routes || 0}
            prefix={<ApiOutlined style={{ color: "#1890ff" }} />}
            valueStyle={{ fontSize: "1.25rem", fontWeight: 600 }}
          />
        </Card>
      </Col>
      <Col xs={24} sm={12} md={4} style={{ flex: 1 }}>
        <Card bordered={false} bodyStyle={{ padding: "16px" }}>
          <Statistic
            title="Connection Rate"
            value={data?.connection_rate || "0%"}
            prefix={<CheckCircleOutlined style={{ color: "#52c41a" }} />}
            valueStyle={{ fontSize: "1.25rem", fontWeight: 600 }}
          />
        </Card>
      </Col>
      <Col xs={24} sm={12} md={4} style={{ flex: 1 }}>
        <Card bordered={false} bodyStyle={{ padding: "16px" }}>
          <Statistic
            title="Frontend Only"
            value={categoryStats.frontend_only}
            prefix={<DisconnectOutlined style={{ color: "#ff4d4f" }} />}
            valueStyle={{ fontSize: "1.25rem", fontWeight: 600 }}
          />
        </Card>
      </Col>
      <Col xs={24} sm={12} md={4} style={{ flex: 1 }}>
        <Card bordered={false} bodyStyle={{ padding: "16px" }}>
          <Statistic
            title="Backend Only"
            value={categoryStats.backend_only}
            prefix={<ApiOutlined style={{ color: "#faad14" }} />}
            valueStyle={{ fontSize: "1.25rem", fontWeight: 600 }}
          />
        </Card>
      </Col>
      <Col xs={24} sm={12} md={4} style={{ flex: 1 }}>
        <Card bordered={false} bodyStyle={{ padding: "16px" }}>
          <Statistic
            title="Components"
            value={data?.total_components || 0}
            prefix={<AppstoreOutlined style={{ color: "#722ed1" }} />}
            valueStyle={{ fontSize: "1.25rem", fontWeight: 600 }}
          />
        </Card>
      </Col>
    </Row>
  );
}
