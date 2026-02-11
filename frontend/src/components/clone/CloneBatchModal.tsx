import React, { useState, useEffect, useMemo } from "react";
import { Modal, Steps, Button } from "antd";
import {
  ShoppingOutlined,
  FormOutlined,
  SyncOutlined,
} from "@ant-design/icons";
import { useBatchClone } from "@/hooks/useClone";
import { CloneProgress } from "./CloneProgress";
import { BatchProductSelector } from "./BatchProductSelector";
import { BatchConfiguration } from "./BatchConfiguration";
import type { Product } from "@/types/product";
import type {
  BatchCloneRequest,
  BatchCloneResult,
  CloneResult,
} from "@/types/clone";
import type { PriceAdjustmentType } from "./types";

const { Step } = Steps;

interface CloneBatchModalProps {
  open: boolean;
  onClose: () => void;
  products: Product[];
}

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
    const failedProductIds = failedItems.map((r) => r.source_item_id);
    setSelectedRowKeys(failedProductIds);
    setBatchResult(null);
    setCurrentStep(0);
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
      onCancel={currentStep === 2 ? undefined : onClose}
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
        {currentStep === 0 && (
          <BatchProductSelector
            products={products}
            selectedRowKeys={selectedRowKeys}
            onSelectionChange={setSelectedRowKeys}
            sourcePlatform={sourcePlatform}
          />
        )}
        {currentStep === 1 && (
          <BatchConfiguration
            sourcePlatform={sourcePlatform}
            targetPlatform={targetPlatform}
            onTargetPlatformChange={setTargetPlatform}
            priceAdjType={priceAdjType}
            onPriceAdjTypeChange={setPriceAdjType}
            priceAdjValue={priceAdjValue}
            onPriceAdjValueChange={setPriceAdjValue}
            saveAsDraft={saveAsDraft}
            onSaveAsDraftChange={setSaveAsDraft}
            selectedCount={selectedProducts.length}
          />
        )}
        {currentStep === 2 && batchResult && (
          <CloneProgress
            batchResult={batchResult}
            products={products}
            targetPlatform={targetPlatform}
            onRetry={handleRetry}
            onDone={onClose}
          />
        )}
      </div>

      {renderFooter()}
    </Modal>
  );
};
