import { Progress, Typography, Space, theme } from "antd";
import type { JobProgress } from "@/types/analytics";

const { Text } = Typography;

interface Props {
  jobProgress: JobProgress;
}

export function SyncProgressBar({ jobProgress }: Props) {
  const { token } = theme.useToken();

  const isFailed = jobProgress.status === "failed";
  const isActive = jobProgress.status === "running";

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
        status={isFailed ? "exception" : isActive ? "active" : "normal"}
        size="small"
        strokeColor={token.colorPrimary}
        style={{ marginBottom: 0 }}
      />
    </div>
  );
}
