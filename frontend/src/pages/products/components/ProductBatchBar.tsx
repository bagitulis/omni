import { Button, Space } from "antd";
import { CopyOutlined } from "@ant-design/icons";

interface ProductBatchBarProps {
  selectedRowKeys: React.Key[];
  onBulkSync: () => void;
  onBatchSkuUpdate: () => void;
  onBatchClone: () => void;
  onDeleteSelected: () => void;
}

export function ProductBatchBar({
  selectedRowKeys,
  onBulkSync,
  onBatchSkuUpdate,
  onBatchClone,
  onDeleteSelected,
}: ProductBatchBarProps) {
  if (selectedRowKeys.length === 0) return null;

  return (
    <div
      style={{
        marginBottom: 16,
        padding: "8px 16px",
        background: "#e6f7ff",
        border: "1px solid #91d5ff",
        borderRadius: 4,
        display: "flex",
        alignItems: "center",
        justifyContent: "space-between",
      }}
    >
      <span>Selected {selectedRowKeys.length} items</span>
      <Space>
        <Button size="small" onClick={onBulkSync}>
          Bulk Sync
        </Button>
        <Button size="small" onClick={onBatchSkuUpdate}>
          Update SKUs
        </Button>
        <Button size="small" icon={<CopyOutlined />} onClick={onBatchClone}>
          Batch Clone
        </Button>
        <Button size="small" danger onClick={onDeleteSelected}>
          Delete Selected
        </Button>
      </Space>
    </div>
  );
}
