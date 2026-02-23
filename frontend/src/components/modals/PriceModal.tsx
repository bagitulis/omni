import { useState, useMemo, useEffect } from "react";
import {
  Modal,
  Form,
  Select,
  InputNumber,
  Checkbox,
  Button,
  Alert,
  Typography,
  Steps,
  theme,
  Flex,
} from "antd";
import { Product } from "@/types/product";
import { usePriceUpdate } from "@/hooks/usePricing";
import { PriceUpdateItem } from "@/api/pricing";
import { PricePreviewTable } from "./PricePreviewTable";

interface Props {
  open: boolean;
  onCancel: () => void;
  selectedProducts: Product[];
}

type AdjustmentType =
  | "percentage_increase"
  | "percentage_decrease"
  | "fixed_increase"
  | "fixed_decrease"
  | "set_value";

const ADJUSTMENT_OPTIONS = [
  { label: "Percentage Increase (%)", value: "percentage_increase" },
  { label: "Percentage Decrease (%)", value: "percentage_decrease" },
  { label: "Fixed Amount Increase (+)", value: "fixed_increase" },
  { label: "Fixed Amount Decrease (-)", value: "fixed_decrease" },
  { label: "Set To Value (=)", value: "set_value" },
] as const;

const PLATFORM_OPTIONS = [
  { label: "Shopee", value: "shopee" },
  { label: "Lazada", value: "lazada" },
  { label: "TikTok", value: "tiktok" },
] as const;

function applyAdjustment(
  oldPrice: number,
  type: AdjustmentType,
  amount: number,
): number {
  let result: number;
  switch (type) {
    case "percentage_increase":
      result = oldPrice * (1 + amount / 100);
      break;
    case "percentage_decrease":
      result = oldPrice * (1 - amount / 100);
      break;
    case "fixed_increase":
      result = oldPrice + amount;
      break;
    case "fixed_decrease":
      result = oldPrice - amount;
      break;
    case "set_value":
      result = amount;
      break;
  }
  return Math.max(0, Math.round(result));
}

export function PriceModal({ open, onCancel, selectedProducts }: Props) {
  const { token } = theme.useToken();
  const [step, setStep] = useState(0);
  const [adjustmentType, setAdjustmentType] =
    useState<AdjustmentType>("percentage_increase");
  const [value, setValue] = useState(0);
  const [platforms, setPlatforms] = useState<string[]>([]);

  const { mutate: updatePrices, isPending } = usePriceUpdate();

  useEffect(() => {
    if (open) {
      setStep(0);
      setAdjustmentType("percentage_increase");
      setValue(0);
      setPlatforms([]);
    }
  }, [open]);

  const previewData = useMemo(
    () =>
      selectedProducts.map((product) => {
        const currentPrice = product.price || 0;
        const newPrice = applyAdjustment(currentPrice, adjustmentType, value);
        return { ...product, new_price: newPrice, price_diff: newPrice - currentPrice };
      }),
    [selectedProducts, adjustmentType, value],
  );

  const handleExecute = () => {
    const items: PriceUpdateItem[] = previewData.map((p) => ({
      sku: p.item_sku,
      price: p.new_price,
      platforms: platforms.length > 0 ? platforms : undefined,
    }));
    updatePrices(items, { onSuccess: onCancel });
  };

  const configStep = (
    <Flex vertical gap={16} style={{ paddingBlock: 16 }}>
      <Alert
        message="Bulk Price Update"
        description={`You are about to update prices for ${selectedProducts.length} selected products.`}
        type="info"
        showIcon
      />
      <Form layout="vertical">
        <Form.Item label="Adjustment Type">
          <Select
            value={adjustmentType}
            onChange={setAdjustmentType}
            options={[...ADJUSTMENT_OPTIONS]}
          />
        </Form.Item>
        <Form.Item label="Value">
          <InputNumber
            style={{ width: "100%" }}
            value={value}
            onChange={(val) => setValue(val || 0)}
            min={0}
            precision={0}
          />
        </Form.Item>
        <Form.Item label="Apply to Platforms (Optional)">
          <div style={{ marginBottom: 6 }}>
            <Checkbox
              checked={platforms.length === PLATFORM_OPTIONS.length}
              indeterminate={platforms.length > 0 && platforms.length < PLATFORM_OPTIONS.length}
              onChange={(e) =>
                setPlatforms(
                  e.target.checked
                    ? PLATFORM_OPTIONS.map((p) => p.value)
                    : [],
                )
              }
            >
              Select All
            </Checkbox>
          </div>
          <Checkbox.Group
            options={[...PLATFORM_OPTIONS]}
            value={platforms}
            onChange={(vals) => setPlatforms(vals as string[])}
          />
          <Typography.Text
            type="secondary"
            style={{ display: "block", marginTop: 4, fontSize: token.fontSizeSM }}
          >
            If none selected, prices will be updated in Omni system only (and
            synced later).
          </Typography.Text>
        </Form.Item>
      </Form>
    </Flex>
  );

  const steps = [
    { title: "Configure", content: configStep },
    { title: "Preview", content: <PricePreviewTable data={previewData} /> },
  ];

  return (
    <Modal
      open={open}
      onCancel={onCancel}
      destroyOnHidden
      title="Update Prices"
      width={700}
      footer={[
        <Button key="cancel" onClick={onCancel}>Cancel</Button>,
        step > 0 && (
          <Button key="back" onClick={() => setStep(step - 1)}>Back</Button>
        ),
        step < steps.length - 1 ? (
          <Button key="next" type="primary" onClick={() => setStep(step + 1)}>
            Next
          </Button>
        ) : (
          <Button
            key="submit"
            type="primary"
            danger
            loading={isPending}
            onClick={handleExecute}
          >
            Confirm Update
          </Button>
        ),
      ]}
    >
      <Steps
        current={step}
        items={steps.map((s) => ({ title: s.title }))}
        size="small"
        style={{ marginBottom: 16 }}
      />
      {steps[step].content}
    </Modal>
  );
}
