import { HistoryOutlined, LockOutlined } from "@ant-design/icons";
import { Button, Card, Col, Descriptions, Row, Space, Tag } from "antd";
import type { CredentialPlatformSummary } from "@/api/credentials";
import { useAuthStore } from "@/stores/authStore";
import {
  CREDENTIAL_STATUS_COLOR,
  CREDENTIAL_STATUS_LABEL,
  formatRegion,
} from "./credentialStatus";

interface CredentialAppCredentialsSectionProps {
  platforms: CredentialPlatformSummary[];
  platformNames: Record<string, string>;
  onManualToken: (platform: CredentialPlatformSummary) => void;
  onViewHistory: (platform: CredentialPlatformSummary) => void;
}

export function CredentialAppCredentialsSection({
  platforms,
  platformNames,
  onManualToken,
  onViewHistory,
}: CredentialAppCredentialsSectionProps) {
  const role = useAuthStore((state) => state.user?.role);
  const canManageManualTokens = role === "developer" || role === "admin";

  return (
    <Card title="App Credentials" size="small">
      <Row gutter={[16, 16]}>
        {platforms.map((platform) => (
          <Col xs={24} md={12} key={`app-${platform.platform}`}>
            <Card
              size="small"
              title={`${platformNames[platform.platform] || platform.platform} App`}
            >
              <Space direction="vertical" size={8} style={{ width: "100%" }}>
                <Tag color={CREDENTIAL_STATUS_COLOR[platform.status]}>
                  {CREDENTIAL_STATUS_LABEL[platform.status]}
                </Tag>
                <Descriptions size="small" column={1} bordered>
                  <Descriptions.Item label="Region">
                    {formatRegion(platform.region)}
                  </Descriptions.Item>
                  <Descriptions.Item label="Secret Mask">
                    {platform.secret_mask || platform.app_secret_mask || "configured"}
                  </Descriptions.Item>
                  <Descriptions.Item label="Stores">
                    {platform.stores.length}
                  </Descriptions.Item>
                </Descriptions>
                <Space>
                  {canManageManualTokens && (
                    <Button icon={<LockOutlined />} onClick={() => onManualToken(platform)}>
                      Manual Token
                    </Button>
                  )}
                  <Button icon={<HistoryOutlined />} onClick={() => onViewHistory(platform)}>
                    View History
                  </Button>
                </Space>
              </Space>
            </Card>
          </Col>
        ))}
      </Row>
    </Card>
  );
}
