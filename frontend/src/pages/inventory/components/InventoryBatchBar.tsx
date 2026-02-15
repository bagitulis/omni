import { useMemo, useState } from "react";
import {
  Badge,
  Button,
  Card,
  InputNumber,
  Modal,
  Space,
  Typography,
  message,
  theme,
} from "antd";
import {
  useCheckPlatformStatus,
  useUpdatePrice,
  useUpdateStock,
} from "@/hooks/useInventory";
import type { InventoryRecord } from "@/types/inventory";

interface InventoryBatchBarProps {
  selectedRows: InventoryRecord[];
  onClearSelection: () => void;
  onBatchComplete: () => void;
}

export function InventoryBatchBar({
  selectedRows,
  onClearSelection,
  onBatchComplete,
}: InventoryBatchBarProps) {
  const {
    token: { colorBgElevated, boxShadowSecondary, borderRadius },
  } = theme.useToken();

  const [priceModalOpen, setPriceModalOpen] = useState(false);
  const [priceValue, setPriceValue] = useState<number | null>(null);
  const [processing, setProcessing] = useState(false);

  const updateStockMutation = useUpdateStock();
  const updatePriceMutation = useUpdatePrice();
  const checkPlatformStatusMutation = useCheckPlatformStatus();

  const selectedSkus = useMemo(
    () =>
      selectedRows
        .map((row) => row.key_value)
        .filter((sku): sku is string => Boolean(sku)),
    [selectedRows],
  );

  /** Sync stock from inventory_records to marketplace platforms */
  const handleSyncStock = async () => {
    if (selectedSkus.length === 0) {
      message.error("No valid SKU selected");
      return;
    }

    setProcessing(true);

    const updates = selectedSkus.map((sku) =>
      updateStockMutation.mutateAsync({ sku }),
    );

    const results = await Promise.allSettled(updates);
    const successCount = results.filter((r) => r.status === "fulfilled").length;
    const failedCount = results.length - successCount;

    if (failedCount > 0) {
      message.warning(`${successCount} synced, ${failedCount} failed`);
    } else {
      message.success(`${successCount} items synced to platforms`);
    }

    setProcessing(false);
    onBatchComplete();
  };

  const handlePriceUpdate = async () => {
    if (priceValue == null) {
      message.error("Please input a price value");
      return;
    }

    if (selectedSkus.length === 0) {
      message.error("No valid SKU selected");
      return;
    }

    setProcessing(true);

    const updates = selectedSkus.map((sku) =>
      updatePriceMutation.mutateAsync({ sku, price: priceValue }),
    );

    const results = await Promise.allSettled(updates);
    const successCount = results.filter((r) => r.status === "fulfilled").length;
    const failedCount = results.length - successCount;

    if (failedCount > 0) {
      message.warning(`${successCount} updated, ${failedCount} failed`);
    } else {
      message.success(`${successCount} items updated`);
    }

    setProcessing(false);
    setPriceModalOpen(false);
    setPriceValue(null);
    onBatchComplete();
  };

  const handleCheckPlatformStatus = async () => {
    setProcessing(true);
    try {
      const result = await checkPlatformStatusMutation.mutateAsync();
      message.success(`Platform check completed for ${result.length} records`);
      onBatchComplete();
    } catch (error) {
      message.error(
        (error as Error).message || "Failed to check platform status",
      );
    } finally {
      setProcessing(false);
    }
  };

  if (selectedRows.length === 0) {
    return null;
  }

  return (
    <>
      <div
        style={{
          position: "fixed",
          left: 24,
          right: 24,
          bottom: 16,
          zIndex: 900,
        }}
      >
        <Card
          size="small"
          style={{
            background: colorBgElevated,
            borderRadius,
            boxShadow: boxShadowSecondary,
          }}
          bodyStyle={{
            display: "flex",
            alignItems: "center",
            justifyContent: "space-between",
            gap: 12,
            padding: 12,
          }}
        >
          <Badge count={selectedRows.length}>
            <Typography.Text strong>
              {selectedRows.length} items selected
            </Typography.Text>
          </Badge>

          <Space wrap>
            <Button
              onClick={() => {
                Modal.confirm({
                  title: "Sync Stock to Platforms",
                  content: `Sync inventory stock for ${selectedSkus.length} items to marketplace platforms. Stock values are read from your inventory records.`,
                  okText: "Sync",
                  onOk: handleSyncStock,
                });
              }}
              disabled={processing}
            >
              Sync Stock
            </Button>
            <Button
              onClick={() => setPriceModalOpen(true)}
              disabled={processing}
            >
              Update Price
            </Button>
            <Button onClick={handleCheckPlatformStatus} loading={processing}>
              Check Platform Status
            </Button>
            <Button onClick={onClearSelection} disabled={processing}>
              Clear Selection
            </Button>
          </Space>
        </Card>
      </div>

      <Modal
        title="Batch Update Price"
        open={priceModalOpen}
        onCancel={() => setPriceModalOpen(false)}
        onOk={handlePriceUpdate}
        okText="Update"
        confirmLoading={processing}
      >
        <Space direction="vertical" style={{ width: "100%" }}>
          <Typography.Text type="secondary">
            Apply price value to {selectedSkus.length} selected items.
          </Typography.Text>
          <InputNumber
            min={0}
            value={priceValue}
            onChange={setPriceValue}
            style={{ width: "100%" }}
            placeholder="Enter price value"
          />
        </Space>
      </Modal>
    </>
  );
}
