import { ref } from "vue";
import shopeeDbService from "@/services/shopeeDbService";

export function useProductData(itemId: number) {
  const loading = ref<boolean>(false);
  const productBase = ref<any>(null);
  const models = ref<any[]>([]);
  const variations = ref<any[]>([]);

  const fetchProductDetails = async () => {
    loading.value = true;
    try {
      const [baseResponse, modelsResponse, variationsResponse] = await Promise.all([
        shopeeDbService.getProductBase(itemId),
        shopeeDbService.getProductModels(itemId),
        shopeeDbService.getProductVariations(itemId),
      ]);

      if (baseResponse.success && baseResponse.data) {
        productBase.value = baseResponse.data;
      }
      if (modelsResponse.success && modelsResponse.data) {
        models.value = modelsResponse.data;
      }
      if (variationsResponse.success && variationsResponse.data) {
        variations.value = variationsResponse.data;
      }
    } catch (error) {
      console.error("Error fetching product details:", error);
    } finally {
      loading.value = false;
    }
  };

  const copyItemId = () => {
    navigator.clipboard.writeText(itemId.toString());
  };

  return {
    loading,
    productBase,
    models,
    variations,
    fetchProductDetails,
    copyItemId,
  };
}
