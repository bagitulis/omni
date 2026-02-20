import { useMemo } from "react";
import {
  Table,
  Alert,
  Typography,
  Space,
  Flex,
  theme,
} from "antd";
import { ArrowRightOutlined } from "@ant-design/icons";

interface PreviewRow {
  id?: string | number;
  item_name?: string;
  item_sku?: string;
  image_url?: string;
  price?: number;
  new_price: number;
  price_diff: number;
}

interface PricePreviewTableProps {
  data: PreviewRow[];
}

/**
 * Renders the price change preview table used in the PriceModal.
 * Extracted for SRP — the table rendering is its own concern.
 */
export function PricePreviewTable({ data }: PricePreviewTableProps) {
  const { token } = theme.useToken();

  const columns = useMemo(
    () => [
      {
        title: "Product",
        dataIndex: "item_name",
        width: 200,
        ellipsis: true,
        render: (text: string, record: PreviewRow) => (
          <Space>
            <img
              src={record.image_url || "/placeholder.png"}
              alt=""
              style={{
                width: 32,
                height: 32,
                objectFit: "cover" as const,
                borderRadius: 2,
              }}
            />
            <Flex vertical>
              <span style={{ fontWeight: 500, fontSize: token.fontSizeSM }}>
                {text}
              </span>
              <span
                style={{
                  fontSize: token.fontSizeSM,
                  color: token.colorTextDescription,
                }}
              >
                {record.item_sku}
              </span>
            </Flex>
          </Space>
        ),
      },
      {
        title: "Old Price",
        dataIndex: "price",
        width: 100,
        render: (val: number) => val?.toLocaleString(),
      },
      {
        title: "",
        width: 30,
        render: () => (
          <ArrowRightOutlined
            style={{ color: token.colorTextDescription }}
          />
        ),
      },
      {
        title: "New Price",
        dataIndex: "new_price",
        width: 100,
        render: (val: number) => (
          <span
            style={{
              fontWeight: "bold",
              color: val === 0 ? token.colorError : token.colorSuccess,
            }}
          >
            {val?.toLocaleString()}
          </span>
        ),
      },
      {
        title: "Change",
        dataIndex: "price_diff",
        width: 100,
        render: (val: number) => (
          <Typography.Text
            style={{
              color:
                val > 0
                  ? token.colorSuccess
                  : val < 0
                    ? token.colorError
                    : token.colorTextDescription,
            }}
          >
            {val > 0 ? "+" : ""}
            {val?.toLocaleString()}
          </Typography.Text>
        ),
      },
    ],
    [token],
  );

  return (
    <Flex vertical gap={16} style={{ paddingBlock: 16 }}>
      <Alert
        message="Review Changes"
        description="Please review the price changes below before confirming. Red rows indicate potential issues (e.g. zero price)."
        type="warning"
        showIcon
      />
      <Table
        dataSource={data}
        rowKey="id"
        pagination={{ pageSize: 5 }}
        size="small"
        scroll={{ y: 300 }}
        columns={columns}
      />
    </Flex>
  );
}
