import {
  Alert,
  Button,
  Checkbox,
  Form,
  InputNumber,
  Modal,
  Radio,
  Space,
  Table,
  Typography,
} from "antd";
import { useEffect, useMemo, useState } from "react";
import type { FC } from "react";
import type {
  Platform,
  StockSyncMode,
  UnifiedProductRow,
} from "@/types/shared";

interface StockSyncPayloadItem {
  seller_sku: string;
  stock: number;
  platforms: string[];
}

export interface StockSyncModalProps {
  open: boolean;
  onClose: () => void;
  onSync: (items: StockSyncPayloadItem[]) => void;
  selectedProducts: UnifiedProductRow[];
}

const SUPPORTED_PLATFORMS: Platform[] = ["shopee", "lazada", "tiktok"];

export const StockSyncModal: FC<StockSyncModalProps> = ({
  open,
  onClose,
  onSync,
  selectedProducts,
}) => {
  const [mode, setMode] = useState<StockSyncMode>("uniform");
  const [uniformStock, setUniformStock] = useState<number | null>(null);
  const [uniformPlatforms, setUniformPlatforms] = useState<Platform[]>([]);

  // Keyed by seller_sku
  const [perPlatformData, setPerPlatformData] = useState<
    Record<string, { stock: number | null; platforms: Platform[] }>
  >({});

  // Reset state when modal opens
  useEffect(() => {
    if (open) {
      setMode("uniform");
      setUniformStock(null);

      // Default uniform platforms: all linked platforms across selected products
      const linkedPlatforms = new Set<Platform>();
      selectedProducts.forEach((p) => {
        Object.entries(p.platform_summary).forEach(([platform, status]) => {
          if (status === "linked") {
            linkedPlatforms.add(platform as Platform);
          }
        });
      });
      setUniformPlatforms(Array.from(linkedPlatforms));

      // Initialize per-platform data
      const initialData: Record<
        string,
        { stock: number | null; platforms: Platform[] }
      > = {};

      selectedProducts.forEach((p) => {
        p.skus.forEach((sku) => {
          const skuLinkedPlatforms = sku.platform_links
            .filter((link) => link.sync_status !== "error")
            .map((link) => link.platform);

          initialData[sku.seller_sku] = {
            stock: sku.stock,
            platforms: skuLinkedPlatforms,
          };
        });
      });
      setPerPlatformData(initialData);
    }
  }, [open, selectedProducts]);

  const handleSync = () => {
    const payload: StockSyncPayloadItem[] = [];

    if (mode === "uniform") {
      if (uniformStock === null || uniformPlatforms.length === 0) return;

      selectedProducts.forEach((p) => {
        p.skus.forEach((sku) => {
          const skuPlatforms = sku.platform_links.map((l) => l.platform);
          const validPlatforms = uniformPlatforms.filter((up) =>
            skuPlatforms.includes(up),
          );

          if (validPlatforms.length > 0) {
            payload.push({
              seller_sku: sku.seller_sku,
              stock: uniformStock,
              platforms: validPlatforms,
            });
          }
        });
      });
    } else {
      Object.entries(perPlatformData).forEach(([sku, data]) => {
        if (data.stock !== null && data.platforms.length > 0) {
          payload.push({
            seller_sku: sku,
            stock: data.stock,
            platforms: data.platforms,
          });
        }
      });
    }

    onSync(payload);
    onClose();
  };

  const isValid = useMemo(() => {
    if (mode === "uniform") {
      return (
        uniformStock !== null &&
        uniformStock >= 0 &&
        uniformPlatforms.length > 0
      );
    }
    return Object.values(perPlatformData).some(
      (d) => d.stock !== null && d.stock >= 0 && d.platforms.length > 0,
    );
  }, [mode, uniformStock, uniformPlatforms, perPlatformData]);

  const columns = [
    {
      title: "Product / SKU",
      dataIndex: "sku",
      key: "sku",
      render: (_: unknown, record: { seller_sku: string; title: string }) => (
        <Space direction="vertical" size={0}>
          <Typography.Text strong>{record.seller_sku}</Typography.Text>
          <Typography.Text type="secondary" style={{ fontSize: 12 }}>
            {record.title.substring(0, 30)}...
          </Typography.Text>
        </Space>
      ),
    },
    {
      title: "Stock",
      key: "stock",
      width: 120,
      render: (_: unknown, record: { seller_sku: string }) => (
        <InputNumber
          min={0}
          value={perPlatformData[record.seller_sku]?.stock}
          onChange={(val) => {
            setPerPlatformData((prev) => ({
              ...prev,
              [record.seller_sku]: {
                ...prev[record.seller_sku],
                stock: val,
              },
            }));
          }}
        />
      ),
    },
    {
      title: "Platforms",
      key: "platforms",
      render: (
        _: unknown,
        record: { seller_sku: string; available_platforms: Platform[] },
      ) => (
        <Checkbox.Group
          options={record.available_platforms.map((p) => ({
            label: p.toUpperCase(),
            value: p,
          }))}
          value={perPlatformData[record.seller_sku]?.platforms}
          onChange={(checkedValues) => {
            setPerPlatformData((prev) => ({
              ...prev,
              [record.seller_sku]: {
                ...prev[record.seller_sku],
                platforms: checkedValues as Platform[],
              },
            }));
          }}
        />
      ),
    },
  ];

  const flatData = useMemo(() => {
    return selectedProducts.flatMap((p) =>
      p.skus.map((s) => ({
        key: s.seller_sku,
        seller_sku: s.seller_sku,
        title: p.title,
        stock: s.stock,
        available_platforms: s.platform_links.map((l) => l.platform),
      })),
    );
  }, [selectedProducts]);

  return (
    <Modal
      title={`Sync Stock (${selectedProducts.length} products selected)`}
      open={open}
      onCancel={onClose}
      width={700}
      footer={[
        <Button key="cancel" onClick={onClose}>
          Cancel
        </Button>,
        <Button
          key="submit"
          type="primary"
          onClick={handleSync}
          disabled={!isValid}
        >
          Sync Stock
        </Button>,
      ]}
    >
      <Space direction="vertical" size="large" style={{ width: "100%" }}>
        <Radio.Group
          value={mode}
          onChange={(e) => setMode(e.target.value)}
          optionType="button"
          buttonStyle="solid"
        >
          <Radio.Button value="uniform">Uniform Sync</Radio.Button>
          <Radio.Button value="per_platform">Per-SKU Sync</Radio.Button>
        </Radio.Group>

        {mode === "uniform" ? (
          <Form layout="vertical">
            <Alert
              message="This will apply the same stock quantity to all selected SKUs on the selected platforms."
              type="info"
              showIcon
              style={{ marginBottom: 16 }}
            />

            <Form.Item label="New Stock Quantity" required>
              <InputNumber
                min={0}
                style={{ width: "100%" }}
                value={uniformStock}
                onChange={setUniformStock}
                placeholder="Enter stock quantity"
              />
            </Form.Item>

            <Form.Item label="Target Platforms" required>
              <Checkbox.Group
                options={SUPPORTED_PLATFORMS.map((p) => ({
                  label: p.toUpperCase(),
                  value: p,
                }))}
                value={uniformPlatforms}
                onChange={(vals) => setUniformPlatforms(vals as Platform[])}
              />
            </Form.Item>

            <div style={{ marginTop: 8 }}>
              <Typography.Text type="secondary" style={{ fontSize: 12 }}>
                Note: Only platforms actually linked to a SKU will be updated.
                Selecting a platform here does not link it if it's not already
                linked.
              </Typography.Text>
            </div>
          </Form>
        ) : (
          <Table
            dataSource={flatData}
            columns={columns}
            pagination={{ pageSize: 5 }}
            size="small"
            scroll={{ y: 300 }}
          />
        )}
      </Space>
    </Modal>
  );
};
