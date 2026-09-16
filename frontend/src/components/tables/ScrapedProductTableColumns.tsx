import { Space, Tag, Typography } from "antd";
import type { ColumnsType } from "antd/es/table";
import type { ScrapedProduct, ScrapeSource } from "@/api/extensions";

const { Text } = Typography;

/**
 * Render the capture source as a tag.
 *
 * Surfaced deliberately: if the DOM fallback is carrying production traffic,
 * that means the preferred API capture has stopped working, and it should be
 * visible rather than discovered when the fallback also breaks.
 */
export function SourceTag({ source }: { source?: ScrapeSource }) {
  if (source === "network") return <Tag color="green">network API</Tag>;
  if (source === "dom") return <Tag color="orange">DOM fallback</Tag>;
  return <Tag>unknown</Tag>;
}

export function getScrapedProductTableColumns(): ColumnsType<ScrapedProduct> {
  return [
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
      render: (source: ScrapeSource | undefined) => <SourceTag source={source} />,
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
}
