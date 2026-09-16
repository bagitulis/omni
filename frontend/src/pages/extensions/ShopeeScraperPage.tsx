import { useState } from "react";
import { Alert, Card, Typography, message } from "antd";
import { useNavigate } from "react-router-dom";
import { useExtensions, useStartScrape } from "@/hooks/useExtensions";
import {
  useCancelScrapeJob,
  useResumeScrape,
  useScrapeJob,
} from "@/hooks/useScrapeJob";
import type { ScrapeRequest } from "@/api/extensions";
import {
  ScrapeForm,
  type ScrapeFormValues,
  type ScrapeMode,
} from "@/components/extensions/ScrapeForm";
import { ScrapeProgressPanel } from "@/components/extensions/ScrapeProgressPanel";

const { Title, Paragraph } = Typography;

export function ShopeeScraperPage() {
  const [mode, setMode] = useState<ScrapeMode>("search");
  const [error, setError] = useState<string | null>(null);
  // The queued job id drives both the progress poll and the link to its results.
  const [queuedJobId, setQueuedJobId] = useState<string | null>(null);
  const navigate = useNavigate();

  const { data: extensions, failureNotice } = useExtensions();
  const startScrape = useStartScrape();

  const jobQuery = useScrapeJob(queuedJobId);
  const cancelJob = useCancelScrapeJob(queuedJobId ?? "");
  const resumeJob = useResumeScrape(queuedJobId ?? "");

  const connected = (extensions ?? []).filter(
    (e) => e.status === "connected" && e.capabilities?.includes("shopee_scrape"),
  );

  const onSubmit = (values: ScrapeFormValues) => {
    setError(null);
    setQueuedJobId(null);

    const payload: ScrapeRequest = {
      mode: values.mode,
      extension_id: values.extension_id,
      max_pages: values.max_pages,
      max_products: values.max_products,
    };
    if (values.mode === "search") payload.query = values.query?.trim();
    if (values.mode === "shop") payload.shop_url = values.shop_url?.trim();
    if (values.mode === "product") payload.product_url = values.product_url?.trim();

    startScrape.mutate(payload, {
      onSuccess: (data) => {
        setQueuedJobId(data.job_id);
        message.success("Scrape queued");
      },
      onError: (err: Error) => {
        // The backend distinguishes an unreachable extension from a blocked
        // scrape, so the message is shown rather than replaced with a generic one.
        setError(err.message || "Scrape failed");
        message.error("Scrape failed");
      },
    });
  };

  const onCancel = () => {
    cancelJob.mutate(undefined, {
      onSuccess: () => message.success("Cancellation requested"),
      onError: (err: Error) => message.error(err.message || "Failed to cancel"),
    });
  };

  const onResume = () => {
    resumeJob.mutate(undefined, {
      onSuccess: (data) => {
        if (data.resumed === false) {
          message.info("The job is no longer blocked, so nothing was resumed");
          return;
        }
        message.success("Scrape resumed");
      },
      onError: (err: Error) => message.error(err.message || "Failed to resume"),
    });
  };

  return (
    <div style={{ padding: 24, maxWidth: 720 }}>
      <Title level={3} style={{ marginBottom: 4 }}>
        Shopee Scraper
      </Title>
      <Paragraph type="secondary">
        Collection runs in a paired browser, so Shopee sees a normal session. It
        prefers Shopee's own API responses and falls back to reading the page
        only when those are unavailable.
      </Paragraph>

      {failureNotice && (
        <Alert
          style={{ marginBottom: 16 }}
          type="error"
          showIcon
          message="Extension list may be stale"
          description={failureNotice}
        />
      )}

      {connected.length === 0 && (
        <Alert
          style={{ marginBottom: 16 }}
          type="warning"
          showIcon
          message="No connected extension"
          description="Pair a browser extension on the Installed Extensions page before starting a scrape."
        />
      )}

      <Card>
        <ScrapeForm
          connected={connected}
          mode={mode}
          onModeChange={setMode}
          onSubmit={onSubmit}
          submitting={startScrape.isPending}
        />
      </Card>

      {error && (
        <Alert
          style={{ marginTop: 16 }}
          type="error"
          showIcon
          message="Scrape failed"
          description={error}
        />
      )}

      {queuedJobId && (
        <ScrapeProgressPanel
          jobId={queuedJobId}
          jobDetail={jobQuery.data}
          loadError={
            jobQuery.isError
              ? (jobQuery.error as Error)?.message || "Failed to read the job status"
              : null
          }
          onCancel={onCancel}
          cancelling={cancelJob.isPending}
          onResume={onResume}
          resuming={resumeJob.isPending}
          onViewResults={() =>
            navigate(
              `/extensions/results?job_id=${encodeURIComponent(queuedJobId)}`,
            )
          }
        />
      )}
    </div>
  );
}

export default ShopeeScraperPage;
