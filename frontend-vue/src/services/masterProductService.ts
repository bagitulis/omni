/**
 * Master Product Service
 * Handles API calls for Master Product management
 */
import apiService from "./api";

// Types matching backend models
export interface MasterProduct {
  id: number;
  tenant_id: string;
  title: string;
  description: string;
  images: string[];
  status: string;
  created_at: string;
  updated_at: string;
  skus?: MasterProductSku[];
}

export interface MasterProductSku {
  id: number;
  tenant_id: string;
  master_product_id: number;
  seller_sku: string;
  variant_name: string;
  variant_data: Record<string, any>;
  price: number;
  stock: number;
  created_at: string;
  updated_at: string;
  platform_links?: MasterProductPlatformLink[];
}

export interface MasterProductPlatformLink {
  id: number;
  master_product_id: number;
  master_sku_id?: number;
  platform: "shopee" | "tiktok" | "lazada";
  platform_product_id?: string;
  platform_sku_id?: string;
  platform_item_id?: number;
  sync_status: string;
  last_synced_at?: string;
}

export interface CreateMasterProductInput {
  title: string;
  description: string;
  images?: string[];
  status?: string;
  skus?: CreateSkuInput[];
}

export interface CreateSkuInput {
  seller_sku: string;
  variant_name?: string;
  variant_data?: Record<string, any>;
  price: number;
  stock: number;
}

export interface UpdateMasterProductInput {
  title?: string;
  description?: string;
  images?: string[];
  status?: string;
}

export interface UpdateSkuInput {
  price?: number;
  stock?: number;
}

export interface BatchUpdateSkuInput {
  sku_ids: number[];
  price?: number;
  stock?: number;
}

export interface BatchUpdateSkuResult {
  updated: number;
  failed: number;
  skus: MasterProductSku[];
  errors?: string[];
}

export interface ListFilter {
  page?: number;
  limit?: number;
  status?: string;
  search?: string;
}

export interface ListResponse {
  success: boolean;
  data: MasterProduct[];
  meta?: {
    total: number;
    page: number;
    page_size: number;
  };
  error?: string;
}

export interface SingleResponse {
  success: boolean;
  data: MasterProduct;
  error?: string;
}

export interface ImportPreviewResponse {
  title: string;
  description: string;
  images: string[];
  skus: {
    seller_sku: string;
    variant_name: string;
    price: number;
    stock: number;
  }[];
  source_platform: string;
  source_item_id: number;
}

export interface ImportResponse {
  master_product: MasterProduct;
  skus_imported: number;
  platform_link_created: boolean;
}

export interface MappingStatus {
  master_product_id: number;
  title: string;
  skus: SkuMappingInfo[];
  platforms: {
    shopee: { linked: number; total: number };
    tiktok: { linked: number; total: number };
    lazada: { linked: number; total: number };
  };
}

export interface SkuMappingInfo {
  master_sku_id: number;
  seller_sku: string;
  variant_name: string;
  platform_links: PlatformLinkInfo[];
}

export interface PlatformLinkInfo {
  platform: string;
  platform_item_id: string;
  platform_sku_id?: string;
  sync_status: string;
  last_synced_at?: string;
}

export interface AutoMapResult {
  seller_sku: string;
  matches: {
    platform: string;
    platform_item_id: string;
    platform_sku_id: string;
    product_name: string;
  }[];
}

class MasterProductService {
  private readonly basePath = "/master-products";

  /**
   * Get list of master products with pagination
   */
  async list(filter: ListFilter = {}): Promise<ListResponse> {
    try {
      const params: Record<string, any> = {};
      if (filter.page) params.page = filter.page;
      if (filter.limit) params.limit = filter.limit;
      if (filter.status) params.status = filter.status;
      if (filter.search) params.search = filter.search;

      const response = await apiService.get<ListResponse>(this.basePath, {
        params,
      });
      return response;
    } catch (error) {
      console.error("❌ Error fetching master products:", error);
      throw error;
    }
  }

  /**
   * Get a single master product by ID
   */
  async getById(id: number): Promise<SingleResponse> {
    try {
      const response = await apiService.get<SingleResponse>(
        `${this.basePath}/${id}`,
      );
      return response;
    } catch (error) {
      console.error(`❌ Error fetching master product ${id}:`, error);
      throw error;
    }
  }

  /**
   * Create a new master product
   */
  async create(input: CreateMasterProductInput): Promise<SingleResponse> {
    try {
      const response = await apiService.post<SingleResponse>(
        this.basePath,
        input,
      );
      return response;
    } catch (error) {
      console.error("❌ Error creating master product:", error);
      throw error;
    }
  }

  /**
   * Update an existing master product
   */
  async update(
    id: number,
    input: UpdateMasterProductInput,
  ): Promise<SingleResponse> {
    try {
      // Use PATCH or PUT - api.ts has patch method but backend expects PUT
      const response = await apiService.client.put<SingleResponse>(
        `${this.basePath}/${id}`,
        input,
      );
      return response.data;
    } catch (error) {
      console.error(`❌ Error updating master product ${id}:`, error);
      throw error;
    }
  }

  /**
   * Delete a master product
   */
  async delete(
    id: number,
  ): Promise<{ success: boolean; data?: any; error?: string }> {
    try {
      const response = await apiService.delete(`${this.basePath}/${id}`);
      return response;
    } catch (error) {
      console.error(`❌ Error deleting master product ${id}:`, error);
      throw error;
    }
  }

  /**
   * Update a single SKU's price and/or stock
   * PUT /api/master-products/skus/:id
   */
  async updateSku(
    skuId: number,
    input: UpdateSkuInput,
  ): Promise<MasterProductSku> {
    try {
      const response = await apiService.client.put<{
        success: boolean;
        data: MasterProductSku;
      }>(`${this.basePath}/skus/${skuId}`, input);
      return response.data.data;
    } catch (error) {
      console.error(`❌ Error updating SKU ${skuId}:`, error);
      throw error;
    }
  }

  /**
   * Batch update multiple SKUs' price and/or stock at once
   * PUT /api/master-products/skus/batch
   */
  async batchUpdateSkus(
    input: BatchUpdateSkuInput,
  ): Promise<BatchUpdateSkuResult> {
    try {
      const response = await apiService.client.put<{
        success: boolean;
        data: BatchUpdateSkuResult;
      }>(`${this.basePath}/skus/batch`, input);
      return response.data.data;
    } catch (error) {
      console.error(`❌ Error batch updating SKUs:`, error);
      throw error;
    }
  }

  /**
   * Helper: Get first image URL from product
   */
  getFirstImage(product: MasterProduct): string | null {
    if (!product.images || product.images.length === 0) {
      return null;
    }
    // Images are stored as string array: ["url1", "url2", ...]
    return product.images[0] || null;
  }

  /**
   * Helper: Get all image URLs from product
   */
  getAllImages(product: MasterProduct): string[] {
    if (!product.images || product.images.length === 0) {
      return [];
    }
    return product.images.filter(Boolean);
  }

  /**
   * Helper: Format price for display
   */
  formatPrice(price: number): string {
    return new Intl.NumberFormat("id-ID", {
      style: "currency",
      currency: "IDR",
      minimumFractionDigits: 0,
    }).format(price || 0);
  }

  /**
   * Helper: Get status display class
   */
  getStatusClass(status: string): string {
    switch (status) {
      case "active":
        return "status-active";
      case "draft":
        return "status-draft";
      case "inactive":
        return "status-inactive";
      default:
        return "status-unknown";
    }
  }

  /**
   * Helper: Get platform icon
   */
  getPlatformIcon(platform: string): string {
    switch (platform) {
      case "shopee":
        return "🟠";
      case "tiktok":
        return "⬛";
      case "lazada":
        return "🔵";
      default:
        return "⚪";
    }
  }

  /**
   * Preview import - shows what will be imported from platform
   * GET /api/master-products/import/preview?platform=shopee&item_id=123
   */
  async previewImport(
    platform: string,
    itemId: string,
  ): Promise<ImportPreviewResponse> {
    try {
      const response = await apiService.get<{
        success: boolean;
        data: ImportPreviewResponse;
      }>(`${this.basePath}/import/preview`, {
        params: { platform, item_id: itemId },
      });
      return response.data;
    } catch (error) {
      console.error("❌ Error previewing import:", error);
      throw error;
    }
  }

  /**
   * Execute import from platform to Master Product
   * POST /api/master-products/import
   */
  async importProduct(
    platform: string,
    itemId: string,
  ): Promise<ImportResponse> {
    try {
      const response = await apiService.post<{
        success: boolean;
        data: ImportResponse;
      }>(`${this.basePath}/import`, {
        platform,
        item_id: itemId,
      });
      return response.data;
    } catch (error) {
      console.error("❌ Error importing product:", error);
      throw error;
    }
  }

  /**
   * Get mapping status for a master product
   * GET /api/master-products/:id/mapping
   */
  async getMappingStatus(productId: number): Promise<MappingStatus> {
    const response = await apiService.get<{
      success: boolean;
      data: MappingStatus;
    }>(`${this.basePath}/${productId}/mapping`);
    return response.data;
  }

  /**
   * Auto-map SKU to platform products
   * POST /api/master-products/mapping/auto
   */
  async autoMapSku(sellerSku: string): Promise<AutoMapResult> {
    const response = await apiService.post<{
      success: boolean;
      data: AutoMapResult;
    }>(`${this.basePath}/mapping/auto`, { seller_sku: sellerSku });
    return response.data;
  }

  /**
   * Manual link SKU to platform product
   * POST /api/master-products/mapping/link
   */
  async manualLink(
    masterSkuId: number,
    platform: string,
    platformItemId: string,
    platformSkuId?: string,
  ): Promise<void> {
    await apiService.post(`${this.basePath}/mapping/link`, {
      master_sku_id: masterSkuId,
      platform,
      platform_item_id: platformItemId,
      platform_sku_id: platformSkuId,
    });
  }

  /**
   * Unlink SKU from platform
   * DELETE /api/master-products/mapping/link
   */
  async unlinkSku(masterSkuId: number, platform: string): Promise<void> {
    await apiService.delete(`${this.basePath}/mapping/link`, {
      data: { master_sku_id: masterSkuId, platform },
    });
  }

  /**
   * Sync product to platform
   * POST /api/master-products/:id/sync
   */
  async sync(
    productId: number,
    targetPlatform: string,
  ): Promise<{ skus_synced: number; platform: string }> {
    const response = await apiService.post<{ success: boolean; data: any }>(
      `${this.basePath}/${productId}/sync`,
      { target_platform: targetPlatform },
    );
    return response.data;
  }

  /**
   * Get sync status for product
   * GET /api/master-products/:id/sync-status
   */
  async getSyncStatus(productId: number): Promise<any> {
    const response = await apiService.get<{ success: boolean; data: any }>(
      `${this.basePath}/${productId}/sync-status`,
    );
    return response.data;
  }
}

export default new MasterProductService();
