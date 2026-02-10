import React, { useState, useEffect } from "react";
import {
  Modal,
  Steps,
  Button,
  Input,
  Select,
  Radio,
  Form,
  Switch,
  InputNumber,
  Space,
  Card,
  Descriptions,
  Alert,
  Result,
  Badge,
  Spin,
  theme,
} from "antd";
import {
  SearchOutlined,
  ShoppingOutlined,
  ShopOutlined,
  ArrowRightOutlined,
  CheckCircleOutlined,
  FormOutlined,
} from "@ant-design/icons";
import {
  useProductData,
  useAvailableTargets,
  useClonePreview,
  useCloneProduct,
} from "@/hooks/useClone";
import { ClonePreview } from "./ClonePreview";
import type { CloneRequest } from "@/types/clone";

interface CloneProductModalProps {
  open: boolean;
  onClose: () => void;
  initialSku?: string;
  initialPlatform?: string;
}

const { Step } = Steps;
const { Option } = Select;

export const CloneProductModal: React.FC<CloneProductModalProps> = ({
  open,
  onClose,
  initialSku = "",
  initialPlatform = "shopee",
}) => {
  const { token } = theme.useToken();
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

  // Step Content Renderers
  const renderSourceStep = () => (
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
            onPressEnter={handleSearch}
          />
          <Button
            type="primary"
            icon={<SearchOutlined />}
            onClick={handleSearch}
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
                {new Intl.NumberFormat("id-ID", {
                  style: "currency",
                  currency: "IDR",
                }).format(productData.price)}
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

  const renderTargetStep = () => {
    if (isLoadingTargets)
      return <Spin style={{ display: "block", margin: "20px auto" }} />;
    if (targetsError)
      return <Alert message="Error loading targets" type="error" showIcon />;

    const availableTargets = ["shopee", "lazada", "tiktok"].filter(
      (p) => p !== sourcePlatform,
    );

    return (
      <div style={{ padding: "20px 0" }}>
        <Alert
          message="Select Target Platform"
          description="Choose where you want to copy this product to."
          type="info"
          style={{ marginBottom: 16 }}
        />
        <Radio.Group
          onChange={(e) => setTargetPlatform(e.target.value)}
          value={targetPlatform}
          style={{ width: "100%" }}
        >
          <Space direction="vertical" style={{ width: "100%" }}>
            {availableTargets.map((platform) => {
              const isExists =
                targetsData?.status?.[
                  platform as keyof typeof targetsData.status
                ];
              return (
                <Radio
                  value={platform}
                  disabled={isExists}
                  key={platform}
                  style={{
                    width: "100%",
                    padding: "12px",
                    border: `1px solid ${token.colorBorder}`,
                    borderRadius: token.borderRadius,
                    marginBottom: 8,
                  }}
                >
                  <Space>
                    <ShopOutlined />
                    <span style={{ textTransform: "capitalize" }}>
                      {platform}
                    </span>
                    {isExists && <Badge status="error" text="Already exists" />}
                  </Space>
                </Radio>
              );
            })}
          </Space>
        </Radio.Group>
      </div>
    );
  };

  const renderPreviewStep = () => (
    <div style={{ padding: "20px 0" }}>
      {previewError && (
        <Alert
          message="Preview Failed"
          description={previewError.message}
          type="error"
          showIcon
          style={{ marginBottom: 16 }}
        />
      )}

      <ClonePreview
        sourceProduct={productData}
        conflictResult={previewData}
        targetPlatform={targetPlatform}
        loading={isLoadingPreview}
      />

      {!isLoadingPreview && !previewError && (
        <Card size="small" title="Configuration" style={{ marginTop: 16 }}>
          <Form layout="vertical">
            <Form.Item label="New Price (Optional)">
              <InputNumber
                style={{ width: "100%" }}
                defaultValue={productData?.price}
                onChange={(val) => setNewPrice(val === null ? undefined : val)}
                formatter={(value) =>
                  `Rp ${value}`.replace(/\B(?=(\d{3})+(?!\d))/g, ".")
                }
                parser={(value) => Number(value!.replace(/Rp\s?|(\.*)/g, ""))}
              />
            </Form.Item>
            <Form.Item label="Status">
              <Space>
                <span>Publish Immediately</span>
                <Switch
                  checked={!saveAsDraft}
                  onChange={(checked) => setSaveAsDraft(!checked)}
                />
              </Space>
            </Form.Item>
          </Form>
        </Card>
      )}
    </div>
  );

  const renderResultStep = () => {
    if (isCloning) {
      return (
        <Result
          icon={<Spin size="large" />}
          title="Cloning Product..."
          subTitle="Please wait while we copy your product to the target platform."
        />
      );
    }
    if (isCloneSuccess) {
      return (
        <Result
          status="success"
          title="Product Cloned Successfully!"
          subTitle={`Product has been copied to ${targetPlatform}.`}
          extra={[
            <Button type="primary" key="close" onClick={onClose}>
              Done
            </Button>,
            <Button
              key="clone-another"
              onClick={() => {
                setCurrentStep(0);
                resetClone();
              }}
            >
              Clone Another
            </Button>,
          ]}
        />
      );
    }
    if (isCloneError) {
      return (
        <Result
          status="error"
          title="Clone Failed"
          subTitle={cloneError?.message}
          extra={[
            <Button type="primary" key="retry" onClick={handleClone}>
              Retry
            </Button>,
            <Button key="back" onClick={handlePrev}>
              Back to Configuration
            </Button>,
          ]}
        />
      );
    }
    // Default confirmation state before clicking clone
    return (
      <div style={{ padding: "20px 0", textAlign: "center" }}>
        <Alert
          message="Ready to Clone"
          description="Please review your settings before proceeding."
          type="info"
          showIcon
          style={{ marginBottom: 24, textAlign: "left" }}
        />
        <Descriptions
          bordered
          column={1}
          size="small"
          style={{ marginBottom: 24, textAlign: "left" }}
        >
          <Descriptions.Item label="Source">
            {sourcePlatform} ({sku})
          </Descriptions.Item>
          <Descriptions.Item label="Target">{targetPlatform}</Descriptions.Item>
          <Descriptions.Item label="Price">
            {new Intl.NumberFormat("id-ID", {
              style: "currency",
              currency: "IDR",
            }).format(newPrice ?? productData?.price ?? 0)}
          </Descriptions.Item>
          <Descriptions.Item label="Status">
            {saveAsDraft ? "Draft" : "Published"}
          </Descriptions.Item>
        </Descriptions>
      </div>
    );
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
        {currentStep === 0 && renderSourceStep()}
        {currentStep === 1 && renderTargetStep()}
        {currentStep === 2 && renderPreviewStep()}
        {currentStep === 3 && renderResultStep()}
      </div>

      {renderFooter()}
    </Modal>
  );
};
