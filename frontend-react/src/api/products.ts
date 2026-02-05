// import apiClient from "./client";
import { Product, ProductListResponse } from "@/types/product";

// Mock data generator
const generateMockProducts = (count: number): Product[] => {
  return Array.from({ length: count }).map((_, index) => ({
    item_id: `prod_${index + 1}`,
    item_name: `Product ${index + 1} - Sample Item`,
    item_sku: `SKU-${1000 + index}`,
    price: 10000 + index * 500,
    stock: index % 5 === 0 ? 0 : index % 3 === 0 ? 5 : 100, // Mix of out of stock, low stock, normal
    platform: index % 2 === 0 ? "shopee" : "tokopedia",
    status: index % 4 === 0 ? "inactive" : "active",
    image_url: `https://picsum.photos/seed/${index}/200`,
    created_at: new Date().toISOString(),
    updated_at: new Date().toISOString(),
  }));
};

const MOCK_PRODUCTS = generateMockProducts(50);

export async function getProducts(params?: {
  page?: number;
  limit?: number;
  status?: string;
  search?: string;
  platform?: string;
}): Promise<ProductListResponse> {
  // Simulate API delay
  await new Promise((resolve) => setTimeout(resolve, 800));

  let filtered = [...MOCK_PRODUCTS];

  if (params?.status && params.status !== "all" && params.status !== "") {
    filtered = filtered.filter((p) => p.status === params.status);
  }

  if (params?.platform && params.platform !== "all" && params.platform !== "") {
    filtered = filtered.filter((p) => p.platform === params.platform);
  }

  if (params?.search) {
    const q = params.search.toLowerCase();
    filtered = filtered.filter(
      (p) =>
        p.item_name.toLowerCase().includes(q) ||
        p.item_sku.toLowerCase().includes(q),
    );
  }

  const page = params?.page || 1;
  const limit = params?.limit || 10;
  const start = (page - 1) * limit;
  const end = start + limit;

  return {
    products: filtered.slice(start, end),
    total: filtered.length,
    page,
    page_size: limit,
  };
}

export async function deleteProduct(_id: string): Promise<boolean> {
  // Simulate API delay
  await new Promise((resolve) => setTimeout(resolve, 500));
  return true;
}
