import { useState } from "react";
import {
  Alert,
  Button,
  Card,
  Form,
  Input,
  InputNumber,
  Select,
  Space,
  Typography,
  message,
} from "antd";
import { PlayCircleOutlined } from "@ant-design/icons";
import { useExtensions, useStartScrape } from "@/hooks/useExtensions";
import type { ScrapeRequest } from "@/api/extensions";

const { Title, Paragraph, Text } = Typography;

type ScrapeMode = "search" | "shop" | "product";

interface FormValues {
  mode: ScrapeMode;
  query?: string;
  shop_url?: string;
  product_url?: string;
  max_pages?: number;
  max_products?: number;
  extension_id: string;
}

/** What each mode collects, shown so the form is self-explanatory. */
const MODE_HELP: Record<ScrapeMode, string> = {
  search:
    "Collects products matching a keyword, paging through the search results.",
  shop: "Collects every product listed in a single Shopee shop.",
  product: "Collects the details of one product page.",
};

export function ShopeeScraperPage() {
  const [form] = Form.useForm<FormValues>();
  const [mode, setMode] = useState<ScrapeMode>("search");
  const [result, setResult] = useState<string | null>(null);
  const [error, setError] = useState<string | null>(null);

  const { data: extensions } = useExtensions();
  const startScrape = useStartScrape();

  const connected = (extensions ?? []).filter(
    (e) => e.status === "connected" && e.capabilities?.includes("shopee_scrape"),
  );

  const onSubmit = (values: FormValues) => {
    setError(null);
    setResult(null);

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
        setResult(data.summary || "Scrape completed");
        message.success("Scrape completed");
      },
      onError: (err: Error) => {
        // The backend distinguishes an unreachable extension from a blocked
        // scrape, so the message is shown rather than replaced with a generic one.
        setError(err.message || "Scrape failed");
        message.error("Scrape failed");
      },
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
        <Form
          form={form}
          layout="vertical"
          initialValues={{ mode: "search", max_pages: 2, max_products: 0 }}
          onFinish={onSubmit}
          onValuesChange={(changed) => {
            if (changed.mode) setMode(changed.mode as ScrapeMode);
          }}
        >
          <Form.Item
            name="extension_id"
            label="Extension"
            rules={[{ required: true, message: "Select an extension" }]}
          >
            <Select
              placeholder="Select a connected extension"
              options={connected.map((e) => ({
                value: e.extension_id,
                label: `${e.hostname || "browser"} — ${e.extension_id.slice(0, 12)}…`,
              }))}
              notFoundContent="No connected extension with Shopee capability"
            />
          </Form.Item>

          <Form.Item name="mode" label="Mode" rules={[{ required: true }]}>
            <Select
              options={[
                { value: "search", label: "Keyword search" },
                { value: "shop", label: "Shop listing" },
                { value: "product", label: "Single product" },
              ]}
            />
          </Form.Item>

          <Paragraph type="secondary" style={{ marginTop: -12 }}>
            {MODE_HELP[mode]}
          </Paragraph>

          {mode === "search" && (
            <Form.Item
              name="query"
              label="Keyword"
              rules={[{ required: true, message: "Enter a keyword to search for" }]}
            >
              <Input placeholder="e.g. mechanical keyboard" allowClear />
            </Form.Item>
          )}

          {mode === "shop" && (
            <Form.Item
              name="shop_url"
              label="Shop URL"
              rules={[{ required: true, message: "Enter the shop URL" }]}
            >
              <Input placeholder="https://shopee.co.id/your-shop" allowClear />
            </Form.Item>
          )}

          {mode === "product" && (
            <Form.Item
              name="product_url"
              label="Product URL"
              rules={[{ required: true, message: "Enter the product URL" }]}
            >
              <Input
                placeholder="https://shopee.co.id/product/123/456"
                allowClear
              />
            </Form.Item>
          )}

          <Space size="large" align="start">
            <Form.Item
              name="max_pages"
              label="Max pages"
              tooltip="0 means no explicit limit."
            >
              <InputNumber min={0} max={200} />
            </Form.Item>
            <Form.Item
              name="max_products"
              label="Max products"
              tooltip="0 means no explicit limit."
            >
              <InputNumber min={0} max={10000} />
            </Form.Item>
          </Space>

          <Form.Item style={{ marginBottom: 0 }}>
            <Space>
              <Button
                type="primary"
                htmlType="submit"
                icon={<PlayCircleOutlined />}
                loading={startScrape.isPending}
                disabled={connected.length === 0}
              >
                Start scrape
              </Button>
              <Text type="secondary">
                Results appear on the Results page.
              </Text>
            </Space>
          </Form.Item>
        </Form>
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

      {result && (
        <Alert
          style={{ marginTop: 16 }}
          type="success"
          showIcon
          message="Scrape completed"
          description={<Text code>{result}</Text>}
        />
      )}
    </div>
  );
}

export default ShopeeScraperPage;
