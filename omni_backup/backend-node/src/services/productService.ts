import { Logger } from "winston";
import { getLogger } from "../utils/logger";
import { PrismaClient } from "@prisma/client";

/**
 * Product Service - Base class for product operations
 * Handles product fetching, caching, and database operations
 */
export abstract class ProductService {
  protected logger: Logger;
  protected prisma: PrismaClient;

  constructor(prisma: PrismaClient) {
    this.logger = getLogger("ProductService");
    this.prisma = prisma;
  }

  /**
   * Validate pagination parameters
   */
  protected validatePagination(
    offset: number = 0,
    limit: number = 100
  ): { offset: number; limit: number } {
    const maxLimit = 1000;
    offset = Math.max(0, offset);
    limit = Math.min(Math.max(1, limit), maxLimit);
    return { offset, limit };
  }

  abstract getProducts(params: any): Promise<any>;
  abstract getMasterProductsFromDb(params: any): Promise<any>;
}

