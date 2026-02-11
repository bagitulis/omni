import React from "react";
import { Alert, Card, Form, InputNumber, Space, Switch } from "antd";
import { CloneConfigurationStepProps } from "./types";
import { ClonePreview } from "./ClonePreview";
import { currencyFormatter, currencyParser } from "./cloneHelpers";

export const CloneConfigurationStep: React.FC<CloneConfigurationStepProps> = ({
  previewError,
  productData,
  previewData,
  targetPlatform,
  isLoadingPreview,
  setNewPrice,
  saveAsDraft,
  setSaveAsDraft,
}) => {
  return (
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
                formatter={currencyFormatter}
                parser={currencyParser}
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
};
