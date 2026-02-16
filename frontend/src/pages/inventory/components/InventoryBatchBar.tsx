import { useMemo, useState } from "react";
import {
  Badge,
  Button,
  Card,
  Grid,
  InputNumber,
  Modal,
  Space,
  Typography,
  message,
  theme,
} from "antd";
import {
  useCheckPlatformStatus,
  useUpdatePriceBatch,
  useUpdateStockBatch,
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
  const screens = Grid.useBreakpoint();
  const isMobile = !screens.md;

  const {
    token: { colorBgElevated, boxShadowSecondary, borderRadius },
  } = theme.useToken();

  const [priceModalOpen, setPriceModalOpen] = useState(false);
  const [priceValue, setPriceValue] = useState<number | null>(null);
  const [processing, setProcessing] = useState(false);

  const updateStockBatchMutation = useUpdateStockBatch();
  const updatePriceBatchMutation = useUpdatePriceBatch();
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
    try {
      await updateStockBatchMutation.mutateAsync({ skus: selectedSkus });
      message.success(`${selectedSkus.length} items synced to platforms`);
      onBatchComplete();
    } catch (error) {
      message.error((error as Error).message || "Failed to sync stock");
    } finally {
      setProcessing(false);
    }
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
    try {
      const result = await updatePriceBatchMutation.mutateAsync(
        selectedSkus.map((sku) => ({ sku, price: priceValue })),
      );

      if (result.failed > 0) {
        message.warning(
          `${result.successful} updated, ${result.failed} failed`,
        );
      } else {
        message.success(`${result.successful} items updated`);
      }

      setPriceModalOpen(false);
      setPriceValue(null);
      onBatchComplete();
    } catch (error) {
      message.error((error as Error).message || "Failed to update prices");
    } finally {
      setProcessing(false);
    }
  };

  const handleCheckPlatformStatus = async () => {
    setProcessing(true);
    try {
      const result =
        await checkPlatformStatusMutation.mutateAsync(selectedSkus);
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
          left: isMobile ? 8 : 24,
          right: isMobile ? 8 : 24,
          bottom: isMobile ? 8 : 16,
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
            flexDirection: isMobile ? "column" : "row",
            alignItems: isMobile ? "stretch" : "center",
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

          <Space wrap size={8} style={{ width: isMobile ? "100%" : "auto" }}>
            <Button
              onClick={() => {
                Modal.confirm({
                  title: "Sync Stock to Platforms",
                  content: `Sync inventory stock for ${selectedSkus.length} items to marketplace platforms. Stock values are read from your inventory records.`,
                  okText: "Sync",
                  onOk: handleSyncStock,
                });
              }}
              style={isMobile ? { flex: 1, minWidth: 120 } : undefined}
              disabled={processing}
            >
              Sync Stock
            </Button>
            <Button
              onClick={() => setPriceModalOpen(true)}
              style={isMobile ? { flex: 1, minWidth: 120 } : undefined}
              disabled={processing}
            >
              Update Price
            </Button>
            <Button
              onClick={handleCheckPlatformStatus}
              loading={processing}
              style={isMobile ? { flex: 1, minWidth: 120 } : undefined}
            >
              Check Platform Status
            </Button>
            <Button
              onClick={onClearSelection}
              style={isMobile ? { flex: 1, minWidth: 120 } : undefined}
              disabled={processing}
            >
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
