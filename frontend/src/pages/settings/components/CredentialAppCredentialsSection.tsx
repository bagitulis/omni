import {
  EditOutlined,
  HistoryOutlined,
  LockOutlined,
  PlusOutlined,
  ReloadOutlined,
} from "@ant-design/icons";
import { Button, Card, Col, Descriptions, Row, Space, Tag, Typography } from "antd";
import type { CredentialPlatformSummary } from "@/api/credentials";
import {
  CREDENTIAL_STATUS_COLOR,
  CREDENTIAL_STATUS_LABEL,
  formatRegion,
} from "./credentialStatus";
import type { AppCredentialDrawerMode } from "./AppCredentialFormDrawer";

interface CredentialAppCredentialsSectionProps {
  platforms: CredentialPlatformSummary[];
  platformNames: Record<string, string>;
  privilegedActionReason: string | null;
  manualSavingPlatform: string | null;
  onManualToken: (platform: CredentialPlatformSummary) => void;
  onViewHistory: (platform: CredentialPlatformSummary) => void;
  onManageAppCredential: (
    platform: CredentialPlatformSummary,
    mode: AppCredentialDrawerMode,
  ) => void;
}

function formatRelativeTime(dateStr?: string): string | null {
  if (!dateStr) return null;
  const date = new Date(dateStr);
  const now = new Date();
  const diffMs = now.getTime() - date.getTime();
  const diffMins = Math.floor(diffMs / 60000);
  if (diffMins < 1) return "just now";
  if (diffMins < 60) return `${diffMins}m ago`;
  const diffHours = Math.floor(diffMins / 60);
  if (diffHours < 24) return `${diffHours}h ago`;
  const diffDays = Math.floor(diffHours / 24);
  if (diffDays < 30) return `${diffDays}d ago`;
  return date.toLocaleDateString();
}

export function CredentialAppCredentialsSection({
  platforms,
  platformNames,
  privilegedActionReason,
  manualSavingPlatform,
  onManualToken,
  onViewHistory,
  onManageAppCredential,
}: CredentialAppCredentialsSectionProps) {
  const canManage = !privilegedActionReason;

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
                <Space>
                  <Tag color={CREDENTIAL_STATUS_COLOR[platform.status]}>
                    {CREDENTIAL_STATUS_LABEL[platform.status]}
                  </Tag>
                  <Tag
                    color={platform.app_configured ? "green" : "red"}
                    style={{ margin: 0 }}
                  >
                    {platform.app_configured ? "App Configured" : "Not Configured"}
                  </Tag>
                </Space>
                <Descriptions size="small" column={1} bordered>
                  <Descriptions.Item label="Region">
                    {formatRegion(platform.region)}
                  </Descriptions.Item>
                  <Descriptions.Item label="Secret Mask">
                    {platform.secret_mask || platform.app_secret_mask || "—"}
                  </Descriptions.Item>
                  <Descriptions.Item label="Stores">
                    {platform.stores.length}
                  </Descriptions.Item>
                  {platform.audit_summary?.last_event_at && (
                    <Descriptions.Item label="Last Updated">
                      <Typography.Text type="secondary" style={{ fontSize: 12 }}>
                        {formatRelativeTime(platform.audit_summary.last_event_at)}
                        {platform.audit_summary.last_event_type && (
                          <span>
                            {" "}
                            ({platform.audit_summary.last_event_type.replace(/_/g, " ")})
                          </span>
                        )}
                      </Typography.Text>
                    </Descriptions.Item>
                  )}
                </Descriptions>
                <Space wrap>
                  {!platform.app_configured ? (
                    <Button
                      type="primary"
                      icon={<PlusOutlined />}
                      disabled={!canManage}
                      title={privilegedActionReason || "Configure app credentials"}
                      onClick={() => onManageAppCredential(platform, "setup")}
                    >
                      Configure
                    </Button>
                  ) : (
                    <>
                      <Button
                        icon={<EditOutlined />}
                        disabled={!canManage}
                        title={privilegedActionReason || "Edit app credentials"}
                        onClick={() => onManageAppCredential(platform, "setup")}
                      >
                        Edit
                      </Button>
                      <Button
                        icon={<ReloadOutlined />}
                        disabled={!canManage}
                        title={privilegedActionReason || "Rotate credentials"}
                        onClick={() => onManageAppCredential(platform, "rotate")}
                      >
                        Rotate
                      </Button>
                    </>
                  )}
                  <Button
                    icon={<LockOutlined />}
                    disabled={
                      !canManage || manualSavingPlatform === platform.platform
                    }
                    loading={manualSavingPlatform === platform.platform}
                    title={privilegedActionReason || undefined}
                    onClick={() => onManualToken(platform)}
                  >
                    Manual Token
                  </Button>
                  <Button
                    icon={<HistoryOutlined />}
                    onClick={() => onViewHistory(platform)}
                  >
                    History
                  </Button>
                </Space>
                {privilegedActionReason && (
                  <Typography.Text type="secondary" style={{ fontSize: 12 }}>
                    {privilegedActionReason}
                  </Typography.Text>
                )}
              </Space>
            </Card>
          </Col>
        ))}
      </Row>
    </Card>
  );
}
