/**
 * useProductCreate Composable
 * Handles product creation for TikTok Shop, Shopee, and Lazada
 */

import { ref, computed } from "vue";
import api from "@/services/api";
import type {
  Category, CategoryAttribute, Brand, DeliveryOption,
  CreateProductForm, CreateProductResult, Platform, ProductSku
} from "./productCreateTypes";

export * from "./productCreateTypes";

function createDefaultForm(): CreateProductForm {
  return {
    title: "",
    description: "",
    categoryId: "",
    brandId: "",
    attributes: {},
    images: [],
    videoUrl: "",
    hasVariants: false,
    variantTypes: [],
    skus: [{ sellerSku: "", price: 0, stock: 0 }],
    packageWeight: 0,
    packageDimensions: { length: 0, width: 0, height: 0 },
    deliveryOptionIds: [],
    saveMode: "LISTING",
  };
}

export function useProductCreate(platform: Platform = "tiktok") {
  // State
  const loading = ref(false);
  const error = ref<string | null>(null);
  const categories = ref<Category[]>([]);
  const categoryAttributes = ref<CategoryAttribute[]>([]);
  const brands = ref<Brand[]>([]);
  const deliveryOptions = ref<DeliveryOption[]>([]);
  const recommendedCategory = ref<Category | null>(null);
  const popularCategories = ref<Category[]>([]);
  const form = ref<CreateProductForm>(createDefaultForm());

  // Computed
  const isFormValid = computed(() => {
    const f = form.value;
    return f.title && f.description && f.categoryId && 
           f.images.length > 0 && f.skus.length > 0 &&
           f.skus.every(s => s.sellerSku && s.price > 0) && f.packageWeight > 0;
  });

  // Helper for GET API calls
  async function apiGet<T>(url: string): Promise<T | null> {
    loading.value = true;
    error.value = null;
    try {
      const data = await api.get<T>(url);
      return data;
    } catch (err: any) {
      error.value = err.message || "API Error";
      return null;
    } finally {
      loading.value = false;
    }
  }

  // Helper for POST API calls
  async function apiPost<T>(url: string, payload: any): Promise<T | null> {
    loading.value = true;
    error.value = null;
    try {
      const data = await api.post<T>(url, payload);
      return data;
    } catch (err: any) {
      error.value = err.message || "API Error";
      return null;
    } finally {
      loading.value = false;
    }
  }

  // Category methods
  async function fetchCategories(parentId?: string) {
    const url = `/api/${platform}/categories${parentId ? `?parent_id=${parentId}` : ""}`;
    const data = await apiGet<{ success: boolean; categories: Category[] }>(url);
    categories.value = data?.categories || [];
  }

  async function searchCategories(keyword: string) {
    if (!keyword || keyword.length < 2) return;
    const url = `/api/${platform}/categories/search?keyword=${encodeURIComponent(keyword)}`;
    const data = await apiGet<{ success: boolean; categories: Category[] }>(url);
    categories.value = data?.categories || [];
  }

  async function fetchCategoryAttributes(categoryId: string) {
    const url = `/api/${platform}/categories/${categoryId}/attributes`;
    const data = await apiGet<{ success: boolean; attributes: CategoryAttribute[] }>(url);
    categoryAttributes.value = data?.attributes || [];
  }

  // Recommended & Popular categories
  async function fetchRecommendedCategory(title: string, description?: string) {
    if (!title || title.length < 3) return;
    const url = `/api/${platform}/categories/recommend`;
    const data = await apiPost<{ success: boolean; category: Category | null }>(
      url, { title, description }
    );
    recommendedCategory.value = data?.category || null;
  }

  async function fetchPopularCategories() {
    const url = `/api/${platform}/categories/popular`;
    const data = await apiGet<{ success: boolean; categories: Category[] }>(url);
    popularCategories.value = data?.categories || [];
  }

  // Brand & Delivery
  async function fetchBrands(categoryId: string) {
    const url = `/api/${platform}/brands?category_id=${categoryId}`;
    const data = await apiGet<{ success: boolean; brands: Brand[] }>(url);
    brands.value = data?.brands || [];
  }

  async function fetchDeliveryOptions() {
    const url = `/api/${platform}/delivery-options`;
    const data = await apiGet<{ success: boolean; deliveryOptions: DeliveryOption[] }>(url);
    deliveryOptions.value = data?.deliveryOptions || [];
  }

  // Image upload
  async function uploadImage(imageUrl: string): Promise<string | null> {
    const url = `/api/${platform}/images/upload`;
    const data = await apiPost<{ success: boolean; imageUrl: string }>(url, { image_url: imageUrl });
    return data?.imageUrl || null;
  }

  // Create product
  async function createProduct(): Promise<CreateProductResult> {
    const f = form.value;
    const url = `/api/${platform}/products/create`;
    
    const skuPayload = f.skus.map(s => ({
      seller_sku: s.sellerSku,
      price: s.price,
      stock: s.stock,
      variant_label: s.variantLabel,
    }));

    const payload = {
      title: f.title,
      description: f.description,
      category_id: f.categoryId,
      brand_id: f.brandId || undefined,
      attributes: f.attributes,
      images: f.images,
      video_url: f.videoUrl || undefined,
      has_variants: f.hasVariants,
      variant_types: f.variantTypes,
      skus: skuPayload,
      package_weight: f.packageWeight,
      package_dimensions: f.packageDimensions,
      delivery_option_ids: f.deliveryOptionIds,
      save_mode: f.saveMode,
    };

    const data = await apiPost<{ success: boolean; productId: string; message?: string }>(url, payload);

    return data?.success
      ? { success: true, productId: data.productId, message: data.message || "Created" }
      : { success: false, message: error.value || "Failed" };
  }

  // Form helpers
  const resetForm = () => { form.value = createDefaultForm(); error.value = null; };
  const addSku = () => form.value.skus.push({ sellerSku: "", price: 0, stock: 0 });
  const removeSku = (i: number) => { if (form.value.skus.length > 1) form.value.skus.splice(i, 1); };
  const addImage = (url: string) => { if (url && !form.value.images.includes(url)) form.value.images.push(url); };
  const removeImage = (i: number) => form.value.images.splice(i, 1);
  const reorderImages = (images: string[]) => { form.value.images = images; };

  // Variant helpers
  function generateSkusFromVariants(combinations: string[][]) {
    const newSkus: ProductSku[] = combinations.map((combo, idx) => ({
      sellerSku: `SKU-${idx + 1}`,
      price: 0,
      stock: 0,
      variantLabel: combo.join(" - "),
    }));
    form.value.skus = newSkus.length > 0 ? newSkus : [{ sellerSku: "", price: 0, stock: 0 }];
  }

  function bulkUpdateSkus(field: string, value: number) {
    form.value.skus.forEach(sku => {
      if (field === "price") sku.price = value;
      else if (field === "stock") sku.stock = value;
    });
  }

  return {
    loading, error, categories, categoryAttributes, brands, deliveryOptions,
    recommendedCategory, popularCategories, form, isFormValid,
    fetchCategories, searchCategories, fetchCategoryAttributes,
    fetchRecommendedCategory, fetchPopularCategories,
    fetchBrands, fetchDeliveryOptions,
    uploadImage, createProduct, resetForm, addSku, removeSku, addImage, removeImage, reorderImages,
    generateSkusFromVariants, bulkUpdateSkus,
  };
}
