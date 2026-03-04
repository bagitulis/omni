import { Request, Response } from "express";
import { PlatformAuthService } from "../services/platformAuthService";
import { getPrisma } from "../services/prismaClient";
import { getLogger } from "../utils/logger";

const logger = getLogger("PlatformAuthController");

type PlatformType = "shopee" | "tiktok" | "lazada";

/**
 * Platform OAuth Controller
 * Handles OAuth authorization flow for marketplace platforms
 */
export class PlatformAuthController {
  private authService: PlatformAuthService;

  constructor() {
    this.authService = new PlatformAuthService();
  }

  /**
   * Initiate OAuth flow - return authorization URL for frontend to open
   */
  async initiateOAuth(req: Request, res: Response): Promise<void> {
    try {
      const platform = req.params.platform as PlatformType;

      const tenantId = req.headers["x-tenant-id"] as string;
      if (!tenantId) {
        res
          .status(400)
          .json({ success: false, error: "Missing x-tenant-id header" });
        return;
      }

      const redirectUrl = (req.query.redirect as string) || "/settings/webhook";

      if (!["shopee", "tiktok", "lazada"].includes(platform)) {
        res.status(400).json({
          success: false,
          error: `Invalid platform: ${platform}`,
        });
        return;
      }

      logger.info(`Initiating OAuth for ${platform}, tenant: ${tenantId}`);

      const authUrl = await this.authService.generateAuthUrl(
        platform,
        tenantId,
        redirectUrl
      );

      // Return JSON with authUrl for frontend to handle
      res.json({
        success: true,
        data: {
          authUrl,
          platform,
        },
      });
    } catch (error: any) {
      logger.error(`OAuth initiation failed: ${error.message}`);
      res.status(500).json({
        success: false,
        error: error.message,
      });
    }
  }

  /**
   * Handle OAuth callback from platform
   * Supports direct callback from Shopee/TikTok/Lazada (no state required)
   */
  async handleCallback(
    req: Request,
    res: Response,
    platform: PlatformType
  ): Promise<void> {
    try {
      const { code, state, shop_id, error, error_description } = req.query;

      logger.info(
        `OAuth callback for ${platform}, code: ${
          code ? "present" : "missing"
        }, shop_id: ${shop_id || "none"}`
      );

      // Handle error from platform
      if (error) {
        logger.error(
          `OAuth error from ${platform}: ${error_description || error}`
        );
        res.redirect(
          `/settings/webhook?error=${encodeURIComponent(
            String(error_description || error)
          )}&platform=${platform}`
        );
        return;
      }

      if (!code) {
        res.redirect(
          `/settings/webhook?error=missing_code&platform=${platform}`
        );
        return;
      }

      // Exchange code for tokens (state is optional for direct callbacks)
      const result = await this.authService.exchangeCodeForTokens(
        platform,
        String(code),
        state ? String(state) : null,
        shop_id ? String(shop_id) : undefined
      );

      if (result.success) {
        logger.info(`OAuth success for ${platform}`);
        // Redirect to homepage with success message
        res.redirect(`/?auth_success=true&platform=${platform}`);
      } else {
        logger.error(`OAuth token exchange failed: ${result.error}`);
        res.redirect(
          `/?auth_error=${encodeURIComponent(
            result.error || "unknown"
          )}&platform=${platform}`
        );
      }
    } catch (error: any) {
      logger.error(`OAuth callback error: ${error.message}`);
      res.redirect(
        `/?auth_error=${encodeURIComponent(error.message)}&platform=${platform}`
      );
    }
  }

  /**
   * Get connection status for all platforms
   */
  async getConnectionStatus(req: Request, res: Response): Promise<void> {
    try {
      const tenantId = req.headers["x-tenant-id"] as string;
      if (!tenantId) {
        res
          .status(400)
          .json({ success: false, error: "Missing x-tenant-id header" });
        return;
      }

      const statuses = await this.authService.getAllConnectionStatus(tenantId);

      res.json({
        success: true,
        data: statuses,
      });
    } catch (error: any) {
      logger.error(`Failed to get connection status: ${error.message}`);
      res.status(500).json({
        success: false,
        error: error.message,
      });
    }
  }

  /**
   * Get status for specific platform
   */
  async getPlatformStatus(req: Request, res: Response): Promise<void> {
    try {
      const platform = req.params.platform as PlatformType;

      const tenantId = req.headers["x-tenant-id"] as string;
      if (!tenantId) {
        res
          .status(400)
          .json({ success: false, error: "Missing x-tenant-id header" });
        return;
      }

      if (!["shopee", "tiktok", "lazada"].includes(platform)) {
        res.status(400).json({
          success: false,
          error: `Invalid platform: ${platform}`,
        });
        return;
      }

      const status = await this.authService.getPlatformStatus(
        platform,
        tenantId
      );

      res.json({
        success: true,
        data: status,
      });
    } catch (error: any) {
      logger.error(`Failed to get platform status: ${error.message}`);
      res.status(500).json({
        success: false,
        error: error.message,
      });
    }
  }

  /**
   * Disconnect platform (revoke tokens)
   */
  async disconnectPlatform(req: Request, res: Response): Promise<void> {
    try {
      const platform = req.params.platform as PlatformType;

      const tenantId = req.headers["x-tenant-id"] as string;
      if (!tenantId) {
        res
          .status(400)
          .json({ success: false, error: "Missing x-tenant-id header" });
        return;
      }

      if (!["shopee", "tiktok", "lazada"].includes(platform)) {
        res.status(400).json({
          success: false,
          error: `Invalid platform: ${platform}`,
        });
        return;
      }

      await this.authService.disconnectPlatform(platform, tenantId);

      res.json({
        success: true,
        message: `${platform} disconnected successfully`,
      });
    } catch (error: any) {
      logger.error(`Failed to disconnect platform: ${error.message}`);
      res.status(500).json({
        success: false,
        error: error.message,
      });
    }
  }

  /**
   * Get OAuth logs
   */
  async getOAuthLogs(req: Request, res: Response): Promise<void> {
    try {
      const tenantId = req.headers["x-tenant-id"] as string;
      if (!tenantId) {
        res
          .status(400)
          .json({ success: false, error: "Missing x-tenant-id header" });
        return;
      }

      const limit = parseInt(req.query.limit as string) || 20;
      const prisma = getPrisma(tenantId);

      const logs = await prisma.oAuthLog.findMany({
        where: { tenantId },
        orderBy: { createdAt: "desc" },
        take: limit,
      });

      res.json({
        success: true,
        data: logs.map((log) => ({
          id: log.id,
          platform: log.platform,
          eventType: log.eventType,
          status: log.status,
          createdAt: log.createdAt,
          processedAt: log.processedAt,
          metadata: log.metadata ? JSON.parse(log.metadata) : null,
        })),
      });
    } catch (error: any) {
      logger.error(`Failed to get OAuth logs: ${error.message}`);
      res.status(500).json({
        success: false,
        error: error.message,
      });
    }
  }
}
