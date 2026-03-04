/**
 * Webhook Controller
 * SRP: Handle incoming webhooks from marketplace platforms
 * Delegates log queries to WebhookLogController
 */

import { Request, Response } from "express";
import { WebhookService } from "../services/webhookService";
import { getLogger } from "../utils/logger";
import { WebhookLogController } from "./webhookLogController";

const logger = getLogger("WebhookController");

export class WebhookController {
  private webhookService: WebhookService;
  private logController: WebhookLogController;

  constructor() {
    this.webhookService = new WebhookService();
    this.logController = new WebhookLogController(this.webhookService);
  }

  /**
   * Handle Shopee webhook
   */
  async handleShopeeWebhook(req: Request, res: Response, tenantId?: string): Promise<void> {
    try {
      logger.info(`Shopee webhook received${tenantId ? ` for tenant: ${tenantId}` : ""}`);

      const payload = req.body || {};
      const headers = this.extractHeaders(req);

      // Handle Shopee verification push
      if (payload.code === 0 && payload.data?.verify_info) {
        logger.info("Shopee verification push received");
        res.status(200).json({ code: 0, message: "success" });
        return;
      }

      // Construct webhook URL for signature verification
      const protocol = req.headers["x-forwarded-proto"] || "https";
      const host = req.headers["host"] || "yndigital.my.id";
      const webhookUrl = `${protocol}://${host}/api/webhooks/${tenantId}/shopee`;

      const isValid = await this.webhookService.verifyShopeeSignature(
        payload, headers, tenantId, req.rawBody, webhookUrl
      );

      if (!isValid) {
        logger.warn("Invalid Shopee webhook signature");
        res.status(401).json({ error: "Invalid signature" });
        return;
      }

      try {
        await this.webhookService.processWebhook("shopee", payload, headers, tenantId);
      } catch (processError: any) {
        logger.warn(`Webhook processing warning: ${processError.message}`);
      }

      res.status(200).json({ code: 0, message: "success" });
    } catch (error: any) {
      logger.error(`Shopee webhook error: ${error.message}`);
      res.status(200).json({ code: 0, message: "success" });
    }
  }

  /**
   * Handle TikTok webhook
   */
  async handleTiktokWebhook(req: Request, res: Response, tenantId?: string): Promise<void> {
    try {
      const payload = req.body;
      const headers = this.extractHeaders(req);

      logger.info(`TikTok webhook received${tenantId ? ` for tenant: ${tenantId}` : ""}: ${payload?.type || "unknown"}`);

      const isValid = await this.webhookService.verifyTiktokSignature(req.body, headers);
      if (!isValid) {
        logger.warn("Invalid TikTok webhook signature");
        res.status(401).json({ error: "Invalid signature" });
        return;
      }

      await this.webhookService.processWebhook("tiktok", payload, headers, tenantId);
      res.status(200).json({ success: true });
    } catch (error: any) {
      logger.error(`TikTok webhook error: ${error.message}`);
      res.status(500).json({ error: error.message });
    }
  }

  /**
   * Handle Lazada webhook
   */
  async handleLazadaWebhook(req: Request, res: Response, tenantId?: string): Promise<void> {
    try {
      const payload = req.body;
      const headers = this.extractHeaders(req);

      logger.info(`Lazada webhook received${tenantId ? ` for tenant: ${tenantId}` : ""}: ${payload?.type || "unknown"}`);

      const isValid = await this.webhookService.verifyLazadaSignature(req.body, headers);
      if (!isValid) {
        logger.warn("Invalid Lazada webhook signature");
        res.status(401).json({ error: "Invalid signature" });
        return;
      }

      await this.webhookService.processWebhook("lazada", payload, headers, tenantId);
      res.status(200).json({ success: true });
    } catch (error: any) {
      logger.error(`Lazada webhook error: ${error.message}`);
      res.status(500).json({ error: error.message });
    }
  }

  // Delegate log operations
  async getWebhookLogs(req: Request, res: Response): Promise<void> {
    return this.logController.getWebhookLogs(req, res);
  }

  async getWebhookLogsByPlatform(req: Request, res: Response): Promise<void> {
    return this.logController.getWebhookLogsByPlatform(req, res);
  }

  async getWebhookStats(req: Request, res: Response): Promise<void> {
    return this.logController.getWebhookStats(req, res);
  }

  /**
   * Test webhook endpoint
   */
  async testWebhook(req: Request, res: Response): Promise<void> {
    try {
      const platform = req.params.platform;
      logger.info(`Test webhook for ${platform}`);
      res.json({ success: true, message: `${platform} webhook endpoint is working`, timestamp: new Date().toISOString() });
    } catch (error: any) {
      res.status(500).json({ success: false, error: error.message });
    }
  }

  private extractHeaders(req: Request): Record<string, string> {
    const relevantHeaders = ["authorization", "x-shopee-signature", "x-tiktok-signature", "x-lazada-signature", "content-type", "user-agent"];
    const headers: Record<string, string> = {};
    for (const header of relevantHeaders) {
      const value = req.headers[header];
      if (value) headers[header] = Array.isArray(value) ? value[0] : value;
    }
    return headers;
  }
}
