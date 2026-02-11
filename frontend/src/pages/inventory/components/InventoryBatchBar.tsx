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

  const [stockModalOpen, setStockModalOpen] = useState(false);
  const [priceModalOpen, setPriceModalOpen] = useState(false);
  const [stockValue, setStockValue] = useState<number | null>(null);
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

  const runBatchUpdate = async (
    mode: "stock" | "price",
    value: number,
    onDone: () => void,
  ) => {
    if (selectedSkus.length === 0) {
      message.error("No valid SKU selected");
      return;
    }

    setProcessing(true);

    const updates = selectedSkus.map((sku) => {
      if (mode === "stock") {
        return updateStockMutation.mutateAsync({ sku, stock: value });
      }
      return updatePriceMutation.mutateAsync({ sku, price: value });
    });

    const results = await Promise.allSettled(updates);
    const successCount = results.filter(
      (result) => result.status === "fulfilled",
    ).length;
    const failedCount = results.length - successCount;

    if (failedCount > 0) {
      message.warning(`${successCount} updated, ${failedCount} failed`);
    } else {
      message.success(`${successCount} items updated`);
    }

    setProcessing(false);
    onDone();
    onBatchComplete();
  };

  const handleStockUpdate = async () => {
    if (stockValue == null) {
      message.error("Please input a stock value");
      return;
    }
    await runBatchUpdate("stock", stockValue, () => {
      setStockModalOpen(false);
      setStockValue(null);
    });
  };

  const handlePriceUpdate = async () => {
    if (priceValue == null) {
      message.error("Please input a price value");
      return;
    }
    await runBatchUpdate("price", priceValue, () => {
      setPriceModalOpen(false);
      setPriceValue(null);
    });
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
              onClick={() => setStockModalOpen(true)}
              disabled={processing}
            >
              Update Stock
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
        title="Batch Update Stock"
        open={stockModalOpen}
        onCancel={() => setStockModalOpen(false)}
        onOk={handleStockUpdate}
        okText="Update"
        confirmLoading={processing}
      >
        <Space direction="vertical" style={{ width: "100%" }}>
          <Typography.Text type="secondary">
            Apply stock value to {selectedSkus.length} selected items.
          </Typography.Text>
          <InputNumber
            min={0}
            value={stockValue}
            onChange={setStockValue}
            style={{ width: "100%" }}
            placeholder="Enter stock value"
          />
        </Space>
      </Modal>

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
