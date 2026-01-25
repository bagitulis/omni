/**
 * Lazada Product Create Controller
 * Handles Lazada product creation endpoints
 * Max 300 lines - AGENTS.MD compliant
 */

import { Request, Response } from "express";
import { PrismaClient } from "@prisma/client";
import { ProductControllerBase } from "./base/ProductControllerBase";
import { LazadaAPIClient } from "../api/clients/lazadaAPIClient";
import { LazadaConfigManager } from "../config/managers/lazadaConfigManager";
import { LazadaCategoryService } from "../services/lazadaCategoryService";
import { LazadaProductCreateService } from "../services/lazadaProductCreateService";

export class LazadaProductCreateController extends ProductControllerBase {
  private configManager: LazadaConfigManager;
  private categoryService: LazadaCategoryService;
  private productCreateService: LazadaProductCreateService;

  constructor(
    prisma: PrismaClient,
    apiClient: LazadaAPIClient,
    configManager: LazadaConfigManager
  ) {
    super(prisma);
    this.configManager = configManager;
    this.categoryService = new LazadaCategoryService(apiClient);
    this.productCreateService = new LazadaProductCreateService(apiClient);
  }

  /** GET /api/lazada/categories */
  async getCategories(req: Request, res: Response): Promise<any> {
    try {
      await this.configManager.loadConfig();
      const parentId = req.query.parent_id ? parseInt(req.query.parent_id as string) : undefined;
      const categories = await this.categoryService.getFlatCategories(parentId);
      return this.sendSuccess(res, { success: true, categories, total: categories.length });
    } catch (error: any) {
      return this.sendError(res, error);
    }
  }

  /** GET /api/lazada/categories/:categoryId/attributes */
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

  /** GET /api/lazada/categories/search */
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

  /** GET /api/lazada/categories/recommend */
  async recommendCategory(req: Request, res: Response): Promise<any> {
    try {
      await this.configManager.loadConfig();
      const productTitle = req.query.title as string || req.query.product_name as string;
      if (!productTitle) return res.status(400).json({ success: false, error: "title required" });
      const category = await this.categoryService.recommendCategory(productTitle);
      return this.sendSuccess(res, { success: true, category });
    } catch (error: any) {
      return this.sendError(res, error);
    }
  }

  /** GET /api/lazada/categories/popular */
  async getPopularCategories(_req: Request, res: Response): Promise<any> {
    try {
      await this.configManager.loadConfig();
      const categories = await this.categoryService.getPopularCategories();
      return this.sendSuccess(res, { success: true, categories, total: categories.length });
    } catch (error: any) {
      return this.sendError(res, error);
    }
  }

  /** GET /api/lazada/brands */
  async getBrands(req: Request, res: Response): Promise<any> {
    try {
      await this.configManager.loadConfig();
      const keyword = req.query.keyword as string || req.query.name as string || "";
      const brands = await this.productCreateService.getBrands(keyword);
      return this.sendSuccess(res, { success: true, brands, total: brands.length });
    } catch (error: any) {
      return this.sendError(res, error);
    }
  }

  /** POST /api/lazada/images/upload */
  async uploadImage(req: Request, res: Response): Promise<any> {
    try {
      await this.configManager.loadConfig();
      const { image_url } = req.body;
      if (!image_url) return res.status(400).json({ success: false, error: "image_url required" });
      const imageUrl = await this.productCreateService.migrateImage(image_url);
      if (!imageUrl) return res.status(400).json({ success: false, error: "Upload failed" });
      return this.sendSuccess(res, { success: true, imageUrl });
    } catch (error: any) {
      return this.sendError(res, error);
    }
  }

  /** POST /api/lazada/products/create */
  async createProduct(req: Request, res: Response): Promise<any> {
    try {
      await this.configManager.loadConfig();

      const result = await this.productCreateService.createProduct({
        name: req.body.title || req.body.name,
        description: req.body.description,
        categoryId: parseInt(req.body.category_id),
        brandName: req.body.brand_name,
        images: req.body.images || [],
        quantity: parseInt(req.body.stock) || parseInt(req.body.skus?.[0]?.stock) || 0,
        price: parseFloat(req.body.price) || parseFloat(req.body.skus?.[0]?.price) || 0,
        sellerSku: req.body.sku || req.body.skus?.[0]?.seller_sku,
        packageWeight: parseFloat(req.body.package_weight) / 1000 || 0.1, // Convert grams to kg
        packageDimensions: req.body.package_dimensions,
      });

      return this.sendSuccess(res, result, result.success ? 201 : 400);
    } catch (error: any) {
      return this.sendError(res, error);
    }
  }
}
