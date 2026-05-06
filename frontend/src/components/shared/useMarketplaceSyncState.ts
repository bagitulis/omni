import { useCallback, useEffect, useMemo, useState } from "react";
import { modal } from "@/components/AntStaticApi";
import type { Platform, UnifiedProductRow } from "@/types/shared";
import {
  type PricePerPlatformConfig,
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

// ─── Types ───────────────────────────────────────────────────────────────────

type SyncTab = "stock" | "price";
type SyncMode = "uniform" | "per_platform";

interface UseMarketplaceSyncStateParams {
  open: boolean;
  selectedProducts: UnifiedProductRow[];
  defaultTab: SyncTab;
  onStockSync: (items: Array<{ seller_sku: string; stock: number; platforms: Platform[] }>) => Promise<void>;
  onPriceSync: (items: Array<{ seller_sku: string; prices: Record<Platform, number>; platforms: Platform[] }>) => Promise<void>;
  onClose: () => void;
}

export function useMarketplaceSyncState({
  open,
  selectedProducts,
  defaultTab,
  onStockSync,
  onPriceSync,
  onClose,
}: UseMarketplaceSyncStateParams) {
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

  const productNameBySku = useMemo(() => {
    const map: Record<string, string> = {};
    for (const product of selectedProducts) {
      for (const sku of product.skus) {
        map[sku.seller_sku] = product.title;
      }
    }
    return map;
  }, [selectedProducts]);

  const currentPricesBySku = useMemo(() => {
    const map: Record<string, Record<Platform, number>> = {};
    for (const product of selectedProducts) {
      for (const sku of product.skus) {
        const prices: Record<Platform, number> = { shopee: 0, tiktok: 0, lazada: 0 };
        for (const pp of sku.platform_prices ?? []) {
          prices[pp.platform] = pp.platform_price;
        }
        map[sku.seller_sku] = prices;
      }
    }
    return map;
  }, [selectedProducts]);

  const currentStockBySku = useMemo(() => {
    const map: Record<string, Record<Platform, number>> = {};
    for (const product of selectedProducts) {
      for (const sku of product.skus) {
        const stocks: Record<Platform, number> = { shopee: 0, tiktok: 0, lazada: 0 };
        for (const pp of sku.platform_prices ?? []) {
          stocks[pp.platform] = pp.platform_stock;
        }
        map[sku.seller_sku] = stocks;
      }
    }
    return map;
  }, [selectedProducts]);

  // ─── Initialize on open ──────────────────────────────────────────────────

  useEffect(() => {
    if (!open) return;
    setStockPerPlatformConfig(
      getDefaultPerPlatformConfig(selectedProducts, linkedPlatformsBySku),
    );
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
      applyPriceRecommendations(prev, priceRecommendations, currentPricesBySku),
    );
  };

  // ─── Columns ─────────────────────────────────────────────────────────────

  const stockColumns = useMemo(
    () => getPerPlatformColumns(linkedPlatformsBySku, stockPerPlatformConfig, setStockPerPlatformConfig, stockRecommendations, productNameBySku, currentStockBySku),
    [linkedPlatformsBySku, stockPerPlatformConfig, stockRecommendations, productNameBySku, currentStockBySku],
  );

  const stockRows = useMemo(
    () => toPerPlatformRows(stockPerPlatformConfig),
    [stockPerPlatformConfig],
  );

  const priceColumns = useMemo(
    () => getPricePerPlatformColumns(linkedPlatformsBySku, pricePerPlatformConfig, setPricePerPlatformConfig, priceRecommendations, productNameBySku, currentPricesBySku),
    [linkedPlatformsBySku, pricePerPlatformConfig, priceRecommendations, productNameBySku, currentPricesBySku],
  );

  const priceRows = useMemo(
    () => Object.keys(pricePerPlatformConfig).map((sku) => ({ key: sku, sku })),
    [pricePerPlatformConfig],
  );

  // ─── Summary ─────────────────────────────────────────────────────────────

  const currentValidCount = activeTab === "stock" ? validStockItems.length : validPriceItems.length;
  const currentOpCount = activeTab === "stock" ? stockOperationCount : priceOperationCount;

  return {
    activeTab,
    setActiveTab,
    loading,
    handleSync,
    // Stock
    stockMode,
    setStockMode,
    stockRecoLoading,
    applyStockRecommendations,
    stockRecommendations,
    uniformStock,
    setUniformStock,
    stockUniformPlatforms,
    setStockUniformPlatforms,
    stockColumns,
    stockRows,
    lockedSkuCount,
    totalLockedQty,
    // Price
    priceMode,
    setPriceMode,
    priceRecoLoading,
    applyPriceRecos,
    priceRecommendations,
    uniformPrice,
    setUniformPrice,
    priceUniformPlatforms,
    setPriceUniformPlatforms,
    priceColumns,
    priceRows,
    // Summary
    currentValidCount,
    currentOpCount,
    validStockItems,
    validPriceItems,
  };
}
