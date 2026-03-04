import { Request, Response } from "express";
import { getGlobalConfigService } from "../services/globalConfigService";
import { getPrisma } from "../services/prismaClient";
import { getLogger } from "../utils/logger";

const logger = getLogger("ShopSetupController");

/**
 * Shop Setup Controller
 * Handles global credentials and shop connection management
 */
export class ShopSetupController {
  /**
   * Get global credentials status (masked)
   * Only shows if credentials are configured, not the actual values
   */
  async getCredentialsStatus(_req: Request, res: Response): Promise<void> {
    try {
      const globalConfig = getGlobalConfigService();
      const shopee = await globalConfig.getShopeeCredentials();
      const tiktok = await globalConfig.getTiktokCredentials();

      res.json({
        success: true,
        data: {
          shopee: {
            hasPartnerId: Boolean(shopee.partnerId),
            hasPartnerKey: Boolean(shopee.partnerKey),
            hasPushPartnerKey: Boolean(shopee.pushPartnerKey),
            partnerId: shopee.partnerId || null,
          },
          tiktok: {
            hasAppKey: Boolean(tiktok.appKey),
            hasAppSecret: Boolean(tiktok.appSecret),
            appKey: tiktok.appKey || null,
          },
        },
      });
    } catch (error: unknown) {
      const message = error instanceof Error ? error.message : "Unknown error";
      logger.error(`Failed to get credentials status: ${message}`);
      res.status(500).json({ success: false, error: message });
    }
  }

  /**
   * Save global credentials (Owner only)
   */
  async saveCredentials(req: Request, res: Response): Promise<void> {
    try {
      const {
        platform,
        partnerId,
        partnerKey,
        pushPartnerKey,
        appKey,
        appSecret,
      } = req.body;
      const globalConfig = getGlobalConfigService();

      if (platform === "shopee") {
        if (partnerId)
          await globalConfig.setConfig("shopee", "partnerId", partnerId, false);
        if (partnerKey)
          await globalConfig.setConfig(
            "shopee",
            "partnerKey",
            partnerKey,
            true
          );
        if (pushPartnerKey !== undefined) {
          await globalConfig.setConfig(
            "shopee",
            "pushPartnerKey",
            pushPartnerKey,
            true
          );
        }
        logger.info("Shopee credentials updated");
      } else if (platform === "tiktok") {
        if (appKey)
          await globalConfig.setConfig("tiktok", "appKey", appKey, false);
        if (appSecret)
          await globalConfig.setConfig("tiktok", "appSecret", appSecret, true);
        logger.info("TikTok credentials updated");
      } else {
        res.status(400).json({ success: false, error: "Invalid platform" });
        return;
      }

      res.json({ success: true, message: `${platform} credentials saved` });
    } catch (error: unknown) {
      const message = error instanceof Error ? error.message : "Unknown error";
      logger.error(`Failed to save credentials: ${message}`);
      res.status(500).json({ success: false, error: message });
    }
  }

  /**
   * Get shop connection status for current tenant
   */
  async getShopStatus(req: Request, res: Response): Promise<void> {
    try {
      const tenantId = req.headers["x-tenant-id"] as string;
      if (!tenantId) {
        res.status(400).json({ success: false, error: "Tenant ID required" });
        return;
      }

      const prisma = getPrisma(tenantId);
      const shopeeConfig = await prisma.platformConfig.findMany({
        where: { platform: "shopee" },
      });

      const configMap = new Map(
        shopeeConfig.map((c) => [c.configKey, c.configValue])
      );
      const shopId = configMap.get("shopId") || "";
      const accessToken = configMap.get("accessToken") || "";
      const tokenExpiry = configMap.get("tokenExpiry") || "";

      const isConnected = Boolean(shopId && accessToken);
      const expiryDate = tokenExpiry ? new Date(tokenExpiry) : null;
      const isExpired = expiryDate ? expiryDate < new Date() : true;

      res.json({
        success: true,
        data: {
          shopee: {
            connected: isConnected,
            shopId: shopId || null,
            tokenValid: isConnected && !isExpired,
            expiresAt: expiryDate?.toISOString() || null,
          },
        },
      });
    } catch (error: unknown) {
      const message = error instanceof Error ? error.message : "Unknown error";
      logger.error(`Failed to get shop status: ${message}`);
      res.status(500).json({ success: false, error: message });
    }
  }

  /**
   * Disconnect shop from platform
   */
  async disconnectShop(req: Request, res: Response): Promise<void> {
    try {
      const tenantId = req.headers["x-tenant-id"] as string;
      const { platform } = req.params;

      if (!tenantId) {
        res.status(400).json({ success: false, error: "Tenant ID required" });
        return;
      }

      const prisma = getPrisma(tenantId);

      // Remove shop-specific config (keep global credentials)
      await prisma.platformConfig.deleteMany({
        where: {
          platform,
          configKey: {
            in: ["shopId", "accessToken", "refreshToken", "tokenExpiry"],
          },
        },
      });

      logger.info(`Disconnected ${platform} for tenant ${tenantId}`);
      res.json({ success: true, message: `${platform} disconnected` });
    } catch (error: unknown) {
      const message = error instanceof Error ? error.message : "Unknown error";
      logger.error(`Failed to disconnect shop: ${message}`);
      res.status(500).json({ success: false, error: message });
    }
  }
}
