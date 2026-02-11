import { useState } from "react";
import type { Key } from "react";
import {
  Table,
  Button,
  Space,
  Popconfirm,
  message,
  Typography,
  Empty,
} from "antd";
import { useInventory } from "@/hooks/useInventory";
import { useBatchDeleteInventoryWholesale } from "@/hooks/useWholesale";

export function DeleteTab() {
  const { data, isLoading } = useInventory();
  const records = data?.records || [];
  const [selectedSkus, setSelectedSkus] = useState<string[]>([]);
  const { mutate: batchDelete, isPending } = useBatchDeleteInventoryWholesale();

  const handleDelete = () => {
    batchDelete(selectedSkus, {
      onSuccess: () => {
        message.success(
          `Deleted wholesale tiers for ${selectedSkus.length} items`,
        );
        setSelectedSkus([]);
      },
      onError: (error) => {
        message.error(`Failed to delete: ${error.message}`);
      },
    });
  };

  const columns = [
    {
      title: "SKU",
      dataIndex: "key_value",
      key: "sku",
    },
    {
      title: "Product Name",
      dataIndex: ["data", "Product Name"],
      key: "name",
    },
  ];

  return (
    <div style={{ display: "flex", flexDirection: "column", gap: 16 }}>
      <Typography.Text type="warning">
        Select items to remove their wholesale pricing configurations.
      </Typography.Text>
      <Space>
        <Popconfirm
          title="Are you sure?"
          description={`Delete wholesale settings for ${selectedSkus.length} items?`}
          onConfirm={handleDelete}
          disabled={selectedSkus.length === 0}
        >
          <Button
            type="primary"
            danger
            disabled={selectedSkus.length === 0}
            loading={isPending}
          >
            Delete Selected ({selectedSkus.length})
          </Button>
        </Popconfirm>
      </Space>

      <Table
        dataSource={records}
        columns={columns}
        rowKey="key_value"
        loading={isLoading}
        rowSelection={{
          selectedRowKeys: selectedSkus,
          onChange: (keys: Key[]) => setSelectedSkus(keys.map(String)),
        }}
        pagination={{ pageSize: 20 }}
        locale={{
          emptyText: (
            <Empty description="No inventory items available for delete actions" />
          ),
        }}
      />
    </div>
  );
}
