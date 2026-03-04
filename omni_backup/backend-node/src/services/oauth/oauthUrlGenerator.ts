import crypto from "crypto";
import { getGlobalConfigService } from "../globalConfigService";

/**
 * OAuth URL Generator
 * Generates platform-specific OAuth authorization URLs
 * Uses GlobalConfigService to fetch credentials from system.db
 */

/**
 * Generate Shopee OAuth URL
 * Uses global credentials from system.db
 */
export async function generateShopeeAuthUrl(
  state: string,
  redirectUrl: string
): Promise<string> {
  const globalConfig = getGlobalConfigService();
  const { partnerId, partnerKey } = await globalConfig.getShopeeCredentials();

  if (!partnerId || !partnerKey) {
    throw new Error("Shopee credentials not configured in system.db");
  }

  const timestamp = Math.floor(Date.now() / 1000);
  const path = "/api/v2/shop/auth_partner";
  const baseString = `${partnerId}${path}${timestamp}`;

  const sign = crypto
    .createHmac("sha256", partnerKey)
    .update(baseString)
    .digest("hex");

  const params = new URLSearchParams({
    partner_id: partnerId,
    timestamp: timestamp.toString(),
    sign,
    redirect: redirectUrl,
    state,
  });

  return `https://partner.shopeemobile.com${path}?${params.toString()}`;
}

/**
 * Generate TikTok OAuth URL
 * Uses global credentials from system.db
 */
export async function generateTiktokAuthUrl(
  state: string,
  redirectUrl: string
): Promise<string> {
  const globalConfig = getGlobalConfigService();
  const { appKey } = await globalConfig.getTiktokCredentials();

  if (!appKey) {
    throw new Error("TikTok credentials not configured in system.db");
  }

  const params = new URLSearchParams({
    app_key: appKey,
    state,
    redirect_uri: redirectUrl,
  });

  return `https://auth.tiktok-shops.com/oauth/authorize?${params.toString()}`;
}

/**
 * Generate Lazada OAuth URL
 * Uses global credentials from system.db
 */
export async function generateLazadaAuthUrl(
  state: string,
  redirectUrl: string
): Promise<string> {
  const globalConfig = getGlobalConfigService();
  const lazadaAppKey = await globalConfig.getConfig("lazada", "appKey");

  if (!lazadaAppKey) {
    throw new Error("Lazada credentials not configured in system.db");
  }

  const params = new URLSearchParams({
    response_type: "code",
    client_id: lazadaAppKey,
    state,
    redirect_uri: redirectUrl,
    force_auth: "true",
  });

  return `https://auth.lazada.com/oauth/authorize?${params.toString()}`;
}
