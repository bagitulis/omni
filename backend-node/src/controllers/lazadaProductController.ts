import { Request, Response } from "express";
import { PrismaClient } from "@prisma/client";
import { ProductControllerBase } from "./base/ProductControllerBase";
import { LazadaProductService } from "../services/lazadaProductService";
import { LazadaAPIClient } from "../api/clients/lazadaAPIClient";
import { LazadaConfigManager } from "../config/managers/lazadaConfigManager";

/**
 * Lazada Product Controller
 * Handles all Lazada product-related endpoints
 */
export class LazadaProductController extends ProductControllerBase {
  private apiClient: LazadaAPIClient;
  private configManager: LazadaConfigManager;

  constructor(
    prisma: PrismaClient,
    apiClient: LazadaAPIClient,
    configManager: LazadaConfigManager,
  ) {
    super(prisma);
    this.apiClient = apiClient;
    this.configManager = configManager;
  }

  private getProductService(tenantId: string): LazadaProductService {
    return new LazadaProductService(
      this.prisma,
      this.apiClient,
      this.configManager,
      tenantId,
    );
  }

  private extractTenantId(req: Request, res: Response): string | null {
    // Extract tenantId from headers or query
    const tenantId =
      (req.headers["x-tenant-id"] as string) || (req.query.tenantId as string);

    if (!tenantId) {
      res.status(401).json({
        success: false,
        error: "Missing tenant_id - authentication required",
      });
      return null;
    }

    return tenantId;
  }

  /**
   * GET /api/lazada/products
   * Get Lazada products from API
   */
  async getProducts(req: Request, res: Response): Promise<any> {
    try {
      // Extract tenant ID
      const tenantId = this.extractTenantId(req, res);
      if (!tenantId) return;

      const productService = this.getProductService(tenantId);

      // Ensure config is loaded
      await this.configManager.loadConfig();

      const filter = (req.query.filter as string) || "live";
      const offset = parseInt(req.query.offset as string) || 0;
      const limit = parseInt(req.query.limit as string) || 50;

      const result = await productService.getProducts({
        filter,
        offset,
        limit,
      });

      return this.sendSuccess(res, result, result.success ? 200 : 400);
    } catch (error: any) {
      return this.sendError(res, error);
    }
  }

  /**
   * GET /api/lazada/products/db
   * Get master products from database
   */
  async getMasterProductsFromDb(req: Request, res: Response): Promise<any> {
    try {
      const tenantId = this.extractTenantId(req, res);
      if (!tenantId) return;

      const productService = this.getProductService(tenantId);

      const offset = parseInt(req.query.offset as string) || 0;
      const limit = parseInt(req.query.limit as string) || 100;

      const result = await productService.getMasterProductsFromDb({
        offset,
        limit,
      });

      // Transform response to flatten SKUs - one row per SKU
      if (result.success && result.products) {
        const flattened: any[] = [];

        result.products.forEach((product: any) => {
          // If product has SKUs, create one row per SKU
          if (
            product.skus &&
            Array.isArray(product.skus) &&
            product.skus.length > 0
          ) {
            product.skus.forEach((sku: any) => {
              flattened.push({
                item_id: product.itemId,
                sku_id: sku.skuId,
                sku_name: sku.sellerSku || `SKU-${sku.skuId}`,
                variant_name: sku.variantName || "-",
                variant_data: sku.variantData
                  ? typeof sku.variantData === "string"
                    ? sku.variantData
                    : JSON.stringify(sku.variantData)
                  : "-",
                item_name: product.name,
                price: sku.price || product.price || 0,
                quantity: sku.quantity || 0,
                status: product.status,
                updated_at: product.updatedAt,
                _original: {
                  product,
                  sku,
                },
              });
            });
          } else {
            // Fallback if no SKUs
            flattened.push({
              item_id: product.itemId,
              sku_id: "-",
              sku_name: "-",
              variant_name: "-",
              variant_data: "-",
              item_name: product.name,
              price: product.price,
              quantity: product.quantity,
              status: product.status,
              updated_at: product.updatedAt,
              _original: product,
            });
          }
        });

        result.products = flattened;
      }

      return this.sendSuccess(res, result, result.success ? 200 : 400);
    } catch (error: any) {
      return this.sendError(res, error);
    }
  }

  /**
   * GET /api/lazada/products/list/db
   * Get product list from database (quick lookup)
   */
  async getProductListFromDb(req: Request, res: Response): Promise<any> {
    try {
      const tenantId = this.extractTenantId(req, res);
      if (!tenantId) return;

      const productService = this.getProductService(tenantId);

      const offset = parseInt(req.query.offset as string) || 0;
      const limit = parseInt(req.query.limit as string) || 100;

      const result = await productService.getProductListFromDb({
        offset,
        limit,
      });

      return this.sendSuccess(res, result, result.success ? 200 : 400);
    } catch (error: any) {
      return this.sendError(res, error);
    }
  }

  /**
   * GET /api/lazada/product/:itemId
   * Get product detail by Item ID from database
   */
  async getProductDetail(req: Request, res: Response): Promise<any> {
    try {
      const tenantId = this.extractTenantId(req, res);
      if (!tenantId) return;

      const productService = this.getProductService(tenantId);

      const itemId = req.params.itemId as string;

      if (!itemId) {
        return this.sendError(res, { message: "Item ID is required" }, 400);
      }

      const result = await productService.getProductDetail({
        itemId,
      });

      return this.sendSuccess(res, result, result.success ? 200 : 400);
    } catch (error: any) {
      return this.sendError(res, error);
    }
  }
}
