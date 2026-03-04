/**
 * Service Initialization Module
 * Handles startup initialization of all background services
 */
import { backendLogger } from "../utils/backendLogger";
import { initializeLogsService } from "../services/logsService";
import { initializeMonitoring } from "../systemIntegration";
import { getTenantJobDatabase } from "../utils/jobDb";
import { getJobManager } from "../services/JobManager";
import { getJobExecutor } from "../services/JobExecutor";
import { getAutoFunctionScheduler } from "../services/AutoFunctionScheduler";
import { initializeJobHandlers } from "../services/jobHandlers";
import { tenantContext } from "../utils/tenantContext";

// NOTE: These are now lazy-loaded per tenant request, not at startup
// import { OrderSyncService } from "../services/orderSyncService";
// import { TokenManager } from "../services/tokenManager";
// import { loadAllPlatformConfigs } from "../services/platformFactory";

let servicesInitialized = false;

/**
 * Initialize all backend services
 * Called once during server startup
 */
export async function initializeServices(): Promise<void> {
  if (servicesInitialized) {
    backendLogger.info("Services", "Services already initialized, skipping...");
    return;
  }

  try {
    backendLogger.info("Services", "🚀 Initializing backend services...");

    // 1. Initialize logs service
    initializeLogsService();
    backendLogger.info("Services", "✅ Logs service initialized");

    // 2. Initialize monitoring systems (queue, metrics, alerts)
    initializeMonitoring();
    backendLogger.info("Services", "✅ Monitoring systems initialized");

    // 3. Initialize job databases for ALL valid tenants (excluding system/global)
    const validTenants = tenantContext.getTenantsWithJobDatabases();
    for (const tenant of validTenants) {
      getTenantJobDatabase(tenant);
    }
    backendLogger.info(
      "Services",
      `✅ Job databases initialized for ${
        validTenants.length
      } tenant(s): ${validTenants.join(", ")}`,
    );

    // 4. Initialize job system components
    getJobManager(); // Initialize singleton
    const jobExecutor = getJobExecutor();
    const scheduler = getAutoFunctionScheduler();

    // Register job handlers
    initializeJobHandlers();
    backendLogger.info("Services", "✅ Job handlers registered");

    // Start job executor
    jobExecutor.start();
    backendLogger.info("Services", "✅ Job executor started");

    // Start scheduler
    scheduler.start();
    backendLogger.info("Services", "✅ Auto-function scheduler started");

    // 5. Platform config, OrderSync, and TokenManager are now LAZY-LOADED
    // They require tenant context which is only available during HTTP requests
    // Each service will initialize when first accessed with proper tenant context
    backendLogger.info(
      "Services",
      "ℹ️ Platform configs will be lazy-loaded per tenant request",
    );

    servicesInitialized = true;
    backendLogger.info("Services", "🎉 All services initialized successfully");
  } catch (error) {
    backendLogger.error("Services", "❌ Failed to initialize services:", error);
    throw error;
  }
}

/**
 * Gracefully shutdown all services
 * Called during server shutdown
 */
export async function shutdownServices(): Promise<void> {
  try {
    backendLogger.info("Services", "🛑 Shutting down services...");

    // Stop job executor
    const jobExecutor = getJobExecutor();
    jobExecutor.stop();
    backendLogger.info("Services", "✅ Job executor stopped");

    // Stop scheduler
    const scheduler = getAutoFunctionScheduler();
    scheduler.stop();
    backendLogger.info("Services", "✅ Scheduler stopped");

    servicesInitialized = false;
    backendLogger.info("Services", "👋 Services shutdown complete");
  } catch (error) {
    backendLogger.error("Services", "❌ Error during shutdown:", error);
    throw error;
  }
}

/**
 * Check if services are initialized
 */
export function areServicesInitialized(): boolean {
  return servicesInitialized;
}
