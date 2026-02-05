import { Table, Tag, Tooltip, Empty } from "antd";
import {
  ExclamationCircleOutlined,
  CheckCircleOutlined,
} from "@ant-design/icons";
import type { ColumnsType } from "antd/es/table";
import type { ImportRow } from "@/types/product";
import "./ImportPreviewTable.css";

interface ImportPreviewTableProps {
  data: ImportRow[];
  loading?: boolean;
}

export function ImportPreviewTable({
  data,
  loading = false,
}: ImportPreviewTableProps) {
  const columns: ColumnsType<ImportRow> = [
    {
      title: "Row",
      dataIndex: "row_number",
      key: "row_number",
      width: 60,
      align: "center",
    },
    {
      title: "Status",
      dataIndex: "valid",
      key: "status",
      width: 80,
      align: "center",
      render: (valid: boolean) =>
        valid ? (
          <Tag icon={<CheckCircleOutlined />} color="success">
            Valid
          </Tag>
        ) : (
          <Tag icon={<ExclamationCircleOutlined />} color="error">
            Invalid
          </Tag>
        ),
    },
    {
      title: "Item Name",
      dataIndex: "item_name",
      key: "item_name",
      ellipsis: true,
    },
    {
      title: "SKU",
      dataIndex: "item_sku",
      key: "item_sku",
      width: 120,
    },
    {
      title: "Price",
      dataIndex: "price",
      key: "price",
      width: 100,
      render: (price: number) => `Rp ${price.toLocaleString("id-ID")}`,
    },
    {
      title: "Stock",
      dataIndex: "stock",
      key: "stock",
      width: 80,
      align: "center",
    },
    {
      title: "Errors",
      dataIndex: "errors",
      key: "errors",
      width: 200,
      render: (errors: string[]) => {
        if (errors.length === 0) return "-";
        return (
          <Tooltip title={errors.join("\n")}>
            <span className="error-badge">{errors.length} error(s)</span>
          </Tooltip>
        );
      },
    },
  ];

  if (data.length === 0) {
    return (
      <div className="import-preview-table">
        <Empty description="No data to preview. Upload a file to get started." />
      </div>
    );
  }

  return (
    <div className="import-preview-table">
      <Table
        columns={columns}
        dataSource={data}
        rowKey="row_number"
        loading={loading}
        pagination={{
          pageSize: 10,
          showSizeChanger: true,
          showTotal: (total) => `Total ${total} rows`,
        }}
        scroll={{ x: 1000 }}
        rowClassName={(record) => {
          if (!record.valid) return "row-invalid";
          return "row-valid";
        }}
        size="small"
      />
    </div>
  );
}
