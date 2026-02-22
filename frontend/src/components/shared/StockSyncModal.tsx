import { useEffect, useMemo, useState, type FC } from "react";
import { Modal, Radio, theme } from "antd";
import type { Platform, UnifiedProductRow } from "@/types/shared";
import {
  type PerPlatformConfig,
  type PerPlatformRow,
  getDefaultPerPlatformConfig,
  getPerPlatformColumns,
} from "./stockSyncColumns";
import {
  applyRecommendationsToConfig,
  buildLinkedPlatformsBySku,
  buildStockSyncItems,
  type StockSyncMode,
  toPerPlatformRows,
} from "./stockSyncModalUtils";
import {
  buildStockRecommendations,
  type StockRecommendationMap,
} from "./stockSyncRecommendations";
import {
  PerPlatformStockSection,
  UniformStockSection,
} from "./stockSyncModalSections";
import { PlatformComparisonPanel } from "./PlatformComparisonPanel";

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

export const StockSyncModal: FC<StockSyncModalProps> = ({
  open,
  onClose,
  onSync,
  selectedProducts,
}) => {
  const { token } = theme.useToken();
  const [mode, setMode] = useState<StockSyncMode>("uniform");
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
  const [recommendations, setRecommendations] =
    useState<StockRecommendationMap>({});
  const [recommendationsLoading, setRecommendationsLoading] = useState(false);

  const linkedPlatformsBySku = useMemo(
    () => buildLinkedPlatformsBySku(selectedProducts),
    [selectedProducts],
  );

  useEffect(() => {
    if (!open) {
      return;
    }

    setPerPlatformConfig(
      getDefaultPerPlatformConfig(selectedProducts, linkedPlatformsBySku),
    );
  }, [linkedPlatformsBySku, open, selectedProducts]);

  useEffect(() => {
    if (!open) {
      setRecommendations({});
      return;
    }

    let isMounted = true;
    setRecommendationsLoading(true);

    void buildStockRecommendations(selectedProducts)
      .then((nextRecommendations) => {
        if (!isMounted) {
          return;
        }

        setRecommendations(nextRecommendations);
      })
      .finally(() => {
        if (!isMounted) {
          return;
        }

        setRecommendationsLoading(false);
      });

    return () => {
      isMounted = false;
    };
  }, [open, selectedProducts]);

  const syncItems = useMemo(
    () =>
      buildStockSyncItems(
        mode,
        selectedProducts,
        linkedPlatformsBySku,
        uniformPlatforms,
        uniformStock,
        perPlatformConfig,
      ),
    [
      linkedPlatformsBySku,
      mode,
      perPlatformConfig,
      selectedProducts,
      uniformPlatforms,
      uniformStock,
    ],
  );

  const validItems = useMemo(
    () =>
      syncItems.filter((item) => item.platforms.length > 0 && item.stock >= 0),
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

  const validSkuCount = useMemo(
    () => new Set(validItems.map((item) => item.seller_sku)).size,
    [validItems],
  );

  const skippedSkuCount = Math.max(0, totalSelectedSkus - validSkuCount);

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

  const applyInventoryRecommendations = () => {
    setPerPlatformConfig((previous) =>
      applyRecommendationsToConfig(
        previous,
        recommendations,
        linkedPlatformsBySku,
      ),
    );
  };

  const perPlatformColumns = useMemo(
    () =>
      getPerPlatformColumns(
        linkedPlatformsBySku,
        perPlatformConfig,
        setPerPlatformConfig,
        recommendations,
      ),
    [linkedPlatformsBySku, perPlatformConfig, recommendations],
  );

  const perPlatformData: PerPlatformRow[] = useMemo(
    () => toPerPlatformRows(perPlatformConfig),
    [perPlatformConfig],
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
      <PlatformComparisonPanel products={selectedProducts} mode="stock" />

      <Radio.Group
        value={mode}
        onChange={(event) => setMode(event.target.value as StockSyncMode)}
        style={{ marginBottom: 16 }}
      >
        <Radio.Button value="uniform">Uniform (same stock all)</Radio.Button>
        <Radio.Button value="per_platform">
          Per Platform (different stock)
        </Radio.Button>
      </Radio.Group>

      {mode === "uniform" ? (
        <UniformStockSection
          uniformStock={uniformStock}
          onUniformStockChange={setUniformStock}
          uniformPlatforms={uniformPlatforms}
          onUniformPlatformChange={(platform, checked) => {
            setUniformPlatforms((previous) => ({
              ...previous,
              [platform]: checked,
            }));
          }}
          skippedSkuCount={skippedSkuCount}
          totalSelectedSkus={totalSelectedSkus}
          validSkuCount={validItems.length}
          operationCount={operationCount}
          secondaryTextColor={token.colorTextSecondary}
        />
      ) : (
        <PerPlatformStockSection
          recommendationsLoading={recommendationsLoading}
          onApplyRecommendations={applyInventoryRecommendations}
          perPlatformColumns={perPlatformColumns}
          perPlatformData={perPlatformData}
        />
      )}
    </Modal>
  );
};
