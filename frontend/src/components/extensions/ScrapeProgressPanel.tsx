import { Alert, Button, Card, Descriptions, Progress, Space, Tag, Typography } from "antd";
import { StopOutlined } from "@ant-design/icons";
import {
  extractBlocker,
  parseScrapeSummary,
  scrapeStatusTag,
  shouldStopPolling,
  type ScrapeJob,
} from "@/api/scrapeJob";
import { ScrapeBlockerAlert } from "./ScrapeBlockerAlert";

const { Text } = Typography;

export interface ScrapeProgressPanelProps {
  jobId: string;
  jobDetail: ScrapeJob | null | undefined;
  loadError: string | null;
  onCancel: () => void;
  cancelling: boolean;
  onResume: () => void;
  resuming: boolean;
  onViewResults: () => void;
}

/**
 * Progress for a live run.
 *
 * The backend reports no percentage for a scrape whose page count is unknown, so
 * an indeterminate run shows an active bar rather than a fabricated number.
 */
function progressFor(jobDetail: ScrapeJob) {
  const percent = jobDetail.progress_percent ?? 0;
  if (jobDetail.status === "completed") return { percent: 100, status: "success" as const };
  if (jobDetail.status === "failed") return { percent, status: "exception" as const };
  if (jobDetail.status === "cancelled" || jobDetail.status === "blocked") {
    return { percent, status: "normal" as const };
  }
  return { percent, status: "active" as const };
}

export function ScrapeProgressPanel({
  jobId,
  jobDetail,
  loadError,
  onCancel,
  cancelling,
  onResume,
  resuming,
  onViewResults,
}: ScrapeProgressPanelProps) {
  const status = jobDetail?.status;
  const tag = scrapeStatusTag(status ?? "pending");
  const summary = parseScrapeSummary(jobDetail?.result_data);
  const blocker = extractBlocker(jobDetail ?? undefined);
  const settled = shouldStopPolling(status);
  const progress = jobDetail ? progressFor(jobDetail) : null;

  return (
    <Card
      style={{ marginTop: 16 }}
      title={
        <Space>
          <Text>Job</Text>
          <Text code>{jobId}</Text>
          <Tag color={tag.color}>{tag.label}</Tag>
        </Space>
      }
      extra={
        <Space>
          {!settled && (
            <Button
              danger
              size="small"
              icon={<StopOutlined />}
              loading={cancelling}
              onClick={onCancel}
            >
              Cancel
            </Button>
          )}
          <Button size="small" type="link" onClick={onViewResults}>
            View results
          </Button>
        </Space>
      }
    >
      <Space direction="vertical" size={12} style={{ width: "100%" }}>
        {loadError && (
          <Alert type="error" showIcon message="Job status unavailable" description={loadError} />
        )}

        {progress && (
          <Progress percent={progress.percent} status={progress.status} size="small" />
        )}

        {jobDetail?.progress_message && (
          <Text type="secondary">{jobDetail.progress_message}</Text>
        )}

        {blocker && (
          <ScrapeBlockerAlert
            blocker={blocker}
            onResume={onResume}
            resuming={resuming}
            resumeFromPage={summary?.resume_from_page}
          />
        )}

        {jobDetail?.status === "failed" && jobDetail.error_message && (
          <Alert
            type="error"
            showIcon
            message="Scrape failed"
            description={jobDetail.error_message}
          />
        )}

        {summary && (
          <Descriptions size="small" column={2} bordered>
            <Descriptions.Item label="Products">{summary.products ?? 0}</Descriptions.Item>
            <Descriptions.Item label="Pages">{summary.pages ?? 0}</Descriptions.Item>
            <Descriptions.Item label="Mode">{summary.mode ?? "—"}</Descriptions.Item>
            <Descriptions.Item label="Stop reason">{summary.reason ?? "—"}</Descriptions.Item>
          </Descriptions>
        )}

        {!settled && !loadError && (
          <Text type="secondary">
            Updating automatically until the run finishes.
          </Text>
        )}
      </Space>
    </Card>
  );
}

export default ScrapeProgressPanel;
