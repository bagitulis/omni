/**
 * useProductClone Composable
 * Handles product cloning logic between platforms (Shopee ↔ Lazada ↔ TikTok)
 */

import { ref, computed } from "vue";
import api from "@/services/api";

export type Platform = "shopee" | "lazada" | "tiktok";

export interface CloneProductData {
  title: string;
  description: string;
  images: string[];
  skus: Array<{
    sellerSku: string;
    price: number;
    stock: number;
    variantLabel?: string;
  }>;
  weight?: number;
  packageDimensions?: {
    length: number;
    width: number;
    height: number;
  };
  categoryId?: string;
  brandId?: string;
  attributes?: Record<string, any>;
}

export interface CloneTargetsResponse {
  sku: string;
  sources: Platform[];
  targets: Platform[];
  status: {
    shopee: boolean;
    lazada: boolean;
    tiktok: boolean;
  };
}

export function useProductClone() {
  // State
  const showCloneModal = ref(false);
  const cloneSourceSku = ref("");
  const cloneSourcePlatform = ref<Platform | "">("");
  const cloneLoading = ref(false);
  const cloneError = ref<string | null>(null);
  const productData = ref<CloneProductData | null>(null);
  const availableTargets = ref<Platform[]>([]);
  const selectedTargets = ref<Platform[]>([]);

  // Computed
  const hasProductData = computed(() => !!productData.value);
  const canClone = computed(() => hasProductData.value && selectedTargets.value.length > 0);

  /**
   * Open clone modal for a given SKU and source platform
   */
  function openCloneModal(sku: string, sourcePlatform: Platform) {
    cloneSourceSku.value = sku;
    cloneSourcePlatform.value = sourcePlatform;
    showCloneModal.value = true;
    cloneError.value = null;
    productData.value = null;
    selectedTargets.value = [];

    // Fetch product data and available targets
    fetchCloneData(sku, sourcePlatform);
  }

  /**
   * Close clone modal and reset state
   */
  function closeCloneModal() {
    showCloneModal.value = false;
    cloneSourceSku.value = "";
    cloneSourcePlatform.value = "";
    productData.value = null;
    availableTargets.value = [];
    selectedTargets.value = [];
    cloneError.value = null;
  }

  /**
   * Fetch product data from source platform and available clone targets
   */
  async function fetchCloneData(sku: string, sourcePlatform: Platform) {
    cloneLoading.value = true;
    cloneError.value = null;

    try {
      const [productResponse, targetsResponse] = await Promise.all([
        api.get<{ success: boolean; product: CloneProductData }>(`/api/clone/product-data?platform=${sourcePlatform}&sku=${encodeURIComponent(sku)}`),
        api.get<{ success: boolean } & CloneTargetsResponse>(`/api/clone/available-targets?sku=${encodeURIComponent(sku)}`),
      ]);

      if (productResponse?.product) {
        productData.value = productResponse.product;
      } else {
        cloneError.value = "Product data not found";
      }

      if (targetsResponse?.targets) {
        availableTargets.value = targetsResponse.targets;
      }
    } catch (err: any) {
      cloneError.value = err.message || "Failed to fetch clone data";
    } finally {
      cloneLoading.value = false;
    }
  }

  /**
   * Toggle target platform selection
   */
  function toggleTarget(platform: Platform) {
    const idx = selectedTargets.value.indexOf(platform);
    if (idx >= 0) {
      selectedTargets.value.splice(idx, 1);
    } else {
      selectedTargets.value.push(platform);
    }
  }

  /**
   * Get pre-filled form data for AddProductModal
   */
  function getPrefilledFormData() {
    if (!productData.value) return null;

    const p = productData.value;
    return {
      title: p.title,
      description: p.description,
      images: p.images,
      skus: p.skus.map((s) => ({
        sellerSku: s.sellerSku,
        price: s.price,
        stock: s.stock,
        variantLabel: s.variantLabel,
      })),
      packageWeight: p.weight || 0,
      packageDimensions: p.packageDimensions || { length: 0, width: 0, height: 0 },
      hasVariants: p.skus.length > 1,
    };
  }

  /**
   * Get platform display name and color
   */
  function getPlatformInfo(platform: Platform) {
    const info: Record<Platform, { name: string; color: string; bgColor: string }> = {
      shopee: { name: "Shopee", color: "#f53d2d", bgColor: "#fff5f5" },
      lazada: { name: "Lazada", color: "#0f146d", bgColor: "#e0e0ff" },
      tiktok: { name: "TikTok", color: "#333", bgColor: "#e8e8e8" },
    };
    return info[platform];
  }

  return {
    // State
    showCloneModal,
    cloneSourceSku,
    cloneSourcePlatform,
    cloneLoading,
    cloneError,
    productData,
    availableTargets,
    selectedTargets,
    // Computed
    hasProductData,
    canClone,
    // Methods
    openCloneModal,
    closeCloneModal,
    fetchCloneData,
    toggleTarget,
    getPrefilledFormData,
    getPlatformInfo,
  };
}
