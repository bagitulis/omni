import { Drawer, Empty, Modal, Spin } from "antd";
import { CloneBatchModal } from "@/components/clone/CloneBatchModal";
import { CloneProductModal } from "@/components/clone/CloneProductModal";
import { PriceSyncModal } from "@/components/shared/PriceSyncModal";
import { WholesaleMpqModal } from "@/components/shared/WholesaleMpqModal";
import { ClonePreviewDiff } from "@/components/shared/ClonePreviewDiff";
import { SkuMappingPanel } from "@/components/shared/SkuMappingPanel";
import { StockSyncModal } from "@/components/shared/StockSyncModal";
import { MarketplaceSyncModal } from "@/components/shared/MarketplaceSyncModal";
import type { MasterProduct, Product } from "@/types/product";
import type { Platform, UnifiedProductRow } from "@/types/shared";

interface UnifiedProductsModalsProps {
  selectedRecords: UnifiedProductRow[];
  stockSyncProducts: UnifiedProductRow[];
  priceSyncProducts: UnifiedProductRow[];
  selectedLegacyProducts: Product[];
  selectedProduct: UnifiedProductRow | null;
  wholesaleMpqDefaultTab: "wholesale" | "mpq";
  stockSyncOpen: boolean;
  priceSyncOpen: boolean;
  wholesaleMpqOpen: boolean;
  clonePreviewOpen: boolean;
  cloneModalOpen: boolean;
  batchCloneOpen: boolean;
  skuMappingOpen: boolean;
  skuMappingLoading: boolean;
  skuMappingProduct: MasterProduct | null;
  onStockSyncClose: () => void;
  onPriceSyncClose: () => void;
  onWholesaleMpqClose: () => void;
  onClonePreviewClose: () => void;
  onClonePreviewContinue: () => void;
  onCloneModalClose: () => void;
  onBatchCloneClose: () => void;
  onSkuMappingClose: () => void;
  onPriceSync: (
    items: Array<{
      seller_sku: string;
      price: number;
      platforms: Platform[];
    }>,
  ) => Promise<void>;
  onStockSync: (
    items: Array<{
      seller_sku: string;
      stock: number;
      platforms: Platform[];
    }>,
  ) => Promise<void>;
  onSkuMappingUpdate: () => void;
  // New unified marketplace sync modal
  marketplaceSyncOpen: boolean;
  marketplaceSyncProducts: UnifiedProductRow[];
  onMarketplaceSyncClose: () => void;
  onPerPlatformPriceSync: (
    items: Array<{
      seller_sku: string;
      prices: Record<Platform, number>;
      platforms: Platform[];
    }>,
  ) => Promise<void>;
}

export function UnifiedProductsModals({
  selectedRecords,
  stockSyncProducts,
  priceSyncProducts,
  selectedLegacyProducts,
  selectedProduct,
  wholesaleMpqDefaultTab,
  stockSyncOpen,
  priceSyncOpen,
  wholesaleMpqOpen,
  clonePreviewOpen,
  cloneModalOpen,
  batchCloneOpen,
  skuMappingOpen,
  skuMappingLoading,
  skuMappingProduct,
  onStockSyncClose,
  onPriceSyncClose,
  onWholesaleMpqClose,
  onClonePreviewClose,
  onClonePreviewContinue,
  onCloneModalClose,
  onBatchCloneClose,
  onSkuMappingClose,
  onPriceSync,
  onStockSync,
  onSkuMappingUpdate,
  marketplaceSyncOpen,
  marketplaceSyncProducts,
  onMarketplaceSyncClose,
  onPerPlatformPriceSync,
}: UnifiedProductsModalsProps) {
  return (
    <>
      <StockSyncModal
        open={stockSyncOpen}
        onClose={onStockSyncClose}
        onSync={onStockSync}
        selectedProducts={stockSyncProducts}
      />

      <PriceSyncModal
        open={priceSyncOpen}
        onClose={onPriceSyncClose}
        onSync={onPriceSync}
        selectedProducts={priceSyncProducts}
      />

      <MarketplaceSyncModal
        open={marketplaceSyncOpen}
        onClose={onMarketplaceSyncClose}
        onStockSync={onStockSync}
        onPriceSync={onPerPlatformPriceSync}
        selectedProducts={marketplaceSyncProducts}
      />

      <WholesaleMpqModal
        open={wholesaleMpqOpen}
        onClose={onWholesaleMpqClose}
        selectedRecords={selectedRecords}
        defaultTab={wholesaleMpqDefaultTab}
      />

      <Modal
        title="Clone Preview"
        open={clonePreviewOpen}
        onCancel={onClonePreviewClose}
        onOk={onClonePreviewContinue}
        okText="Continue to Clone"
        width={1000}
      >
        {selectedProduct ? (
          <ClonePreviewDiff
            source={{
              title: selectedProduct.title,
              description: selectedProduct.description,
              price: selectedProduct.primary_price,
              stock: selectedProduct.primary_stock,
              images: selectedProduct.images,
              sku: selectedProduct.primary_sku,
            }}
            target={{
              platform: "shopee",
              title:
                selectedProduct.title.length > 120
                  ? selectedProduct.title.slice(0, 120)
                  : selectedProduct.title,
              description: selectedProduct.description,
              price: selectedProduct.primary_price,
              stock: selectedProduct.primary_stock,
              images: selectedProduct.images,
            }}
            differences={
              selectedProduct.title.length > 120
                ? [
                    {
                      field: "title",
                      source_value: selectedProduct.title,
                      target_value: selectedProduct.title.slice(0, 120),
                      reason: "Title truncated to 120 chars for Shopee",
                    },
                  ]
                : []
            }
            warnings={[]}
          />
        ) : (
          <Empty description="No product selected" />
        )}
      </Modal>

      <CloneProductModal
        open={cloneModalOpen}
        onClose={onCloneModalClose}
        initialSku={selectedProduct?.primary_sku || ""}
        initialPlatform="shopee"
      />

      <CloneBatchModal
        open={batchCloneOpen}
        onClose={onBatchCloneClose}
        products={selectedLegacyProducts}
      />

      <Drawer
        open={skuMappingOpen}
        onClose={onSkuMappingClose}
        title="SKU Mapping"
        width={960}
      >
        {skuMappingLoading ? (
          <Spin />
        ) : skuMappingProduct ? (
          <SkuMappingPanel
            masterProduct={skuMappingProduct}
            onUpdate={onSkuMappingUpdate}
          />
        ) : (
          <Empty description="Product not loaded" />
        )}
      </Drawer>
    </>
  );
}
