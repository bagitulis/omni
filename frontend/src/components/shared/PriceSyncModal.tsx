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
import {
  getDefaultPricePerPlatformConfig,
  getPricePerPlatformColumns,
  PRICE_PLATFORM_OPTIONS,
  type PricePerPlatformConfig,
  type PricePerPlatformRow,
} from "./priceSyncColumns";
import { PlatformComparisonPanel } from "./PlatformComparisonPanel";

interface PriceSyncModalProps {
  open: boolean;
  onClose: () => void;
  onSync: (
    items: Array<{
      seller_sku: string;
      price: number;
      platforms: Platform[];
    }>,
  ) => Promise<void>;
  selectedProducts: UnifiedProductRow[];
}

type SyncMode = "uniform" | "per_platform";

export const PriceSyncModal: FC<PriceSyncModalProps> = ({
  open,
  onClose,
  onSync,
  selectedProducts,
}) => {
  const { token } = theme.useToken();
  const [mode, setMode] = useState<SyncMode>("uniform");
  const [uniformPrice, setUniformPrice] = useState<number>(0);
  const [uniformPlatforms, setUniformPlatforms] = useState<
    Record<Platform, boolean>
  >({
    shopee: true,
    tiktok: true,
    lazada: true,
  });
  const [perPlatformConfig, setPerPlatformConfig] =
    useState<PricePerPlatformConfig>({});
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
    if (!open) return;

    const config = getDefaultPricePerPlatformConfig(
      selectedProducts,
      linkedPlatformsBySku,
    );
    setPerPlatformConfig(config);

    const firstPrice = selectedProducts[0]?.skus[0]?.price ?? 0;
    setUniformPrice(firstPrice);
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
          price: uniformPrice,
          platforms: platforms.filter(
            (platform) => linkedPlatformsBySku[sku.seller_sku]?.[platform],
          ),
        })),
      );
    }

    return Object.entries(perPlatformConfig).map(([sku, config]) => ({
      seller_sku: sku,
      price: config.price,
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
    uniformPrice,
  ]);

  const validItems = useMemo(
    () =>
      syncItems.filter((item) => item.platforms.length > 0 && item.price > 0),
    [syncItems],
  );

  const totalSelectedSkus = useMemo(
    () =>
      selectedProducts.reduce((sum, product) => sum + product.skus.length, 0),
    [selectedProducts],
  );

  const operationCount = useMemo(
    () => validItems.reduce((sum, item) => sum + item.platforms.length, 0),
    [validItems],
  );

  const handleSync = async () => {
    if (validItems.length === 0) return;

    setLoading(true);
    try {
      await onSync(validItems);
      onClose();
    } finally {
      setLoading(false);
    }
  };

  const perPlatformColumns = useMemo(
    () =>
      getPricePerPlatformColumns(
        linkedPlatformsBySku,
        perPlatformConfig,
        setPerPlatformConfig,
      ),
    [linkedPlatformsBySku, perPlatformConfig],
  );

  const perPlatformData: PricePerPlatformRow[] = Object.entries(
    perPlatformConfig,
  ).map(([sku, config]) => ({
    key: sku,
    sku,
    price: config.price,
  }));

  return (
    <Modal
      title="Sync Price to Marketplaces"
      open={open}
      onCancel={onClose}
      onOk={handleSync}
      confirmLoading={loading}
      width={mode === "per_platform" ? 720 : 480}
      okText={`Sync ${validItems.length} SKUs`}
      okButtonProps={{ disabled: validItems.length === 0 }}
    >
      <PlatformComparisonPanel products={selectedProducts} mode="price" />

      <Radio.Group
        value={mode}
        onChange={(event) => setMode(event.target.value as SyncMode)}
        style={{ marginBottom: 16 }}
      >
        <Radio.Button value="uniform">Uniform (same price all)</Radio.Button>
        <Radio.Button value="per_platform">
          Per Platform (different price)
        </Radio.Button>
      </Radio.Group>

      {mode === "uniform" ? (
        <div>
          <Alert
            type="info"
            showIcon
            style={{ marginBottom: 16 }}
            message="All selected SKUs will be synced with the same price value to checked platforms."
          />

          <div style={{ marginBottom: 16 }}>
            <div style={{ display: "block", marginBottom: 4, fontWeight: 500 }}>
              Price
            </div>
            <InputNumber
              min={0}
              value={uniformPrice}
              onChange={(value) => setUniformPrice(value ?? 0)}
              style={{ width: 220 }}
            />
          </div>

          <div>
            <div style={{ display: "block", marginBottom: 4, fontWeight: 500 }}>
              Target Platforms
            </div>
            <Space>
              {PRICE_PLATFORM_OPTIONS.map((platformOption) => (
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
            <Tag>{totalSelectedSkus} selected SKUs</Tag>
            <Tag>{validItems.length} ready SKUs</Tag>
            <Tag color="processing">{operationCount} sync operations</Tag>
          </div>
        </div>
      ) : (
        <div>
          <Alert
            message="Set price and target platforms individually per SKU."
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
