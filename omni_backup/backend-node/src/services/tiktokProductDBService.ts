/**
 * TikTok Product DB Service
 * SRP: Facade for TikTok product database operations
 * Delegates write operations to TiktokProductWriteService
 */

import { PrismaClient } from "@prisma/client";
import { Logger } from "winston";
import { getLogger } from "../utils/logger";
import { TiktokProductWriteService } from "./tiktokProductWriteService";

interface ProductData {
  id: string;
  title?: string;
  product_name?: string;
  description?: string;
  status?: string;
  price?: any;
  quantity?: number;
  cover_image_url?: string;
  images?: string[];
  skus?: any[];
}

export class TiktokProductDBService {
  private prisma: PrismaClient;
  private logger: Logger;
  private writeService: TiktokProductWriteService;

  constructor(prisma: PrismaClient) {
    this.prisma = prisma;
    this.logger = getLogger("TiktokProductDBService");
    this.writeService = new TiktokProductWriteService(prisma);
  }

  async getProductList(offset: number, limit: number): Promise<any> {
    const validOffset = Math.max(0, offset);
    const validLimit = Math.min(Math.max(1, limit), 1000);

    this.logger.info(`🔄 Retrieving TikTok products | Limit: ${validLimit}, Offset: ${validOffset}`);

    const [products, total] = await Promise.all([
      (this.prisma as any).tiktokProduct.findMany({
        skip: validOffset,
        take: validLimit,
        orderBy: { createdAt: "desc" },
        include: { skus: true },
      }),
      (this.prisma as any).tiktokProduct.count(),
    ]);

    this.logger.info(`✅ Retrieved ${products.length} products from DB`);
    return { products, total, offset: validOffset, limit: validLimit };
  }

  async getMasterProducts(offset: number, limit: number): Promise<any> {
    this.logger.info(`🔄 Getting TikTok master products | Offset: ${offset}, Limit: ${limit}`);

    const products = await (this.prisma as any).tiktokProduct.findMany({
      skip: offset,
      take: limit,
      orderBy: { createdAt: "desc" },
      include: { skus: true },
    });

    const total = await (this.prisma as any).tiktokProduct.count();
    const flattened = this.flattenProductsToSkus(products);

    this.logger.info(`✅ Retrieved ${flattened.length} product SKUs (Total Products: ${total})`);
    return { flattened, total, offset, limit };
  }

  private flattenProductsToSkus(products: any[]): any[] {
    const flattened: any[] = [];
    products.forEach((product: any) => {
      if (product.skus?.length > 0) {
        product.skus.forEach((sku: any) => {
          flattened.push({
            item_id: product.productId,
            sku_id: sku.skuId,
            sku_name: sku.sellerSku || `SKU-${sku.skuId}`,
            variant_name: sku.variantName || "-",
            item_name: product.name,
            price: sku.price || product.price || 0,
            quantity: sku.quantity || product.quantity || 0,
            status: product.status,
            updated_at: product.updatedAt,
            _original: { product, sku },
          });
        });
      } else {
        flattened.push({
          item_id: product.productId,
          sku_id: "-",
          sku_name: "-",
          variant_name: "-",
          item_name: product.name,
          price: product.price || 0,
          quantity: product.quantity || 0,
          status: product.status,
          updated_at: product.updatedAt,
          _original: product,
        });
      }
    });
    return flattened;
  }

  async getProductById(productId: string): Promise<any> {
    this.logger.info(`🔄 Retrieving product by ID | ID: ${productId}`);

    const product = await (this.prisma as any).tiktokProduct.findFirst({
      where: { productId },
      include: { skus: true },
    });

    if (!product) throw new Error("Product not found");

    this.logger.info(`✅ Retrieved product ${productId} from DB`);
    return product;
  }

  async searchProducts(query: string, offset: number, limit: number): Promise<any> {
    const validOffset = Math.max(0, offset);
    const validLimit = Math.min(Math.max(1, limit), 1000);

    if (!query) throw new Error("Search query required");

    this.logger.info(`🔄 Searching TikTok products | Query: "${query}"`);

    const [products, total] = await Promise.all([
      (this.prisma as any).tiktokProduct.findMany({
        where: { OR: [{ name: { contains: query } }, { description: { contains: query } }] },
        skip: validOffset,
        take: validLimit,
        orderBy: { createdAt: "desc" },
        include: { skus: true },
      }),
      (this.prisma as any).tiktokProduct.count({
        where: { OR: [{ name: { contains: query } }, { description: { contains: query } }] },
      }),
    ]);

    this.logger.info(`✅ Found ${products.length} products matching "${query}"`);
    return { products, total, offset: validOffset, limit: validLimit, query };
  }

  async getProductsByStatus(status: string, offset: number, limit: number): Promise<any> {
    const validOffset = Math.max(0, offset);
    const validLimit = Math.min(Math.max(1, limit), 1000);

    this.logger.info(`🔄 Retrieving TikTok products by status | Status: ${status}`);

    const [products, total] = await Promise.all([
      (this.prisma as any).tiktokProduct.findMany({
        where: { status },
        skip: validOffset,
        take: validLimit,
        orderBy: { createdAt: "desc" },
        include: { skus: true },
      }),
      (this.prisma as any).tiktokProduct.count({ where: { status } }),
    ]);

    this.logger.info(`✅ Retrieved ${products.length} products with status ${status}`);
    return { products, total, offset: validOffset, limit: validLimit, status };
  }

  async getStatistics(): Promise<any> {
    this.logger.info("🔄 Retrieving TikTok database statistics");

    const [totalProducts, statusCounts, totalQuantity, recentProducts] = await Promise.all([
      (this.prisma as any).tiktokProduct.count(),
      (this.prisma as any).tiktokProduct.groupBy({ by: ["status"], _count: true }),
      (this.prisma as any).tiktokProduct.aggregate({ _sum: { quantity: true } }),
      (this.prisma as any).tiktokProduct.findMany({ take: 5, orderBy: { updatedAt: "desc" } }),
    ]);

    const statusBreakdown: Record<string, number> = {};
    statusCounts.forEach((item: any) => { statusBreakdown[item.status] = item._count; });

    this.logger.info("✅ Database statistics retrieved");

    return {
      total_products: totalProducts,
      status_breakdown: statusBreakdown,
      total_quantity: totalQuantity._sum.quantity || 0,
      recent_updates: recentProducts.length,
    };
  }

  // Delegate write operations
  async deleteProduct(productId: string): Promise<any> {
    return this.writeService.deleteProduct(productId);
  }

  async saveProducts(products: ProductData[], cleanBeforeSave = false): Promise<void> {
    return this.writeService.saveProducts(products, cleanBeforeSave);
  }
}
