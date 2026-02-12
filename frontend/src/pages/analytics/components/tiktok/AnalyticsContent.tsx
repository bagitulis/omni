import { Card, Empty, Spin, Typography } from "antd";
import { VideoCameraOutlined } from "@ant-design/icons";
import { SyncProgressBar } from "@/components/analytics/common";
import type {
  TiktokReconciliationResult,
  TiktokShippingFeeResult,
  JobProgress,
} from "@/types/analytics";
import { TiktokPriceSummaryCards } from "./TiktokPriceSummaryCards";
import { TiktokPriceResultsTable } from "./TiktokPriceResultsTable";
import { TiktokShippingFeeSummary } from "./TiktokShippingFeeSummary";
import { TiktokShippingFeeTable } from "./TiktokShippingFeeTable";

const { Text } = Typography;

interface AnalyticsContentProps {
  activeTab: "price" | "shipping";
  jobProgress: JobProgress | null;
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
        <TiktokPriceSummaryCards summary={reconciliationData.summary} />
        <div style={{ marginTop: 24 }}>
          <TiktokPriceResultsTable
            skuGroups={reconciliationData.sku_groups}
            loading={reconciliationLoading}
          />
        </div>
      </>
    );
  }

  if (activeTab === "shipping" && shippingFeeData) {
    return (
      <>
        <TiktokShippingFeeSummary summary={shippingFeeData.summary} />
        <div style={{ marginTop: 24 }}>
          <TiktokShippingFeeTable
            orders={shippingFeeData.orders}
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
