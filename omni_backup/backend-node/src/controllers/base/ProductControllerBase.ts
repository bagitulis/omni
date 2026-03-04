import { Response } from "express";
import { Logger } from "winston";
import { getLogger } from "../../utils/logger";
import { PrismaClient } from "@prisma/client";

/**
 * Product Controller Base Class
 * Provides common functionality for all product controllers
 */
export abstract class ProductControllerBase {
  protected logger: Logger;
  protected prisma: PrismaClient;

  constructor(prisma: PrismaClient) {
    this.logger = getLogger("ProductController");
    this.prisma = prisma;
  }

  /**
   * Send error response
   */
  protected sendError(res: Response, error: any, statusCode: number = 500) {
    this.logger.error(`Error: ${error.message}`);
    return res.status(statusCode).json({
      success: false,
      error: error.message,
    });
  }

  /**
   * Send success response
   */
  protected sendSuccess(res: Response, data: any, statusCode: number = 200) {
    return res.status(statusCode).json(data);
  }
}
