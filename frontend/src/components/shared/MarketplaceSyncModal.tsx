import {
  Badge,
  Button,
  Modal,
  Space,
  Tabs,
  Typography,
  theme,
} from "antd";
import type { ColumnType } from "antd/es/table";
import type { FC } from "react";
import type { Platform, UnifiedProductRow } from "@/types/shared";
import type { PricePerPlatformRow } from "./priceSyncColumns";
import type { PerPlatformRow } from "./stockSyncColumns";
import { StockTabContent, PriceTabContent } from "./MarketplaceSyncTabContent";
import { useMarketplaceSyncState } from "./useMarketplaceSyncState";

// ─── Types ───────────────────────────────────────────────────────────────────

type SyncTab = "stock" | "price";

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

  const state = useMarketplaceSyncState({
    open,
    selectedProducts,
    defaultTab,
    onStockSync,
    onPriceSync,
    onClose,
  });

  // ─── Tab Content ──────────────────────────────────────────────────────────

  const stockTabContent = (
    <StockTabContent
      lockedSkuCount={state.lockedSkuCount}
      totalLockedQty={state.totalLockedQty}
      stockMode={state.stockMode}
      setStockMode={state.setStockMode}
      stockRecoLoading={state.stockRecoLoading}
      onApplyStockRecommendations={state.applyStockRecommendations}
      stockRecoCount={Object.keys(state.stockRecommendations).length}
      uniformStock={state.uniformStock}
      setUniformStock={state.setUniformStock}
      stockUniformPlatforms={state.stockUniformPlatforms}
      setStockUniformPlatforms={(p, c) => state.setStockUniformPlatforms((prev) => ({ ...prev, [p]: c }))}
      stockColumns={state.stockColumns as ColumnType<PerPlatformRow | PricePerPlatformRow>[]}
      stockRows={state.stockRows}
    />
  );

  const priceTabContent = (
    <PriceTabContent
      priceMode={state.priceMode}
      setPriceMode={state.setPriceMode}
      priceRecoLoading={state.priceRecoLoading}
      onApplyPriceRecommendations={state.applyPriceRecos}
      priceRecoCount={Object.keys(state.priceRecommendations).length}
      uniformPrice={state.uniformPrice}
      setUniformPrice={state.setUniformPrice}
      priceUniformPlatforms={state.priceUniformPlatforms}
      setPriceUniformPlatforms={(p, c) => state.setPriceUniformPlatforms((prev) => ({ ...prev, [p]: c }))}
      priceColumns={state.priceColumns as ColumnType<PerPlatformRow | PricePerPlatformRow>[]}
      priceRows={state.priceRows}
    />
  );

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
            {state.currentValidCount} SKUs • {state.currentOpCount} sync operations
          </Typography.Text>
          <Space>
            <Button onClick={onClose}>Cancel</Button>
            <Button
              type="primary"
              onClick={state.handleSync}
              loading={state.loading}
              disabled={state.currentValidCount === 0}
            >
              {state.activeTab === "stock" ? `Push Stock (${state.currentValidCount})` : `Push Prices (${state.currentValidCount})`}
            </Button>
          </Space>
        </div>
      }
    >

      <Tabs
        activeKey={state.activeTab}
        onChange={(key) => state.setActiveTab(key as SyncTab)}
        items={[
          {
            key: "stock",
            label: (
              <span>
                📦 Stock{" "}
                <Badge
                  count={state.validStockItems.length}
                  size="small"
                  style={{ backgroundColor: state.validStockItems.length > 0 ? token.colorPrimary : token.colorTextQuaternary }}
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
                  count={state.validPriceItems.length}
                  size="small"
                  style={{ backgroundColor: state.validPriceItems.length > 0 ? token.colorPrimary : token.colorTextQuaternary }}
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
