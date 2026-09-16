import { useState } from "react";
import { useSearchParams } from "react-router-dom";
import {
  Alert,
  Button,
  Card,
  Empty,
  Input,
  Space,
  Tag,
  Typography,
  message,
} from "antd";
import { SearchOutlined, ReloadOutlined } from "@ant-design/icons";
import { useScrapedProducts } from "@/hooks/useExtensions";
import { useResumeScrape, useScrapeJob } from "@/hooks/useScrapeJob";
import { extractBlocker, parseScrapeSummary, scrapeStatusTag } from "@/api/scrapeJob";
import { summariseSources } from "@/api/extensions";
import { ScrapedProductTable } from "@/components/tables/ScrapedProductTable";
import { ScrapeBlockerAlert } from "@/components/extensions/ScrapeBlockerAlert";

const { Title, Paragraph, Text } = Typography;

export function ExtensionResultsPage() {
  // The job id is read from the URL so the scraper page can link straight to a
  // queued job's results, and so the view is shareable.
  const [searchParams, setSearchParams] = useSearchParams();
  const urlJobId = searchParams.get("job_id") ?? "";

  // The URL is the source of truth for the job id. Initialising state from it
  // once would ignore later navigations: arriving from the scraper while already
  // on this page changes only the query string, so the component does not
  // remount and the state would keep showing the previous job.
  const [pendingJobId, setPendingJobId] = useState(urlJobId);
  const [page, setPage] = useState(1);
  const [lastUrlJobId, setLastUrlJobId] = useState(urlJobId);

  if (urlJobId !== lastUrlJobId) {
    // Sync during render rather than in an effect, so the table never renders
    // one frame of the previous job's data under the new job's heading.
    setLastUrlJobId(urlJobId);
    setPendingJobId(urlJobId);
    setPage(1);
  }

  const jobId = urlJobId;

  // A page change must reset when the job changes; the sync above handles it.
  const applyJobId = (next: string) => {
    const trimmed = next.trim();
    setPendingJobId(trimmed);
    setPage(1);
    setSearchParams(trimmed ? { job_id: trimmed } : {});
  };

  const { data, isLoading, isError, error, refetch, isFetching } =
    useScrapedProducts(jobId, page, 50);

  const jobQuery = useScrapeJob(jobId);
  const resumeJob = useResumeScrape(jobId);

  const jobDetail = jobQuery.data;
  const blocker = extractBlocker(jobDetail ?? undefined);
  const summary = parseScrapeSummary(jobDetail?.result_data);
  const statusTag = jobDetail ? scrapeStatusTag(jobDetail.status) : null;

  const onResume = () => {
    resumeJob.mutate(undefined, {
      onSuccess: (resumed) => {
        if (resumed.resumed === false) {
          message.info("The job is no longer blocked, so nothing was resumed");
          return;
        }
        message.success("Scrape resumed");
      },
      onError: (err: Error) => message.error(err.message || "Failed to resume"),
    });
  };

  const products = data?.products ?? [];

  return (
    <div style={{ padding: 24 }}>
      <Title level={3} style={{ marginBottom: 4 }}>
        Scrape Results
      </Title>
      <Paragraph type="secondary">
        Products collected by a scrape run. Enter a job id to narrow the list, or
        leave it blank to see everything collected for this tenant.
      </Paragraph>

      <Card style={{ marginBottom: 16 }}>
        <Space>
          <Input
            placeholder="Job id (optional)"
            value={pendingJobId}
            onChange={(e) => setPendingJobId(e.target.value)}
            onPressEnter={() => applyJobId(pendingJobId)}
            style={{ width: 320 }}
            allowClear
          />
          <Button
            type="primary"
            icon={<SearchOutlined />}
            onClick={() => applyJobId(pendingJobId)}
            disabled={!pendingJobId.trim()}
          >
            Load
          </Button>
          {jobId && <Button onClick={() => applyJobId("")}>Clear</Button>}
          <Button
            icon={<ReloadOutlined />}
            loading={isFetching}
            disabled={!jobId}
            onClick={() => void refetch()}
          >
            Refresh
          </Button>
        </Space>
      </Card>

      {blocker && (
        <div style={{ marginBottom: 16 }}>
          <ScrapeBlockerAlert
            blocker={blocker}
            onResume={onResume}
            resuming={resumeJob.isPending}
            resumeFromPage={summary?.resume_from_page}
          />
        </div>
      )}

      {isError && (
        <Alert
          style={{ marginBottom: 16 }}
          type="error"
          showIcon
          message="Failed to load results"
          description={(error as Error)?.message}
        />
      )}

      <Card
        title={
          jobId ? (
            <Space>
              <Text>Job</Text>
              <Text code>{jobId}</Text>
              {statusTag && <Tag color={statusTag.color}>{statusTag.label}</Tag>}
              {products.length > 0 && (
                <Text type="secondary">— {summariseSources(products)}</Text>
              )}
            </Space>
          ) : (
            "All results"
          )
        }
      >
        {!jobId ? (
          <Empty description="Enter a job id to load its results." />
        ) : (
          <ScrapedProductTable
            products={products}
            loading={isLoading}
            pagination={{
              current: data?.page ?? page,
              pageSize: data?.page_size ?? 50,
              total: data?.total ?? 0,
              onChange: setPage,
            }}
          />
        )}
      </Card>
    </div>
  );
}

export default ExtensionResultsPage;
