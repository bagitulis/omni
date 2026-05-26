import { Alert, Typography } from "antd";

const { Text, Title } = Typography;

interface PlatformsTabHeaderProps {
  tenantBanner: string;
  isTenantDataReady: boolean;
  actionStatus: string | null;
}

export function PlatformsTabHeader({
  tenantBanner,
  isTenantDataReady,
  actionStatus,
}: PlatformsTabHeaderProps) {
  return (
    <>
      <div style={{ marginBottom: 16 }}>
        <Title level={5} style={{ margin: 0, fontSize: 14 }}>
          Credential Management
        </Title>
        <Text type="secondary" style={{ fontSize: 12 }}>
          Separate store connections from app credentials and review masked credential history
        </Text>
      </div>
      <Alert
        showIcon
        type={isTenantDataReady ? "info" : "warning"}
        message="Tenant credential context"
        description={tenantBanner}
        style={{ marginBottom: 16 }}
      />
      {actionStatus && (
        <Alert
          showIcon
          type="success"
          message={actionStatus}
          style={{ marginBottom: 16 }}
        />
      )}
    </>
  );
}
