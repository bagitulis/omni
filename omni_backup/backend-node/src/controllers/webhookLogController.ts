/**
 * Webhook Log Controller
 * SRP: Handle webhook log queries and statistics
 */

import { Request, Response } from "express";
import { WebhookService } from "../services/webhookService";
import { getLogger } from "../utils/logger";
import { getDbManager } from "../services/dbConnectionManager";

const logger = getLogger("WebhookLogController");

export class WebhookLogController {
  private webhookService: WebhookService;

  constructor(webhookService: WebhookService) {
    this.webhookService = webhookService;
  }

  /**
   * Validate tenant from request header
   */
  private validateTenant(req: Request, res: Response): string | null {
    const dbManager = getDbManager();
    const tenants = dbManager.getAllTenants();
    const headerTenantId = req.headers["x-tenant-id"] as string;

    if (!headerTenantId) {
      logger.warn("No tenant ID provided in request header");
      res.status(400).json({
        success: false,
        error: "Tenant ID required. Please login again.",
        code: "TENANT_REQUIRED",
      });
      return null;
    }

    if (!tenants.includes(headerTenantId)) {
      logger.warn(`Invalid tenant ID: ${headerTenantId}`);
      res.status(400).json({
        success: false,
        error: `Invalid tenant: ${headerTenantId}`,
        code: "INVALID_TENANT",
      });
      return null;
    }

    return headerTenantId;
  }

  /**
   * Get webhook logs with pagination
   */
  async getWebhookLogs(req: Request, res: Response): Promise<void> {
    try {
      const tenantId = this.validateTenant(req, res);
      if (!tenantId) return;

      logger.info(`Fetching webhook logs for tenant: ${tenantId}`);

      const page = parseInt(req.query.page as string) || 1;
      const limit = parseInt(req.query.limit as string) || 50;
      const status = req.query.status as string;

      const logs = await this.webhookService.getWebhookLogs(tenantId, {
        page,
        limit,
        status,
      });

      res.json({
        success: true,
        data: logs.items,
        pagination: {
          page,
          limit,
          total: logs.total,
          totalPages: Math.ceil(logs.total / limit),
        },
      });
    } catch (error: any) {
      logger.error(`Failed to get webhook logs: ${error.message}`);
      res.status(500).json({ success: false, error: error.message });
    }
  }

  /**
   * Get webhook logs filtered by platform
   */
  async getWebhookLogsByPlatform(req: Request, res: Response): Promise<void> {
    try {
      const tenantId = this.validateTenant(req, res);
      if (!tenantId) return;

      const platform = req.params.platform;
      const page = parseInt(req.query.page as string) || 1;
      const limit = parseInt(req.query.limit as string) || 50;

      const logs = await this.webhookService.getWebhookLogsByPlatform(
        tenantId,
        platform,
        { page, limit }
      );

      res.json({
        success: true,
        data: logs.items,
        pagination: {
          page,
          limit,
          total: logs.total,
          totalPages: Math.ceil(logs.total / limit),
        },
      });
    } catch (error: any) {
      logger.error(`Failed to get platform webhook logs: ${error.message}`);
      res.status(500).json({ success: false, error: error.message });
    }
  }

  /**
   * Get webhook statistics
   */
  async getWebhookStats(req: Request, res: Response): Promise<void> {
    try {
      const tenantId = this.validateTenant(req, res);
      if (!tenantId) return;

      const stats = await this.webhookService.getWebhookStats(tenantId);

      res.json({ success: true, data: stats });
    } catch (error: any) {
      logger.error(`Failed to get webhook stats: ${error.message}`);
      res.status(500).json({ success: false, error: error.message });
    }
  }
}
