/**
 * Shopee Product Create Controller
 * Handles Shopee product creation endpoints
 * Max 300 lines - AGENTS.MD compliant
 */

import { Request, Response } from "express";
import { PrismaClient } from "@prisma/client";
import { ProductControllerBase } from "./base/ProductControllerBase";
import { ShopeeAPIClient } from "../api/clients/shopeeAPIClient";
import { ShopeeConfigManager } from "../config/managers/shopeeConfigManager";
import { ShopeeCategoryService } from "../services/shopeeCategoryService";
import { ShopeeProductCreateService } from "../services/shopeeProductCreateService";

export class ShopeeProductCreateController extends ProductControllerBase {
  private configManager: ShopeeConfigManager;
  private categoryService: ShopeeCategoryService;
  private productCreateService: ShopeeProductCreateService;

  constructor(
    prisma: PrismaClient,
    apiClient: ShopeeAPIClient,
    configManager: ShopeeConfigManager
  ) {
    super(prisma);
    this.configManager = configManager;
    this.categoryService = new ShopeeCategoryService(apiClient);
    this.productCreateService = new ShopeeProductCreateService(apiClient);
  }

  /** GET /api/shopee/categories */
  async getCategories(req: Request, res: Response): Promise<any> {
    try {
      await this.configManager.loadConfig();
      const parentId = req.query.parent_id ? parseInt(req.query.parent_id as string) : undefined;
      const categories = await this.categoryService.getCategoryTree(parentId);
      return this.sendSuccess(res, { success: true, categories, total: categories.length });
    } catch (error: any) {
      return this.sendError(res, error);
    }
  }

  /** GET /api/shopee/categories/:categoryId/attributes */
  async getCategoryAttributes(req: Request, res: Response): Promise<any> {
    try {
      await this.configManager.loadConfig();
      const categoryId = parseInt(req.params.categoryId);
      const attributes = await this.categoryService.getCategoryAttributes(categoryId);
      return this.sendSuccess(res, { success: true, attributes, total: attributes.length });
    } catch (error: any) {
      return this.sendError(res, error);
    }
  }

  /** GET /api/shopee/categories/search */
  async searchCategories(req: Request, res: Response): Promise<any> {
    try {
      await this.configManager.loadConfig();
      const keyword = req.query.keyword as string;
      if (!keyword) return res.status(400).json({ success: false, error: "Keyword required" });
      const categories = await this.categoryService.searchCategories(keyword);
      return this.sendSuccess(res, { success: true, categories, total: categories.length });
    } catch (error: any) {
      return this.sendError(res, error);
    }
  }

  /** GET /api/shopee/categories/recommend */
  async recommendCategory(req: Request, res: Response): Promise<any> {
    try {
      await this.configManager.loadConfig();
      const itemName = req.query.item_name as string || req.query.title as string;
      const coverImage = req.query.cover_image as string;
      if (!itemName) return res.status(400).json({ success: false, error: "item_name required" });
      const category = await this.categoryService.recommendCategory(itemName, coverImage);
      return this.sendSuccess(res, { success: true, category });
    } catch (error: any) {
      return this.sendError(res, error);
    }
  }

  /** GET /api/shopee/categories/popular */
  async getPopularCategories(_req: Request, res: Response): Promise<any> {
    try {
      await this.configManager.loadConfig();
      const categories = await this.categoryService.getPopularCategories();
      return this.sendSuccess(res, { success: true, categories, total: categories.length });
    } catch (error: any) {
      return this.sendError(res, error);
    }
  }

  /** GET /api/shopee/brands */
  async getBrands(req: Request, res: Response): Promise<any> {
    try {
      await this.configManager.loadConfig();
      const categoryId = parseInt(req.query.category_id as string);
      if (!categoryId) return res.status(400).json({ success: false, error: "category_id required" });
      const brands = await this.productCreateService.getBrands(categoryId);
      return this.sendSuccess(res, { success: true, brands, total: brands.length });
    } catch (error: any) {
      return this.sendError(res, error);
    }
  }

  /** GET /api/shopee/logistics */
  async getLogistics(_req: Request, res: Response): Promise<any> {
    try {
      await this.configManager.loadConfig();
      const logistics = await this.productCreateService.getLogistics();
      return this.sendSuccess(res, { success: true, logistics, total: logistics.length });
    } catch (error: any) {
      return this.sendError(res, error);
    }
  }

  /** POST /api/shopee/images/upload */
  async uploadImage(req: Request, res: Response): Promise<any> {
    try {
      await this.configManager.loadConfig();
      const { image_url } = req.body;
      if (!image_url) return res.status(400).json({ success: false, error: "image_url required" });
      const imageId = await this.productCreateService.uploadImage(image_url);
      if (!imageId) return res.status(400).json({ success: false, error: "Upload failed" });
      return this.sendSuccess(res, { success: true, imageId });
    } catch (error: any) {
      return this.sendError(res, error);
    }
  }

  /** POST /api/shopee/products/create */
  async createProduct(req: Request, res: Response): Promise<any> {
    try {
      await this.configManager.loadConfig();

      const result = await this.productCreateService.createProduct({
        item_name: req.body.item_name || req.body.title,
        description: req.body.description,
        category_id: parseInt(req.body.category_id),
        brand: req.body.brand_id ? { brand_id: parseInt(req.body.brand_id), original_brand_name: "" } : undefined,
        images: req.body.images || [],
        normal_stock: parseInt(req.body.stock) || parseInt(req.body.skus?.[0]?.stock) || 0,
        original_price: parseFloat(req.body.price) || parseFloat(req.body.skus?.[0]?.price) || 0,
        seller_sku: req.body.sku || req.body.skus?.[0]?.seller_sku,
        weight: parseFloat(req.body.weight) || parseFloat(req.body.package_weight) / 1000 || 0.1,
        dimension: req.body.package_dimensions,
        logistic_info: req.body.logistics,
        condition: req.body.condition || "NEW",
      });

      return this.sendSuccess(res, result, result.success ? 201 : 400);
    } catch (error: any) {
      return this.sendError(res, error);
    }
  }
}
