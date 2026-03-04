// Stock Update Services - Export all components
export { ProductIdFetcher, type PlatformProductIds } from "./productIdFetcher";
export {
  ShopeeStockUpdater,
  type ShopeeStockUpdateRequest,
  type ShopeeStockUpdateResult,
} from "./shopeeStockUpdater";
export {
  LazadaStockUpdater,
  type LazadaStockUpdateRequest,
  type LazadaStockUpdateResult,
} from "./lazadaStockUpdater";
export {
  TiktokStockUpdater,
  type TiktokStockUpdateRequest,
  type TiktokStockUpdateResult,
} from "./tiktokStockUpdater";
export {
  StockUpdateOrchestrator,
  type StockUpdateItem,
  type StockUpdateResult,
  type BatchStockUpdateResult,
  type PlatformResult,
} from "./stockUpdateOrchestrator";

import { ProductIdFetcher } from "./productIdFetcher";
import { ShopeeStockUpdater } from "./shopeeStockUpdater";
import { LazadaStockUpdater } from "./lazadaStockUpdater";
import { TiktokStockUpdater } from "./tiktokStockUpdater";
import { StockUpdateOrchestrator } from "./stockUpdateOrchestrator";
import { getDbManager } from "../dbConnectionManager";
import { ShopeeAPIClient } from "../../api/clients/shopeeAPIClient";
import { LazadaAPIClient } from "../../api/clients/lazadaAPIClient";
import { TiktokAPIClient } from "../../api/clients/tiktokAPIClient";
import { ShopeeConfigManager } from "../../config/managers/shopeeConfigManager";
import { LazadaConfigManager } from "../../config/managers/lazadaConfigManager";
import { TiktokConfigManager } from "../../config/managers/tiktokConfigManager";
import { getLogger } from "../../utils/logger";

const logger = getLogger("StockService");

/**
 * Create tenant-aware StockUpdateOrchestrator
 * Creates all dependencies with proper tenant context
 *
 * @param tenantId - Tenant ID for loading correct credentials
 */
export async function getStockUpdateOrchestrator(
  tenantId: string
): Promise<StockUpdateOrchestrator> {
  logger.info(`[${tenantId}] Creating stock update orchestrator`);

  // Create tenant-specific config managers
  const shopeeConfigManager = new ShopeeConfigManager(tenantId);
  const lazadaConfigManager = new LazadaConfigManager(tenantId);
  const tiktokConfigManager = new TiktokConfigManager(tenantId);

  // Load configs from tenant database
  await shopeeConfigManager.loadConfig();
  await lazadaConfigManager.loadConfig();
  await tiktokConfigManager.loadConfig();

  // Create API clients with loaded configs
  const shopeeAPIClient = new ShopeeAPIClient(shopeeConfigManager);
  const lazadaAPIClient = new LazadaAPIClient(lazadaConfigManager);
  const tiktokAPIClient = new TiktokAPIClient(tiktokConfigManager);

  // Create prisma for tenant
  const prisma = getDbManager().getConnection(tenantId);
  const productIdFetcher = new ProductIdFetcher(prisma);

  // Create updaters
  const shopeeUpdater = new ShopeeStockUpdater(
    shopeeAPIClient,
    shopeeConfigManager
  );
  const lazadaUpdater = new LazadaStockUpdater(
    lazadaAPIClient,
    lazadaConfigManager
  );
  const tiktokUpdater = new TiktokStockUpdater(
    tiktokAPIClient,
    tiktokConfigManager
  );

  return new StockUpdateOrchestrator(
    productIdFetcher,
    shopeeUpdater,
    lazadaUpdater,
    tiktokUpdater
  );
}
