import {
  Card,
  Typography,
  Select,
  InputNumber,
  Button,
  Row,
  Col,
  Alert,
  Statistic,
  theme,
  Divider,
} from "antd";
import {
  RocketOutlined,
  LineChartOutlined,
  DollarOutlined,
  InfoCircleOutlined,
} from "@ant-design/icons";
import { useState, useMemo } from "react";
import {
  useProductsFromAds,
  useBudgetSimulation,
} from "../../hooks/useAnalyticsIntelligence";
import type { ProductFromAds } from "../../api/analyticsIntelligence";

const { Title, Text } = Typography;
const { useToken } = theme;

export const BudgetSimulatorPage = () => {
  const { token } = useToken();
  const [selectedProductId, setSelectedProductId] = useState<string>("");
  const [targetRoas, setTargetRoas] = useState<number>(5.0);
  const [budgetPerDay, setBudgetPerDay] = useState<number>(100000);
  const [periodDays, setPeriodDays] = useState<number>(7);

  const { data: products, isLoading: loadingProducts } = useProductsFromAds();
  const {
    mutate: runSimulation,
    isPending: simulating,
    data: simulationResult,
    error: simulationError,
  } = useBudgetSimulation();

  const selectedProduct = useMemo(
    () => products?.find((p) => p.product_id === selectedProductId),
    [products, selectedProductId],
  );

  const handleRunSimulation = () => {
    if (!selectedProductId) return;

    runSimulation({
      product_id: selectedProductId,
      target_roas: targetRoas,
      budget_per_day: budgetPerDay,
      period_days: periodDays,
    });
  };

  const formatCurrency = (value: number) =>
    new Intl.NumberFormat("id-ID", {
      style: "currency",
      currency: "IDR",
      maximumFractionDigits: 0,
    }).format(value);

  const formatRoas = (value: number) => `${value.toFixed(2)}x`;

  const getFeasibilityColor = (feasibility: string) => {
    switch (feasibility) {
      case "ACHIEVABLE":
        return token.colorSuccess;
      case "DIFFICULT":
        return token.colorWarning;
      case "NOT_ACHIEVABLE":
        return token.colorError;
      default:
        return token.colorTextSecondary;
    }
  };

  return (
    <div style={{ padding: 24 }}>
      <div style={{ marginBottom: 24 }}>
        <Title level={2} style={{ margin: 0 }}>
          <RocketOutlined style={{ marginRight: 8 }} />
          Budget Simulator
        </Title>
        <Text type="secondary">
          Predict ROAS and optimize your ad spend with AI-driven projections
        </Text>
      </div>

      <Row gutter={[24, 24]}>
        {/* Input Section */}
        <Col xs={24} lg={8}>
          <Card
            title="Simulation Parameters"
            style={{ borderRadius: token.borderRadiusLG }}
          >
            <div style={{ display: "flex", flexDirection: "column", gap: 16 }}>
              <div>
                <Text strong>Select Product</Text>
                <Select
                  style={{ width: "100%", marginTop: 8 }}
                  placeholder="Select a product"
                  loading={loadingProducts}
                  value={selectedProductId}
                  onChange={setSelectedProductId}
                  showSearch
                  filterOption={(input, option) =>
                    (option?.label as string)
                      .toLowerCase()
                      .includes(input.toLowerCase())
                  }
                  options={products?.map((p: ProductFromAds) => ({
                    value: p.product_id,
                    label: `${p.product_name} (${formatRoas(p.avg_roas)})`,
                  }))}
                />
              </div>

              <div>
                <Text strong>Target ROAS</Text>
                <InputNumber
                  style={{ width: "100%", marginTop: 8 }}
                  min={0.1}
                  max={50}
                  step={0.1}
                  value={targetRoas}
                  onChange={(v) => setTargetRoas(v || 5.0)}
                  addonAfter="x"
                />
                {selectedProduct && (
                  <Text type="secondary" style={{ fontSize: 12 }}>
                    Current ROAS: {formatRoas(selectedProduct.avg_roas)}
                  </Text>
                )}
              </div>

              <div>
                <Text strong>Budget per Day</Text>
                <InputNumber
                  style={{ width: "100%", marginTop: 8 }}
                  min={10000}
                  step={10000}
                  value={budgetPerDay}
                  onChange={(v) => setBudgetPerDay(v || 100000)}
                  formatter={(value) =>
                    `Rp ${value}`.replace(/\B(?=(\d{3})+(?!\d))/g, ",")
                  }
                  parser={(value) =>
                    Number(value!.replace(/Rp\s?|(,*)/g, "")) as number
                  }
                  addonBefore="Rp"
                />
              </div>

              <div>
                <Text strong>Period</Text>
                <Select
                  style={{ width: "100%", marginTop: 8 }}
                  value={periodDays}
                  onChange={setPeriodDays}
                  options={[
                    { value: 7, label: "7 days" },
                    { value: 14, label: "14 days" },
                    { value: 30, label: "30 days" },
                  ]}
                />
              </div>

              <Button
                type="primary"
                size="large"
                onClick={handleRunSimulation}
                loading={simulating}
                disabled={!selectedProductId}
                block
                style={{ marginTop: 16 }}
              >
                Calculate Projection
              </Button>

              {simulationError && (
                <Alert
                  type="error"
                  message="Simulation failed"
                  description={(simulationError as any).message}
                  showIcon
                />
              )}
            </div>
          </Card>
        </Col>

        {/* Results Section */}
        <Col xs={24} lg={16}>
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
              <div
                style={{ display: "flex", flexDirection: "column", gap: 24 }}
              >
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
                  <Text
                    type="secondary"
                    style={{ marginTop: 8, display: "block" }}
                  >
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

                {simulationResult.alternatives &&
                  simulationResult.alternatives.length > 0 && (
                    <div>
                      <Divider orientation="left">Alternatives</Divider>
                      <ul style={{ paddingLeft: 20 }}>
                        {simulationResult.alternatives.map((alt, idx) => (
                          <li key={idx} style={{ marginBottom: 8 }}>
                            <Text>
                              {alt.target_roas && alt.required_budget ? (
                                <>
                                  To reach ROAS{" "}
                                  <Text strong>
                                    {formatRoas(alt.target_roas)}
                                  </Text>
                                  , you need{" "}
                                  <Text strong>
                                    {formatCurrency(alt.required_budget)}/day
                                  </Text>
                                </>
                              ) : alt.budget && alt.expected_roas ? (
                                <>
                                  Budget{" "}
                                  <Text strong>
                                    {formatCurrency(alt.budget)}/day
                                  </Text>{" "}
                                  yields{" "}
                                  <Text strong>
                                    {formatRoas(alt.expected_roas)}
                                  </Text>
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
        </Col>
      </Row>
    </div>
  );
};

export default BudgetSimulatorPage;
