import { Button, Spin, Alert, Typography, theme } from "antd";
import { ReloadOutlined } from "@ant-design/icons";
import { useAnalyticsHub } from "@/hooks/useAnalyticsHub";
import { KPISection } from "@/components/analytics/hub/KPISection";
import { QuickActions } from "@/components/analytics/hub/QuickActions";
import { PlatformComparison } from "@/components/analytics/hub/PlatformComparison";
import { ActionSummary } from "@/components/analytics/hub/ActionSummary";

const { Title, Text } = Typography;
const { useToken } = theme;

export const AnalyticsHubPage = () => {
  const { token } = useToken();
  const { kpi, summary, isLoading, error, refetch } = useAnalyticsHub();

  if (error) {
    return (
      <div style={{ padding: 24 }}>
        <Alert
          message="Error Loading Data"
          description={(error as Error).message}
          type="error"
          showIcon
          action={
            <Button size="small" onClick={() => refetch()}>
              Retry
            </Button>
          }
        />
      </div>
    );
  }

  return (
    <div style={{ padding: 24 }}>
      {/* Header */}
      <div
        style={{
          marginBottom: 24,
          display: "flex",
          justifyContent: "space-between",
          alignItems: "center",
        }}
      >
        <div>
          <Title
            level={2}
            style={{ margin: 0, fontSize: token.fontSizeHeading2 }}
          >
            Analytics Hub
          </Title>
          <Text type="secondary">Unified insights across all platforms</Text>
        </div>
        <Button
          icon={<ReloadOutlined />}
          onClick={() => refetch()}
          loading={isLoading}
        >
          Refresh Data
        </Button>
      </div>

      <Spin spinning={isLoading}>
        <KPISection kpi={kpi} summary={summary} />
        <QuickActions />
        <PlatformComparison summary={summary} />
        <ActionSummary kpi={kpi} />
      </Spin>
    </div>
  );
};

export default AnalyticsHubPage;
