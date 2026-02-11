import React from "react";
import { Alert, Button, Descriptions, Result, Spin } from "antd";
import { CloneResultStepProps } from "./types";
import { formatCurrency } from "./cloneHelpers";

export const CloneResultStep: React.FC<CloneResultStepProps> = ({
  isCloning,
  isCloneSuccess,
  isCloneError,
  cloneError,
  targetPlatform,
  sourcePlatform,
  sku,
  newPrice,
  productData,
  saveAsDraft,
  onClose,
  onReset,
  onRetry,
  onBack,
}) => {
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
              onReset();
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
          <Button type="primary" key="retry" onClick={onRetry}>
            Retry
          </Button>,
          <Button key="back" onClick={onBack}>
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
          {formatCurrency(newPrice ?? productData?.price ?? 0)}
        </Descriptions.Item>
        <Descriptions.Item label="Status">
          {saveAsDraft ? "Draft" : "Published"}
        </Descriptions.Item>
      </Descriptions>
    </div>
  );
};
