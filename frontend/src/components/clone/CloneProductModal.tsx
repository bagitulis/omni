import React, { useState, useEffect } from "react";
import { Modal, Steps, Button } from "antd";
import {
  ShoppingOutlined,
  ShopOutlined,
  FormOutlined,
  ArrowRightOutlined,
  CheckCircleOutlined,
} from "@ant-design/icons";
import {
  useProductData,
  useAvailableTargets,
  useClonePreview,
  useCloneProduct,
} from "@/hooks/useClone";
import type { CloneRequest } from "@/types/clone";
import { CloneSourceStep } from "./CloneSourceStep";
import { CloneTargetStep } from "./CloneTargetStep";
import { CloneConfigurationStep } from "./CloneConfigurationStep";
import { CloneResultStep } from "./CloneResultStep";
import { CloneProductModalProps } from "./types";

const { Step } = Steps;

export const CloneProductModal: React.FC<CloneProductModalProps> = ({
  open,
  onClose,
  initialSku = "",
  initialPlatform = "shopee",
}) => {
  const [currentStep, setCurrentStep] = useState(0);

  // Form States
  const [sourcePlatform, setSourcePlatform] = useState(initialPlatform);
  const [sku, setSku] = useState(initialSku);
  const [targetPlatform, setTargetPlatform] = useState("");
  const [newPrice, setNewPrice] = useState<number | undefined>(undefined);
  const [saveAsDraft, setSaveAsDraft] = useState(true);

  // Search Trigger (to avoid fetching on every keystroke)
  const [searchTrigger, setSearchTrigger] = useState<{
    platform: string;
    sku: string;
  } | null>(initialSku ? { platform: initialPlatform, sku: initialSku } : null);

  // Queries
  const {
    data: productData,
    isLoading: isLoadingProduct,
    error: productError,
  } = useProductData(searchTrigger?.platform || "", searchTrigger?.sku || "");

  const {
    data: targetsData,
    isLoading: isLoadingTargets,
    error: targetsError,
  } = useAvailableTargets(productData ? sku : "");

  const {
    data: previewData,
    isLoading: isLoadingPreview,
    error: previewError,
  } = useClonePreview({
    source_platform: sourcePlatform,
    target_platform: targetPlatform,
    source_item_id: productData?.item_id || "",
    sku: sku,
  });

  // Mutation
  const {
    mutate: clone,
    isPending: isCloning,
    isSuccess: isCloneSuccess,
    isError: isCloneError,
    error: cloneError,
    reset: resetClone,
  } = useCloneProduct();

  // Reset when opening
  useEffect(() => {
    if (open) {
      setCurrentStep(0);
      setSku(initialSku);
      setSourcePlatform(initialPlatform);
      setTargetPlatform("");
      setNewPrice(undefined);
      setSaveAsDraft(true);
      resetClone();
      if (initialSku) {
        setSearchTrigger({ platform: initialPlatform, sku: initialSku });
      } else {
        setSearchTrigger(null);
      }
    }
  }, [open, initialSku, initialPlatform, resetClone]);

  // Handlers
  const handleSearch = () => {
    if (sku && sourcePlatform) {
      setSearchTrigger({ platform: sourcePlatform, sku });
    }
  };

  const handleNext = () => {
    setCurrentStep(currentStep + 1);
  };

  const handlePrev = () => {
    setCurrentStep(currentStep - 1);
  };

  const handleClone = () => {
    if (!productData) return;

    const request: CloneRequest = {
      source_platform: sourcePlatform,
      target_platform: targetPlatform,
      source_item_id: productData.item_id,
      sku: sku,
      category_id: productData.category_id,
      update_price: newPrice !== undefined && newPrice !== productData.price,
      new_price: newPrice,
      save_as_draft: saveAsDraft,
      use_inventory: true,
    };

    clone(request);
  };

  // Footer Actions
  const renderFooter = () => {
    if (isCloneSuccess || (currentStep === 3 && isCloning)) return null;

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
            disabled={!productData || isLoadingProduct}
          >
            Next
          </Button>
        )}

        {currentStep === 1 && (
          <Button
            type="primary"
            onClick={handleNext}
            disabled={!targetPlatform}
          >
            Next
          </Button>
        )}

        {currentStep === 2 && (
          <Button
            type="primary"
            onClick={handleNext}
            disabled={isLoadingPreview || !!previewError}
          >
            Next
          </Button>
        )}

        {currentStep === 3 && (
          <Button type="primary" onClick={handleClone} loading={isCloning}>
            Confirm & Clone
          </Button>
        )}
      </div>
    );
  };

  return (
    <Modal
      open={open}
      onCancel={onClose}
      width={700}
      title="Clone Product Wizard"
      footer={null}
      maskClosable={false}
      destroyOnClose
    >
      <Steps current={currentStep} size="small">
        <Step title="Source" icon={<ShoppingOutlined />} />
        <Step title="Target" icon={<ShopOutlined />} />
        <Step title="Configure" icon={<FormOutlined />} />
        <Step
          title="Confirm"
          icon={
            isCloneSuccess ? <CheckCircleOutlined /> : <ArrowRightOutlined />
          }
        />
      </Steps>

      <div className="steps-content">
        {currentStep === 0 && (
          <CloneSourceStep
            sourcePlatform={sourcePlatform}
            setSourcePlatform={setSourcePlatform}
            sku={sku}
            setSku={setSku}
            onSearch={handleSearch}
            isLoadingProduct={isLoadingProduct}
            productError={productError}
            productData={productData}
          />
        )}
        {currentStep === 1 && (
          <CloneTargetStep
            targetPlatform={targetPlatform}
            setTargetPlatform={setTargetPlatform}
            sourcePlatform={sourcePlatform}
            isLoadingTargets={isLoadingTargets}
            targetsError={targetsError}
            targetsData={targetsData}
          />
        )}
        {currentStep === 2 && (
          <CloneConfigurationStep
            previewError={previewError}
            productData={productData}
            previewData={previewData}
            targetPlatform={targetPlatform}
            isLoadingPreview={isLoadingPreview}
            setNewPrice={setNewPrice}
            saveAsDraft={saveAsDraft}
            setSaveAsDraft={setSaveAsDraft}
          />
        )}
        {currentStep === 3 && (
          <CloneResultStep
            isCloning={isCloning}
            isCloneSuccess={isCloneSuccess}
            isCloneError={isCloneError}
            cloneError={cloneError}
            targetPlatform={targetPlatform}
            sourcePlatform={sourcePlatform}
            sku={sku}
            newPrice={newPrice}
            productData={productData}
            saveAsDraft={saveAsDraft}
            onClose={onClose}
            onReset={() => {
              setCurrentStep(0);
              resetClone();
            }}
            onRetry={handleClone}
            onBack={handlePrev}
          />
        )}
      </div>

      {renderFooter()}
    </Modal>
  );
};
