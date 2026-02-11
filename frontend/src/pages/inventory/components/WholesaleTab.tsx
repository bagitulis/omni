import { useState } from "react";
import { Table, Button, Space, Tag } from "antd";
import { useInventory } from "@/hooks/useInventory";
import type { InventoryRecord } from "@/types/inventory";
import { WholesaleUpdateModal } from "@/components/modals/WholesaleUpdateModal";
import { WholesaleBatchDeleteModal } from "@/components/modals/WholesaleBatchDeleteModal";
import { useInventoryWholesaleTiers } from "@/hooks/useWholesale";

function WholesaleTiersList({ sku }: { sku: string }) {
  const { data: tiers, isLoading } = useInventoryWholesaleTiers(sku);

  if (isLoading) return <span>Loading...</span>;
  if (!tiers || tiers.length === 0)
    return <span style={{ color: "#999" }}>No tiers</span>;

  return (
    <Space direction="vertical" size="small">
      {tiers.map((tier, index) => (
        <Tag key={index} color="blue">
          Qty {tier.min_qty}+ : ${tier.price}
        </Tag>
      ))}
    </Space>
  );
}

export function WholesaleTab() {
  const { data, isLoading } = useInventory();
  const [selectedSkus, setSelectedSkus] = useState<string[]>([]);
  const [isUpdateModalOpen, setIsUpdateModalOpen] = useState(false);
  const [isDeleteModalOpen, setIsDeleteModalOpen] = useState(false);

  const columns = [
    {
      title: "Product Name",
      dataIndex: ["data", "Product Name"],
      key: "name",
      render: (text: string) => text || "Unknown Product",
    },
    {
      title: "SKU",
      dataIndex: "key_value",
      key: "sku",
    },
    {
      title: "Wholesale Tiers",
      key: "tiers",
      render: (_: unknown, record: InventoryRecord) => (
        <WholesaleTiersList sku={record.key_value} />
      ),
    },
  ];

  const rowSelection = {
    selectedRowKeys: selectedSkus,
    onChange: (selectedRowKeys: React.Key[]) => {
      setSelectedSkus(selectedRowKeys as string[]);
    },
  };

  return (
    <div style={{ display: "flex", flexDirection: "column", gap: 16 }}>
      <Space>
        <Button
          type="primary"
          disabled={selectedSkus.length === 0}
          onClick={() => setIsUpdateModalOpen(true)}
        >
          Batch Update Wholesale
        </Button>
        <Button
          danger
          disabled={selectedSkus.length === 0}
          onClick={() => setIsDeleteModalOpen(true)}
        >
          Batch Delete Wholesale
        </Button>
      </Space>

      <Table
        dataSource={data?.records || []}
        columns={columns}
        rowKey="key_value"
        loading={isLoading}
        rowSelection={rowSelection}
        pagination={{ pageSize: 20 }}
      />

      <WholesaleUpdateModal
        open={isUpdateModalOpen}
        onClose={() => setIsUpdateModalOpen(false)}
        selectedSkus={selectedSkus}
      />

      <WholesaleBatchDeleteModal
        open={isDeleteModalOpen}
        onClose={() => setIsDeleteModalOpen(false)}
        selectedSkus={selectedSkus}
      />
    </div>
  );
}
