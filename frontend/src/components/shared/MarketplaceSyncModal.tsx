import { useCallback, useEffect, useMemo, useState } from "react";
import {
  Alert,
  Badge,
  Button,
  Checkbox,
  InputNumber,
  Modal,
  Radio,
  Space,
  Table,
  Tabs,
  Tag,
  Tooltip,
  Typography,
  theme,
} from "antd";
import { BulbOutlined } from "@ant-design/icons";
import type { ColumnType } from "antd/es/table";
import type { FC } from "react";
import { modal } from "@/components/AntStaticApi";
import type { Platform, UnifiedProductRow } from "@/types/shared";
import {
  type PricePerPlatformConfig,
  type PricePerPlatformRow,
  getDefaultPricePerPlatformConfig,
  getPricePerPlatformColumns,
} from "./priceSyncColumns";
import {
  applyPriceRecommendations,
  buildPriceRecommendations,
  type PriceRecommendationMap,
} from "./priceSyncRecommendations";
import {
  type PerPlatformConfig,
  type PerPlatformRow,
  PLATFORM_OPTIONS,
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
  type LockedStockMap,
} from "./stockSyncRecommendations";
import { PlatformComparisonPanel } from "./PlatformComparisonPanel";

// ─── Types ───────────────────────────────────────────────────────────────────

type SyncTab = "stock" | "price";
type SyncMode = "uniform" | "per_platform";

export interface MarketplaceSyncModalProps {
  open: boolean;
  onClose: () => void;
  onStockSync: (
    items: Array<{ seller_sku: string; stock: number; platforms: Platform[] }>,
  ) => Promise<void>;
  onPriceSync: (
    items: Array<{
      seller_sku: string;
      prices: Record<Platform, number>;
      platforms: Platform[];
    }>,
  ) => Promise<void>;
  selectedProducts: UnifiedProductRow[];
  defaultTab?: SyncTab;
}

// ─── Component ───────────────────────────────────────────────────────────────

export const MarketplaceSyncModal: FC<MarketplaceSyncModalProps> = ({
  open,
  onClose,
  onStockSync,
  onPriceSync,
  selectedProducts,
  defaultTab = "stock",
}) => {
  const { token } = theme.useToken();
  const [activeTab, setActiveTab] = useState<SyncTab>(defaultTab);
  const [loading, setLoading] = useState(false);

  // ─── Stock State ─────────────────────────────────────────────────────────
  const [stockMode, setStockMode] = useState<SyncMode>("per_platform");
  const [uniformStock, setUniformStock] = useState<number>(0);
  const [stockUniformPlatforms, setStockUniformPlatforms] = useState<
    Record<Platform, boolean>
  >({ shopee: true, tiktok: true, lazada: true });
  const [stockPerPlatformConfig, setStockPerPlatformConfig] =
    useState<PerPlatformConfig>({});
  const [stockRecommendations, setStockRecommendations] =
    useState<StockRecommendationMap>({});
  const [stockRecoLoading, setStockRecoLoading] = useState(false);
  const [lockedStockMap, setLockedStockMap] = useState<LockedStockMap>({});

  // ─── Price State ─────────────────────────────────────────────────────────
  const [priceMode, setPriceMode] = useState<SyncMode>("per_platform");
  const [uniformPrice, setUniformPrice] = useState<number>(0);
  const [priceUniformPlatforms, setPriceUniformPlatforms] = useState<
    Record<Platform, boolean>
  >({ shopee: true, tiktok: true, lazada: true });
  const [pricePerPlatformConfig, setPricePerPlatformConfig] =
    useState<PricePerPlatformConfig>({});
  const [priceRecommendations, setPriceRecommendations] =
    useState<PriceRecommendationMap>({});
  const [priceRecoLoading, setPriceRecoLoading] = useState(false);

  // ─── Shared Computed ─────────────────────────────────────────────────────

  const linkedPlatformsBySku = useMemo(
    () => buildLinkedPlatformsBySku(selectedProducts),
    [selectedProducts],
  );


  // ─── Initialize on open ──────────────────────────────────────────────────

  useEffect(() => {
    if (!open) return;
    // Stock defaults
    setStockPerPlatformConfig(
      getDefaultPerPlatformConfig(selectedProducts, linkedPlatformsBySku),
    );
    // Price defaults
    setPricePerPlatformConfig(
      getDefaultPricePerPlatformConfig(selectedProducts, linkedPlatformsBySku),
    );
    const firstPrice = selectedProducts[0]?.skus[0]?.price ?? 0;
    setUniformPrice(firstPrice);
  }, [open, selectedProducts, linkedPlatformsBySku]);

  // ─── Load Stock Recommendations ─────────────────────────────────────────

  useEffect(() => {
    if (!open) {
      setStockRecommendations({});
      setLockedStockMap({});
      return;
    }
    let mounted = true;
    setStockRecoLoading(true);
    void buildStockRecommendations(selectedProducts, linkedPlatformsBySku)
      .then((recs) => {
        if (!mounted) return;
        setStockRecommendations(recs);
        const locked: Record<string, number> = {};
        for (const [sku, rec] of Object.entries(recs)) {
          if (rec.lockedQty && rec.lockedQty > 0) locked[sku] = rec.lockedQty;
        }
        setLockedStockMap(locked);
      })
      .finally(() => {
        if (mounted) setStockRecoLoading(false);
      });
    return () => { mounted = false; };
  }, [open, selectedProducts, linkedPlatformsBySku]);

  // ─── Load Price Recommendations ─────────────────────────────────────────

  useEffect(() => {
    if (!open) {
      setPriceRecommendations({});
      return;
    }
    let mounted = true;
    setPriceRecoLoading(true);
    void buildPriceRecommendations(selectedProducts, linkedPlatformsBySku)
      .then((recs) => {
        if (!mounted) return;
        setPriceRecommendations(recs);
        const firstSku = selectedProducts[0]?.skus[0]?.seller_sku;
        if (firstSku && recs[firstSku]) {
          setUniformPrice(recs[firstSku].base_price);
        }
      })
      .finally(() => {
        if (mounted) setPriceRecoLoading(false);
      });
    return () => { mounted = false; };
  }, [open, selectedProducts, linkedPlatformsBySku]);

  // ─── Stock Sync Items ────────────────────────────────────────────────────

  const stockSyncItems = useMemo(
    () =>
      buildStockSyncItems(
        stockMode as StockSyncMode,
        selectedProducts,
        linkedPlatformsBySku,
        stockUniformPlatforms,
        uniformStock,
        stockPerPlatformConfig,
      ),
    [linkedPlatformsBySku, stockMode, selectedProducts, stockUniformPlatforms, uniformStock, stockPerPlatformConfig],
  );

  const validStockItems = useMemo(
    () => stockSyncItems.filter((i) => i.platforms.length > 0 && i.stock >= 0),
    [stockSyncItems],
  );

  const stockOperationCount = useMemo(
    () => validStockItems.reduce((sum, i) => sum + i.platforms.length, 0),
    [validStockItems],
  );

  // ─── Price Sync Items ────────────────────────────────────────────────────

  const priceSyncItems = useMemo(() => {
    if (priceMode === "uniform") {
      const platforms = (Object.entries(priceUniformPlatforms) as [Platform, boolean][])
        .filter(([, on]) => on)
        .map(([p]) => p);
      return selectedProducts.flatMap((product) =>
        product.skus.map((sku) => ({
          seller_sku: sku.seller_sku,
          prices: { shopee: uniformPrice, tiktok: uniformPrice, lazada: uniformPrice } as Record<Platform, number>,
          platforms: platforms.filter((p) => linkedPlatformsBySku[sku.seller_sku]?.[p]),
        })),
      );
    }
    return Object.entries(pricePerPlatformConfig).map(([sku, config]) => ({
      seller_sku: sku,
      prices: config.prices,
      platforms: (Object.entries(config.platforms) as [Platform, boolean][])
        .filter(([p, on]) => on && linkedPlatformsBySku[sku]?.[p])
        .map(([p]) => p),
    }));
  }, [priceMode, priceUniformPlatforms, uniformPrice, selectedProducts, linkedPlatformsBySku, pricePerPlatformConfig]);

  const validPriceItems = useMemo(
    () => priceSyncItems.filter((i) => i.platforms.length > 0 && i.platforms.some((p) => i.prices[p] > 0)),
    [priceSyncItems],
  );

  const priceOperationCount = useMemo(
    () => validPriceItems.reduce((sum, i) => sum + i.platforms.length, 0),
    [validPriceItems],
  );

  // ─── Handlers ────────────────────────────────────────────────────────────

  const lockedSkuCount = Object.keys(lockedStockMap).length;
  const totalLockedQty = Object.values(lockedStockMap).reduce((s, q) => s + q, 0);

  const handleSync = useCallback(async () => {
    setLoading(true);
    try {
      if (activeTab === "stock") {
        if (stockMode === "uniform" && uniformStock === 0) {
          await new Promise<void>((resolve, reject) => {
            modal.confirm({
              title: "Set stock to 0?",
              content: `This will set stock to 0 for ${validStockItems.length} SKUs. Products with 0 stock may be hidden from marketplace listings.`,
              okText: "Yes, set to 0",
              okType: "danger",
              zIndex: 1100,
              onOk: () => { resolve(); },
              onCancel: () => { reject(new Error("cancelled")); },
            });
          });
        }
        await onStockSync(validStockItems);
      } else {
        await onPriceSync(validPriceItems);
      }
      onClose();
    } catch (err) {
      if ((err as Error).message === "cancelled") return;
      throw err;
    } finally {
      setLoading(false);
    }
  }, [activeTab, stockMode, uniformStock, validStockItems, validPriceItems, onStockSync, onPriceSync, onClose]);

  const applyStockRecommendations = () => {
    setStockPerPlatformConfig((prev) =>
      applyRecommendationsToConfig(prev, stockRecommendations),
    );
  };

  const applyPriceRecos = () => {
    setPricePerPlatformConfig((prev) =>
      applyPriceRecommendations(prev, priceRecommendations),
    );
  };

  // ─── Stock Columns ───────────────────────────────────────────────────────

  const stockColumns = useMemo(
    () => getPerPlatformColumns(linkedPlatformsBySku, stockPerPlatformConfig, setStockPerPlatformConfig, stockRecommendations),
    [linkedPlatformsBySku, stockPerPlatformConfig, stockRecommendations],
  );

  const stockRows: PerPlatformRow[] = useMemo(
    () => toPerPlatformRows(stockPerPlatformConfig),
    [stockPerPlatformConfig],
  );

  // ─── Price Columns ───────────────────────────────────────────────────────

  const priceColumns = useMemo(
    () => getPricePerPlatformColumns(linkedPlatformsBySku, pricePerPlatformConfig, setPricePerPlatformConfig, priceRecommendations),
    [linkedPlatformsBySku, pricePerPlatformConfig, priceRecommendations],
  );

  const priceRows: PricePerPlatformRow[] = useMemo(
    () => Object.keys(pricePerPlatformConfig).map((sku) => ({ key: sku, sku })),
    [pricePerPlatformConfig],
  );

  // ─── Render Helpers ──────────────────────────────────────────────────────

  const renderModeToggle = (
    mode: SyncMode,
    setMode: (m: SyncMode) => void,
    recoLoading: boolean,
    onApply: () => void,
    recoCount: number,
    tooltipText: string,
  ) => (
    <div style={{ display: "flex", justifyContent: "space-between", alignItems: "center", marginBottom: 12 }}>
      <Radio.Group value={mode} onChange={(e) => setMode(e.target.value as SyncMode)}>
        <Radio.Button value="uniform">Uniform</Radio.Button>
        <Radio.Button value="per_platform">Per Platform</Radio.Button>
      </Radio.Group>
      {mode === "per_platform" && (
        <Tooltip title={tooltipText}>
          <Button
            icon={<BulbOutlined />}
            size="small"
            onClick={onApply}
            disabled={recoLoading || recoCount === 0}
            loading={recoLoading}
          >
            Fill from Inventory
          </Button>
        </Tooltip>
      )}
    </div>
  );

  const renderUniformSection = (
    value: number,
    onChange: (v: number) => void,
    platforms: Record<Platform, boolean>,
    onPlatformChange: (p: Platform, checked: boolean) => void,
    label: string,
  ) => (
    <div>
      <div style={{ marginBottom: 16 }}>
        <Typography.Text strong style={{ display: "block", marginBottom: 4 }}>
          {label}
        </Typography.Text>
        <InputNumber
          min={0}
          value={value}
          onChange={(v) => onChange(v ?? 0)}
          style={{ width: 200 }}
          formatter={(v) => v != null ? `${Number(v).toLocaleString("id-ID")}` : ""}
          parser={(v) => Number((v ?? "").replace(/\./g, "")) || 0}
        />
      </div>
      <div>
        <Typography.Text strong style={{ display: "block", marginBottom: 4 }}>
          Target Platforms
        </Typography.Text>
        <Space>
          {PLATFORM_OPTIONS.map((opt) => (
            <Checkbox
              key={opt.key}
              checked={platforms[opt.key]}
              onChange={(e) => onPlatformChange(opt.key, e.target.checked)}
            >
              {opt.label}
            </Checkbox>
          ))}
        </Space>
      </div>
    </div>
  );

  const renderPerPlatformTable = (
    columns: ColumnType<PerPlatformRow | PricePerPlatformRow>[],
    data: (PerPlatformRow | PricePerPlatformRow)[],
  ) => (
    <Table
      columns={columns}
      dataSource={data}
      size="small"
      pagination={false}
      scroll={{ y: 340 }}
      bordered
    />
  );

  // ─── Tab Items ───────────────────────────────────────────────────────────

  const stockTabContent = (
    <div>
      {lockedSkuCount > 0 && (
        <Alert
          type="warning"
          showIcon
          style={{ marginBottom: 12 }}
          message={
            <span>
              🔒 <strong>Lock Stock Active</strong> —{" "}
              <Tag color="error" style={{ borderRadius: 3 }}>{lockedSkuCount} SKUs</Tag>
              <Tag color="warning" style={{ borderRadius: 3 }}>{totalLockedQty} pcs locked</Tag>
              <Typography.Text type="secondary" style={{ fontSize: 12 }}>
                Recommendations auto-adjusted
              </Typography.Text>
            </span>
          }
        />
      )}
      {renderModeToggle(
        stockMode, setStockMode, stockRecoLoading, applyStockRecommendations,
        Object.keys(stockRecommendations).length,
        "Auto-fills from inventory. Locked orders deducted, allocated by ratio settings.",
      )}
      {stockMode === "uniform"
        ? renderUniformSection(uniformStock, setUniformStock, stockUniformPlatforms, (p, c) => setStockUniformPlatforms((prev) => ({ ...prev, [p]: c })), "Stock Quantity")
        : renderPerPlatformTable(stockColumns as ColumnType<PerPlatformRow | PricePerPlatformRow>[], stockRows)
      }
    </div>
  );

  const priceTabContent = (
    <div>
      {renderModeToggle(
        priceMode, setPriceMode, priceRecoLoading, applyPriceRecos,
        Object.keys(priceRecommendations).length,
        "Auto-fills per-platform prices from inventory sheet (HARGA_SHOPEE, HARGA_TIKTOK, HARGA_LAZADA).",
      )}
      {priceMode === "uniform"
        ? renderUniformSection(uniformPrice, setUniformPrice, priceUniformPlatforms, (p, c) => setPriceUniformPlatforms((prev) => ({ ...prev, [p]: c })), "Price")
        : renderPerPlatformTable(priceColumns as ColumnType<PerPlatformRow | PricePerPlatformRow>[], priceRows)
      }
    </div>
  );

  // ─── Summary ─────────────────────────────────────────────────────────────

  const currentValidCount = activeTab === "stock" ? validStockItems.length : validPriceItems.length;
  const currentOpCount = activeTab === "stock" ? stockOperationCount : priceOperationCount;

  // ─── Render ──────────────────────────────────────────────────────────────

  return (
    <Modal
      title="Push to Marketplace"
      open={open}
      onCancel={onClose}
      width={900}
      footer={
        <div style={{ display: "flex", justifyContent: "space-between", alignItems: "center" }}>
          <Typography.Text type="secondary" style={{ fontSize: 12 }}>
            {currentValidCount} SKUs • {currentOpCount} sync operations
          </Typography.Text>
          <Space>
            <Button onClick={onClose}>Cancel</Button>
            <Button
              type="primary"
              onClick={handleSync}
              loading={loading}
              disabled={currentValidCount === 0}
            >
              {activeTab === "stock" ? `Push Stock (${currentValidCount})` : `Push Prices (${currentValidCount})`}
            </Button>
          </Space>
        </div>
      }
    >
      <PlatformComparisonPanel
        products={selectedProducts}
        mode={activeTab === "stock" ? "stock" : "price"}
      />

      <Tabs
        activeKey={activeTab}
        onChange={(key) => setActiveTab(key as SyncTab)}
        items={[
          {
            key: "stock",
            label: (
              <span>
                📦 Stock{" "}
                <Badge
                  count={validStockItems.length}
                  size="small"
                  style={{ backgroundColor: validStockItems.length > 0 ? token.colorPrimary : token.colorTextQuaternary }}
                />
              </span>
            ),
            children: stockTabContent,
          },
          {
            key: "price",
            label: (
              <span>
                💰 Price{" "}
                <Badge
                  count={validPriceItems.length}
                  size="small"
                  style={{ backgroundColor: validPriceItems.length > 0 ? token.colorPrimary : token.colorTextQuaternary }}
                />
              </span>
            ),
            children: priceTabContent,
          },
        ]}
      />
    </Modal>
  );
};
