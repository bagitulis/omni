/**
 * Shopee Webhook Event Router
 * Routes webhook events to appropriate category tables based on push code
 */

import { getLogger } from "../../utils/logger";
import { getPushCodeInfo } from "../../constants/shopeePushCodes";
import {
  ProductWebhookRepository,
  MarketingWebhookRepository,
  ReturnWebhookRepository,
  ShopeeSystemWebhookRepository,
  WebchatWebhookRepository,
  FBSWebhookRepository,
} from "./repositories";
import { WebhookOrderEventRepository } from "./webhookOrderEventRepository";

const logger = getLogger("ShopeeWebhookRouter");

// Category to handler mapping
type CategoryHandler = (eventType: string, payload: any) => Promise<void>;

const categoryHandlers: Record<string, CategoryHandler> = {
  Order: handleOrderEvent,
  Product: handleProductEvent,
  Marketing: handleMarketingEvent,
  Return: handleReturnEvent,
  Shopee: handleShopeeSystemEvent,
  Webchat: handleWebchatEvent,
  FBS: handleFBSEvent,
};

/**
 * Route Shopee webhook event to appropriate category handler
 */
export async function routeShopeeEvent(
  code: number,
  eventType: string,
  payload: any
): Promise<void> {
  const pushInfo = getPushCodeInfo(code);
  const category = pushInfo.category;

  logger.info(`📨 Routing ${eventType} (code: ${code}) to ${category} handler`);

  const handler = categoryHandlers[category];
  if (handler) {
    await handler(eventType, payload);
  } else {
    logger.warn(`No handler for category: ${category}, event: ${eventType}`);
  }
}

/**
 * Handle Order category events
 */
async function handleOrderEvent(
  eventType: string,
  payload: any
): Promise<void> {
  const repo = new WebhookOrderEventRepository();
  const data = payload?.data || {};

  await repo.insertEvent({
    platform: "shopee",
    eventType,
    orderSn: data.ordersn || data.order_sn || "",
    shopId: payload.shop_id?.toString(),
    newStatus: data.status,
    fulfillmentStatus: data.fulfillment_status,
    packageNumber: data.package_number,
    payload,
  });

  logger.info(
    `📦 Order event saved: ${eventType} - ${data.ordersn || data.order_sn}`
  );
}

/**
 * Handle Product category events
 */
async function handleProductEvent(
  eventType: string,
  payload: any
): Promise<void> {
  const repo = new ProductWebhookRepository();
  const data = payload?.data || {};

  await repo.insertEvent({
    eventType,
    shopId: payload.shop_id?.toString(),
    itemId: data.item_id?.toString(),
    variationId: data.variation_id?.toString(),
    action: data.action,
    payload,
  });

  logger.info(`📦 Product event saved: ${eventType} - item ${data.item_id}`);
}

/**
 * Handle Marketing category events
 */
async function handleMarketingEvent(
  eventType: string,
  payload: any
): Promise<void> {
  const repo = new MarketingWebhookRepository();
  const data = payload?.data || {};

  await repo.insertEvent({
    eventType,
    shopId: payload.shop_id?.toString(),
    itemId: data.item_id?.toString(),
    promotionId: data.promotion_id?.toString(),
    promotionType: data.promotion_type,
    action: data.action,
    payload,
  });

  logger.info(
    `📢 Marketing event saved: ${eventType} - promo ${data.promotion_id}`
  );
}

/**
 * Handle Return category events
 */
async function handleReturnEvent(
  eventType: string,
  payload: any
): Promise<void> {
  const repo = new ReturnWebhookRepository();
  const data = payload?.data || {};

  await repo.insertEvent({
    eventType,
    shopId: payload.shop_id?.toString(),
    orderSn: data.ordersn || data.order_sn,
    returnSn: data.return_sn || data.returnsn,
    status: data.status,
    reason: data.reason,
    payload,
  });

  logger.info(`🔄 Return event saved: ${eventType} - return ${data.return_sn}`);
}

/**
 * Handle Shopee system category events
 */
async function handleShopeeSystemEvent(
  eventType: string,
  payload: any
): Promise<void> {
  const repo = new ShopeeSystemWebhookRepository();
  const data = payload?.data || {};

  await repo.insertEvent({
    eventType,
    shopId: payload.shop_id?.toString(),
    action: data.action,
    expiryTime: data.expire_time?.toString(),
    penaltyPoints: data.penalty_points,
    payload,
  });

  logger.info(
    `🏪 Shopee system event saved: ${eventType} - shop ${payload.shop_id}`
  );
}

/**
 * Handle Webchat category events
 */
async function handleWebchatEvent(
  eventType: string,
  payload: any
): Promise<void> {
  const repo = new WebchatWebhookRepository();
  const data = payload?.data || {};

  await repo.insertEvent({
    eventType,
    shopId: payload.shop_id?.toString(),
    conversationId: data.conversation_id?.toString(),
    messageType: data.message_type,
    senderId: data.from_id?.toString() || data.sender_id?.toString(),
    payload,
  });

  logger.info(
    `💬 Webchat event saved: ${eventType} - conv ${data.conversation_id}`
  );
}

/**
 * Handle FBS category events
 */
async function handleFBSEvent(eventType: string, payload: any): Promise<void> {
  const repo = new FBSWebhookRepository();
  const data = payload?.data || {};

  await repo.insertEvent({
    eventType,
    shopId: payload.shop_id?.toString(),
    itemId: data.item_id?.toString(),
    skuId: data.model_id?.toString() || data.sku_id?.toString(),
    stockChange: data.sellable_stock || data.stock_change,
    invoiceNumber: data.invoice_number,
    payload,
  });

  logger.info(`📦 FBS event saved: ${eventType} - item ${data.item_id}`);
}
