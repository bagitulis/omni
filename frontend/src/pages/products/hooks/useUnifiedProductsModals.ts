import { useState } from "react";
import type { MasterProduct } from "@/types/product";
import type { UnifiedProductRow } from "@/types/shared";

/**
 * Manages all modal open/close state and modal-specific display state
 * for the unified products page.
 */
export function useUnifiedProductsModals() {
  const [stockSyncOpen, setStockSyncOpen] = useState(false);
  const [cloneModalOpen, setCloneModalOpen] = useState(false);
  const [batchCloneOpen, setBatchCloneOpen] = useState(false);
  const [skuMappingOpen, setSkuMappingOpen] = useState(false);
  const [wholesaleMpqOpen, setWholesaleMpqOpen] = useState(false);
  const [wholesaleMpqDefaultTab, setWholesaleMpqDefaultTab] = useState<
    "wholesale" | "mpq"
  >("wholesale");
  const [batchPriceOpen, setBatchPriceOpen] = useState(false);
  const [clonePreviewOpen, setClonePreviewOpen] = useState(false);
  const [batchPriceValue, setBatchPriceValue] = useState(0);
  const [selectedProduct, setSelectedProduct] =
    useState<UnifiedProductRow | null>(null);
  const [skuMappingProduct, setSkuMappingProduct] =
    useState<MasterProduct | null>(null);
  const [skuMappingLoading, setSkuMappingLoading] = useState(false);

  return {
    stockSyncOpen,
    setStockSyncOpen,
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
    batchPriceOpen,
    setBatchPriceOpen,
    clonePreviewOpen,
    setClonePreviewOpen,
    batchPriceValue,
    setBatchPriceValue,
    selectedProduct,
    setSelectedProduct,
    skuMappingProduct,
    setSkuMappingProduct,
    skuMappingLoading,
    setSkuMappingLoading,
  };
}
