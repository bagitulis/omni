import apiService from "./api";

class ShopeeDbService {
  async getProductList(
    limit: number = 100,
    offset: number = 0,
    status?: string | null
  ): Promise<any> {
    try {
      const params: Record<string, any> = { limit, offset };
      if (status) params.status = status;

      const response = await apiService.get("/shopee/db/products/list", {
        params,
      });
      return response;
    } catch (error) {
      console.error("❌ Error fetching product list:", error);
      throw error;
    }
  }

  async getProductBaseList(
    limit: number = 100,
    offset: number = 0
  ): Promise<any> {
    try {
      const params = { limit, offset };
      const response = await apiService.get("/shopee/db/products/base", {
        params,
      });
      return response;
    } catch (error) {
      console.error("❌ Error fetching product base list:", error);
      throw error;
    }
  }

  async getProductModelList(
    limit: number = 100,
    offset: number = 0
  ): Promise<any> {
    try {
      const params = { limit, offset };
      const response = await apiService.get("/shopee/db/products/model", {
        params,
      });
      return response;
    } catch (error) {
      console.error("❌ Error fetching product model list:", error);
      throw error;
    }
  }

  async getProductBase(itemId: number | string): Promise<any> {
    try {
      const response = await apiService.get(
        `/shopee/db/products/base/${itemId}`
      );
      return response;
    } catch (error) {
      console.error(`❌ Error fetching product base for ${itemId}:`, error);
      throw error;
    }
  }

  async getProductModels(itemId: number | string): Promise<any> {
    try {
      const response = await apiService.get(
        `/shopee/db/products/models/${itemId}`
      );
      return response;
    } catch (error) {
      console.error(`❌ Error fetching product models for ${itemId}:`, error);
      throw error;
    }
  }

  async getProductVariations(itemId: number | string): Promise<any> {
    try {
      const response = await apiService.get(
        `/shopee/db/products/variations/${itemId}`
      );
      return response;
    } catch (error) {
      console.error(
        `❌ Error fetching product variations for ${itemId}:`,
        error
      );
      throw error;
    }
  }

  async getProductFull(itemId: number | string): Promise<any> {
    try {
      const response = await apiService.get(
        `/shopee/db/products/full/${itemId}`
      );
      return response;
    } catch (error) {
      console.error(`❌ Error fetching full product for ${itemId}:`, error);
      throw error;
    }
  }

  async searchBySku(sku: string): Promise<any> {
    try {
      if (!sku || sku.trim() === "") {
        throw new Error("SKU cannot be empty");
      }

      const response = await apiService.get("/shopee/db/products/search", {
        params: { sku: sku.trim() },
      });
      return response;
    } catch (error) {
      console.error(`❌ Error searching by SKU ${sku}:`, error);
      throw error;
    }
  }

  async getProductsByStatus(status: string): Promise<any> {
    try {
      if (!status) {
        throw new Error("Status is required");
      }

      const response = await apiService.get(
        `/shopee/db/products/status/${status}`
      );
      return response;
    } catch (error) {
      console.error(`❌ Error fetching products with status ${status}:`, error);
      throw error;
    }
  }

  async getDatabaseStats(): Promise<any> {
    try {
      const response = await apiService.get("/shopee/db/stats");
      return response;
    } catch (error) {
      console.error("❌ Error fetching database stats:", error);
      throw error;
    }
  }

  async getUnprocessedItems(): Promise<any> {
    try {
      const response = await apiService.get("/shopee/db/sync/unprocessed");
      return response;
    } catch (error) {
      console.error("❌ Error fetching unprocessed items:", error);
      throw error;
    }
  }

  async getItemsWithoutModels(): Promise<any> {
    try {
      const response = await apiService.get("/shopee/db/sync/no-models");
      return response;
    } catch (error) {
      console.error("❌ Error fetching items without models:", error);
      throw error;
    }
  }

  async getSyncLogs(limit: number = 100): Promise<any> {
    try {
      const response = await apiService.get("/shopee/db/sync/logs", {
        params: { limit },
      });
      return response;
    } catch (error) {
      console.error("❌ Error fetching sync logs:", error);
      throw error;
    }
  }

  async batchFetchProductBase(): Promise<any> {
    try {
      const response = await apiService.post(
        "/shopee/db/products/base/batch-fetch",
        {}
      );
      return response;
    } catch (error) {
      console.error("❌ Error in batch fetch product base:", error);
      throw error;
    }
  }

  async fetchProductListFromApi(): Promise<any> {
    try {
      const response = await apiService.post(
        "/shopee/db/products/list/fetch-from-api",
        { item_status: "NORMAL", page_size: 100 }
      );
      return response;
    } catch (error) {
      console.error("❌ Error in fetch product list from API:", error);
      throw error;
    }
  }

  async batchFetchProductModels(): Promise<any> {
    try {
      const response = await apiService.post(
        "/shopee/db/products/model/batch-fetch",
        {}
      );
      return response;
    } catch (error) {
      console.error("❌ Error in batch fetch product models:", error);
      throw error;
    }
  }

  formatProductForDisplay(product: any): Record<string, any> {
    return {
      itemId: product.item_id,
      name: product.item_name || "N/A",
      sku: product.item_sku || "N/A",
      price: product.current_price || product.price || 0,
      originalPrice: product.original_price || 0,
      stock: product.seller_stock || product.stock || 0,
      status: product.item_status || "UNKNOWN",
      hasModels: product.has_model || false,
      category: product.category_id || null,
      brand: product.brand || "N/A",
      description: product.description || "",
      images: Array.isArray(product.image_urls)
        ? product.image_urls
        : product.image_urls
          ? product.image_urls.split(",")
          : [],
      updatedAt: product.updated_at || new Date().toISOString(),
    };
  }

  formatPrice(price: number): string {
    return new Intl.NumberFormat("id-ID", {
      style: "currency",
      currency: "IDR",
    }).format(price || 0);
  }

  formatNumber(num: number): string {
    return new Intl.NumberFormat("id-ID").format(num || 0);
  }
}

export default new ShopeeDbService();
