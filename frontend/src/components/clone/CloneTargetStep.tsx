import React from "react";
import { Alert, Radio, Space, Badge, Spin, theme } from "antd";
import { ShopOutlined } from "@ant-design/icons";
import { CloneTargetStepProps } from "./types";

export const CloneTargetStep: React.FC<CloneTargetStepProps> = ({
  targetPlatform,
  setTargetPlatform,
  sourcePlatform,
  isLoadingTargets,
  targetsError,
  targetsData,
}) => {
  const { token } = theme.useToken();

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
