/**
 * Lazada Product Create Service
 * Handles product creation on Lazada
 * Max 300 lines - AGENTS.MD compliant
 */

import { LazadaAPIClient } from "../api/clients/lazadaAPIClient";

export interface LazadaCreateProductRequest {
  name: string;
  description: string;
  categoryId: number;
  brandName?: string;
  images: string[];
  price: number;
  quantity: number;
  sellerSku?: string;
  packageWeight: number; // in kg
  packageDimensions?: { length: number; width: number; height: number };
}

interface CreateProductResult {
  success: boolean;
  itemId?: string;
  message: string;
}

export class LazadaProductCreateService {
  private apiClient: LazadaAPIClient;

  constructor(apiClient: LazadaAPIClient) {
    this.apiClient = apiClient;
  }

  /**
   * Migrate image to Lazada CDN
   * POST /image/migrate
   * Per Lazada SDK: payload must be in params, not body
   */
  async migrateImage(imageUrl: string): Promise<string | null> {
    try {
      // Lazada expects XML payload for image migration
      const xmlPayload = `<?xml version="1.0" encoding="UTF-8"?>
<Request>
  <Image>
    <Url>${imageUrl}</Url>
  </Image>
</Request>`;

      const response = await this.apiClient.request(
        "/image/migrate",
        "POST",
        { payload: xmlPayload }, // Must be in params for signature
        {},
      );

      // Response format: { data: { image: { url: "..." } } } or { data: { images: { image: [...] } } }
      if (response?.data?.image?.url) {
        return response.data.image.url;
      }
      if (response?.data?.images?.image?.[0]?.url) {
        return response.data.images.image[0].url;
      }

      console.log("Image migrate response:", JSON.stringify(response));
      return imageUrl; // Fallback to original if migration fails
    } catch (error: any) {
      console.log("Image migrate error:", error.message);
      return imageUrl;
    }
  }

  /**
   * Get brand list
   * GET /category/brands/query
   */
  async getBrands(
    keyword: string,
    offset = 0,
    limit = 100,
  ): Promise<Array<{ id: string; name: string }>> {
    try {
      const response = await this.apiClient.request(
        "/category/brands/query",
        "GET",
        { name: keyword, offset, limit },
      );

      if (response?.data?.module) {
        return response.data.module.map((brand: any) => ({
          id: String(brand.brand_id),
          name: brand.name,
        }));
      }
      return [];
    } catch {
      return [];
    }
  }

  /**
   * Create product on Lazada using XML format
   * POST /product/create
   * IMPORTANT: Per Lazada SDK, for POST requests, body must be part of query string
   * so that signature calculation includes the payload
   */
  async createProduct(
    request: LazadaCreateProductRequest,
  ): Promise<CreateProductResult> {
    if (!request.name || !request.categoryId || request.images.length === 0) {
      return { success: false, message: "Missing required fields" };
    }

    try {
      // Prepare images (migrate if needed)
      const imageUrls = await this.prepareImages(request.images);

      // Build XML payload (Lazada uses XML for product creation)
      const xmlPayload = this.buildProductXml(request, imageUrls);

      // Per Lazada SDK: POST body must be sent as query param for signature
      // The signature calculation includes params, so payload must be in params
      const response = await this.apiClient.request(
        "/product/create",
        "POST",
        { payload: xmlPayload }, // Send as param, not body!
        {}, // Empty body
      );

      // Log full response for debugging
      console.log(
        "[LazadaProductCreate] Response:",
        JSON.stringify(response, null, 2),
      );

      if (response?.data?.item_id) {
        return {
          success: true,
          itemId: String(response.data.item_id),
          message: "Product created successfully",
        };
      }

      // Extract detailed error message
      const errorMsg =
        response?.message ||
        response?.detail?.[0]?.message ||
        response?.code ||
        JSON.stringify(response);

      return {
        success: false,
        message: `E${response?.code || 500}: ${errorMsg}`,
      };
    } catch (error: any) {
      console.log(
        "[LazadaProductCreate] Error:",
        error.message,
        error.response?.data,
      );
      return {
        success: false,
        message: error.message || "Create product failed",
      };
    }
  }

  /**
   * Build XML payload for Lazada product creation
   */
  private buildProductXml(
    request: LazadaCreateProductRequest,
    imageUrls: string[],
  ): string {
    const images = imageUrls
      .map((url) => `<Image>${this.escapeXml(url)}</Image>`)
      .join("");
    const sku = request.sellerSku || `SKU-${Date.now()}`;

    return `<?xml version="1.0" encoding="UTF-8"?>
<Request>
  <Product>
    <PrimaryCategory>${request.categoryId}</PrimaryCategory>
    <Attributes>
      <name>${this.escapeXml(request.name)}</name>
      <description><![CDATA[${request.description}]]></description>
      <brand>${this.escapeXml(request.brandName || "No Brand")}</brand>
      <warranty_type>No Warranty</warranty_type>
    </Attributes>
    <Skus>
      <Sku>
        <SellerSku>${this.escapeXml(sku)}</SellerSku>
        <quantity>${request.quantity}</quantity>
        <price>${request.price}</price>
        <package_weight>${request.packageWeight}</package_weight>
        ${
          request.packageDimensions
            ? `
        <package_length>${request.packageDimensions.length}</package_length>
        <package_width>${request.packageDimensions.width}</package_width>
        <package_height>${request.packageDimensions.height}</package_height>
        `
            : ""
        }
        <Images>${images}</Images>
      </Sku>
    </Skus>
  </Product>
</Request>`;
  }

  /**
   * Escape XML special characters
   */
  private escapeXml(str: string): string {
    return str
      .replace(/&/g, "&amp;")
      .replace(/</g, "&lt;")
      .replace(/>/g, "&gt;")
      .replace(/"/g, "&quot;")
      .replace(/'/g, "&apos;");
  }

  /**
   * Prepare images - migrate URLs to Lazada CDN
   */
  private async prepareImages(images: string[]): Promise<string[]> {
    const migratedUrls: string[] = [];

    for (const img of images) {
      const migrated = await this.migrateImage(img);
      if (migrated) migratedUrls.push(migrated);
    }

    return migratedUrls.length > 0 ? migratedUrls : images;
  }
}
