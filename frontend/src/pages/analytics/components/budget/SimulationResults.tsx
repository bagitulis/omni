import {
  Card,
  Typography,
  Row,
  Col,
  Statistic,
  Divider,
  Alert,
  Tag,
  GlobalToken,
} from "antd";
import {
  LineChartOutlined,
  DollarOutlined,
  InfoCircleOutlined,
  ExperimentOutlined,
} from "@ant-design/icons";
import type { SimulationResult } from "@/api/analyticsIntelligence";

const { Title, Text } = Typography;

interface SimulationResultsProps {
  simulationResult: SimulationResult | undefined;
  token: GlobalToken;
  formatCurrency: (value: number) => string;
  formatRoas: (value: number) => string;
  getFeasibilityColor: (feasibility: string) => string;
}

const mlCategoryColors: Record<string, string> = {
  STAR: "green",
  GROWTH: "blue",
  STABLE: "cyan",
  WATCH: "orange",
  PROBLEM: "red",
};

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
          {/* Feasibility Header with ML Category */}
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
            <div
              style={{
                marginTop: 8,
                display: "flex",
                justifyContent: "center",
                gap: 8,
                flexWrap: "wrap",
              }}
            >
              <Text type="secondary">
                {simulationResult.confidence_percent}% Confidence
              </Text>
              {simulationResult.ml_category && (
                <Tag
                  color={
                    mlCategoryColors[simulationResult.ml_category] || "default"
                  }
                >
                  {simulationResult.ml_category}
                </Tag>
              )}
              <Tag
                icon={<ExperimentOutlined />}
                color={
                  simulationResult.simulation_mode === "FUNNEL"
                    ? "purple"
                    : "default"
                }
              >
                {simulationResult.simulation_mode === "FUNNEL"
                  ? "Funnel Model"
                  : "Yield Model"}
              </Tag>
            </div>
          </div>

          <Row gutter={[16, 16]}>
            <Col xs={12} sm={8}>
              <Card size="small">
                <Statistic
                  title="Current ROAS"
                  value={simulationResult.current_roas}
                  precision={2}
                  suffix="x"
                />
              </Card>
            </Col>
            <Col xs={12} sm={8}>
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
            <Col xs={24} sm={8}>
              <Card size="small">
                <Statistic
                  title="Current Daily Spend"
                  value={simulationResult.current_daily_spend}
                  precision={0}
                  formatter={(value) => formatCurrency(Number(value))}
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
                    key={`alt-${idx}-${alt.target_roas}-${alt.required_budget}`}
                    style={{ marginBottom: 8 }}
                  >
                    <Text>
                      To reach ROAS{" "}
                      <Text strong>{formatRoas(alt.target_roas)}</Text>, budget{" "}
                      <Text strong>
                        {formatCurrency(alt.required_budget)}/day
                      </Text>{" "}
                      → expected ROAS{" "}
                      <Text strong>{formatRoas(alt.expected_roas)}</Text>
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
