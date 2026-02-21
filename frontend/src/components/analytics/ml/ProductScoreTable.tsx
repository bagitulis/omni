import { Table, Tag, Progress, theme } from "antd";
import { ThunderboltOutlined } from "@ant-design/icons";
import type { ColumnsType } from "antd/es/table";
import dayjs from "dayjs";
import type { Product } from "./types";
import { getScoreColor, getScoreTag } from "./types";

const { useToken } = theme;

interface Props {
  products: Product[];
  loading: boolean;
  onRowClick?: (product: Product) => void;
}

export const ProductScoreTable = ({ products, loading, onRowClick }: Props) => {
  const { token } = useToken();

  const columns: ColumnsType<Product> = [
    {
      title: "SKU",
      dataIndex: "sku",
      key: "sku",
      width: 100,
      render: (text) => (
        <span style={{ fontSize: 12, fontWeight: 500 }}>{text}</span>
      ),
    },
    {
      title: "Product Name",
      dataIndex: "name",
      key: "name",
      width: 200,
      render: (text) => <span style={{ fontSize: 12 }}>{text}</span>,
    },
    {
      title: "Score",
      dataIndex: "score",
      key: "score",
      width: 120,
      render: (score) => (
        <div style={{ display: "flex", alignItems: "center", gap: 8 }}>
          <Progress
            type="circle"
            percent={score}
            width={40}
            strokeColor={getScoreColor(score)}
            format={(percent) => (
              <span style={{ fontSize: 10, fontWeight: 600 }}>{percent}</span>
            )}
          />
          <Tag color={getScoreTag(score)} style={{ fontSize: 11 }}>
            {score >= 80 ? "High" : score >= 50 ? "Medium" : "Low"}
          </Tag>
        </div>
      ),
    },
    {
      title: "AI Recommendation",
      dataIndex: "recommendation",
      key: "recommendation",
      width: 250,
      render: (text) => {
        if (!text) {
          return (
            <span
              style={{
                fontSize: 12,
                color: token.colorTextSecondary,
                fontStyle: "italic",
              }}
            >
              No recommendation available
            </span>
          );
        }
        return (
          <span style={{ fontSize: 12, lineHeight: 1.4 }}>
            <ThunderboltOutlined
              style={{ marginRight: 6, color: token.colorPrimary }}
            />
            {text}
          </span>
        );
      },
    },
    {
      title: "Last Updated",
      dataIndex: "last_updated",
      key: "last_updated",
      width: 100,
      render: (date) => {
        if (!date) {
          return (
            <span style={{ fontSize: 11, color: token.colorTextSecondary }}>
              —
            </span>
          );
        }
        return (
          <span style={{ fontSize: 11, color: token.colorTextSecondary }}>
            {dayjs(date).format("MMM DD, YYYY HH:mm")}
          </span>
        );
      },
    },
  ];

  return (
    <Table
      columns={columns}
      dataSource={products}
      rowKey="id"
      loading={loading}
      pagination={{
        defaultPageSize: 10,
        showSizeChanger: true,
        showTotal: (total) => (
          <span style={{ fontSize: 12, color: token.colorTextSecondary }}>
            {total} products
          </span>
        ),
      }}
      style={{ fontSize: 12 }}
      bordered
      onRow={(record) => ({
        onClick: () => onRowClick?.(record),
        style: { cursor: onRowClick ? "pointer" : "default" },
      })}
    />
  );
};
