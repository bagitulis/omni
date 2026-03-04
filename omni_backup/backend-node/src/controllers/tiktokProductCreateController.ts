/**
 * TikTok Product Create Controller
 * Handles product creation, category, brand, and image upload endpoints
 * Split from TiktokProductController for SRP compliance
 * Max 300 lines - AGENTS.MD compliant
 */

import { Request, Response } from "express";
import { TiktokCategoryService } from "../services/tiktokCategoryService";
import {
  TiktokProductCreateService,
  CreateProductRequest,
} from "../services/tiktokProductCreateService";
import { TiktokAPIClient } from "../api/clients/tiktokAPIClient";
import { TiktokConfigManager } from "../config/managers/tiktokConfigManager";

export class TiktokProductCreateController {
  private categoryService: TiktokCategoryService;
  private productCreateService: TiktokProductCreateService;
  private configManager: TiktokConfigManager;

  constructor(apiClient: TiktokAPIClient, configManager: TiktokConfigManager) {
    this.configManager = configManager;
    this.categoryService = new TiktokCategoryService(apiClient);
    this.productCreateService = new TiktokProductCreateService(apiClient);
  }

  /** Helper to send success response */
  private sendSuccess(res: Response, data: any, status: number = 200): any {
    return res.status(status).json(data);
  }

  /** Helper to send error response */
  private sendError(res: Response, error: any): any {
    return res.status(500).json({ success: false, error: error.message });
  }

  // ============================================
  // CATEGORY ENDPOINTS
  // ============================================

  /** GET /api/tiktok/categories */
  async getCategories(req: Request, res: Response): Promise<any> {
    try {
      await this.configManager.loadConfig();
      const parentId = req.query.parent_id as string | undefined;
      const categories = await this.categoryService.getCategoryTree(parentId);
      return this.sendSuccess(res, {
        success: true,
        categories,
        total: categories.length,
      });
    } catch (error: any) {
      return this.sendError(res, error);
    }
  }

  /** GET /api/tiktok/categories/:categoryId/attributes */
  async getCategoryAttributes(req: Request, res: Response): Promise<any> {
    try {
      await this.configManager.loadConfig();
      const { categoryId } = req.params;
      const attributes =
        await this.categoryService.getCategoryAttributes(categoryId);
      return this.sendSuccess(res, {
        success: true,
        attributes,
        total: attributes.length,
      });
    } catch (error: any) {
      return this.sendError(res, error);
    }
  }

  /** GET /api/tiktok/categories/search?keyword=... */
  async searchCategories(req: Request, res: Response): Promise<any> {
    try {
      await this.configManager.loadConfig();
      const keyword = req.query.keyword as string;
      if (!keyword) {
        return res
          .status(400)
          .json({ success: false, error: "Keyword is required" });
      }
      const categories = await this.categoryService.searchCategories(keyword);
      return this.sendSuccess(res, {
        success: true,
        categories,
        total: categories.length,
      });
    } catch (error: any) {
      return this.sendError(res, error);
    }
  }

  /** POST /api/tiktok/categories/recommend */
  async recommendCategory(req: Request, res: Response): Promise<any> {
    try {
      await this.configManager.loadConfig();
      const { title, images } = req.body;
      if (!title) {
        return res
          .status(400)
          .json({ success: false, error: "title required" });
      }
      if (!images || images.length === 0) {
        return res.status(400).json({
          success: false,
          error:
            "At least one image URI required for TikTok category recommendation",
        });
      }
      const category = await this.categoryService.recommendCategory(
        title,
        images,
      );
      return this.sendSuccess(res, { success: true, category });
    } catch (error: any) {
      return this.sendError(res, error);
    }
  }

  /** GET /api/tiktok/categories/popular */
  async getPopularCategories(_req: Request, res: Response): Promise<any> {
    try {
      await this.configManager.loadConfig();
      const categories = await this.categoryService.getPopularCategories();
      return this.sendSuccess(res, {
        success: true,
        categories,
        total: categories.length,
      });
    } catch (error: any) {
      return this.sendError(res, error);
    }
  }

  // ============================================
  // BRAND & DELIVERY ENDPOINTS
  // ============================================

  /** GET /api/tiktok/brands?category_id=... */
  async getBrands(req: Request, res: Response): Promise<any> {
    try {
      await this.configManager.loadConfig();
      const categoryId = req.query.category_id as string;
      if (!categoryId) {
        return res
          .status(400)
          .json({ success: false, error: "category_id is required" });
      }
      const brands = await this.productCreateService.getBrands(categoryId);
      return this.sendSuccess(res, {
        success: true,
        brands,
        total: brands.length,
      });
    } catch (error: any) {
      return this.sendError(res, error);
    }
  }

  /** GET /api/tiktok/delivery-options */
  async getDeliveryOptions(_req: Request, res: Response): Promise<any> {
    try {
      await this.configManager.loadConfig();
      const deliveryOptions =
        await this.productCreateService.getDeliveryOptions();
      return this.sendSuccess(res, {
        success: true,
        deliveryOptions,
        total: deliveryOptions.length,
      });
    } catch (error: any) {
      return this.sendError(res, error);
    }
  }

  /** GET /api/tiktok/warehouses */
  async getWarehouses(_req: Request, res: Response): Promise<any> {
    try {
      await this.configManager.loadConfig();
      const warehouses = await this.productCreateService.getWarehouses();
      return this.sendSuccess(res, {
        success: true,
        warehouses,
        total: warehouses.length,
      });
    } catch (error: any) {
      return this.sendError(res, error);
    }
  }

  // ============================================
  // IMAGE UPLOAD ENDPOINT
  // ============================================

  /** POST /api/tiktok/images/upload */
  async uploadImage(req: Request, res: Response): Promise<any> {
    try {
      await this.configManager.loadConfig();
      const { image_url } = req.body;
      if (!image_url) {
        return res
          .status(400)
          .json({ success: false, error: "image_url is required" });
      }
      const hostedUrl = await this.productCreateService.uploadImage(image_url);
      if (!hostedUrl) {
        return res
          .status(400)
          .json({ success: false, error: "Failed to upload image" });
      }
      return this.sendSuccess(res, { success: true, imageUrl: hostedUrl });
    } catch (error: any) {
      return this.sendError(res, error);
    }
  }

  // ============================================
  // PRODUCT CREATE ENDPOINT
  // ============================================

  /** POST /api/tiktok/products/create */
  async createProduct(req: Request, res: Response): Promise<any> {
    try {
      await this.configManager.loadConfig();

      const productRequest: CreateProductRequest = {
        title: req.body.title,
        description: req.body.description,
        categoryId: req.body.category_id,
        brandId: req.body.brand_id,
        images: req.body.images || [],
        skus: this.parseSkus(req.body.skus),
        packageWeight: parseInt(req.body.package_weight) || 500,
        packageDimensions: this.parsePackageDimensions(
          req.body.package_dimensions,
        ),
        deliveryOptionIds: req.body.delivery_option_ids,
        saveMode: req.body.save_mode || "AS_DRAFT",
      };

      const result =
        await this.productCreateService.createProduct(productRequest);
      return this.sendSuccess(res, result, result.success ? 201 : 400);
    } catch (error: any) {
      return this.sendError(res, error);
    }
  }

  // ============================================
  // PRIVATE HELPERS
  // ============================================

  private parseSkus(skus: any[]): CreateProductRequest["skus"] {
    if (!skus || !Array.isArray(skus)) return [];
    return skus.map((sku: any) => ({
      sellerSku: sku.seller_sku || sku.sellerSku,
      price: parseFloat(sku.price) || 0,
      stock: parseInt(sku.stock) || 0,
      salesAttributes: sku.sales_attributes || sku.salesAttributes,
    }));
  }

  private parsePackageDimensions(
    dims: any,
  ): CreateProductRequest["packageDimensions"] {
    if (!dims) return undefined;
    return {
      length: parseFloat(dims.length) || 0,
      width: parseFloat(dims.width) || 0,
      height: parseFloat(dims.height) || 0,
    };
  }
}
