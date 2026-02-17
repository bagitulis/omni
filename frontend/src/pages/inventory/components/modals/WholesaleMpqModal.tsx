import { useMemo } from "react";
import { Alert, Modal, Tabs, Button, Typography } from "antd";
import { WholesaleTab } from "@/pages/inventory/components/WholesaleTab";
import { MpqTab } from "@/pages/inventory/components/MpqTab";
import type { InventoryRecord } from "@/types/inventory";
import { extractBulkPricingItems } from "@/pages/inventory/utils/bulkPricingItems";

interface WholesaleMpqModalProps {
  open: boolean;
  onClose: () => void;
  selectedRecords: InventoryRecord[];
}

export function WholesaleMpqModal({
  open,
  onClose,
  selectedRecords,
}: WholesaleMpqModalProps) {
  const { items: selectedItems, skipped_skus } = useMemo(
    () => extractBulkPricingItems(selectedRecords),
    [selectedRecords],
  );

  const items = [
    {
      key: "wholesale",
      label: "Wholesale",
      children: <WholesaleTab items={selectedItems} />,
    },
    {
      key: "mpq",
      label: "MPQ",
      children: <MpqTab items={selectedItems} />,
    },
  ];

  return (
    <Modal
      title="Bulk Pricing"
      open={open}
      onCancel={onClose}
      destroyOnClose
      width={900}
      footer={[
        <Button key="close" onClick={onClose}>
          Close
        </Button>,
      ]}
      styles={{ body: { height: "600px", overflowY: "auto" } }}
      centered
    >
      {selectedRecords.length === 0 ? (
        <Alert
          type="warning"
          showIcon
          style={{ marginBottom: 12 }}
          message="Select at least one inventory row first"
          description="Bulk Pricing uses SKU and price from the selected inventory rows."
        />
      ) : null}

      {skipped_skus.length > 0 ? (
        <Alert
          type="info"
          showIcon
          style={{ marginBottom: 12 }}
          message={`${skipped_skus.length} SKU skipped`}
          description={`Missing/invalid price for: ${skipped_skus.slice(0, 5).join(", ")}${skipped_skus.length > 5 ? "..." : ""}`}
        />
      ) : null}

      <Typography.Text
        type="secondary"
        style={{ display: "block", marginBottom: 12 }}
      >
        Valid items: {selectedItems.length}
      </Typography.Text>

      <Tabs defaultActiveKey="wholesale" items={items} />
    </Modal>
  );
}
