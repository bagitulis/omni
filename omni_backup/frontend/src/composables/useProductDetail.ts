import { ref } from "vue";
import { getAuthHeaders } from "@/utils/apiHeaders";

interface ProductDetail {
  itemId: number;
  name: string;
  brand?: string;
  status: string;
  description?: string;
  skus?: Array<{
    sellerSku: string;
    variantName?: string;
    price: number;
    quantity: number;
  }>;
}

export function useProductDetail(apiBaseUrl: string) {
  const detailItemId = ref("");
  const detailLoading = ref(false);
  const selectedProduct = ref<ProductDetail | null>(null);

  async function fetchProductDetail() {
    if (!detailItemId.value) return;
    detailLoading.value = true;
    try {
      const response = await fetch(`${apiBaseUrl}/product/${detailItemId.value}`, {
        method: "GET",
        headers: getAuthHeaders(),
      });
      const data = await response.json();
      if (data.success) {
        selectedProduct.value = data.product;
        return { success: true };
      } else {
        return { success: false, error: data.error || "Product not found" };
      }
    } catch {
      return { success: false, error: "Error fetching product" };
    } finally {
      detailLoading.value = false;
    }
  }

  function clearProductDetail() {
    selectedProduct.value = null;
    detailItemId.value = "";
  }

  return {
    detailItemId,
    detailLoading,
    selectedProduct,
    fetchProductDetail,
    clearProductDetail,
  };
}
