/**
 * Webhook Event Handler
 * Routes webhook events to category-specific tables
 * Shopee uses category routing, TikTok/Lazada use generic order event table
 */

import { getLogger } from "../../utils/logger";
import {
  WebhookOrderEventRepository,
  WebhookOrderEventData,
} from "./webhookOrderEventRepository";
import { routeShopeeEvent } from "./shopeeWebhookRouter";
import { getPushCodeInfo } from "../../constants/shopeePushCodes";

const logger = getLogger("WebhookEventHandler");

type PlatformType = "shopee" | "tiktok" | "lazada";

/**
 * Extract event data from TikTok webhook payload
 */
function extractTiktokEventData(
  eventType: string,
  payload: any
): WebhookOrderEventData {
  const data = payload?.data || {};
  const orderSn = data.order_id || "";
  const shopId = payload.shop_id?.toString() || "";

  return {
    platform: "tiktok",
    eventType,
    orderSn,
    shopId,
    newStatus: data.order_status,
    payload,
  };
}

/**
 * Extract event data from Lazada webhook payload
 */
function extractLazadaEventData(
  eventType: string,
  payload: any
): WebhookOrderEventData {
  const data = payload?.data || payload || {};
  const orderSn = data.order_id?.toString() || "";

  return {
    platform: "lazada",
    eventType,
    orderSn,
    newStatus: data.status,
    payload,
  };
}

/**
 * Handle Shopee webhook - routes to category-specific tables
 */
async function handleShopeeEvent(
  eventType: string,
  payload: any
): Promise<void> {
  const code = payload?.code as number;

  if (code === undefined) {
    logger.warn("Shopee webhook missing code field");
    return;
  }

  const pushInfo = getPushCodeInfo(code);
  logger.info(`📨 Shopee ${pushInfo.category}/${eventType} (code: ${code})`);

  await routeShopeeEvent(code, eventType, payload);
}

/**
 * Handle TikTok webhook event - store to WebhookOrderEvent
 */
async function handleTiktokEvent(
  eventType: string,
  payload: any
): Promise<void> {
  const repo = new WebhookOrderEventRepository();
  const eventData = extractTiktokEventData(eventType, payload);

  if (!eventData.orderSn) {
    logger.debug(`No order_id in TikTok event: ${eventType}`);
    return;
  }

  await repo.insertEvent(eventData);
  logger.info(
    `📋 TikTok order ${eventData.orderSn} status: ${eventData.newStatus}`
  );
}

/**
 * Handle Lazada webhook event - store to WebhookOrderEvent
 */
async function handleLazadaEvent(
  eventType: string,
  payload: any
): Promise<void> {
  const repo = new WebhookOrderEventRepository();
  const eventData = extractLazadaEventData(eventType, payload);

  if (!eventData.orderSn) {
    logger.debug(`No order_id in Lazada event: ${eventType}`);
    return;
  }

  await repo.insertEvent(eventData);
  logger.info(
    `📋 Lazada order ${eventData.orderSn} status: ${eventData.newStatus}`
  );
}

/**
 * Get Shopee event type from webhook code
 */
export function getShopeeEventType(code: number): string {
  const info = getPushCodeInfo(code);
  return info.name;
}

/**
 * Main handler for webhook events
 * Routes Shopee to category tables, TikTok/Lazada to order event table
 */
export async function handleWebhookEvent(
  platform: PlatformType,
  eventType: string,
  payload: unknown,
  tenantId?: string
): Promise<void> {
  logger.info(
    `🔔 Processing ${platform} event: ${eventType}${
      tenantId ? ` for tenant: ${tenantId}` : ""
    }`
  );

  try {
    switch (platform) {
      case "shopee":
        await handleShopeeEvent(eventType, payload);
        break;
      case "tiktok":
        await handleTiktokEvent(eventType, payload);
        break;
      case "lazada":
        await handleLazadaEvent(eventType, payload);
        break;
      default:
        logger.warn(`Unknown platform: ${platform}`);
    }
  } catch (error: any) {
    logger.error(`Failed to process ${platform} event: ${error.message}`);
  }
}

export default { handleWebhookEvent, getShopeeEventType };
