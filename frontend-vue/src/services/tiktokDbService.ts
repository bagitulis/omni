import apiService from "./api";

class TiktokDbService {
  async searchProducts(
    status: string = "ACTIVATE", // Default to ACTIVATE only
    pageSize: number = 100,
    limit?: number
  ): Promise<any> {
    try {
      const response = await apiService.post("/tiktok/products/search", {
        status,
        page_size: pageSize,
        limit,
        sync_to_db: true,
      });

      return response;
    } catch (error) {
      console.error("❌ Error searching products:", error);
      throw error;
    }
  }

  async getAllProducts(
    status: string = "ACTIVATE", // Default to ACTIVATE only
    pageSize: number = 100
  ): Promise<any> {
    try {
      const response = await apiService.post("/tiktok/products/get-all", {
        status,
        page_size: pageSize,
        sync_to_db: true,
      });
      return response;
    } catch (error) {
      console.error("❌ Error getting all products:", error);
      throw error;
    }
  }

  async getProductDetail(productId: string | number): Promise<any> {
    try {
      const response = await apiService.get(`/tiktok/products/${productId}`, {
        params: { sync_to_db: true },
      });
      return response;
    } catch (error) {
      console.error("❌ Error getting product detail:", error);
      throw error;
    }
  }

  async refreshProductData(status: string = "ALL"): Promise<any> {
    try {
      const response = await apiService.post("/tiktok/products/refresh-all", {
        status,
        page_size: 100,
        fetch_details: true,
      });
      return response;
    } catch (error) {
      console.error("❌ Error refreshing product data:", error);
      throw error;
    }
  }

  async getProductListFromDB(): Promise<any> {
    try {
      const response = await apiService.get("/tiktok/db/products/list");
      // Map 'data' field to 'products' for consistency with UI expectations
      return {
        ...response,
        products: response.data || [],
      };
    } catch (error) {
      console.error("❌ Error getting product list from DB:", error);
      throw error;
    }
  }

  async getMasterProductsFromDB(): Promise<any> {
    try {
      const response = await apiService.get("/tiktok/db/products/master");
      return response;
    } catch (error) {
      console.error("❌ Error getting master products from DB:", error);
      throw error;
    }
  }

  async searchProductsInDB(
    query: string,
    field: string = "name"
  ): Promise<any> {
    try {
      const response = await apiService.get(
        `/tiktok/db/search?q=${encodeURIComponent(query)}&field=${field}`
      );
      return {
        ...response,
        products: response.data || [],
      };
    } catch (error) {
      console.error("❌ Error searching products in DB:", error);
      throw error;
    }
  }

  async getProductsByStatus(status: string): Promise<any> {
    try {
      const response = await apiService.get(
        `/tiktok/db/products/status/${status}`
      );
      return {
        ...response,
        products: response.data || [],
      };
    } catch (error) {
      console.error("❌ Error getting products by status:", error);
      throw error;
    }
  }

  async getProductById(productId: string): Promise<any> {
    try {
      const response = await apiService.get(`/tiktok/db/products/${productId}`);
      return response;
    } catch (error) {
      console.error("❌ Error getting product by ID:", error);
      throw error;
    }
  }

  async getStatistics(): Promise<any> {
    try {
      const response = await apiService.get("/tiktok/db/statistics");
      return response;
    } catch (error) {
      console.error("❌ Error getting statistics:", error);
      throw error;
    }
  }

  async deleteProduct(productId: string): Promise<any> {
    try {
      const response = await apiService.delete(
        `/tiktok/db/products/${productId}`
      );
      return response;
    } catch (error) {
      console.error("❌ Error deleting product:", error);
      throw error;
    }
  }
}

export default new TiktokDbService();
