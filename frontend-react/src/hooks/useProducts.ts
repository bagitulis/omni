import {
  useQuery,
  useMutation,
  useQueryClient,
  keepPreviousData,
} from "@tanstack/react-query";
import { getProducts, deleteProduct } from "@/api/products";
import { ProductListResponse } from "@/types/product";

interface UseProductsParams {
  page?: number;
  limit?: number;
  status?: string;
  search?: string;
  platform?: string;
}

export function useProducts(params?: UseProductsParams) {
  return useQuery<ProductListResponse>({
    queryKey: ["products", params],
    queryFn: () => getProducts(params),
    staleTime: 60 * 1000, // 1 minute
    placeholderData: keepPreviousData,
  });
}

export function useDeleteProduct() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (id: string) => deleteProduct(id),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["products"] });
    },
  });
}
