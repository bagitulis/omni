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
  ShoppingCartOutlined,
} from "@ant-design/icons";
import type { SimulationResult } from "@/api/analyticsIntelligence";
import { FunnelProjection } from "./FunnelProjection";
import { ScaleUpAnalysis } from "./ScaleUpAnalysis";
import { CalendarEvents } from "./CalendarEvents";

const { Title, Text } = Typography;

interface SimulationResultsProps {
  simulationResult: SimulationResult | undefined;
  periodDays: number;
  targetRoas: number;
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
  periodDays,
  targetRoas,
  token,
  formatCurrency,
  formatRoas,
  getFeasibilityColor,
}: SimulationResultsProps) => {
  if (!simulationResult) {
    return (
      <Card
        style={{ borderRadius: token.borderRadiusLG, height: "100%" }}
        styles={{ body: { height: "100%" } }}
      >
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
      </Card>
    );
  }

  const r = simulationResult;
  const roasChange = r.current_roas > 0
    ? ((r.projected_roas - r.current_roas) / r.current_roas) * 100
    : 0;

  return (
    <Card
      style={{ borderRadius: token.borderRadiusLG }}
      styles={{ body: { padding: "20px" } }}
    >
      <div style={{ display: "flex", flexDirection: "column", gap: 16 }}>
        {/* Feasibility Header */}
        <div
          style={{
            padding: "16px 24px",
            background: `${getFeasibilityColor(r.feasibility)}15`,
            borderRadius: token.borderRadiusLG,
            border: `1px solid ${getFeasibilityColor(r.feasibility)}`,
            textAlign: "center",
          }}
        >
          <Title
            level={3}
            style={{
              margin: 0,
              color: getFeasibilityColor(r.feasibility),
            }}
          >
            {r.feasibility}
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
              {r.confidence_percent}% Confidence
            </Text>
            {r.ml_category && (
              <Tag
                color={mlCategoryColors[r.ml_category] || "default"}
              >
                {r.ml_category}
              </Tag>
            )}
            <Tag
              icon={<ExperimentOutlined />}
              color={r.simulation_mode === "FUNNEL" ? "purple" : "default"}
            >
              {r.simulation_mode === "FUNNEL" ? "Funnel Model" : "Yield Model"}
            </Tag>
            {r.calendar_multiplier && r.calendar_multiplier !== 1.0 && (
              <Tag color={r.calendar_multiplier >= 1.0 ? "green" : "orange"}>
                Seasonal {r.calendar_multiplier.toFixed(2)}x
              </Tag>
            )}
          </div>
        </div>

        {/* Key Metrics: ROAS, Revenue, Orders */}
        <Row gutter={[12, 12]}>
          <Col xs={24} sm={8}>
            <Card size="small">
              <Statistic
                title="ROAS"
                value={r.projected_roas}
                precision={2}
                suffix="x"
                valueStyle={{
                  color: r.projected_roas >= r.current_roas
                    ? token.colorSuccess
                    : token.colorError,
                }}
                prefix={
                  r.trend_prediction === "UP" ? (
                    <LineChartOutlined />
                  ) : undefined
                }
              />
              <div style={{ marginTop: 4 }}>
                <Text type="secondary" style={{ fontSize: 11 }}>
                  Current: {formatRoas(r.current_roas)} →{" "}
                  <Text
                    style={{
                      fontSize: 11,
                      color: roasChange >= 0 ? token.colorSuccess : token.colorError,
                    }}
                  >
                    {roasChange >= 0 ? "+" : ""}{roasChange.toFixed(1)}%
                  </Text>
                </Text>
              </div>
            </Card>
          </Col>
          <Col xs={12} sm={8}>
            <Card size="small">
              <Statistic
                title="Revenue (Daily)"
                value={r.funnel?.projected_revenue || r.projected_roas * r.current_daily_spend}
                precision={0}
                formatter={(value) => formatCurrency(Number(value))}
                prefix={<DollarOutlined />}
                valueStyle={{ fontSize: 18 }}
              />
              {r.total_projected_revenue && (
                <Text type="secondary" style={{ fontSize: 11 }}>
                  Total ({periodDays}d): {formatCurrency(r.total_projected_revenue)}
                </Text>
              )}
            </Card>
          </Col>
          <Col xs={12} sm={8}>
            <Card size="small">
              <Statistic
                title="Orders (Daily)"
                value={r.funnel?.projected_orders || 0}
                precision={1}
                prefix={<ShoppingCartOutlined />}
                valueStyle={{ fontSize: 18 }}
              />
              {r.total_projected_orders != null && r.total_projected_orders > 0 && (
                <Text type="secondary" style={{ fontSize: 11 }}>
                  Total ({periodDays}d): {r.total_projected_orders.toFixed(0)} orders
                </Text>
              )}
            </Card>
          </Col>
        </Row>

        {/* Budget Summary */}
        <Row gutter={[12, 12]}>
          <Col xs={12}>
            <Card size="small">
              <Statistic
                title="Current Daily Spend"
                value={r.current_daily_spend}
                precision={0}
                formatter={(value) => formatCurrency(Number(value))}
              />
            </Card>
          </Col>
          <Col xs={12}>
            <Card size="small">
              <Statistic
                title={r.optimal_label || "Optimal Daily Budget"}
                value={r.optimal_budget}
                precision={0}
                prefix={<DollarOutlined />}
                formatter={(value) => formatCurrency(Number(value))}
              />
            </Card>
          </Col>
        </Row>

        {/* Total Budget for Period */}
        {r.total_budget && r.total_budget > 0 && (
          <div
            style={{
              padding: "8px 16px",
              background: token.colorBgLayout,
              borderRadius: token.borderRadius,
              display: "flex",
              justifyContent: "space-between",
              flexWrap: "wrap",
              gap: 8,
            }}
          >
            <Text type="secondary" style={{ fontSize: 12 }}>
              Total Budget ({periodDays} days): <Text strong>{formatCurrency(r.total_budget)}</Text>
            </Text>
            {r.total_projected_revenue && (
              <Text type="secondary" style={{ fontSize: 12 }}>
                Projected Revenue: <Text strong style={{ color: token.colorSuccess }}>{formatCurrency(r.total_projected_revenue)}</Text>
              </Text>
            )}
          </div>
        )}

        {/* Scale-Up Analysis (over-performing products) */}
        {r.max_safe_budget != null && r.max_safe_budget > 0 && (
          <ScaleUpAnalysis
            targetRoas={targetRoas}
            maxSafeBudget={r.max_safe_budget}
            scaleFactor={r.scale_factor || 0}
            optimalLabel={r.optimal_label || "Max Safe Budget"}
            token={token}
            formatCurrency={formatCurrency}
            formatRoas={formatRoas}
          />
        )}

        {/* Funnel Projection */}
        {r.funnel && (
          <FunnelProjection
            funnel={r.funnel}
            token={token}
            formatCurrency={formatCurrency}
          />
        )}

        {/* Calendar Events */}
        {r.calendar_events && r.calendar_events.length > 0 && (
          <CalendarEvents
            events={r.calendar_events}
            multiplier={r.calendar_multiplier || 1.0}
            periodDays={periodDays}
            token={token}
          />
        )}

        {/* Recommendation */}
        <Alert
          message="Recommendation"
          description={r.recommendation}
          type="info"
          showIcon
          icon={<InfoCircleOutlined />}
        />

        {/* Alternatives */}
        {(r.alternatives ?? []).length > 0 && (
          <div>
            <Divider orientation="left">Alternatives</Divider>
            <ul style={{ paddingLeft: 20 }}>
              {(r.alternatives ?? []).map((alt, idx) => (
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
    </Card>
  );
};
