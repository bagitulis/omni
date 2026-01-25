import { getLogger } from "../utils/logger";
import {
  verifyShopeeSignature,
  verifyTiktokSignature,
  verifyLazadaSignature,
} from "./webhook/webhookVerification";
import {
  logWebhook,
  updateWebhookStatus,
  getWebhookLogs,
  getWebhookLogsByPlatform,
  getWebhookStats,
  WebhookLogEntry,
  PaginatedResult,
  WebhookStats,
} from "./webhook/webhookLogger";
import { handleWebhookEvent } from "./webhook/webhookEventHandler";
import { getPushCodeName } from "../constants/shopeePushCodes";

const logger = getLogger("WebhookService");

type PlatformType = "shopee" | "tiktok" | "lazada";

/**
 * Webhook Service
 * Handles webhook verification, processing, and logging
 */
export class WebhookService {
  /**
   * Verify Shopee webhook signature
   * @param payload - Parsed request body
   * @param headers - Request headers
   * @param tenantId - Optional tenant ID to get tenant-specific partner key
   * @param rawBody - Raw request body string for signature verification
   * @param webhookUrl - Full webhook URL for signature computation
   */
  async verifyShopeeSignature(
    payload: unknown,
    headers: Record<string, string>,
    tenantId?: string,
    rawBody?: string,
    webhookUrl?: string
  ): Promise<boolean> {
    return verifyShopeeSignature(
      payload,
      headers,
      tenantId,
      rawBody,
      webhookUrl
    );
  }

  /**
   * Verify TikTok webhook signature
   */
  async verifyTiktokSignature(
    payload: unknown,
    headers: Record<string, string>
  ): Promise<boolean> {
    return verifyTiktokSignature(payload, headers);
  }

  /**
   * Verify Lazada webhook signature
   */
  async verifyLazadaSignature(
    payload: unknown,
    headers: Record<string, string>
  ): Promise<boolean> {
    return verifyLazadaSignature(payload, headers);
  }

  /**
   * Process incoming webhook
   * @param tenantId - Optional tenant ID from URL path
   */
  async processWebhook(
    platform: PlatformType,
    payload: unknown,
    headers: Record<string, string>,
    tenantId?: string
  ): Promise<{ success: boolean; logId?: string; error?: string }> {
    // Extract event type
    const eventType = this.extractEventType(platform, payload);

    // Log webhook with tenant ID
    const logId = await logWebhook(
      platform,
      eventType,
      payload,
      headers,
      tenantId
    );

    try {
      // Process based on event type - updates database
      await this.handleEvent(platform, eventType, payload, tenantId);
      await updateWebhookStatus(logId, "processed");

      return { success: true, logId };
    } catch (error: any) {
      await updateWebhookStatus(logId, "failed", error.message);
      logger.error(`Webhook processing failed: ${error.message}`);
      return { success: false, logId, error: error.message };
    }
  }

  /**
   * Extract event type from payload
   * For Shopee: converts numeric code to readable name (e.g., 30 -> "package_fulfillment_status_push")
   */
  private extractEventType(platform: PlatformType, payload: unknown): string {
    const p = payload as Record<string, unknown>;

    switch (platform) {
      case "shopee": {
        const code = p.code as number;
        if (code !== undefined) {
          // Return readable name from push code mapping
          return getPushCodeName(code);
        }
        return "unknown";
      }
      case "tiktok":
        return (p.type as string) || "unknown";
      case "lazada":
        return (p.event_type as string) || "unknown";
      default:
        return "unknown";
    }
  }

  /**
   * Handle specific webhook events
   * @param tenantId - Tenant ID for database operations
   */
  private async handleEvent(
    platform: PlatformType,
    eventType: string,
    payload: unknown,
    tenantId?: string
  ): Promise<void> {
    // Use the dedicated webhook event handler to process and update database
    await handleWebhookEvent(platform, eventType, payload, tenantId);
  }

  /**
   * Get webhook logs with pagination
   */
  async getWebhookLogs(
    tenantId: string,
    options: { page: number; limit: number; status?: string }
  ): Promise<PaginatedResult<WebhookLogEntry>> {
    return getWebhookLogs(tenantId, options);
  }

  /**
   * Get webhook logs filtered by platform
   */
  async getWebhookLogsByPlatform(
    tenantId: string,
    platform: string,
    options: { page: number; limit: number }
  ): Promise<PaginatedResult<WebhookLogEntry>> {
    return getWebhookLogsByPlatform(tenantId, platform, options);
  }

  /**
   * Get webhook statistics
   */
  async getWebhookStats(tenantId: string): Promise<WebhookStats> {
    return getWebhookStats(tenantId);
  }
}
