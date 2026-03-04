/**
 * Shopee Product Detail Controller
 * SRP: Handle individual product detail queries
 */

import { Request, Response } from "express";
import { PrismaClient } from "@prisma/client";
import { ProductControllerBase } from "./base/ProductControllerBase";

export class ShopeeProductDetailController extends ProductControllerBase {
  constructor(prisma: PrismaClient) {
    super(prisma);
  }

  async getProductBase(req: Request, res: Response): Promise<any> {
    try {
      const itemId = req.params.itemId;
      const product = await this.prisma.shopeeProduct.findFirst({
        where: { itemId: BigInt(itemId) },
      });

      if (!product) {
        return this.sendError(res, new Error("Product not found"), 404);
      }

      return this.sendSuccess(res, {
        success: true,
        data: {
          item_id: Number(product.itemId),
          item_name: product.name,
          description: product.description,
          image: product.image,
          price: product.price,
          stock: product.quantity,
          item_status: product.status,
        },
      });
    } catch (error: any) {
      return this.sendError(res, error);
    }
  }

  async getProductModels(req: Request, res: Response): Promise<any> {
    try {
      const itemId = req.params.itemId;
      const product = await this.prisma.shopeeProduct.findFirst({
        where: { itemId: BigInt(itemId) },
        include: { skus: true },
      });

      if (!product) {
        return this.sendError(res, new Error("Product not found"), 404);
      }

      const mappedModels = product.skus.map((s: any) => ({
        item_id: Number(product.itemId),
        model_id: Number(s.modelId),
        model_sku: s.sellerSku,
        model_name: s.variantName,
        price: s.price,
        stock: s.quantity,
      }));

      return this.sendSuccess(res, { success: true, data: mappedModels });
    } catch (error: any) {
      return this.sendError(res, error);
    }
  }

  async getProductVariations(req: Request, res: Response): Promise<any> {
    try {
      const itemId = req.params.itemId;
      const product = await this.prisma.shopeeProduct.findFirst({
        where: { itemId: BigInt(itemId) },
        include: { skus: true },
      });

      if (!product) {
        return this.sendError(res, new Error("Product not found"), 404);
      }

      return this.sendSuccess(res, {
        success: true,
        data: product.skus.map((s: any) => ({
          variation_id: Number(s.modelId),
          variation_sku: s.sellerSku,
          variation_name: s.variantName,
          price: s.price,
          stock: s.quantity,
        })),
      });
    } catch (error: any) {
      return this.sendError(res, error);
    }
  }

  async getProductFull(req: Request, res: Response): Promise<any> {
    try {
      const itemId = req.params.itemId;
      const product = await this.prisma.shopeeProduct.findFirst({
        where: { itemId: BigInt(itemId) },
        include: { skus: true },
      });

      if (!product) {
        return this.sendError(res, new Error("Product not found"), 404);
      }

      return this.sendSuccess(res, {
        success: true,
        data: {
          item_id: Number(product.itemId),
          item_name: product.name,
          description: product.description,
          image: product.image,
          price: product.price,
          stock: product.quantity,
          item_status: product.status,
          models: product.skus.map((s: any) => ({
            model_id: Number(s.modelId),
            model_sku: s.sellerSku,
            model_name: s.variantName,
            price: s.price,
            stock: s.quantity,
          })),
        },
      });
    } catch (error: any) {
      return this.sendError(res, error);
    }
  }

  async getProductDetail(req: Request, res: Response): Promise<any> {
    try {
      const itemId = req.params.itemId;
      const product = await this.prisma.shopeeProduct.findFirst({
        where: { itemId: BigInt(itemId) },
        include: { skus: true },
      });

      if (!product) {
        return this.sendSuccess(res, {
          success: false,
          error: "Product not found",
        });
      }

      return this.sendSuccess(res, {
        success: true,
        product: {
          itemId: product.itemId.toString(),
          itemName: product.name,
          description: product.description,
          status: product.status,
          image: product.image,
          models: product.skus.map((sku: any) => ({
            modelId: sku.modelId.toString(),
            modelName: sku.variantName,
            modelSku: sku.sellerSku,
            currentPrice: sku.price,
            originalPrice: sku.price,
            sellerStock: sku.quantity,
            shopeeStock: sku.quantity,
          })),
        },
      });
    } catch (error: any) {
      return this.sendError(res, error);
    }
  }
}
