import {
  Button,
  Form,
  Input,
  InputNumber,
  Select,
  Space,
  Typography,
} from "antd";
import { PlayCircleOutlined } from "@ant-design/icons";
import type { OmniExtension } from "@/api/extensions";

const { Paragraph, Text } = Typography;

export type ScrapeMode = "search" | "shop" | "product";

export interface ScrapeFormValues {
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

export interface ScrapeFormProps {
  connected: OmniExtension[];
  mode: ScrapeMode;
  onModeChange: (mode: ScrapeMode) => void;
  onSubmit: (values: ScrapeFormValues) => void;
  submitting: boolean;
}

export function ScrapeForm({
  connected,
  mode,
  onModeChange,
  onSubmit,
  submitting,
}: ScrapeFormProps) {
  const [form] = Form.useForm<ScrapeFormValues>();

  return (
    <Form
      form={form}
      layout="vertical"
      initialValues={{ mode: "search", max_pages: 2, max_products: 0 }}
      onFinish={onSubmit}
      onValuesChange={(changed) => {
        if (changed.mode) onModeChange(changed.mode as ScrapeMode);
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
          <Input placeholder="https://shopee.co.id/product/123/456" allowClear />
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
            loading={submitting}
            disabled={connected.length === 0}
          >
            Start scrape
          </Button>
          <Text type="secondary">Progress appears below once queued.</Text>
        </Space>
      </Form.Item>
    </Form>
  );
}

export default ScrapeForm;
