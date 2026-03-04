/**
 * Platform Services Factory
 * Single Responsibility: Centralized creation of config managers and API clients
 * Eliminates code duplication across route files
 */

import { ShopeeConfigManager } from "../config/managers/shopeeConfigManager";
import { LazadaConfigManager } from "../config/managers/lazadaConfigManager";
import { TiktokConfigManager } from "../config/managers/tiktokConfigManager";
import { ShopeeAPIClient } from "../api/clients/shopeeAPIClient";
import { LazadaAPIClient } from "../api/clients/lazadaAPIClient";
import { TiktokAPIClient } from "../api/clients/tiktokAPIClient";
import { getLogger } from "../utils/logger";

const logger = getLogger("PlatformFactory");

/**
 * Platform configuration managers
 */
export interface ConfigManagers {
  shopee: ShopeeConfigManager;
  lazada: LazadaConfigManager;
  tiktok: TiktokConfigManager;
}

/**
 * Platform API clients
 */
export interface ApiClients {
  shopee: ShopeeAPIClient;
  lazada: LazadaAPIClient;
  tiktok: TiktokAPIClient;
}

/**
 * Combined platform services
 */
export interface PlatformServices {
  configManagers: ConfigManagers;
  apiClients: ApiClients;
}

// Singleton instances
let configManagers: ConfigManagers | null = null;
let apiClients: ApiClients | null = null;
let initialized = false;

/**
 * Initialize all config managers (singleton)
 */
function initConfigManagers(): ConfigManagers {
  if (configManagers) {
    return configManagers;
  }

  configManagers = {
    shopee: new ShopeeConfigManager(),
    lazada: new LazadaConfigManager(),
    tiktok: new TiktokConfigManager(),
  };

  return configManagers;
}

/**
 * Initialize all API clients (singleton)
 * Requires config managers to be initialized first
 */
function initApiClients(managers: ConfigManagers): ApiClients {
  if (apiClients) {
    return apiClients;
  }

  apiClients = {
    shopee: new ShopeeAPIClient(managers.shopee),
    lazada: new LazadaAPIClient(managers.lazada),
    tiktok: new TiktokAPIClient(managers.tiktok),
  };

  return apiClients;
}

/**
 * Load all platform configurations
 * Should be called once at application startup
 */
export async function loadAllPlatformConfigs(): Promise<void> {
  if (initialized) {
    return;
  }

  const managers = initConfigManagers();

  try {
    await Promise.all([
      managers.shopee.loadConfig(),
      managers.lazada.loadConfig(),
      managers.tiktok.loadConfig(),
    ]);
    initialized = true;
    logger.info("✅ All platform configurations loaded");
  } catch (error) {
    logger.error(`❌ Error loading platform configs: ${error}`);
    throw error;
  }
}

/**
 * Get config managers singleton
 */
export function getConfigManagers(): ConfigManagers {
  if (!configManagers) {
    configManagers = initConfigManagers();
  }
  return configManagers;
}

/**
 * Get API clients singleton
 */
export function getApiClients(): ApiClients {
  const managers = getConfigManagers();
  if (!apiClients) {
    apiClients = initApiClients(managers);
  }
  return apiClients;
}

/**
 * Get all platform services (config managers + API clients)
 */
export function getPlatformServices(): PlatformServices {
  return {
    configManagers: getConfigManagers(),
    apiClients: getApiClients(),
  };
}

/**
 * Check if platforms are initialized
 */
export function isPlatformInitialized(): boolean {
  return initialized;
}

/**
 * Reset factory (for testing purposes)
 */
export function resetPlatformFactory(): void {
  configManagers = null;
  apiClients = null;
  initialized = false;
}
