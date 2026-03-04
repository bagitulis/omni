import { Router, Request, Response } from "express";
import { PlatformAuthController } from "../controllers/platformAuthController";
import { getLogger } from "../utils/logger";

const router = Router();
const logger = getLogger("PlatformAuthRoutes");
const authController = new PlatformAuthController();

/**
 * Platform OAuth Routes
 * Handles OAuth authorization flow for marketplace platforms
 *
 * Flow:
 * 1. Frontend calls /api/platform-auth/:platform/authorize
 * 2. Backend generates OAuth URL and redirects user
 * 3. User authorizes on platform (Shopee/TikTok/Lazada)
 * 4. Platform redirects to /api/platform-auth/:platform/callback
 * 5. Backend exchanges code for tokens
 * 6. Backend redirects to frontend with success/error
 */

// ============================================
// Get OAuth URLs (for display in settings)
// ============================================
router.get("/urls", async (_req: Request, res: Response) => {
  try {
    const baseUrl = process.env.APP_BASE_URL || "https://yndigital.my.id";

    const urls = {
      shopee: {
        oauthCallback: `${baseUrl}/api/platform-auth/shopee/callback`,
        webhook: `${baseUrl}/api/webhooks/shopee`,
      },
      tiktok: {
        oauthCallback: `${baseUrl}/api/platform-auth/tiktok/callback`,
        webhook: `${baseUrl}/api/webhooks/tiktok`,
      },
      lazada: {
        oauthCallback: `${baseUrl}/api/platform-auth/lazada/callback`,
        webhook: `${baseUrl}/api/webhooks/lazada`,
      },
    };

    res.json({
      success: true,
      data: urls,
    });
  } catch (error: any) {
    logger.error(`Failed to get URLs: ${error.message}`);
    res.status(500).json({
      success: false,
      error: error.message,
    });
  }
});

// ============================================
// Initiate OAuth Authorization
// ============================================
router.get("/:platform/authorize", async (req: Request, res: Response) => {
  await authController.initiateOAuth(req, res);
});

// ============================================
// OAuth Callback Handlers
// ============================================
router.get("/shopee/callback", async (req: Request, res: Response) => {
  await authController.handleCallback(req, res, "shopee");
});

router.get("/tiktok/callback", async (req: Request, res: Response) => {
  await authController.handleCallback(req, res, "tiktok");
});

router.get("/lazada/callback", async (req: Request, res: Response) => {
  await authController.handleCallback(req, res, "lazada");
});

// ============================================
// Connection Status
// ============================================
router.get("/status", async (req: Request, res: Response) => {
  await authController.getConnectionStatus(req, res);
});

router.get("/:platform/status", async (req: Request, res: Response) => {
  await authController.getPlatformStatus(req, res);
});

// ============================================
// Disconnect Platform
// ============================================
router.post("/:platform/disconnect", async (req: Request, res: Response) => {
  await authController.disconnectPlatform(req, res);
});

// ============================================
// OAuth Logs
// ============================================
router.get("/logs", async (req: Request, res: Response) => {
  await authController.getOAuthLogs(req, res);
});

export default router;
