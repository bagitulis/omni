import React from "react";
import {
  Space,
  Select,
  Input,
  Button,
  Alert,
  Card,
  Badge,
  Descriptions,
} from "antd";
import { SearchOutlined } from "@ant-design/icons";
import { CloneSourceStepProps } from "./types";
import { formatCurrency } from "./cloneHelpers";

const { Option } = Select;

export const CloneSourceStep: React.FC<CloneSourceStepProps> = ({
  sourcePlatform,
  setSourcePlatform,
  sku,
  setSku,
  onSearch,
  isLoadingProduct,
  productError,
  productData,
}) => {
  return (
    <div style={{ padding: "20px 0" }}>
      <Space direction="vertical" style={{ width: "100%" }} size="large">
        <Space.Compact style={{ width: "100%" }}>
          <Select
            value={sourcePlatform}
            onChange={setSourcePlatform}
            style={{ width: "120px" }}
          >
            <Option value="shopee">Shopee</Option>
            <Option value="lazada">Lazada</Option>
            <Option value="tiktok">TikTok</Option>
          </Select>
          <Input
            placeholder="Enter Product SKU"
            value={sku}
            onChange={(e) => setSku(e.target.value)}
            onPressEnter={onSearch}
          />
          <Button
            type="primary"
            icon={<SearchOutlined />}
            onClick={onSearch}
            loading={isLoadingProduct}
          >
            Search
          </Button>
        </Space.Compact>

        {productError && (
          <Alert
            message="Product not found"
            description={productError.message}
            type="error"
            showIcon
          />
        )}

        {productData && (
          <Card
            size="small"
            title={productData.name}
            extra={<Badge status="success" text="Found" />}
          >
            <Descriptions size="small" column={2}>
              <Descriptions.Item label="Price">
                {formatCurrency(productData.price)}
              </Descriptions.Item>
              <Descriptions.Item label="Stock">
                {productData.stock}
              </Descriptions.Item>
              <Descriptions.Item label="Images">
                {productData.images.length}
              </Descriptions.Item>
            </Descriptions>
          </Card>
        )}
      </Space>
    </div>
  );
};
