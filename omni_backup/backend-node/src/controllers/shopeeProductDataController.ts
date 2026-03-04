/**
 * Shopee Product Data Controller
 * SRP: Handle product list/data retrieval
 * Detail queries delegated to ShopeeProductDetailController
 */

import { Request, Response } from "express";
import { PrismaClient } from "@prisma/client";
import { ProductControllerBase } from "./base/ProductControllerBase";
import { ShopeeProductService } from "../services/shopeeProductService";
import { ShopeeAPIClient } from "../api/clients/shopeeAPIClient";
import { ShopeeConfigManager } from "../config/managers/shopeeConfigManager";
import { ShopeeProductDetailController } from "./shopeeProductDetailController";

export class ShopeeProductDataController extends ProductControllerBase {
  private productService: ShopeeProductService;
  private configManager: ShopeeConfigManager;
  private detailController: ShopeeProductDetailController;

  constructor(prisma: PrismaClient, apiClient: ShopeeAPIClient, configManager: ShopeeConfigManager) {
    super(prisma);
    this.configManager = configManager;
    this.productService = new ShopeeProductService(prisma, apiClient, configManager);
    this.detailController = new ShopeeProductDetailController(prisma);
  }

  async getProducts(req: Request, res: Response): Promise<any> {
    try {
      await this.configManager.loadConfig();
      const itemStatus = (req.query.itemStatus as string) || "NORMAL";
      const offset = parseInt(req.query.offset as string) || 0;
      const limit = parseInt(req.query.limit as string) || 50;

      const result = await this.productService.getProducts({ itemStatus, offset, limit });
      return this.sendSuccess(res, result, result.success ? 200 : 400);
    } catch (error: any) {
      return this.sendError(res, error);
    }
  }

  async getMasterProductsFromDb(_req: Request, res: Response): Promise<any> {
    try {
      res.setHeader("Cache-Control", "no-store");
      res.setHeader("Pragma", "no-cache");
      res.setHeader("Expires", "0");

      const result = await this.productService.getMasterProductsFromDbFormatted();
      return this.sendSuccess(res, result, result.success ? 200 : 400);
    } catch (error: any) {
      return this.sendError(res, error);
    }
  }

  async getMasterProductList(req: Request, res: Response): Promise<any> {
    try {
      const limit = parseInt(req.query.limit as string) || 100;
      const offset = parseInt(req.query.offset as string) || 0;

      const result = await this.productService.getMasterProductsList({ limit, offset });
      return this.sendSuccess(res, result, result.success ? 200 : 400);
    } catch (error: any) {
      return this.sendError(res, error);
    }
  }

  async getProductList(req: Request, res: Response): Promise<any> {
    try {
      const offset = parseInt(req.query.offset as string) || 0;
      const limit = parseInt(req.query.limit as string) || 100;
      const status = req.query.status as string;

      const where: any = {};
      if (status) where.status = status;

      const products = await this.prisma.shopeeProduct.findMany({ where, skip: offset, take: limit, orderBy: { updatedAt: "desc" } });
      const total = await this.prisma.shopeeProduct.count({ where });

      const mappedProducts = products.map((p) => ({
        item_id: Number(p.itemId),
        item_status: p.status,
        update_time: Math.floor(p.updatedAt.getTime() / 1000),
      }));

      return this.sendSuccess(res, { success: true, data: mappedProducts, total });
    } catch (error: any) {
      return this.sendError(res, error);
    }
  }

  async getProductBaseList(req: Request, res: Response): Promise<any> {
    try {
      const offset = parseInt(req.query.offset as string) || 0;
      const limit = parseInt(req.query.limit as string) || 100;

      const products = await this.prisma.shopeeProduct.findMany({ skip: offset, take: limit, orderBy: { updatedAt: "desc" } });
      const total = await this.prisma.shopeeProduct.count();

      const mappedProducts = products.map((p) => ({
        item_id: Number(p.itemId),
        item_name: p.name,
        description: p.description,
        image: p.image,
        price: p.price,
        stock: p.quantity,
        item_status: p.status,
        create_time: Math.floor(p.createdAt.getTime() / 1000),
        update_time: Math.floor(p.updatedAt.getTime() / 1000),
      }));

      return this.sendSuccess(res, { success: true, data: mappedProducts, total });
    } catch (error: any) {
      return this.sendError(res, error);
    }
  }

  async getProductModelList(req: Request, res: Response): Promise<any> {
    try {
      const offset = parseInt(req.query.offset as string) || 0;
      const limit = parseInt(req.query.limit as string) || 100;

      const skus = await this.prisma.shopeeSku.findMany({ skip: offset, take: limit, include: { product: true }, orderBy: { updatedAt: "desc" } });
      const total = await this.prisma.shopeeSku.count();

      const mappedModels = skus.map((s) => ({
        item_id: Number(s.product.itemId),
        model_id: Number(s.modelId),
        model_sku: s.sellerSku,
        model_name: s.variantName,
        price: s.price,
        stock: s.quantity,
        update_time: Math.floor(s.updatedAt.getTime() / 1000),
      }));

      return this.sendSuccess(res, { success: true, data: mappedModels, total });
    } catch (error: any) {
      return this.sendError(res, error);
    }
  }

  // Delegate detail operations
  async getProductBase(req: Request, res: Response): Promise<any> {
    return this.detailController.getProductBase(req, res);
  }

  async getProductModels(req: Request, res: Response): Promise<any> {
    return this.detailController.getProductModels(req, res);
  }

  async getProductVariations(req: Request, res: Response): Promise<any> {
    return this.detailController.getProductVariations(req, res);
  }

  async getProductFull(req: Request, res: Response): Promise<any> {
    return this.detailController.getProductFull(req, res);
  }

  async getProductDetail(req: Request, res: Response): Promise<any> {
    return this.detailController.getProductDetail(req, res);
  }
}
