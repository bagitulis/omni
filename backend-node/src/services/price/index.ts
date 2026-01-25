/**
 * Price Update Services
 * RESPONSIBILITY: Export all price update services for easier imports
 */

export {
  ShopeePriceUpdater,
  ShopeePriceUpdateRequest,
  ShopeePriceUpdateResult,
} from "./shopeePriceUpdater";

export {
  LazadaPriceUpdater,
  LazadaPriceUpdateRequest,
  LazadaPriceUpdateResult,
} from "./lazadaPriceUpdater";

export {
  TiktokPriceUpdater,
  TiktokPriceUpdateRequest,
  TiktokPriceUpdateResult,
} from "./tiktokPriceUpdater";

export {
  PriceUpdateOrchestrator,
  PriceUpdateItem,
  PlatformPriceResult,
  PriceUpdateResult,
  BatchPriceUpdateResult,
} from "./priceUpdateOrchestrator";

import { ProductIdFetcher } from "../stock/productIdFetcher";
import { ShopeePriceUpdater } from "./shopeePriceUpdater";
import { LazadaPriceUpdater } from "./lazadaPriceUpdater";
import { TiktokPriceUpdater } from "./tiktokPriceUpdater";
import { PriceUpdateOrchestrator } from "./priceUpdateOrchestrator";
import { getDbManager } from "../dbConnectionManager";
import { ShopeeAPIClient } from "../../api/clients/shopeeAPIClient";
import { LazadaAPIClient } from "../../api/clients/lazadaAPIClient";
import { TiktokAPIClient } from "../../api/clients/tiktokAPIClient";
import { ShopeeConfigManager } from "../../config/managers/shopeeConfigManager";
import { LazadaConfigManager } from "../../config/managers/lazadaConfigManager";
import { TiktokConfigManager } from "../../config/managers/tiktokConfigManager";
import { getLogger } from "../../utils/logger";

const logger = getLogger("PriceService");

/**
 * Create tenant-aware PriceUpdateOrchestrator
 * Creates all dependencies with proper tenant context
 *
 * @param tenantId - Tenant ID for loading correct credentials
 */
export async function getPriceUpdateOrchestrator(
  tenantId: string
): Promise<PriceUpdateOrchestrator> {
  logger.info(`[${tenantId}] Creating price update orchestrator`);

  // Create tenant-specific config managers
  const shopeeConfig = new ShopeeConfigManager(tenantId);
  const lazadaConfig = new LazadaConfigManager(tenantId);
  const tiktokConfig = new TiktokConfigManager(tenantId);

  // Load configs from tenant database
  await shopeeConfig.loadConfig();
  await lazadaConfig.loadConfig();
  await tiktokConfig.loadConfig();

  // Create API clients with loaded configs
  const shopeeApiClient = new ShopeeAPIClient(shopeeConfig);
  const lazadaApiClient = new LazadaAPIClient(lazadaConfig);
  const tiktokApiClient = new TiktokAPIClient(tiktokConfig);

  // Create prisma for tenant
  const prisma = getDbManager().getConnection(tenantId);
  const productIdFetcher = new ProductIdFetcher(prisma);

  // Create updaters
  const shopeeUpdater = new ShopeePriceUpdater(shopeeApiClient, shopeeConfig);
  const lazadaUpdater = new LazadaPriceUpdater(lazadaApiClient, lazadaConfig);
  const tiktokUpdater = new TiktokPriceUpdater(tiktokApiClient, tiktokConfig);

  return new PriceUpdateOrchestrator(
    productIdFetcher,
    shopeeUpdater,
    lazadaUpdater,
    tiktokUpdater
  );
}
