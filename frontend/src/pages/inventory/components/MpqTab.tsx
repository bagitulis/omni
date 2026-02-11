import { useState, useMemo } from "react";
import type { Key } from "react";
import { Table, Button, Switch, Space, Empty } from "antd";
import { useInventory } from "@/hooks/useInventory";
import type { InventoryRecord } from "@/types/inventory";
import {
  useInventoryMpqSettings,
  useUpdateInventoryMpqSettings,
} from "@/hooks/useWholesale";
import { WholesaleMpqModal } from "@/components/modals/WholesaleMpqModal";

export function MpqTab() {
  const { data: inventory, isLoading: isInventoryLoading } = useInventory();
  const records = inventory?.records || [];
  const { data: mpqSettings, isLoading: isMpqLoading } =
    useInventoryMpqSettings();
  const { mutate: updateMpq } = useUpdateInventoryMpqSettings();

  const [selectedSkus, setSelectedSkus] = useState<string[]>([]);
  const [isModalOpen, setIsModalOpen] = useState(false);

  const mpqMap = useMemo(() => {
    const map = new Map();
    if (mpqSettings) {
      mpqSettings.forEach((setting) => {
        map.set(setting.sku, setting);
      });
    }
    return map;
  }, [mpqSettings]);

  const handleToggle = (sku: string, checked: boolean, currentQty: number) => {
    updateMpq([{ sku, min_purchase_qty: currentQty, enabled: checked }]);
  };

  const columns = [
    {
      title: "SKU",
      dataIndex: "key_value",
      key: "sku",
    },
    {
      title: "Min Purchase Qty",
      key: "mpq",
      render: (_: unknown, record: InventoryRecord) => {
        const setting = mpqMap.get(record.key_value);
        return setting ? setting.min_purchase_qty : "-";
      },
    },
    {
      title: "Status",
      key: "status",
      render: (_: unknown, record: InventoryRecord) => {
        const setting = mpqMap.get(record.key_value);
        const enabled = setting?.enabled || false;
        const qty = setting?.min_purchase_qty || 1;
        return (
          <Switch
            checked={enabled}
            onChange={(checked) => handleToggle(record.key_value, checked, qty)}
          />
        );
      },
    },
  ];

  return (
    <div style={{ display: "flex", flexDirection: "column", gap: 16 }}>
      <Space>
        <Button
          type="primary"
          disabled={selectedSkus.length === 0}
          onClick={() => setIsModalOpen(true)}
        >
          Batch Update MPQ
        </Button>
      </Space>

      <Table
        dataSource={records}
        columns={columns}
        rowKey="key_value"
        loading={isInventoryLoading || isMpqLoading}
        rowSelection={{
          selectedRowKeys: selectedSkus,
          onChange: (keys: Key[]) => setSelectedSkus(keys.map(String)),
        }}
        pagination={{ pageSize: 20 }}
        locale={{
          emptyText: (
            <Empty description="No inventory items available for MPQ configuration" />
          ),
        }}
      />

      <WholesaleMpqModal
        open={isModalOpen}
        onClose={() => setIsModalOpen(false)}
        selectedSkus={selectedSkus}
      />
    </div>
  );
}
