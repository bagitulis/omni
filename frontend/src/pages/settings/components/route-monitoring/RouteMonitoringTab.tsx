import { Card, Col, Row, Space, Typography, theme } from "antd";
import { MonitorOutlined } from "@ant-design/icons";
import type { RouteConfig } from "@/types/routeConfig";
import { RouteFlowControls } from "../RouteFlowControls";
import { RouteManagementCacheSettings } from "../RouteManagementCacheSettings";
import { RouteErrorStats } from "./RouteErrorStats";
import { RouteExcludedRoutesManager } from "./RouteExcludedRoutesManager";
import { RouteHealthStatus } from "./RouteHealthStatus";
import { RoutePerformanceTable } from "./RoutePerformanceTable";
import { RouteQueueSection } from "./RouteQueueSection";
import { RouteSlowRoutesSection } from "./RouteSlowRoutesSection";
import { RouteStatesSection } from "./RouteStatesSection";

const { Title, Text } = Typography;

interface RouteMonitoringTabProps {
  routes: RouteConfig[];
}

export function RouteMonitoringTab({ routes }: RouteMonitoringTabProps) {
  const { token } = theme.useToken();

  return (
    <Space direction="vertical" size={16} style={{ width: "100%" }}>
      <Card size="small" style={{ borderRadius: token.borderRadius }}>
        <Space direction="vertical" size={4}>
          <Space>
            <MonitorOutlined style={{ color: token.colorPrimary }} />
            <Title level={5} style={{ margin: 0 }}>
              Route Monitoring
            </Title>
          </Space>
          <Text type="secondary">
            Monitoring data is computed from the current route configuration
            state.
          </Text>
        </Space>
      </Card>

      <RouteHealthStatus routes={routes} />
      <RouteErrorStats routes={routes} />
      <RoutePerformanceTable routes={routes} />
      <RouteQueueSection routes={routes} />
      <RouteStatesSection routes={routes} />
      <RouteSlowRoutesSection routes={routes} />

      <Row gutter={[16, 16]}>
        <Col xs={24} lg={12}>
          <RouteExcludedRoutesManager />
        </Col>
        <Col xs={24} lg={12}>
          <RouteFlowControls />
        </Col>
      </Row>

      <RouteManagementCacheSettings />
    </Space>
  );
}
