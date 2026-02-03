import { ref } from "vue";

/**
 * Interface for any item that has a product image
 */
export interface ProductItem {
  sku?: string;
  product_name: string;
  product_image?: string;
  // allow other props
  [key: string]: any;
}

export function useProductImage(platform: string) {
  const imageErrors = ref<Set<string>>(new Set());

  const getItemKey = (item: ProductItem): string => {
    return item.sku || item.product_name || "unknown";
  };

  const getProductImage = (item: ProductItem): string | null => {
    const key = getItemKey(item);

    // Return null if image previously failed to load or no image provided
    if (imageErrors.value.has(key) || !item.product_image) {
      return null;
    }

    const url = item.product_image;

    // Shopee CDN Optimization
    // Format: https://cf.shopee.co.id/file/{image_id}
    if (platform === "shopee" && url.includes("cf.shopee")) {
      // If already has _tn suffix, use as-is
      if (url.includes("_tn")) {
        return url;
      }
      // Add _tn suffix for 60x60 thumbnail if not present
      // Check if it already has an extension
      if (/\.(jpg|jpeg|png|webp)$/i.test(url)) {
        return url.replace(/\.(jpg|jpeg|png|webp)$/i, "_tn.$1");
      }
      return `${url}_tn`;
    }

    // Lazada CDN Optimization (if applicable in future)
    if (platform === "lazada" && url.includes("lazada")) {
      // Lazada logic if known, currently return as is
      return url;
    }

    // TikTok CDN Optimization (if applicable in future)
    if (platform === "tiktok" && url.includes("tiktok")) {
      // TikTok logic if known
      return url;
    }

    return url;
  };

  const onImageError = (item: ProductItem) => {
    const key = getItemKey(item);
    imageErrors.value.add(key);
  };

  const resetErrors = () => {
    imageErrors.value.clear();
  };

  return {
    getProductImage,
    onImageError,
    resetErrors,
  };
}
