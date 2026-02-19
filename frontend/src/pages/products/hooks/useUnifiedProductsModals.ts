import { useState } from "react";
import type { MasterProduct } from "@/types/product";
import type { UnifiedProductRow } from "@/types/shared";

/**
 * Manages all modal open/close state and modal-specific display state
 * for the unified products page.
 */
export function useUnifiedProductsModals() {
  const [stockSyncOpen, setStockSyncOpen] = useState(false);
  const [priceSyncOpen, setPriceSyncOpen] = useState(false);
  const [cloneModalOpen, setCloneModalOpen] = useState(false);
  const [batchCloneOpen, setBatchCloneOpen] = useState(false);
  const [skuMappingOpen, setSkuMappingOpen] = useState(false);
  const [wholesaleMpqOpen, setWholesaleMpqOpen] = useState(false);
  const [wholesaleMpqDefaultTab, setWholesaleMpqDefaultTab] = useState<
    "wholesale" | "mpq"
  >("wholesale");
  const [clonePreviewOpen, setClonePreviewOpen] = useState(false);
  const [stockSyncProducts, setStockSyncProducts] = useState<
    UnifiedProductRow[]
  >([]);
  const [priceSyncProducts, setPriceSyncProducts] = useState<
    UnifiedProductRow[]
  >([]);
  const [selectedProduct, setSelectedProduct] =
    useState<UnifiedProductRow | null>(null);
  const [skuMappingProduct, setSkuMappingProduct] =
    useState<MasterProduct | null>(null);
  const [skuMappingLoading, setSkuMappingLoading] = useState(false);

  return {
    stockSyncOpen,
    setStockSyncOpen,
    priceSyncOpen,
    setPriceSyncOpen,
    cloneModalOpen,
    setCloneModalOpen,
    batchCloneOpen,
    setBatchCloneOpen,
    skuMappingOpen,
    setSkuMappingOpen,
    wholesaleMpqOpen,
    setWholesaleMpqOpen,
    wholesaleMpqDefaultTab,
    setWholesaleMpqDefaultTab,
    clonePreviewOpen,
    setClonePreviewOpen,
    stockSyncProducts,
    setStockSyncProducts,
    priceSyncProducts,
    setPriceSyncProducts,
    selectedProduct,
    setSelectedProduct,
    skuMappingProduct,
    setSkuMappingProduct,
    skuMappingLoading,
    setSkuMappingLoading,
  };
}
