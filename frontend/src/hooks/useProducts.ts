import {
  useQuery,
  useMutation,
  useQueryClient,
  keepPreviousData,
} from "@tanstack/react-query";
import {
  getProducts,
  getProductById,
  createProduct,
  updateProduct,
  deleteProduct,
  syncProduct,
} from "@/api/products";
import type {
  ProductListFilter,
  CreateMasterProductInput,
  UpdateMasterProductInput,
  MasterProduct,
  Product,
} from "@/types/product";

// Helper to transform MasterProduct to Product for table compatibility
function transformToTableProduct(backendProduct: MasterProduct): Product {
  const prices = backendProduct.skus?.map((s) => s.price) || [];
  const price = prices.length > 0 ? Math.min(...prices) : 0;

  const stock = backendProduct.skus?.reduce((sum, s) => sum + s.stock, 0) || 0;

  return {
    ...backendProduct,
    item_id: String(backendProduct.id),
    item_name: backendProduct.title,
    item_sku: backendProduct.skus?.[0]?.seller_sku || `MP-${backendProduct.id}`,
    price: price,
    stock: stock,
    platform: "master",
    image_url: backendProduct.images?.[0] || "",
  };
}

export function useProducts(params: ProductListFilter) {
  return useQuery({
    queryKey: ["products", params],
    queryFn: async () => {
      const response = await getProducts(params);

      const rawData = response.data;
      const products = (Array.isArray(rawData) ? rawData : []).map(transformToTableProduct);

      return {
        products,
        total: response.meta?.total || products.length,
      };
    },
    placeholderData: keepPreviousData,
    staleTime: 30000,
  });
}

export function useProduct(id: string | undefined) {
  return useQuery({
    queryKey: ["product", id],
    queryFn: () => getProductById(id!),
    enabled: !!id,
    staleTime: 30000,
  });
}

export function useCreateProduct() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: (data: CreateMasterProductInput) => createProduct(data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["products"] });
    },
  });
}

export function useUpdateProduct() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: ({
      id,
      data,
    }: {
      id: string | number;
      data: UpdateMasterProductInput;
    }) => updateProduct(id, data),
    onSuccess: (data) => {
      queryClient.invalidateQueries({ queryKey: ["products"] });
      queryClient.invalidateQueries({ queryKey: ["product", String(data.id)] });
    },
  });
}

export function useDeleteProduct() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: (id: string | number) => deleteProduct(id),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["products"] });
    },
  });
}

export function useSyncProduct() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: ({
      id,
      platform,
    }: {
      id: string | number;
      platform?: string;
    }) => syncProduct(id, platform),
    onSuccess: (_, { id }) => {
      queryClient.invalidateQueries({ queryKey: ["product", String(id)] });
    },
  });
}
