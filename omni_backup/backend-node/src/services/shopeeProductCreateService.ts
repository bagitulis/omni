/**
 * Shopee Product Create Service
 * Handles product creation on Shopee
 * Max 300 lines - AGENTS.MD compliant
 */

import { ShopeeAPIClient } from "../api/clients/shopeeAPIClient";

export interface ShopeeCreateProductRequest {
  item_name: string;
  description: string;
  category_id: number;
  brand?: { brand_id: number; original_brand_name: string };
  images: string[];
  normal_stock: number;
  original_price: number;
  seller_sku?: string;
  weight: number; // in kg
  dimension?: { package_length: number; package_width: number; package_height: number };
  logistic_info?: Array<{ logistic_id: number; enabled: boolean }>;
  pre_order?: { is_pre_order: boolean; days_to_ship: number };
  condition?: "NEW" | "USED";
}

interface CreateProductResult {
  success: boolean;
  itemId?: number;
  message: string;
}

export class ShopeeProductCreateService {
  private apiClient: ShopeeAPIClient;

  constructor(apiClient: ShopeeAPIClient) {
    this.apiClient = apiClient;
  }

  /**
   * Upload image to Shopee
   * POST /api/v2/media_space/upload_image
   */
  async uploadImage(imageUrl: string): Promise<string | null> {
    try {
      const response = await this.apiClient.request(
        "/api/v2/media_space/upload_image",
        "POST",
        {},
        { image_url: imageUrl }
      );

      if (response?.response?.image_info?.image_id) {
        return response.response.image_info.image_id;
      }
      return null;
    } catch (error: any) {
      console.error("Shopee image upload failed:", error.message);
      return null;
    }
  }

  /**
   * Get brand list for category
   * GET /api/v2/product/get_brand_list
   */
  async getBrands(categoryId: number, offset = 0, pageSize = 100): Promise<Array<{
    id: string;
    name: string;
  }>> {
    try {
      const response = await this.apiClient.request(
        "/api/v2/product/get_brand_list",
        "GET",
        { category_id: categoryId, offset, page_size: pageSize, status: 1 }
      );

      if (response?.response?.brand_list) {
        return response.response.brand_list.map((brand: any) => ({
          id: String(brand.brand_id),
          name: brand.original_brand_name,
        }));
      }
      return [];
    } catch {
      return [];
    }
  }

  /**
   * Get logistics channels
   * GET /api/v2/logistics/get_channel_list
   */
  async getLogistics(): Promise<Array<{ id: string; name: string }>> {
    try {
      const response = await this.apiClient.request(
        "/api/v2/logistics/get_channel_list",
        "GET"
      );

      if (response?.response?.logistics_channel_list) {
        return response.response.logistics_channel_list
          .filter((ch: any) => ch.enabled)
          .map((ch: any) => ({
            id: String(ch.logistics_channel_id),
            name: ch.logistics_channel_name,
          }));
      }
      return [];
    } catch {
      return [];
    }
  }

  /**
   * Create product on Shopee
   * POST /api/v2/product/add_item
   */
  async createProduct(request: ShopeeCreateProductRequest): Promise<CreateProductResult> {
    // Validate required fields
    if (!request.item_name || !request.category_id || request.images.length === 0) {
      return { success: false, message: "Missing required fields" };
    }

    try {
      // Build image list - upload if URLs, use directly if image_ids
      const imageList = await this.prepareImages(request.images);
      if (imageList.length === 0) {
        return { success: false, message: "Failed to process images" };
      }

      // Build payload per Shopee API v2.0
      const payload = {
        item_name: request.item_name.substring(0, 120), // Max 120 chars
        description: request.description.substring(0, 3000), // Max 3000 chars
        category_id: request.category_id,
        image: { image_id_list: imageList },
        original_price: request.original_price,
        normal_stock: request.normal_stock,
        seller_sku: request.seller_sku || "",
        weight: request.weight,
        dimension: request.dimension || undefined,
        logistic_info: request.logistic_info || [],
        brand: request.brand || undefined,
        pre_order: request.pre_order || { is_pre_order: false },
        condition: request.condition || "NEW",
        item_status: "NORMAL",
      };

      const response = await this.apiClient.request(
        "/api/v2/product/add_item",
        "POST",
        {},
        payload
      );

      if (response?.response?.item_id) {
        return {
          success: true,
          itemId: response.response.item_id,
          message: "Product created successfully",
        };
      }

      return {
        success: false,
        message: response?.message || response?.msg || "Failed to create product",
      };
    } catch (error: any) {
      return { success: false, message: error.message || "Create product failed" };
    }
  }

  /**
   * Prepare images - upload URLs or validate image IDs
   */
  private async prepareImages(images: string[]): Promise<string[]> {
    const imageIds: string[] = [];

    for (const img of images) {
      // If it's a URL, upload it first
      if (img.startsWith("http://") || img.startsWith("https://")) {
        const uploadedId = await this.uploadImage(img);
        if (uploadedId) imageIds.push(uploadedId);
      } else {
        // Assume it's already an image_id
        imageIds.push(img);
      }
    }

    return imageIds;
  }
}
