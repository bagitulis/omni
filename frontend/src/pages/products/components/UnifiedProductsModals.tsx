import {
  Drawer,
  Empty,
  InputNumber,
  Modal,
  Space,
  Spin,
  Typography,
} from "antd";
import { CloneBatchModal } from "@/components/clone/CloneBatchModal";
import { CloneProductModal } from "@/components/clone/CloneProductModal";
import { WholesaleMpqModal } from "@/components/shared/WholesaleMpqModal";
import { ClonePreviewDiff } from "@/components/shared/ClonePreviewDiff";
import { SkuMappingPanel } from "@/components/shared/SkuMappingPanel";
import { StockSyncModal } from "@/components/shared/StockSyncModal";
import type { MasterProduct, Product } from "@/types/product";
import type { Platform, UnifiedProductRow } from "@/types/shared";

interface UnifiedProductsModalsProps {
  selectedRecords: UnifiedProductRow[];
  selectedLegacyProducts: Product[];
  selectedProduct: UnifiedProductRow | null;
  selectedSkus: string[];
  wholesaleMpqDefaultTab: "wholesale" | "mpq";
  batchPriceValue: number;
  stockSyncOpen: boolean;
  wholesaleMpqOpen: boolean;
  batchPriceOpen: boolean;
  clonePreviewOpen: boolean;
  cloneModalOpen: boolean;
  batchCloneOpen: boolean;
  skuMappingOpen: boolean;
  skuMappingLoading: boolean;
  skuMappingProduct: MasterProduct | null;
  onStockSyncClose: () => void;
  onWholesaleMpqClose: () => void;
  onBatchPriceClose: () => void;
  onClonePreviewClose: () => void;
  onClonePreviewContinue: () => void;
  onCloneModalClose: () => void;
  onBatchCloneClose: () => void;
  onSkuMappingClose: () => void;
  onBatchPriceValueChange: (value: number) => void;
  onBatchPriceUpdate: () => Promise<void>;
  onStockSync: (
    items: Array<{
      seller_sku: string;
      stock: number;
      platforms: Platform[];
    }>,
  ) => Promise<void>;
  onSkuMappingUpdate: () => void;
}

export function UnifiedProductsModals({
  selectedRecords,
  selectedLegacyProducts,
  selectedProduct,
  selectedSkus,
  wholesaleMpqDefaultTab,
  batchPriceValue,
  stockSyncOpen,
  wholesaleMpqOpen,
  batchPriceOpen,
  clonePreviewOpen,
  cloneModalOpen,
  batchCloneOpen,
  skuMappingOpen,
  skuMappingLoading,
  skuMappingProduct,
  onStockSyncClose,
  onWholesaleMpqClose,
  onBatchPriceClose,
  onClonePreviewClose,
  onClonePreviewContinue,
  onCloneModalClose,
  onBatchCloneClose,
  onSkuMappingClose,
  onBatchPriceValueChange,
  onBatchPriceUpdate,
  onStockSync,
  onSkuMappingUpdate,
}: UnifiedProductsModalsProps) {
  return (
    <>
      <StockSyncModal
        open={stockSyncOpen}
        onClose={onStockSyncClose}
        onSync={onStockSync}
        selectedProducts={selectedRecords}
      />

      <WholesaleMpqModal
        open={wholesaleMpqOpen}
        onClose={onWholesaleMpqClose}
        selectedRecords={selectedRecords}
        defaultTab={wholesaleMpqDefaultTab}
      />

      <Modal
        title={`Batch Price Update (${selectedSkus.length} SKUs)`}
        open={batchPriceOpen}
        onCancel={onBatchPriceClose}
        onOk={() => {
          void onBatchPriceUpdate();
        }}
        okText="Update Price"
      >
        <Space direction="vertical" style={{ width: "100%" }}>
          <Typography.Text>Set new price for selected SKUs.</Typography.Text>
          <InputNumber
            style={{ width: "100%" }}
            min={1}
            value={batchPriceValue}
            onChange={(value) => onBatchPriceValueChange(value || 0)}
          />
        </Space>
      </Modal>

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
