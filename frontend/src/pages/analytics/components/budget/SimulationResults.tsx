import {
  Card,
  Typography,
  Row,
  Col,
  Statistic,
  Divider,
  Alert,
  GlobalToken,
} from "antd";
import {
  LineChartOutlined,
  DollarOutlined,
  InfoCircleOutlined,
} from "@ant-design/icons";
import type { SimulationResult } from "../../../../api/analyticsIntelligence";

const { Title, Text } = Typography;

interface SimulationResultsProps {
  simulationResult: SimulationResult | undefined;
  token: GlobalToken;
  formatCurrency: (value: number) => string;
  formatRoas: (value: number) => string;
  getFeasibilityColor: (feasibility: string) => string;
}

export const SimulationResults = ({
  simulationResult,
  token,
  formatCurrency,
  formatRoas,
  getFeasibilityColor,
}: SimulationResultsProps) => {
  return (
    <Card
      style={{ borderRadius: token.borderRadiusLG, height: "100%" }}
      styles={{ body: { height: "100%" } }}
    >
      {!simulationResult ? (
        <div
          style={{
            height: "100%",
            minHeight: 400,
            display: "flex",
            flexDirection: "column",
            alignItems: "center",
            justifyContent: "center",
            color: token.colorTextSecondary,
          }}
        >
          <LineChartOutlined style={{ fontSize: 48, marginBottom: 16 }} />
          <Title level={4} style={{ color: token.colorTextSecondary }}>
            No Simulation Data
          </Title>
          <Text type="secondary">
            Select a product and run simulation to see projections
          </Text>
        </div>
      ) : (
        <div style={{ display: "flex", flexDirection: "column", gap: 24 }}>
          {/* Feasibility Header */}
          <div
            style={{
              padding: 24,
              background: `${getFeasibilityColor(
                simulationResult.feasibility,
              )}15`,
              borderRadius: token.borderRadiusLG,
              border: `1px solid ${getFeasibilityColor(
                simulationResult.feasibility,
              )}`,
              textAlign: "center",
            }}
          >
            <Title
              level={3}
              style={{
                margin: 0,
                color: getFeasibilityColor(simulationResult.feasibility),
              }}
            >
              {simulationResult.feasibility}
            </Title>
            <Text type="secondary">
              {simulationResult.confidence_percent}% Confidence Level
            </Text>
          </div>

          <Row gutter={24}>
            <Col span={12}>
              <Card size="small">
                <Statistic
                  title="Current ROAS"
                  value={simulationResult.current_roas}
                  precision={2}
                  suffix="x"
                />
              </Card>
            </Col>
            <Col span={12}>
              <Card size="small">
                <Statistic
                  title="Projected ROAS"
                  value={simulationResult.projected_roas}
                  precision={2}
                  suffix="x"
                  valueStyle={{
                    color:
                      simulationResult.projected_roas >=
                      simulationResult.current_roas
                        ? token.colorSuccess
                        : token.colorError,
                  }}
                  prefix={
                    simulationResult.trend_prediction === "UP" ? (
                      <LineChartOutlined />
                    ) : undefined
                  }
                />
              </Card>
            </Col>
          </Row>

          <Card size="small">
            <Statistic
              title="Optimal Daily Budget"
              value={simulationResult.optimal_budget}
              precision={0}
              prefix={<DollarOutlined />}
              formatter={(value) => formatCurrency(Number(value))}
            />
            <Text type="secondary" style={{ marginTop: 8, display: "block" }}>
              Recommended daily budget to achieve target ROAS
            </Text>
          </Card>

          <Alert
            message="Recommendation"
            description={simulationResult.recommendation}
            type="info"
            showIcon
            icon={<InfoCircleOutlined />}
          />

          {(simulationResult?.alternatives ?? []).length > 0 && (
            <div>
              <Divider orientation="left">Alternatives</Divider>
              <ul style={{ paddingLeft: 20 }}>
                {(simulationResult?.alternatives ?? []).map((alt, idx) => (
                  <li
                    key={`${idx}-${alt.target_roas}-${alt.budget}`}
                    style={{ marginBottom: 8 }}
                  >
                    <Text>
                      {alt.target_roas && alt.required_budget ? (
                        <>
                          To reach ROAS{" "}
                          <Text strong>{formatRoas(alt.target_roas)}</Text>, you
                          need{" "}
                          <Text strong>
                            {formatCurrency(alt.required_budget)}/day
                          </Text>
                        </>
                      ) : alt.budget && alt.expected_roas ? (
                        <>
                          Budget{" "}
                          <Text strong>{formatCurrency(alt.budget)}/day</Text>{" "}
                          yields{" "}
                          <Text strong>{formatRoas(alt.expected_roas)}</Text>
                        </>
                      ) : null}
                    </Text>
                  </li>
                ))}
              </ul>
            </div>
          )}
        </div>
      )}
    </Card>
  );
};
