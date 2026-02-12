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
  ApartmentOutlined,
  DatabaseOutlined,
  DeploymentUnitOutlined,
} from "@ant-design/icons";
import type { RouteConfig } from "@/types/routeConfig";

const { Text } = Typography;

interface RouteStatesSectionProps {
  routes: RouteConfig[];
}

function getPercent(part: number, total: number): number {
  return total === 0 ? 0 : Math.round((part / total) * 100);
}

export function RouteStatesSection({ routes }: RouteStatesSectionProps) {
  const { token } = theme.useToken();
  const total = routes.length;

  const enabled = routes.filter((route) => route.enabled).length;
  const cached = routes.filter((route) => route.caching_enabled).length;
  const queued = routes.filter((route) => route.queue_enabled).length;

  const state_cards = [
    {
      title: "Enabled State",
      icon: <DeploymentUnitOutlined style={{ color: token.colorPrimary }} />,
      active_label: "Enabled",
      inactive_label: "Disabled",
      active_count: enabled,
      inactive_count: total - enabled,
      percent: getPercent(enabled, total),
    },
    {
      title: "Cache State",
      icon: <DatabaseOutlined style={{ color: token.colorPrimary }} />,
      active_label: "Cached",
      inactive_label: "Uncached",
      active_count: cached,
      inactive_count: total - cached,
      percent: getPercent(cached, total),
    },
    {
      title: "Queue State",
      icon: <ApartmentOutlined style={{ color: token.colorPrimary }} />,
      active_label: "Queued",
      inactive_label: "Unqueued",
      active_count: queued,
      inactive_count: total - queued,
      percent: getPercent(queued, total),
    },
  ];

  return (
    <Card
      size="small"
      title="Route States"
      style={{ borderRadius: token.borderRadius }}
    >
      <Row gutter={[12, 12]}>
        {state_cards.map((state) => (
          <Col xs={24} md={8} key={state.title}>
            <Card
              size="small"
              title={
                <Space size={8}>
                  {state.icon}
                  <Text>{state.title}</Text>
                </Space>
              }
              style={{ borderRadius: token.borderRadiusSM }}
            >
              <Space direction="vertical" size={10} style={{ width: "100%" }}>
                <Row gutter={8}>
                  <Col span={12}>
                    <Statistic
                      title={state.active_label}
                      value={state.active_count}
                    />
                  </Col>
                  <Col span={12}>
                    <Statistic
                      title={state.inactive_label}
                      value={state.inactive_count}
                    />
                  </Col>
                </Row>
                <Progress
                  percent={state.percent}
                  strokeColor={token.colorPrimary}
                />
              </Space>
            </Card>
          </Col>
        ))}
      </Row>
    </Card>
  );
}
