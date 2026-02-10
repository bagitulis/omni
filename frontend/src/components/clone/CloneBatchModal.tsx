import React, { useState, useEffect, useMemo } from "react";
import {
  Modal,
  Steps,
  Button,
  Table,
  Radio,
  Form,
  Switch,
  InputNumber,
  Space,
  Select,
  Alert,
  Typography,
  Avatar,
  Tag,
  Card,
} from "antd";
import {
  ShoppingOutlined,
  FormOutlined,
  SyncOutlined,
  WarningOutlined,
} from "@ant-design/icons";
import { useBatchClone } from "@/hooks/useClone";
import { CloneProgress } from "./CloneProgress";
import type { Product } from "@/types/product";
import type {
  BatchCloneRequest,
  BatchCloneResult,
  CloneResult,
} from "@/types/clone";

const { Step } = Steps;
const { Option } = Select;
const { Text } = Typography;

interface CloneBatchModalProps {
  open: boolean;
  onClose: () => void;
  products: Product[];
}

type PriceAdjustmentType =
  | "none"
  | "fixed"
  | "percentage_inc"
  | "percentage_dec"
  | "amount_inc"
  | "amount_dec";

export const CloneBatchModal: React.FC<CloneBatchModalProps> = ({
  open,
  onClose,
  products,
}) => {
  const [currentStep, setCurrentStep] = useState(0);

  // Selection State
  const [selectedRowKeys, setSelectedRowKeys] = useState<React.Key[]>([]);

  // Configuration State
  const [targetPlatform, setTargetPlatform] = useState("");
  const [priceAdjType, setPriceAdjType] = useState<PriceAdjustmentType>("none");
  const [priceAdjValue, setPriceAdjValue] = useState<number | null>(null);
  const [saveAsDraft, setSaveAsDraft] = useState(true);

  // Result State
  const [batchResult, setBatchResult] = useState<BatchCloneResult | null>(null);

  // Reset when opening
  useEffect(() => {
    if (open) {
      setCurrentStep(0);
      setTargetPlatform("");
      setPriceAdjType("none");
      setPriceAdjValue(null);
      setSaveAsDraft(true);
      setBatchResult(null);
      // Pre-select all products
      setSelectedRowKeys(products.map((p) => p.item_id));
    }
  }, [open, products]);

  // Derived Data
  const sourcePlatform = products.length > 0 ? products[0].platform : "";

  const selectedProducts = useMemo(() => {
    return products.filter((p) => selectedRowKeys.includes(p.item_id));
  }, [products, selectedRowKeys]);

  // Mutation
  const { mutate: batchClone, isPending: isCloning } = useBatchClone();

  // Handlers
  const handleNext = () => {
    setCurrentStep(currentStep + 1);
  };

  const handlePrev = () => {
    setCurrentStep(currentStep - 1);
  };

  const handleStartClone = () => {
    if (selectedProducts.length === 0 || !targetPlatform) return;

    let adjType: "fixed" | "percentage" | "amount" | undefined;
    let adjValue: number | undefined;

    switch (priceAdjType) {
      case "fixed":
        adjType = "fixed";
        adjValue = priceAdjValue || 0;
        break;
      case "percentage_inc":
        adjType = "percentage";
        adjValue = priceAdjValue || 0;
        break;
      case "percentage_dec":
        adjType = "percentage";
        adjValue = -(priceAdjValue || 0);
        break;
      case "amount_inc":
        adjType = "amount";
        adjValue = priceAdjValue || 0;
        break;
      case "amount_dec":
        adjType = "amount";
        adjValue = -(priceAdjValue || 0);
        break;
    }

    const request: BatchCloneRequest = {
      source_platform: sourcePlatform,
      target_platform: targetPlatform,
      source_item_ids: selectedProducts.map((p) => p.item_id),
      save_as_draft: saveAsDraft,
      price_adjustment_type: adjType,
      price_adjustment_value: adjValue,
    };

    batchClone(request, {
      onSuccess: (data) => {
        setBatchResult(data);
        setCurrentStep(2); // Move to progress step
      },
    });
  };

  const handleRetry = (failedItems: CloneResult[]) => {
    // Retry logic:
    // We create a new batch request ONLY for the failed items.
    // However, the progress component displays the OLD batch result.
    // If we trigger a new batch, we get a NEW batch result.
    // We should probably just start a new clone process for these items.
    // For simplicity, we'll reset the wizard to Step 0, keeping ONLY the failed items selected.

    // Find products corresponding to failed items
    const failedProductIds = failedItems.map((r) => r.source_item_id);
    setSelectedRowKeys(failedProductIds);
    setBatchResult(null);
    setCurrentStep(0);
  };

  // Step 1: Select Products
  const renderSelectStep = () => {
    const columns = [
      {
        title: "Product",
        dataIndex: "item_name",
        key: "item_name",
        render: (text: string, record: Product) => (
          <Space>
            <Avatar shape="square" src={record.image_url} />
            <Text ellipsis style={{ maxWidth: 300 }}>
              {text}
            </Text>
          </Space>
        ),
      },
      {
        title: "SKU",
        dataIndex: "item_sku",
        key: "item_sku",
      },
      {
        title: "Price",
        dataIndex: "price",
        key: "price",
        render: (val: number) =>
          val
            ? new Intl.NumberFormat("id-ID", {
                style: "currency",
                currency: "IDR",
              }).format(val)
            : "-",
      },
      {
        title: "Platform",
        dataIndex: "platform",
        key: "platform",
        render: (val: string) => <Tag>{val.toUpperCase()}</Tag>,
      },
    ];

    return (
      <div style={{ padding: "20px 0" }}>
        <Alert
          message={`Selected Source: ${sourcePlatform.toUpperCase()}`}
          type="info"
          showIcon
          style={{ marginBottom: 16 }}
        />
        <Table
          rowSelection={{
            selectedRowKeys,
            onChange: (keys) => setSelectedRowKeys(keys),
          }}
          columns={columns}
          dataSource={products}
          rowKey="item_id"
          pagination={{ pageSize: 5 }}
          size="small"
          scroll={{ y: 300 }}
        />
        <div style={{ marginTop: 8, textAlign: "right" }}>
          <Text type="secondary">
            {selectedRowKeys.length} products selected (Max 100)
          </Text>
        </div>
      </div>
    );
  };

  // Step 2: Configure
  const renderConfigureStep = () => {
    const availableTargets = ["shopee", "lazada", "tiktok"].filter(
      (p) => p !== sourcePlatform,
    );

    return (
      <div style={{ padding: "20px 0" }}>
        <Form layout="vertical">
          <Form.Item label="Target Platform" required>
            <Radio.Group
              onChange={(e) => setTargetPlatform(e.target.value)}
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
                onChange={setPriceAdjType}
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
                  onChange={setPriceAdjValue}
                  formatter={(value) =>
                    priceAdjType.startsWith("percentage")
                      ? `${value}%`
                      : `Rp ${value}`.replace(/\B(?=(\d{3})+(?!\d))/g, ".")
                  }
                  parser={(value) =>
                    Number(value!.replace(/Rp\s?|%|(\.*)/g, ""))
                  }
                />
              )}
            </Space>
          </Card>

          <Form.Item label="Status">
            <Space>
              <span>Publish Immediately</span>
              <Switch
                checked={!saveAsDraft}
                onChange={(checked) => setSaveAsDraft(!checked)}
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
          description={`Cloning ${selectedProducts.length} products from ${sourcePlatform} to ${targetPlatform || "..."}`}
          type="warning"
          showIcon
          icon={<WarningOutlined />}
        />
      </div>
    );
  };

  // Step 3: Progress
  const renderProgressStep = () => {
    if (!batchResult) return <div />; // Should not happen

    return (
      <CloneProgress
        batchResult={batchResult}
        products={products}
        targetPlatform={targetPlatform}
        onRetry={handleRetry}
        onDone={onClose}
      />
    );
  };

  // Footer
  const renderFooter = () => {
    if (currentStep === 2) return null; // Progress step has its own actions

    return (
      <div
        style={{
          marginTop: 24,
          display: "flex",
          justifyContent: "flex-end",
          gap: 8,
        }}
      >
        {currentStep > 0 && (
          <Button onClick={handlePrev} disabled={isCloning}>
            Previous
          </Button>
        )}

        {currentStep === 0 && (
          <Button
            type="primary"
            onClick={handleNext}
            disabled={selectedRowKeys.length === 0}
          >
            Next
          </Button>
        )}

        {currentStep === 1 && (
          <Button
            type="primary"
            onClick={handleStartClone}
            loading={isCloning}
            disabled={!targetPlatform}
          >
            Start Batch Clone
          </Button>
        )}
      </div>
    );
  };

  return (
    <Modal
      open={open}
      onCancel={currentStep === 2 ? undefined : onClose} // Prevent closing during progress unless done
      width={800}
      title="Batch Clone Products"
      footer={null}
      maskClosable={false}
      destroyOnClose
    >
      <Steps current={currentStep} size="small">
        <Step title="Select Products" icon={<ShoppingOutlined />} />
        <Step title="Configure" icon={<FormOutlined />} />
        <Step
          title="Processing"
          icon={<SyncOutlined spin={currentStep === 2 && isCloning} />}
        />
      </Steps>

      <div className="steps-content">
        {currentStep === 0 && renderSelectStep()}
        {currentStep === 1 && renderConfigureStep()}
        {currentStep === 2 && renderProgressStep()}
      </div>

      {renderFooter()}
    </Modal>
  );
};
