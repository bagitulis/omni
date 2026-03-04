import crypto from "crypto";
import { getLogger } from "../../utils/logger";
import { getGlobalConfigService } from "../globalConfigService";

const logger = getLogger("WebhookVerification");

/**
 * Get Shopee Push Partner Key for webhook verification
 * Reads from GlobalConfig (system.db) with env var fallback
 * Note: pushPartnerKey is specifically for webhook signature verification
 */
async function getShopeePartnerKey(): Promise<string | null> {
  try {
    const globalConfig = getGlobalConfigService();
    const { pushPartnerKey, partnerKey } =
      await globalConfig.getShopeeCredentials();

    // Prefer pushPartnerKey for webhooks, fallback to partnerKey
    const key = pushPartnerKey || partnerKey;

    if (key) {
      logger.debug("Using Shopee partner key from GlobalConfig");
      return key;
    }
  } catch (error) {
    logger.warn(`Failed to get partner key from GlobalConfig: ${error}`);
  }

  return null;
}

/**
 * Verify Shopee webhook signature
 * Shopee uses: HMAC-SHA256(partner_key, URL + "|" + request_body)
 * See: https://open.shopee.com/developer-guide/18
 *
 * @param payload - Parsed request body (fallback)
 * @param headers - Request headers containing authorization signature
 * @param tenantId - Optional tenant ID to get tenant-specific partner key
 * @param rawBody - Raw request body string (preferred for signature verification)
 * @param webhookUrl - Full webhook URL used for signature computation
 */
export async function verifyShopeeSignature(
  payload: unknown,
  headers: Record<string, string>,
  _tenantId?: string,
  rawBody?: string,
  webhookUrl?: string
): Promise<boolean> {
  const signature = headers["authorization"];

  if (!signature) {
    logger.warn("No Shopee signature provided");
    return process.env.NODE_ENV !== "production";
  }

  const partnerKey = await getShopeePartnerKey();
  if (!partnerKey) {
    // Skip verification if partner key not configured (for initial setup)
    logger.warn(
      "Shopee partner key not configured in GlobalConfig, skipping verification"
    );
    return true;
  }

  try {
    // Use raw body if available, otherwise stringify parsed payload
    const bodyToVerify = rawBody || JSON.stringify(payload);

    // Shopee signature format: HMAC-SHA256(partner_key, URL + "|" + body)
    // URL must be the full callback URL registered with Shopee
    const baseString = webhookUrl
      ? `${webhookUrl}|${bodyToVerify}`
      : bodyToVerify;

    logger.info(
      `🔍 Verifying signature with ${
        rawBody ? "raw body" : "stringified payload"
      }`
    );
    logger.info(`📏 Body length: ${bodyToVerify.length} bytes`);
    logger.info(`🔗 Webhook URL: ${webhookUrl || "NOT PROVIDED"}`);
    logger.info(`🔑 Partner Key: ${partnerKey.substring(0, 20)}...`);

    const computedSignature = crypto
      .createHmac("sha256", partnerKey)
      .update(baseString)
      .digest("hex");

    logger.info(`📥 Expected signature: ${signature}`);
    logger.info(`🔐 Computed signature: ${computedSignature}`);

    const isValid = signature === computedSignature;

    if (!isValid) {
      logger.warn(
        `Signature mismatch! Expected: ${signature}, Got: ${computedSignature}`
      );
    } else {
      logger.info("✅ Shopee webhook signature verified successfully");
    }

    return isValid;
  } catch (error) {
    logger.error(`Signature verification error: ${error}`);
    return false;
  }
}

/**
 * Verify TikTok webhook signature
 */
export async function verifyTiktokSignature(
  payload: unknown,
  headers: Record<string, string>
): Promise<boolean> {
  const signature = headers["x-tiktok-signature"];

  if (!signature) {
    logger.warn("No TikTok signature provided");
    return process.env.NODE_ENV !== "production";
  }

  const appSecret = process.env.TIKTOK_APP_SECRET;
  if (!appSecret) {
    return false;
  }

  try {
    const payloadString = JSON.stringify(payload);
    const computedSignature = crypto
      .createHmac("sha256", appSecret)
      .update(payloadString)
      .digest("hex");

    return signature === computedSignature;
  } catch {
    return false;
  }
}

/**
 * Verify Lazada webhook signature
 */
export async function verifyLazadaSignature(
  payload: unknown,
  headers: Record<string, string>
): Promise<boolean> {
  const signature = headers["x-lazada-signature"];

  if (!signature) {
    logger.warn("No Lazada signature provided");
    return process.env.NODE_ENV !== "production";
  }

  const appSecret = process.env.LAZADA_APP_SECRET;
  if (!appSecret) {
    return false;
  }

  try {
    const payloadString = JSON.stringify(payload);
    const computedSignature = crypto
      .createHmac("sha256", appSecret)
      .update(payloadString)
      .digest("hex");

    return signature === computedSignature;
  } catch {
    return false;
  }
}
