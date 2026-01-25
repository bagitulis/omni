/**
 * TikTok Product Create Service
 * Handles product creation to TikTok Shop via API
 */

import { TiktokAPIClient } from "../api/clients/tiktokAPIClient";
import { Logger } from "winston";
import { getLogger } from "../utils/logger";
import axios from "axios";
import FormData from "form-data";

// Request interfaces
interface CreateProductSku {
  sellerSku: string;
  price: number;
  stock: number;
  salesAttributes?: { attributeId: string; valueId: string }[];
}

interface CreateProductRequest {
  title: string;
  description: string;
  categoryId: string;
  brandId?: string;
  images: string[]; // Image URLs (must be TikTok hosted)
  skus: CreateProductSku[];
  packageWeight: number; // in grams
  packageDimensions?: {
    length: number;
    width: number;
    height: number;
  };
  deliveryOptionIds?: string[];
  saveMode?: "AS_DRAFT" | "LISTING";
}

// Response interfaces
interface CreateProductResponse {
  success: boolean;
  productId?: string;
  skuIds?: string[];
  message: string;
  errors?: { field: string; message: string }[];
}

export class TiktokProductCreateService {
  private apiClient: TiktokAPIClient;
  private logger: Logger;

  constructor(apiClient: TiktokAPIClient) {
    this.apiClient = apiClient;
    this.logger = getLogger("TiktokProductCreateService");
  }

  /**
   * Create a new product on TikTok Shop
   * API: POST /product/202309/products
   */
  async createProduct(
    request: CreateProductRequest,
  ): Promise<CreateProductResponse> {
    this.logger.info(`🔄 Creating TikTok product: ${request.title}`);

    try {
      // Validate required fields
      this.validateRequest(request);

      // Get warehouse ID (required for inventory)
      const warehouses = await this.getWarehouses();
      const warehouseId =
        warehouses.find((w) => w.isDefault)?.id || warehouses[0]?.id;

      if (!warehouseId) {
        this.logger.warn(
          "⚠️ No warehouse found - product may fail without inventory",
        );
      } else {
        this.logger.info(`📦 Using warehouse ID: ${warehouseId}`);
      }

      // Build TikTok API payload with warehouse ID
      const payload = this.buildCreatePayload(request, warehouseId);

      // Call TikTok API
      const endpoint = "/product/202309/products";
      const result = await this.apiClient.request(
        endpoint,
        "POST",
        {},
        payload,
      );

      // Check response
      if (result?.code !== 0) {
        const errorMsg = result?.message || "Unknown error";
        this.logger.error(`❌ TikTok API error: ${errorMsg}`);
        return {
          success: false,
          message: errorMsg,
          errors: result?.data?.errors || [],
        };
      }

      const productId = result.data?.product_id;
      const skuIds = result.data?.skus?.map((s: any) => s.id) || [];

      this.logger.info(`✅ Product created successfully | ID: ${productId}`);

      return {
        success: true,
        productId,
        skuIds,
        message: `Product "${request.title}" created successfully`,
      };
    } catch (error: any) {
      this.logger.error(`❌ Error creating product: ${error.message}`);
      return {
        success: false,
        message: error.message,
        errors: [{ field: "general", message: error.message }],
      };
    }
  }

  /**
   * Upload image to TikTok and get hosted URI
   * API: POST /product/202309/images/upload
   * TikTok requires multipart/form-data with actual file content
   * This method downloads the image and re-uploads it to TikTok CDN
   */
  async uploadImage(
    imageUrl: string,
    useCase: string = "MAIN_IMAGE",
  ): Promise<string | null> {
    this.logger.info(
      `📤 Uploading image to TikTok: ${imageUrl.substring(0, 50)}...`,
    );

    try {
      // Step 1: Download the image from source URL
      const imageResponse = await axios.get(imageUrl, {
        responseType: "arraybuffer",
        timeout: 30000,
        headers: {
          "User-Agent": "Mozilla/5.0 (compatible; TikTokImageUploader/1.0)",
        },
      });

      const imageBuffer = Buffer.from(imageResponse.data);
      const contentType = imageResponse.headers["content-type"] || "image/jpeg";

      // Determine file extension from content type
      const extMap: Record<string, string> = {
        "image/jpeg": "jpg",
        "image/jpg": "jpg",
        "image/png": "png",
        "image/webp": "webp",
      };
      const ext = extMap[contentType] || "jpg";
      const filename = `image_${Date.now()}.${ext}`;

      this.logger.info(
        `📦 Downloaded image: ${imageBuffer.length} bytes, type: ${contentType}`,
      );

      // Step 2: Prepare multipart form data
      const formData = new FormData();
      formData.append("data", imageBuffer, {
        filename,
        contentType,
      });
      formData.append("use_case", useCase);

      // Step 3: Upload to TikTok using raw request (multipart)
      const result = await this.apiClient.uploadImage(formData);

      if (result?.code !== 0 || !result?.data?.uri) {
        this.logger.error(
          `❌ Image upload failed: ${result?.message || "No URI returned"}`,
        );
        return null;
      }

      this.logger.info(`✅ Image uploaded successfully: ${result.data.uri}`);
      return result.data.uri;
    } catch (error: any) {
      this.logger.error(`❌ Error uploading image: ${error.message}`);
      return null;
    }
  }

  /**
   * Get warehouse list for the shop
   * API: GET /logistics/202309/warehouses
   */
  async getWarehouses(): Promise<
    { id: string; name: string; isDefault?: boolean }[]
  > {
    this.logger.info(`🔄 Fetching warehouse list`);

    try {
      const endpoint = "/logistics/202309/warehouses";
      const result = await this.apiClient.request(endpoint, "GET", {});

      if (!result?.data?.warehouses) {
        return [];
      }

      return result.data.warehouses.map((wh: any) => ({
        id: wh.id,
        name: wh.name || wh.warehouse_name,
        isDefault: wh.is_default || false,
      }));
    } catch (error: any) {
      this.logger.error(`❌ Error fetching warehouses: ${error.message}`);
      return [];
    }
  }

  /**
   * Get available delivery options for the shop
   * API: GET /logistics/202309/delivery_options
   */
  async getDeliveryOptions(): Promise<{ id: string; name: string }[]> {
    this.logger.info(`🔄 Fetching delivery options`);

    try {
      const endpoint = "/logistics/202309/delivery_options";
      const result = await this.apiClient.request(endpoint, "GET", {});

      if (!result?.data?.delivery_options) {
        return [];
      }

      return result.data.delivery_options.map((opt: any) => ({
        id: opt.id,
        name: opt.name,
      }));
    } catch (error: any) {
      this.logger.error(`❌ Error fetching delivery options: ${error.message}`);
      return [];
    }
  }

  /**
   * Get available brands for a category
   * API: GET /product/202309/brands
   */
  async getBrands(categoryId: string): Promise<{ id: string; name: string }[]> {
    this.logger.info(`🔄 Fetching brands for category ${categoryId}`);

    try {
      const endpoint = "/product/202309/brands";
      const result = await this.apiClient.request(endpoint, "GET", {
        category_id: categoryId,
      });

      if (!result?.data?.brands) {
        return [];
      }

      return result.data.brands.map((brand: any) => ({
        id: brand.id,
        name: brand.name,
      }));
    } catch (error: any) {
      this.logger.error(`❌ Error fetching brands: ${error.message}`);
      return [];
    }
  }

  /**
   * Validate create product request
   */
  private validateRequest(request: CreateProductRequest): void {
    const errors: string[] = [];

    if (!request.title || request.title.length < 1) {
      errors.push("Product title is required");
    }
    if (!request.description || request.description.length < 1) {
      errors.push("Product description is required");
    }
    if (!request.categoryId) {
      errors.push("Category ID is required");
    }
    if (!request.images || request.images.length === 0) {
      errors.push("At least one product image is required");
    }
    if (!request.skus || request.skus.length === 0) {
      errors.push("At least one SKU is required");
    }
    if (!request.packageWeight || request.packageWeight <= 0) {
      errors.push("Package weight must be greater than 0");
    }

    // Validate each SKU
    request.skus?.forEach((sku, index) => {
      if (!sku.sellerSku) {
        errors.push(`SKU ${index + 1}: Seller SKU is required`);
      }
      if (sku.price <= 0) {
        errors.push(`SKU ${index + 1}: Price must be greater than 0`);
      }
      if (sku.stock < 0) {
        errors.push(`SKU ${index + 1}: Stock cannot be negative`);
      }
    });

    if (errors.length > 0) {
      throw new Error(`Validation failed: ${errors.join("; ")}`);
    }
  }

  /**
   * Build TikTok API create product payload
   * @param request - Product creation request
   * @param warehouseId - Valid TikTok warehouse ID for inventory
   */
  private buildCreatePayload(
    request: CreateProductRequest,
    warehouseId?: string,
  ): Record<string, any> {
    const payload: Record<string, any> = {
      title: request.title,
      description: request.description,
      category_id: request.categoryId,
      category_version: "v2", // V2 required for SEA (ID, SG, MY, VN, TH, PH) and US shops
      main_images: request.images.map((url) => ({ uri: url })),
      package_weight: {
        value: String(request.packageWeight),
        unit: "GRAM",
      },
      // Default product attributes - "Contains Dangerous Goods?" = Tidak (No)
      product_attributes: [
        {
          id: "101734", // Contains Dangerous Goods?
          values: [
            {
              name: "Tidak", // No dangerous goods
            },
          ],
        },
      ],
      skus: request.skus.map((sku) => {
        const skuPayload: Record<string, any> = {
          seller_sku: sku.sellerSku,
          price: {
            amount: String(Math.round(sku.price)), // IDR is already in full amount, no conversion needed
            currency: "IDR",
          },
        };

        // Add inventory with warehouse ID (required by TikTok API)
        if (warehouseId) {
          skuPayload.inventory = [
            {
              warehouse_id: warehouseId,
              quantity: sku.stock >= 0 ? sku.stock : 0,
            },
          ];
        }

        // Only add sales_attributes if they exist
        if (sku.salesAttributes && sku.salesAttributes.length > 0) {
          skuPayload.sales_attributes = sku.salesAttributes.map((attr) => ({
            attribute_id: attr.attributeId,
            value_id: attr.valueId,
          }));
        }

        return skuPayload;
      }),
      save_mode: request.saveMode || "AS_DRAFT", // Use draft mode for testing
    };

    // Add optional fields
    if (request.brandId) {
      payload.brand_id = request.brandId;
    }

    if (request.packageDimensions) {
      payload.package_dimensions = {
        length: String(request.packageDimensions.length),
        width: String(request.packageDimensions.width),
        height: String(request.packageDimensions.height),
        unit: "CENTIMETER",
      };
    }

    if (request.deliveryOptionIds && request.deliveryOptionIds.length > 0) {
      payload.delivery_option_ids = request.deliveryOptionIds;
    }

    return payload;
  }
}

export type { CreateProductRequest, CreateProductResponse, CreateProductSku };
