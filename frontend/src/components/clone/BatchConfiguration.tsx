import React from "react";
import {
  Form,
  Radio,
  Card,
  Select,
  InputNumber,
  Space,
  Switch,
  Alert,
  Typography,
} from "antd";
import { WarningOutlined } from "@ant-design/icons";
import type { PriceAdjustmentType } from "./types";

const { Option } = Select;
const { Text } = Typography;

interface BatchConfigurationProps {
  sourcePlatform: string;
  targetPlatform: string;
  onTargetPlatformChange: (val: string) => void;
  priceAdjType: PriceAdjustmentType;
  onPriceAdjTypeChange: (val: PriceAdjustmentType) => void;
  priceAdjValue: number | null;
  onPriceAdjValueChange: (val: number | null) => void;
  saveAsDraft: boolean;
  onSaveAsDraftChange: (val: boolean) => void;
  selectedCount: number;
}

export const BatchConfiguration: React.FC<BatchConfigurationProps> = ({
  sourcePlatform,
  targetPlatform,
  onTargetPlatformChange,
  priceAdjType,
  onPriceAdjTypeChange,
  priceAdjValue,
  onPriceAdjValueChange,
  saveAsDraft,
  onSaveAsDraftChange,
  selectedCount,
}) => {
  const availableTargets = ["shopee", "lazada", "tiktok"].filter(
    (p) => p !== sourcePlatform,
  );

  return (
    <div style={{ padding: "20px 0" }}>
      <Form layout="vertical">
        <Form.Item label="Target Platform" required>
          <Radio.Group
            onChange={(e) => onTargetPlatformChange(e.target.value)}
            value={targetPlatform}
            optionType="button"
            buttonStyle="solid"
          >
            {availableTargets.map((p) => (
              <Radio.Button key={p} value={p}>
                {p.charAt(0).toUpperCase() + p.slice(1)}
              </Radio.Button>
            ))}
          </Radio.Group>
        </Form.Item>

        <Card
          size="small"
          title="Price Adjustment"
          style={{ marginBottom: 16 }}
        >
          <Space direction="vertical" style={{ width: "100%" }}>
            <Select
              value={priceAdjType}
              onChange={onPriceAdjTypeChange}
              style={{ width: "100%" }}
            >
              <Option value="none">No Change (Keep Original)</Option>
              <Option value="fixed">Set to Fixed Price</Option>
              <Option value="percentage_inc">Increase by Percentage</Option>
              <Option value="percentage_dec">Decrease by Percentage</Option>
              <Option value="amount_inc">Increase by Amount</Option>
              <Option value="amount_dec">Decrease by Amount</Option>
            </Select>

            {priceAdjType !== "none" && (
              <InputNumber
                style={{ width: "100%" }}
                placeholder="Enter value"
                value={priceAdjValue}
                onChange={onPriceAdjValueChange}
                formatter={(value) =>
                  priceAdjType.startsWith("percentage")
                    ? `${value}%`
                    : `Rp ${value}`.replace(/\B(?=(\d{3})+(?!\d))/g, ".")
                }
                parser={(value) => {
                  const val = value?.replace(/Rp\s?|%|(\.*)/g, "");
                  return val ? Number(val) : 0;
                }}
              />
            )}
          </Space>
        </Card>

        <Form.Item label="Status">
          <Space>
            <span>Publish Immediately</span>
            <Switch
              checked={!saveAsDraft}
              onChange={(checked) => onSaveAsDraftChange(!checked)}
            />
          </Space>
          <div style={{ marginTop: 4 }}>
            <Text type="secondary" style={{ fontSize: 12 }}>
              {saveAsDraft
                ? "Products will be saved as Draft."
                : "Products will be live immediately."}
            </Text>
          </div>
        </Form.Item>
      </Form>

      <Alert
        message="Review Selection"
        description={`Cloning ${selectedCount} products from ${sourcePlatform} to ${targetPlatform || "..."}`}
        type="warning"
        showIcon
        icon={<WarningOutlined />}
      />
    </div>
  );
};
