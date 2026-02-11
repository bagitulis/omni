import { Card, Empty, Spin, Typography } from "antd";
import { VideoCameraOutlined } from "@ant-design/icons";
import {
  AnalyticsSummaryCards,
  ReconciliationTable,
  ShippingFeeTable,
  SyncProgressBar,
} from "@/components/analytics/common";
import {
  TiktokReconciliationResult,
  TiktokShippingFeeResult,
} from "@/types/analytics";

const { Text } = Typography;

interface AnalyticsContentProps {
  activeTab: "price" | "shipping";
  jobProgress: any;
  showProgressBar: boolean;
  isLoading: boolean;
  isSynced: boolean;
  hasData: boolean;
  reconciliationData: TiktokReconciliationResult | undefined;
  shippingFeeData: TiktokShippingFeeResult | undefined;
  reconciliationLoading: boolean;
  shippingFeeLoading: boolean;
}

export const AnalyticsContent = ({
  activeTab,
  jobProgress,
  showProgressBar,
  isLoading,
  isSynced,
  hasData,
  reconciliationData,
  shippingFeeData,
  reconciliationLoading,
  shippingFeeLoading,
}: AnalyticsContentProps) => {
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
      <Card style={{ borderRadius: 8 }}>
        <Empty
          image={
            <VideoCameraOutlined style={{ fontSize: 64, color: "#d9d9d9" }} />
          }
          description="Select a period and sync escrow data to start analysis"
        />
      </Card>
    );
  }

  if (activeTab === "price" && reconciliationData) {
    return (
      <>
        <AnalyticsSummaryCards
          type="reconciliation"
          summary={reconciliationData.summary}
        />
        <div style={{ marginTop: 24 }}>
          <ReconciliationTable
            platform="tiktok"
            data={reconciliationData.sku_groups}
            loading={reconciliationLoading}
          />
        </div>
      </>
    );
  }

  if (activeTab === "shipping" && shippingFeeData) {
    return (
      <>
        <AnalyticsSummaryCards
          type="shipping"
          summary={shippingFeeData.summary}
        />
        <div style={{ marginTop: 24 }}>
          <ShippingFeeTable
            platform="tiktok"
            data={shippingFeeData.orders}
            loading={shippingFeeLoading}
          />
        </div>
      </>
    );
  }

  // Default empty state if synced but no data
  return (
    <Card style={{ borderRadius: 8, marginTop: 24 }}>
      <Empty description="No data found for this period" />
    </Card>
  );
};
