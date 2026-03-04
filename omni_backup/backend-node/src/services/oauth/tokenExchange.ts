import crypto from "crypto";
import { getLogger } from "../../utils/logger";
import { getGlobalConfigService } from "../globalConfigService";

const logger = getLogger("TokenExchange");

// API Response Types
interface ShopeeTokenResponse {
  error?: string;
  message?: string;
  access_token?: string;
  refresh_token?: string;
  expire_in?: number;
}

interface TiktokTokenResponse {
  code?: number;
  message?: string;
  data?: {
    access_token: string;
    refresh_token: string;
    access_token_expire_in: number;
  };
}

interface LazadaTokenResponse {
  code?: string;
  message?: string;
  access_token?: string;
  refresh_token?: string;
  expires_in?: number;
}

export interface TokenExchangeResult {
  success: boolean;
  error?: string;
  accessToken?: string;
  refreshToken?: string;
  expiresAt?: Date;
}

/**
 * Exchange Shopee authorization code for tokens
 * Reads credentials from GlobalConfig (system.db) with env var fallback
 */
export async function exchangeShopeeCode(
  code: string,
  shopId: string
): Promise<TokenExchangeResult> {
  const globalConfig = getGlobalConfigService();
  const { partnerId, partnerKey } = await globalConfig.getShopeeCredentials();

  if (!partnerId || !partnerKey) {
    return { success: false, error: "Shopee credentials not configured" };
  }

  const timestamp = Math.floor(Date.now() / 1000);
  const path = "/api/v2/auth/token/get";
  const baseString = `${partnerId}${path}${timestamp}`;

  const sign = crypto
    .createHmac("sha256", partnerKey)
    .update(baseString)
    .digest("hex");

  const response = await fetch(
    `https://partner.shopeemobile.com${path}?partner_id=${partnerId}&timestamp=${timestamp}&sign=${sign}`,
    {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({
        code,
        shop_id: parseInt(shopId),
        partner_id: parseInt(partnerId),
      }),
    }
  );

  const data = (await response.json()) as ShopeeTokenResponse;

  if (data.error) {
    return { success: false, error: data.message || data.error };
  }

  logger.info("Shopee token exchange successful");

  return {
    success: true,
    accessToken: data.access_token,
    refreshToken: data.refresh_token,
    expiresAt: new Date(Date.now() + (data.expire_in || 0) * 1000),
  };
}

/**
 * Exchange TikTok authorization code for tokens
 * Reads credentials from GlobalConfig (system.db) with env var fallback
 */
export async function exchangeTiktokCode(
  code: string
): Promise<TokenExchangeResult> {
  const globalConfig = getGlobalConfigService();
  const { appKey, appSecret } = await globalConfig.getTiktokCredentials();

  if (!appKey || !appSecret) {
    return { success: false, error: "TikTok credentials not configured" };
  }

  const response = await fetch(
    "https://auth.tiktok-shops.com/api/v2/token/get",
    {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({
        app_key: appKey,
        app_secret: appSecret,
        auth_code: code,
        grant_type: "authorized_code",
      }),
    }
  );

  const data = (await response.json()) as TiktokTokenResponse;

  if (data.code !== 0) {
    return { success: false, error: data.message || "Token exchange failed" };
  }

  logger.info("TikTok token exchange successful");

  return {
    success: true,
    accessToken: data.data?.access_token,
    refreshToken: data.data?.refresh_token,
    expiresAt: new Date(
      Date.now() + (data.data?.access_token_expire_in || 0) * 1000
    ),
  };
}

/**
 * Exchange Lazada authorization code for tokens
 * Reads credentials from GlobalConfig (system.db) with env var fallback
 */
export async function exchangeLazadaCode(
  code: string
): Promise<TokenExchangeResult> {
  const globalConfig = getGlobalConfigService();
  const { appKey, appSecret } = await globalConfig.getLazadaCredentials();

  if (!appKey || !appSecret) {
    return { success: false, error: "Lazada credentials not configured" };
  }

  const timestamp = Date.now();
  const params = new URLSearchParams({
    app_key: appKey,
    code,
    sign_method: "sha256",
    timestamp: timestamp.toString(),
  });

  // Generate signature
  const signStr = `${appSecret}app_key${appKey}code${code}sign_methodsha256timestamp${timestamp}${appSecret}`;
  const sign = crypto
    .createHash("sha256")
    .update(signStr)
    .digest("hex")
    .toUpperCase();
  params.append("sign", sign);

  const response = await fetch(
    `https://auth.lazada.com/rest/auth/token/create?${params.toString()}`,
    { method: "GET" }
  );

  const data = (await response.json()) as LazadaTokenResponse;

  if (data.code !== "0") {
    return { success: false, error: data.message || "Token exchange failed" };
  }

  logger.info("Lazada token exchange successful");

  return {
    success: true,
    accessToken: data.access_token,
    refreshToken: data.refresh_token,
    expiresAt: new Date(Date.now() + (data.expires_in || 0) * 1000),
  };
}
