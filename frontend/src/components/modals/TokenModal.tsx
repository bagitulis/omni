import React from "react";
import { Modal, Button, Alert, Space } from "antd";
import { ReloadOutlined, KeyOutlined, FormOutlined } from "@ant-design/icons";

interface TokenModalProps {
  open: boolean;
  onClose: () => void;
  platform: string | null;
  onTokenOperation: (operation: string) => void;
}

export const TokenModal: React.FC<TokenModalProps> = ({
  open,
  onClose,
  platform,
  onTokenOperation,
}) => {
  const displayPlatform = platform
    ? platform.charAt(0).toUpperCase() + platform.slice(1)
    : "";

  return (
    <Modal
      title={`${displayPlatform} Token Management`}
      open={open}
      onCancel={onClose}
      footer={[
        <Button key="close" onClick={onClose}>
          Close
        </Button>,
      ]}
      width={400}
    >
      <Space direction="vertical" size="middle" style={{ width: "100%" }}>
        <Alert
          message={`Manage your ${platform || "platform"} API tokens and authorization codes.`}
          type="info"
          showIcon
        />

        <div style={{ display: "flex", flexDirection: "column", gap: "12px" }}>
          <Button
            block
            icon={<FormOutlined />}
            onClick={() => onTokenOperation("update_code")}
          >
            Update Authorization Code
          </Button>

          <Button
            block
            icon={<KeyOutlined />}
            onClick={() => onTokenOperation("get_token")}
          >
            Get Access Token
          </Button>

          <Button
            block
            icon={<ReloadOutlined />}
            onClick={() => onTokenOperation("refresh_token")}
          >
            Refresh Token
          </Button>
        </div>
      </Space>
    </Modal>
  );
};

export default TokenModal;
