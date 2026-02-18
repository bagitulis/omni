import { useEffect, useMemo, useState } from "react";
import {
  Alert,
  Checkbox,
  InputNumber,
  Modal,
  Radio,
  Space,
  Table,
  Tag,
  theme,
} from "antd";
import type { FC } from "react";
import type { Platform, UnifiedProductRow } from "@/types/shared";

interface StockSyncModalProps {
  open: boolean;
  onClose: () => void;
  onSync: (
    items: Array<{
      seller_sku: string;
      stock: number;
      platforms: Platform[];
    }>,
  ) => Promise<void>;
  selectedProducts: UnifiedProductRow[];
}

type SyncMode = "uniform" | "per_platform";

interface PerPlatformConfig {
  [sku: string]: {
    stock: number;
    platforms: Record<Platform, boolean>;
  };
}

const PLATFORM_OPTIONS: Array<{ key: Platform; label: string }> = [
  { key: "shopee", label: "🟠 Shopee" },
  { key: "tiktok", label: "⬛ TikTok" },
  { key: "lazada", label: "🔵 Lazada" },
];

export const StockSyncModal: FC<StockSyncModalProps> = ({
  open,
  onClose,
  onSync,
  selectedProducts,
}) => {
  const { token } = theme.useToken();
  const [mode, setMode] = useState<SyncMode>("uniform");
  const [uniformStock, setUniformStock] = useState<number>(0);
  const [uniformPlatforms, setUniformPlatforms] = useState<
    Record<Platform, boolean>
  >({
    shopee: true,
    tiktok: true,
    lazada: true,
  });
  const [perPlatformConfig, setPerPlatformConfig] = useState<PerPlatformConfig>(
    {},
  );
  const [loading, setLoading] = useState(false);

  const linkedPlatformsBySku = useMemo(() => {
    const linkedMap: Record<string, Record<Platform, boolean>> = {};

    for (const product of selectedProducts) {
      const linkedPlatforms: Record<Platform, boolean> = {
        shopee: product.platform_summary.shopee !== "not_linked",
        tiktok: product.platform_summary.tiktok !== "not_linked",
        lazada: product.platform_summary.lazada !== "not_linked",
      };

      for (const sku of product.skus) {
        linkedMap[sku.seller_sku] = linkedPlatforms;
      }
    }

    return linkedMap;
  }, [selectedProducts]);

  useEffect(() => {
    if (!open) {
      return;
    }

    const config: PerPlatformConfig = {};
    for (const product of selectedProducts) {
      for (const sku of product.skus) {
        const linkedPlatforms = linkedPlatformsBySku[sku.seller_sku] ?? {
          shopee: false,
          tiktok: false,
          lazada: false,
        };

        config[sku.seller_sku] = {
          stock: sku.stock,
          platforms: { ...linkedPlatforms },
        };
      }
    }

    setPerPlatformConfig(config);
  }, [linkedPlatformsBySku, open, selectedProducts]);

  const syncItems = useMemo(() => {
    if (mode === "uniform") {
      const platforms = (
        Object.entries(uniformPlatforms) as [Platform, boolean][]
      )
        .filter(([, enabled]) => enabled)
        .map(([platform]) => platform);

      return selectedProducts.flatMap((product) =>
        product.skus.map((sku) => ({
          seller_sku: sku.seller_sku,
          stock: uniformStock,
          platforms: platforms.filter(
            (platform) => linkedPlatformsBySku[sku.seller_sku]?.[platform],
          ),
        })),
      );
    }

    return Object.entries(perPlatformConfig).map(([sku, config]) => ({
      seller_sku: sku,
      stock: config.stock,
      platforms: (Object.entries(config.platforms) as [Platform, boolean][])
        .filter(
          ([platform, enabled]) =>
            enabled && linkedPlatformsBySku[sku]?.[platform],
        )
        .map(([platform]) => platform),
    }));
  }, [
    linkedPlatformsBySku,
    mode,
    perPlatformConfig,
    selectedProducts,
    uniformPlatforms,
    uniformStock,
  ]);

  const validItems = useMemo(
    () =>
      syncItems.filter((item) => item.platforms.length > 0 && item.stock >= 0),
    [syncItems],
  );

  const handleSync = async () => {
    if (validItems.length === 0) {
      return;
    }

    setLoading(true);
    try {
      await onSync(validItems);
      onClose();
    } finally {
      setLoading(false);
    }
  };

  const perPlatformColumns = [
    {
      title: "SKU",
      dataIndex: "sku",
      key: "sku",
      width: 150,
    },
    {
      title: "Stock",
      dataIndex: "stock",
      key: "stock",
      width: 100,
      render: (_: unknown, record: { sku: string; stock: number }) => (
        <InputNumber
          size="small"
          min={0}
          value={perPlatformConfig[record.sku]?.stock ?? 0}
          onChange={(value) =>
            setPerPlatformConfig((prev) => {
              const existing = prev[record.sku];
              if (!existing) {
                return prev;
              }

              return {
                ...prev,
                [record.sku]: {
                  ...existing,
                  stock: value ?? 0,
                },
              };
            })
          }
          style={{ width: 80 }}
        />
      ),
    },
    ...PLATFORM_OPTIONS.map((platformOption) => ({
      title: platformOption.label,
      key: platformOption.key,
      width: 100,
      render: (_: unknown, record: { sku: string }) => (
        <Checkbox
          disabled={!linkedPlatformsBySku[record.sku]?.[platformOption.key]}
          checked={
            linkedPlatformsBySku[record.sku]?.[platformOption.key]
              ? (perPlatformConfig[record.sku]?.platforms[platformOption.key] ??
                false)
              : false
          }
          onChange={(event) =>
            setPerPlatformConfig((prev) => {
              const existing = prev[record.sku];
              if (!existing) {
                return prev;
              }

              if (!linkedPlatformsBySku[record.sku]?.[platformOption.key]) {
                return {
                  ...prev,
                  [record.sku]: {
                    ...existing,
                    platforms: {
                      ...existing.platforms,
                      [platformOption.key]: false,
                    },
                  },
                };
              }

              return {
                ...prev,
                [record.sku]: {
                  ...existing,
                  platforms: {
                    ...existing.platforms,
                    [platformOption.key]: event.target.checked,
                  },
                },
              };
            })
          }
        />
      ),
    })),
  ];

  const perPlatformData = Object.entries(perPlatformConfig).map(
    ([sku, config]) => ({
      key: sku,
      sku,
      stock: config.stock,
    }),
  );

  return (
    <Modal
      title="Sync Stock to Marketplaces"
      open={open}
      onCancel={onClose}
      onOk={handleSync}
      confirmLoading={loading}
      width={mode === "per_platform" ? 700 : 480}
      okText={`Sync ${validItems.length} SKUs`}
      okButtonProps={{
        disabled: validItems.length === 0,
      }}
    >
      <Radio.Group
        value={mode}
        onChange={(event) => setMode(event.target.value as SyncMode)}
        style={{ marginBottom: 16 }}
      >
        <Radio.Button value="uniform">Uniform (same stock all)</Radio.Button>
        <Radio.Button value="per_platform">
          Per Platform (different stock)
        </Radio.Button>
      </Radio.Group>

      {mode === "uniform" ? (
        <div>
          <Alert
            message="All selected SKUs will be synced with the same stock value to checked platforms."
            type="info"
            showIcon
            style={{ marginBottom: 16 }}
          />

          <div style={{ marginBottom: 16 }}>
            <div style={{ display: "block", marginBottom: 4, fontWeight: 500 }}>
              Stock Quantity
            </div>
            <InputNumber
              min={0}
              value={uniformStock}
              onChange={(value) => setUniformStock(value ?? 0)}
              style={{ width: 200 }}
            />
          </div>

          <div>
            <div style={{ display: "block", marginBottom: 4, fontWeight: 500 }}>
              Target Platforms
            </div>
            <Space>
              {PLATFORM_OPTIONS.map((platformOption) => (
                <Checkbox
                  key={platformOption.key}
                  checked={uniformPlatforms[platformOption.key]}
                  onChange={(event) =>
                    setUniformPlatforms((previous) => ({
                      ...previous,
                      [platformOption.key]: event.target.checked,
                    }))
                  }
                >
                  {platformOption.label}
                </Checkbox>
              ))}
            </Space>
          </div>

          <div style={{ marginTop: 16, color: token.colorTextSecondary }}>
            <Tag>
              {selectedProducts.reduce(
                (sum, product) => sum + product.skus.length,
                0,
              )}{" "}
              SKUs
            </Tag>
            ×{" "}
            <Tag>
              {Object.values(uniformPlatforms).filter(Boolean).length} platforms
            </Tag>
            =
            <Tag color="processing">
              {selectedProducts.reduce(
                (sum, product) => sum + product.skus.length,
                0,
              ) * Object.values(uniformPlatforms).filter(Boolean).length}{" "}
              sync operations
            </Tag>
          </div>
        </div>
      ) : (
        <div>
          <Alert
            message="Set stock and target platforms individually per SKU."
            type="info"
            showIcon
            style={{ marginBottom: 16 }}
          />
          <Table
            columns={perPlatformColumns}
            dataSource={perPlatformData}
            size="small"
            pagination={false}
            scroll={{ y: 300 }}
            bordered
          />
        </div>
      )}
    </Modal>
  );
};
