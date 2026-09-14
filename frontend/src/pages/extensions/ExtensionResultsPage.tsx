import { useState } from "react";
import { useSearchParams } from "react-router-dom";
import {
  Alert,
  Button,
  Card,
  Empty,
  Input,
  Space,
  Table,
  Tag,
  Typography,
} from "antd";
import type { ColumnsType } from "antd/es/table";
import { SearchOutlined, ReloadOutlined } from "@ant-design/icons";
import { useScrapedProducts } from "@/hooks/useExtensions";
import { summariseSources, type ScrapedProduct } from "@/api/extensions";

const { Title, Paragraph, Text } = Typography;

/**
 * Render the capture source as a tag.
 *
 * Surfaced deliberately: if the DOM fallback is carrying production traffic,
 * that means the preferred API capture has stopped working, and it should be
 * visible rather than discovered when the fallback also breaks.
 */
function SourceTag({ source }: { source?: string }) {
  if (source === "network") return <Tag color="green">network API</Tag>;
  if (source === "dom") return <Tag color="orange">DOM fallback</Tag>;
  return <Tag>unknown</Tag>;
}

export function ExtensionResultsPage() {
  // The job id is read from the URL so the scraper page can link straight to a
  // queued job's results, and so the view is shareable.
  const [searchParams, setSearchParams] = useSearchParams();
  const urlJobId = searchParams.get("job_id") ?? "";

  const [jobId, setJobId] = useState(urlJobId);
  const [pendingJobId, setPendingJobId] = useState(urlJobId);
  const [page, setPage] = useState(1);

  const applyJobId = (next: string) => {
    const trimmed = next.trim();
    setJobId(trimmed);
    setPendingJobId(trimmed);
    setPage(1);
    // Keep the URL in step so a refresh or a shared link shows the same view.
    setSearchParams(trimmed ? { job_id: trimmed } : {});
  };

  const { data, isLoading, isError, error, refetch, isFetching } =
    useScrapedProducts(jobId, page, 50);

  const columns: ColumnsType<ScrapedProduct> = [
    {
      title: "Product",
      dataIndex: "product_name",
      key: "product_name",
      render: (name: string, row) => (
        <Space direction="vertical" size={0}>
          <Text strong>{name}</Text>
          {row.link && (
            <Typography.Link
              href={row.link}
              target="_blank"
              rel="noopener noreferrer"
              style={{ fontSize: 12 }}
            >
              open on Shopee
            </Typography.Link>
          )}
        </Space>
      ),
    },
    {
      title: "Price",
      dataIndex: "price",
      key: "price",
      width: 110,
      render: (price: string) => price || "—",
    },
    {
      title: "Sold",
      dataIndex: "sold",
      key: "sold",
      width: 90,
      render: (sold: string) => sold || "—",
    },
    {
      title: "Source",
      dataIndex: "source",
      key: "source",
      width: 130,
      render: (source: string) => <SourceTag source={source} />,
    },
    {
      title: "Page",
      dataIndex: "page_number",
      key: "page_number",
      width: 70,
    },
    {
      title: "Shop item id",
      dataIndex: "shopee_item_id",
      key: "shopee_item_id",
      width: 130,
      render: (id: string) => id || "—",
    },
  ];

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
          {jobId && (
            <Button onClick={() => applyJobId("")}>Clear</Button>
          )}
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
              {products.length > 0 && (
                <Text type="secondary">
                  — {summariseSources(products)}
                </Text>
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
          <Table
            rowKey="id"
            columns={columns}
            dataSource={products}
            loading={isLoading}
            pagination={{
              current: data?.page ?? page,
              pageSize: data?.page_size ?? 50,
              total: data?.total ?? 0,
              onChange: setPage,
              showSizeChanger: false,
            }}
            locale={{
              emptyText: (
                <Empty description="No products were collected for this job." />
              ),
            }}
          />
        )}
      </Card>
    </div>
  );
}

export default ExtensionResultsPage;
