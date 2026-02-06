import { ref, reactive } from "vue";
import { getAuthHeaders, getApiBaseUrl } from "@/utils/apiHeaders";

interface SkuCheckResult {
  lazada: boolean;
  shopee: boolean;
  tiktok: boolean;
  loading: boolean;
  error?: string;
}

// Cache untuk menghindari request berulang untuk SKU yang sama
const skuCheckCache = reactive<Record<string, SkuCheckResult>>({});

/**
 * Hook to check SKU availability on each platform
 * @param sku - SKU yang akan dicek
 * @returns Object berisi status untuk setiap platform
 */
export const usePlatformSkuCheck = (sku: string) => {
  // Return from cache if already exists
  if (skuCheckCache[sku]) {
    return ref(skuCheckCache[sku]);
  }

  const result = ref<SkuCheckResult>({
    lazada: false,
    shopee: false,
    tiktok: false,
    loading: true,
  });

  // Fetch dari setiap platform
  const checkSkuInPlatforms = async () => {
    try {
      result.value.loading = true;

      // Cek di Lazada
      const lazadaResult = await checkSkuInPlatform("lazada", sku);
      result.value.lazada = lazadaResult;

      // Cek di Shopee
      const shopeeResult = await checkSkuInPlatform("shopee", sku);
      result.value.shopee = shopeeResult;

      // Cek di TikTok
      const tiktokResult = await checkSkuInPlatform("tiktok", sku);
      result.value.tiktok = tiktokResult;

      result.value.loading = false;

      // Cache hasil
      skuCheckCache[sku] = result.value;
    } catch (error) {
      console.error("❌ Error checking SKU:", error);
      result.value.error = "Failed to check SKU availability";
      result.value.loading = false;
    }
  };

  // Auto-call function when composable is used
  checkSkuInPlatforms();

  return result;
};

/**
 * Check SKU availability on one platform
 * @param platform - Nama platform (lazada, shopee, tiktok)
 * @param sku - SKU yang akan dicek
 * @returns true jika SKU ditemukan, false jika tidak
 */
async function checkSkuInPlatform(
  platform: "lazada" | "shopee" | "tiktok",
  sku: string
): Promise<boolean> {
  try {
    const response = await fetch(
      `${getApiBaseUrl(`/${platform}/products`)}/check-sku?sku=${encodeURIComponent(sku)}`,
      {
        method: "GET",
        headers: getAuthHeaders(),
      }
    );

    if (!response.ok) {
      console.warn(`⚠️ Failed to check ${platform} SKU:`, response.status);
      return false;
    }

    const data = await response.json();
    // Support both legacy and new format
    // Legacy: { exists: true }
    // New: { success: true, data: { exists: true } }
    const responseData = data.data || data;
    return responseData.exists === true;
  } catch (error) {
    console.warn(`⚠️ Could not check SKU in ${platform}:`, error);
    return false;
  }
}

/**
 * Clear cache untuk memungkinkan refresh
 */
export const clearSkuCheckCache = () => {
  Object.keys(skuCheckCache).forEach((key) => {
    delete skuCheckCache[key];
  });
};

/**
 * Clear cache untuk satu SKU tertentu
 */
export const clearSkuCheckCacheForSku = (sku: string) => {
  delete skuCheckCache[sku];
};
