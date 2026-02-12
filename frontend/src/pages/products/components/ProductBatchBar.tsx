import { Button, Space, theme } from "antd";
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
  const { token } = theme.useToken();
  if (selectedRowKeys.length === 0) return null;

  return (
    <div
      style={{
        marginBottom: 16,
        padding: "8px 16px",
        background: token.colorFillQuaternary,
        border: `1px solid ${token.colorBorderSecondary}`,
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
