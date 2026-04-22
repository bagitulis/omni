import { Progress, Typography, Space, Alert, theme } from "antd";
import { WarningOutlined } from "@ant-design/icons";
import type { JobProgress } from "@/types/analytics";

const { Text } = Typography;

interface Props {
  jobProgress: JobProgress;
}

export function SyncProgressBar({ jobProgress }: Props) {
  const { token } = theme.useToken();

  const isFailed = jobProgress.status === "failed";
  const isCancelled = jobProgress.status === "cancelled";
  const isActive = jobProgress.status === "running";

  if (isFailed || isCancelled) {
    return (
      <Alert
        type={isFailed ? "error" : "warning"}
        showIcon
        icon={<WarningOutlined />}
        message={isFailed ? "Sync Failed" : "Sync Cancelled"}
        description={
          <Space direction="vertical" size={4}>
            <Text type="secondary">
              {jobProgress.error_message ||
                jobProgress.progress_message ||
                "An error occurred during sync."}
            </Text>
            <Text type="secondary" style={{ fontSize: 12 }}>
              Click <strong>Sync Escrow</strong> or{" "}
              <strong>Force Resync</strong> to try again.
            </Text>
          </Space>
        }
        style={{ borderRadius: token.borderRadius }}
      />
    );
  }

  return (
    <div style={{ width: "100%" }}>
      <Space style={{ width: "100%", justifyContent: "space-between" }}>
        <Text style={{ fontSize: 12 }}>
          {jobProgress.progress_message || "Ready"}
        </Text>
        {jobProgress.total_items > 0 && (
          <Text type="secondary" style={{ fontSize: 12 }}>
            {jobProgress.processed_items} / {jobProgress.total_items} items
          </Text>
        )}
      </Space>
      <Progress
        percent={jobProgress.progress_percent}
        status={isActive ? "active" : "normal"}
        size="small"
        strokeColor={token.colorPrimary}
        style={{ marginBottom: 0 }}
      />
    </div>
  );
}
