import { Card, Col, Row, Statistic, Tag, Typography, theme } from "antd";
import {
  RiseOutlined,
  FallOutlined,
  MinusOutlined,
  StopOutlined,
} from "@ant-design/icons";
import type { KPIData } from "@/types/analytics";

const { Title } = Typography;
const { useToken } = theme;

interface Props {
  kpi?: KPIData;
}

export const ActionSummary = ({ kpi }: Props) => {
  const { token } = useToken();

  return (
    <div style={{ marginBottom: 24 }}>
      <Title level={4}>Action Summary</Title>
      <Card bordered={false}>
        <Row gutter={[16, 16]}>
          <Col xs={12} sm={6}>
            <Statistic
              title="Scale Up"
              value={kpi?.actions.scale_up}
              valueStyle={{ color: token.colorSuccess }}
              prefix={<RiseOutlined />}
              suffix={<Tag color="success">Scale Up</Tag>}
            />
          </Col>
          <Col xs={12} sm={6}>
            <Statistic
              title="Maintain"
              value={kpi?.actions.maintain}
              valueStyle={{ color: token.colorInfo }}
              prefix={<MinusOutlined />}
              suffix={<Tag color="processing">Maintain</Tag>}
            />
          </Col>
          <Col xs={12} sm={6}>
            <Statistic
              title="Reduce"
              value={kpi?.actions.reduce}
              valueStyle={{ color: token.colorWarning }}
              prefix={<FallOutlined />}
              suffix={<Tag color="warning">Reduce</Tag>}
            />
          </Col>
          <Col xs={12} sm={6}>
            <Statistic
              title="Stop"
              value={kpi?.actions.stop}
              valueStyle={{ color: token.colorError }}
              prefix={<StopOutlined />}
              suffix={<Tag color="error">Stop</Tag>}
            />
          </Col>
        </Row>
      </Card>
    </div>
  );
};
