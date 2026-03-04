import { Router, Request, Response } from "express";
import { WebhookController } from "../controllers/webhookController";
import { getLogger } from "../utils/logger";
import { getDbManager } from "../services/dbConnectionManager";

const router = Router();
const webhookController = new WebhookController();
const logger = getLogger("WebhookRoutes");

/**
 * Helper to get first REAL tenant for legacy webhooks
 * ⚠️ EXCLUDES: 'system' (global config only, not a real tenant per AGENTS.MD)
 */
function getFirstRealTenant(): string | undefined {
  const dbManager = getDbManager();
  const tenants = dbManager.getRealTenants();
  return tenants.length > 0 ? tenants[0] : undefined;
}

// Log all webhook requests
router.use((req, _res, next) => {
  logger.info(`📥 Webhook request: ${req.method} ${req.path} from ${req.ip}`);
  next();
});

/**
 * Webhook Routes
 * Handles incoming webhooks from marketplace platforms
 *
 * These endpoints receive real-time notifications about:
 * - Order created/updated/cancelled
 * - Payment status changes
 * - Shipping updates
 * - Return/Refund requests
 *
 * URL Format: /api/webhooks/{tenantId}/{platform}
 * Example: /api/webhooks/bertigamart/shopee
 */

// ============================================
// Tenant-specific Shopee Webhook (Recommended)
// ============================================
router.post("/:tenantId/shopee", async (req: Request, res: Response) => {
  const { tenantId } = req.params;
  logger.info(`🛒 Shopee webhook for tenant: ${tenantId}`);
  await webhookController.handleShopeeWebhook(req, res, tenantId);
});

// ============================================
// Tenant-specific TikTok Webhook
// ============================================
router.post("/:tenantId/tiktok", async (req: Request, res: Response) => {
  const { tenantId } = req.params;
  logger.info(`🎵 TikTok webhook for tenant: ${tenantId}`);
  await webhookController.handleTiktokWebhook(req, res, tenantId);
});

// ============================================
// Tenant-specific Lazada Webhook
// ============================================
router.post("/:tenantId/lazada", async (req: Request, res: Response) => {
  const { tenantId } = req.params;
  logger.info(`🛍️ Lazada webhook for tenant: ${tenantId}`);
  await webhookController.handleLazadaWebhook(req, res, tenantId);
});

// ============================================
// Legacy routes (without tenant - uses first tenant)
// Keep for backward compatibility until Shopee URL is updated
// ============================================
router.post("/shopee", async (req: Request, res: Response) => {
  const firstTenant = getFirstRealTenant();
  if (!firstTenant) {
    logger.error(`❌ Legacy Shopee webhook failed: No real tenant available`);
    res.status(400).json({
      error:
        "No tenant configured. Update webhook URL to: /api/webhooks/{tenantId}/shopee",
    });
    return;
  }
  logger.warn(
    `⚠️ Legacy Shopee webhook (no tenant in URL). ` +
      `Using tenant: ${firstTenant}. ` +
      `Please update webhook URL to: /api/webhooks/{tenantId}/shopee`,
  );
  await webhookController.handleShopeeWebhook(req, res, firstTenant);
});

router.post("/tiktok", async (req: Request, res: Response) => {
  const firstTenant = getFirstRealTenant();
  if (!firstTenant) {
    logger.error(`❌ Legacy TikTok webhook failed: No real tenant available`);
    res.status(400).json({
      error:
        "No tenant configured. Update webhook URL to: /api/webhooks/{tenantId}/tiktok",
    });
    return;
  }
  logger.warn(`⚠️ Legacy TikTok webhook. Using tenant: ${firstTenant}`);
  await webhookController.handleTiktokWebhook(req, res, firstTenant);
});

router.post("/lazada", async (req: Request, res: Response) => {
  const firstTenant = getFirstRealTenant();
  if (!firstTenant) {
    logger.error(`❌ Legacy Lazada webhook failed: No real tenant available`);
    res.status(400).json({
      error:
        "No tenant configured. Update webhook URL to: /api/webhooks/{tenantId}/lazada",
    });
    return;
  }
  logger.warn(`⚠️ Legacy Lazada webhook. Using tenant: ${firstTenant}`);
  await webhookController.handleLazadaWebhook(req, res, firstTenant);
});

// ============================================
// Webhook Logs (for UI display)
// ============================================
router.get("/logs", async (req: Request, res: Response) => {
  await webhookController.getWebhookLogs(req, res);
});

router.get("/logs/:platform", async (req: Request, res: Response) => {
  await webhookController.getWebhookLogsByPlatform(req, res);
});

// ============================================
// Webhook Statistics
// ============================================
router.get("/stats", async (req: Request, res: Response) => {
  await webhookController.getWebhookStats(req, res);
});

// ============================================
// Test Webhook (for debugging)
// ============================================
router.post("/test/:platform", async (req: Request, res: Response) => {
  await webhookController.testWebhook(req, res);
});

export default router;
