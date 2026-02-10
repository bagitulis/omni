import { useState, useMemo, useEffect } from "react";
import {
  Modal,
  Form,
  Select,
  InputNumber,
  Checkbox,
  Button,
  Table,
  Alert,
  Typography,
  Space,
  Steps,
} from "antd";
import { Product } from "@/types/product";
import { usePriceUpdate } from "@/hooks/usePricing";
import { PriceUpdateItem } from "@/api/pricing";
import { ArrowRightOutlined } from "@ant-design/icons";

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

export function PriceModal({ open, onCancel, selectedProducts }: Props) {
  const [step, setStep] = useState<number>(0);
  const [adjustmentType, setAdjustmentType] = useState<AdjustmentType>(
    "percentage_increase",
  );
  const [value, setValue] = useState<number>(0);
  const [platforms, setPlatforms] = useState<string[]>([]);

  const { mutate: updatePrices, isPending } = usePriceUpdate();

  // Reset state when modal opens
  useEffect(() => {
    if (open) {
      setStep(0);
      setAdjustmentType("percentage_increase");
      setValue(0);
      setPlatforms([]);
    }
  }, [open]);

  const calculateNewPrice = useMemo(() => {
    return (oldPrice: number): number => {
      let newPrice = oldPrice;
      switch (adjustmentType) {
        case "percentage_increase":
          newPrice = oldPrice * (1 + value / 100);
          break;
        case "percentage_decrease":
          newPrice = oldPrice * (1 - value / 100);
          break;
        case "fixed_increase":
          newPrice = oldPrice + value;
          break;
        case "fixed_decrease":
          newPrice = oldPrice - value;
          break;
        case "set_value":
          newPrice = value;
          break;
      }
      return Math.max(0, Math.round(newPrice)); // Ensure no negative and integer
    };
  }, [adjustmentType, value]);

  const previewData = useMemo(() => {
    return selectedProducts.map((product) => {
      const currentPrice = product.price || 0;
      const newPrice = calculateNewPrice(currentPrice);
      return {
        ...product,
        new_price: newPrice,
        price_diff: newPrice - currentPrice,
      };
    });
  }, [selectedProducts, calculateNewPrice]);

  const handleExecute = () => {
    const items: PriceUpdateItem[] = previewData.map((p) => ({
      sku: p.item_sku,
      price: p.new_price,
      platforms: platforms.length > 0 ? platforms : undefined,
    }));

    updatePrices(items, {
      onSuccess: () => {
        onCancel();
      },
    });
  };

  const steps = [
    {
      title: "Configure",
      content: (
        <div className="flex flex-col gap-4 py-4">
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
                options={[
                  {
                    label: "Percentage Increase (%)",
                    value: "percentage_increase",
                  },
                  {
                    label: "Percentage Decrease (%)",
                    value: "percentage_decrease",
                  },
                  {
                    label: "Fixed Amount Increase (+)",
                    value: "fixed_increase",
                  },
                  {
                    label: "Fixed Amount Decrease (-)",
                    value: "fixed_decrease",
                  },
                  { label: "Set To Value (=)", value: "set_value" },
                ]}
              />
            </Form.Item>

            <Form.Item label="Value">
              <InputNumber
                className="w-full"
                value={value}
                onChange={(val) => setValue(val || 0)}
                min={0}
                precision={0} // Integer prices usually
              />
            </Form.Item>

            <Form.Item label="Apply to Platforms (Optional)">
              <Checkbox.Group
                options={[
                  { label: "Shopee", value: "shopee" },
                  { label: "Lazada", value: "lazada" },
                  { label: "TikTok", value: "tiktok" },
                ]}
                value={platforms}
                onChange={(vals) => setPlatforms(vals as string[])}
              />
              <Typography.Text type="secondary" className="block mt-1 text-xs">
                If none selected, prices will be updated in Omni system only
                (and synced later).
              </Typography.Text>
            </Form.Item>
          </Form>
        </div>
      ),
    },
    {
      title: "Preview",
      content: (
        <div className="flex flex-col gap-4 py-4">
          <Alert
            message="Review Changes"
            description="Please review the price changes below before confirming. Red rows indicate potential issues (e.g. zero price)."
            type="warning"
            showIcon
          />

          <Table
            dataSource={previewData}
            rowKey="id"
            pagination={{ pageSize: 5 }}
            size="small"
            scroll={{ y: 300 }}
            columns={[
              {
                title: "Product",
                dataIndex: "item_name",
                width: 200,
                ellipsis: true,
                render: (text, record) => (
                  <Space>
                    <img
                      src={record.image_url || "/placeholder.png"}
                      alt=""
                      className="w-8 h-8 object-cover rounded-sm"
                    />
                    <div className="flex flex-col">
                      <span className="font-medium text-xs">{text}</span>
                      <span className="text-xs text-gray-400">
                        {record.item_sku}
                      </span>
                    </div>
                  </Space>
                ),
              },
              {
                title: "Old Price",
                dataIndex: "price",
                width: 100,
                render: (val) => val?.toLocaleString(),
              },
              {
                title: "",
                width: 30,
                render: () => <ArrowRightOutlined className="text-gray-400" />,
              },
              {
                title: "New Price",
                dataIndex: "new_price",
                width: 100,
                render: (val) => (
                  <span
                    className={
                      val === 0
                        ? "text-red-500 font-bold"
                        : "text-green-600 font-bold"
                    }
                  >
                    {val?.toLocaleString()}
                  </span>
                ),
              },
              {
                title: "Change",
                dataIndex: "price_diff",
                width: 100,
                render: (val) => (
                  <span
                    className={
                      val > 0
                        ? "text-green-600"
                        : val < 0
                          ? "text-red-500"
                          : "text-gray-400"
                    }
                  >
                    {val > 0 ? "+" : ""}
                    {val?.toLocaleString()}
                  </span>
                ),
              },
            ]}
          />
        </div>
      ),
    },
  ];

  return (
    <Modal
      open={open}
      onCancel={onCancel}
      title="Update Prices"
      width={700}
      footer={[
        <Button key="cancel" onClick={onCancel}>
          Cancel
        </Button>,
        step > 0 && (
          <Button key="back" onClick={() => setStep(step - 1)}>
            Back
          </Button>
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
        className="mb-4"
      />
      {steps[step].content}
    </Modal>
  );
}
