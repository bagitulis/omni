import { Card, Empty, Spin, Typography, theme } from "antd";
import type { ReactNode } from "react";
import type { JobProgress } from "@/types/analytics";
import { SyncProgressBar } from "@/components/analytics/common";

const { Text } = Typography;

interface AnalyticsContentStateProps {
  isLoading: boolean;
  hasData: boolean;
  isSynced: boolean;
  showProgressBar: boolean;
  jobProgress?: JobProgress | null;
  emptyIcon: ReactNode;
  emptyDescription: string;
  children: ReactNode;
}

export const AnalyticsContentState = ({
  isLoading,
  hasData,
  isSynced,
  showProgressBar,
  jobProgress,
  emptyIcon,
  emptyDescription,
  children,
}: AnalyticsContentStateProps) => {
  const { token } = theme.useToken();

  if (showProgressBar && jobProgress) {
    return (
      <div style={{ marginBottom: 24 }}>
        <SyncProgressBar jobProgress={jobProgress} />
      </div>
    );
  }

  if (isLoading && !hasData) {
    return (
      <div style={{ textAlign: "center", padding: 80 }}>
        <Spin size="large" />
        <div style={{ marginTop: 16 }}>
          <Text type="secondary">Loading analytics data...</Text>
        </div>
      </div>
    );
  }

  if (!isSynced && !showProgressBar) {
    return (
      <Card style={{ borderRadius: token.borderRadius }}>
        <Empty image={emptyIcon} description={emptyDescription} />
      </Card>
    );
  }

  if (hasData) {
    return <>{children}</>;
  }

  // Default empty state if synced but no data
  return (
    <Card style={{ borderRadius: token.borderRadius, marginTop: 24 }}>
      <Empty description="No data found for this period" />
    </Card>
  );
};
