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
  theme,
  Flex,
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
  const { token } = theme.useToken();
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
                style={{ width: "100%" }}
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
              <Typography.Text
                type="secondary"
                style={{
                  display: "block",
                  marginTop: 4,
                  fontSize: token.fontSizeSM,
                }}
              >
                If none selected, prices will be updated in Omni system only
                (and synced later).
              </Typography.Text>
            </Form.Item>
          </Form>
        </Flex>
      ),
    },
    {
      title: "Preview",
      content: (
        <Flex vertical gap={16} style={{ paddingBlock: 16 }}>
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
                      style={{
                        width: 32,
                        height: 32,
                        objectFit: "cover",
                        borderRadius: 2,
                      }}
                    />
                    <Flex vertical>
                      <span
                        style={{ fontWeight: 500, fontSize: token.fontSizeSM }}
                      >
                        {text}
                      </span>
                      <span
                        style={{
                          fontSize: token.fontSizeSM,
                          color: token.colorTextDescription,
                        }}
                      >
                        {record.item_sku}
                      </span>
                    </Flex>
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
                render: () => (
                  <ArrowRightOutlined
                    style={{ color: token.colorTextDescription }}
                  />
                ),
              },
              {
                title: "New Price",
                dataIndex: "new_price",
                width: 100,
                render: (val) => (
                  <span
                    style={{
                      fontWeight: "bold",
                      color: val === 0 ? token.colorError : token.colorSuccess,
                    }}
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
                    style={{
                      color:
                        val > 0
                          ? token.colorSuccess
                          : val < 0
                            ? token.colorError
                            : token.colorTextDescription,
                    }}
                  >
                    {val > 0 ? "+" : ""}
                    {val?.toLocaleString()}
                  </span>
                ),
              },
            ]}
          />
        </Flex>
      ),
    },
  ];

  return (
    <Modal
      open={open}
      onCancel={onCancel}
      destroyOnHidden
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
        style={{ marginBottom: 16 }}
      />
      {steps[step].content}
    </Modal>
  );
}
