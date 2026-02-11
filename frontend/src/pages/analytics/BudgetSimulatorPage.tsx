import { RocketOutlined } from "@ant-design/icons";
import { Col, Row, Typography, theme } from "antd";
import { useMemo, useState } from "react";
import {
  useBudgetSimulation,
  useProductsFromAds,
} from "../../hooks/useAnalyticsIntelligence";
import { SimulationParameters } from "./components/budget/SimulationParameters";
import { SimulationResults } from "./components/budget/SimulationResults";

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
        <Col xs={24} lg={8}>
          <SimulationParameters
            token={token}
            products={products}
            loadingProducts={loadingProducts}
            selectedProductId={selectedProductId}
            setSelectedProductId={setSelectedProductId}
            targetRoas={targetRoas}
            setTargetRoas={setTargetRoas}
            budgetPerDay={budgetPerDay}
            setBudgetPerDay={setBudgetPerDay}
            periodDays={periodDays}
            setPeriodDays={setPeriodDays}
            selectedProduct={selectedProduct}
            handleRunSimulation={handleRunSimulation}
            simulating={simulating}
            simulationError={simulationError}
            formatRoas={formatRoas}
          />
        </Col>

        <Col xs={24} lg={16}>
          <SimulationResults
            simulationResult={simulationResult}
            token={token}
            formatCurrency={formatCurrency}
            formatRoas={formatRoas}
            getFeasibilityColor={getFeasibilityColor}
          />
        </Col>
      </Row>
    </div>
  );
};

export default BudgetSimulatorPage;
