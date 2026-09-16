import { Alert, Button, Space, Typography } from "antd";
import { blockerKindLabel, type ScrapeBlocker } from "@/api/scrapeJob";

const { Text } = Typography;

export interface ScrapeBlockerAlertProps {
  blocker: ScrapeBlocker;
  onResume: () => void;
  resuming: boolean;
  /** The page a resume restarts from, when the summary recorded one. */
  resumeFromPage?: number;
}

/**
 * Tell the operator what stopped the run and offer to resume it.
 *
 * The blocked URL is shown as a link because clearing the obstacle means opening
 * that exact page and solving it by hand; resuming before that just blocks again.
 */
export function ScrapeBlockerAlert({
  blocker,
  onResume,
  resuming,
  resumeFromPage,
}: ScrapeBlockerAlertProps) {
  const page = resumeFromPage ?? blocker.blocked_page;

  return (
    <Alert
      type="warning"
      showIcon
      message={`Scrape blocked — ${blockerKindLabel(blocker.kind)}`}
      description={
        <Space direction="vertical" size={4} style={{ width: "100%" }}>
          <Text>
            Shopee interrupted the run on page {blocker.blocked_page}. Open the
            page below, clear the challenge in the paired browser, then resume.
          </Text>
          <Typography.Link
            href={blocker.url}
            target="_blank"
            rel="noopener noreferrer"
          >
            {blocker.url}
          </Typography.Link>
          <Space>
            <Button size="small" type="primary" loading={resuming} onClick={onResume}>
              Resume from page {page}
            </Button>
            <Text type="secondary">
              Resuming before the challenge is cleared will block again.
            </Text>
          </Space>
        </Space>
      }
    />
  );
}

export default ScrapeBlockerAlert;
