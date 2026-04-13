import {
  Card,
  Select,
  InputNumber,
  Button,
  Alert,
  Typography,
  GlobalToken,
} from "antd";
import type { ProductFromAds } from "@/api/analyticsIntelligence";

const { Text } = Typography;

interface SimulationParametersProps {
  token: GlobalToken;
  products: ProductFromAds[] | undefined;
  loadingProducts: boolean;
  selectedProductId: string;
  setSelectedProductId: (id: string) => void;
  targetRoas: number;
  setTargetRoas: (roas: number) => void;
  budgetPerDay: number;
  setBudgetPerDay: (budget: number) => void;
  periodDays: number;
  setPeriodDays: (days: number) => void;
  selectedProduct: ProductFromAds | undefined;
  handleRunSimulation: () => void;
  simulating: boolean;
  simulationError: Error | null;
  formatRoas: (value: number) => string;
}

export const SimulationParameters = ({
  token,
  products,
  loadingProducts,
  selectedProductId,
  setSelectedProductId,
  targetRoas,
  setTargetRoas,
  budgetPerDay,
  setBudgetPerDay,
  periodDays,
  setPeriodDays,
  selectedProduct,
  handleRunSimulation,
  simulating,
  simulationError,
  formatRoas,
}: SimulationParametersProps) => {
  return (
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
            value={selectedProductId || undefined}
            onChange={setSelectedProductId}
            showSearch
            notFoundContent={
              loadingProducts
                ? "Loading products..."
                : "No products found. Upload ads data first."
            }
            filterOption={(input, option) =>
              (option?.label as string)
                .toLowerCase()
                .includes(input.toLowerCase())
            }
            options={products?.map((p: ProductFromAds) => ({
              value: p.product_id,
              label: `[${(p.source || "ads").toUpperCase()}] ${p.product_name} (${formatRoas(p.avg_roas)})`,
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
              `${value}`.replace(/\B(?=(\d{3})+(?!\d))/g, ",")
            }
            parser={(value) =>
              Number(value!.replace(/,/g, "")) as number
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
            description={simulationError.message || String(simulationError)}
            showIcon
          />
        )}
      </div>
    </Card>
  );
};
