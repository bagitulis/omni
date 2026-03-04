import { getTenantJobDatabase, withRetry } from "../utils/jobDb";
import { tenantContext } from "../utils/tenantContext";
import { getJobManager } from "./JobManager";
import { getTenantManagementService } from "./tenantManagementService";
import { getLogger } from "../utils/logger";
import { IAutoFunctionConfig } from "../types/job.types";
import { AutoFunctionConfigManager } from "./autoFunctionConfigManager";

const logger = getLogger("AutoFunctionScheduler");

export type AutoFunctionType = string;
export type AutoFunctionTrigger = () => Promise<void>;

export class AutoFunctionScheduler {
  private static instance: AutoFunctionScheduler;
  private isRunning: boolean = false;
  private triggers: Map<AutoFunctionType, AutoFunctionTrigger> = new Map();
  private readonly CHECK_INTERVAL_MS: number = 60000;
  private scheduleCheckTimeoutId: NodeJS.Timeout | null = null;
  private configManager: AutoFunctionConfigManager;

  // Circuit breaker for database errors
  private consecutiveErrors: number = 0;
  private lastErrorLog: number = 0;
  private readonly ERROR_LOG_INTERVAL_MS: number = 60000; // Log errors max every 60s

  private constructor() {
    this.configManager = new AutoFunctionConfigManager();
  }

  public static getInstance(): AutoFunctionScheduler {
    if (!AutoFunctionScheduler.instance) {
      AutoFunctionScheduler.instance = new AutoFunctionScheduler();
    }
    return AutoFunctionScheduler.instance;
  }

  public registerTrigger(
    tenantId: string,
    functionName: AutoFunctionType,
    trigger: AutoFunctionTrigger,
  ): void {
    this.triggers.set(functionName, trigger);
    this.configManager.initializeConfig(tenantId, functionName);
    logger.info(
      `✅ Auto-function registered for tenant ${tenantId}: ${functionName}`,
    );
  }

  public start(): void {
    if (this.isRunning) {
      logger.warn("⚠️ Auto-function scheduler already running");
      return;
    }

    this.isRunning = true;
    logger.info("🚀 Auto-function scheduler started");
    this.scheduleCheckWithDynamicInterval();
  }

  public stop(): void {
    this.isRunning = false;
    if (this.scheduleCheckTimeoutId) {
      clearTimeout(this.scheduleCheckTimeoutId);
      this.scheduleCheckTimeoutId = null;
    }
    logger.info("🛑 Auto-function scheduler stopped");
  }

  public isSchedulerRunning(): boolean {
    return this.isRunning;
  }

  private scheduleCheckWithDynamicInterval(): void {
    if (!this.isRunning) return;

    this.checkAndExecute()
      .then(() => {
        // Reset error counter on success
        this.consecutiveErrors = 0;
      })
      .catch((error) => {
        this.consecutiveErrors++;

        // Only log errors periodically to avoid spam
        const now = Date.now();
        if (now - this.lastErrorLog >= this.ERROR_LOG_INTERVAL_MS) {
          logger.error(
            `❌ Auto-function scheduler error (${this.consecutiveErrors}x): ${error}`,
          );
          this.lastErrorLog = now;
        }
      });

    this.scheduleCheckTimeoutId = setTimeout(
      () => this.scheduleCheckWithDynamicInterval(),
      this.CHECK_INTERVAL_MS,
    );
  }

  private async checkAndExecute(): Promise<void> {
    // Use getRealTenantIds() to get only tenants with job databases
    // Excludes: system (GlobalConfig only), default
    const tenantService = getTenantManagementService();
    const realTenants = tenantService.getRealTenantIds();

    for (const tenantId of realTenants) {
      await tenantContext.run(tenantId, () => this.checkTenantAutoFunctions());
    }
  }

  private async checkTenantAutoFunctions(): Promise<void> {
    const tenantId = tenantContext.getTenantId();
    const db = getTenantJobDatabase(tenantId);
    const jobManager = getJobManager();

    try {
      const configs = withRetry(
        () => {
          const stmt = db.prepare(
            "SELECT * FROM auto_functions_config WHERE enabled = 1",
          );
          return stmt.all() as any[];
        },
        { operationName: `checkAutoFunctions-${tenantId}`, maxRetries: 3 },
      );

      const sortedConfigs = configs.sort((a, b) => {
        const timeA = a.next_scheduled_execution
          ? new Date(a.next_scheduled_execution).getTime()
          : Infinity;
        const timeB = b.next_scheduled_execution
          ? new Date(b.next_scheduled_execution).getTime()
          : Infinity;
        return timeA - timeB;
      });

      for (const config of sortedConfigs) {
        if (this.shouldExecute(config)) {
          logger.info(
            `⏰ [${tenantId}] ${config.name} trigger condition met, enqueuing...`,
          );
          this.enqueueAutoFunction(jobManager, config, tenantId);
        }
      }
    } catch (error) {
      logger.error(
        `❌ Failed to check auto-functions for ${tenantId}: ${error}`,
      );
    }
  }

  private shouldExecute(config: any): boolean {
    if (!config.next_scheduled_execution) return true;
    const nextExec = new Date(config.next_scheduled_execution);
    return new Date().getTime() >= nextExec.getTime();
  }

  private enqueueAutoFunction(
    jobManager: any,
    config: any,
    tenantId: string,
  ): void {
    try {
      const jobId = jobManager.enqueueJob(
        config.name,
        { functionName: config.name, tenantId },
        "normal",
      );
      logger.info(
        `📥 Auto-function ENQUEUED: ${config.name} (jobId: ${jobId.substring(
          0,
          8,
        )}..., interval: ${config.interval_minutes}min)`,
      );
    } catch (error: any) {
      logger.error(
        `❌ Failed to enqueue auto-function ${config.name}: ${error.message}`,
      );
    }
  }

  public updateLastExecutedAfterCompletion(
    tenantId: string,
    functionName: string,
  ): void {
    const config = this.configManager.getConfig(tenantId, functionName);
    if (!config) {
      logger.error(
        `❌ Auto-function config not found: ${tenantId}/${functionName}`,
      );
      return;
    }
    this.configManager.updateLastExecuted(
      tenantId,
      functionName,
      config.intervalMinutes,
    );
  }

  public getConfig(
    tenantId: string,
    functionName: string,
  ): IAutoFunctionConfig | null {
    return this.configManager.getConfig(tenantId, functionName);
  }

  public getAllConfigs(tenantId: string): IAutoFunctionConfig[] {
    return this.configManager.getAllConfigs(tenantId);
  }

  public updateConfig(
    tenantId: string,
    functionName: string,
    enabled: boolean,
    intervalMinutes: number,
    startTime?: string,
    endTime?: string,
  ): boolean {
    return this.configManager.updateConfig(
      tenantId,
      functionName,
      enabled,
      intervalMinutes,
      startTime,
      endTime,
    );
  }

  public createOrUpdateConfig(
    tenantId: string,
    functionName: string,
    enabled: boolean,
    intervalMinutes: number,
    startTime?: string,
    endTime?: string,
  ): IAutoFunctionConfig | null {
    return this.configManager.createOrUpdateConfig(
      tenantId,
      functionName,
      enabled,
      intervalMinutes,
      startTime,
      endTime,
    );
  }

  public deleteConfig(tenantId: string, functionName: string): boolean {
    return this.configManager.deleteConfig(tenantId, functionName);
  }

  public enableFunction(tenantId: string, functionName: string): boolean {
    const config = this.configManager.getConfig(tenantId, functionName);
    if (!config) {
      logger.error(
        `❌ Auto-function config not found: ${tenantId}/${functionName}`,
      );
      return false;
    }
    return this.configManager.updateConfig(
      tenantId,
      functionName,
      true,
      config.intervalMinutes,
      config.startTime,
      config.endTime,
    );
  }

  public disableFunction(tenantId: string, functionName: string): boolean {
    const config = this.configManager.getConfig(tenantId, functionName);
    if (!config) {
      logger.error(
        `❌ Auto-function config not found: ${tenantId}/${functionName}`,
      );
      return false;
    }
    return this.configManager.updateConfig(
      tenantId,
      functionName,
      false,
      config.intervalMinutes,
      config.startTime,
      config.endTime,
    );
  }

  public cancelScheduledExecution(
    tenantId: string,
    functionName: string,
  ): boolean {
    const config = this.configManager.getConfig(tenantId, functionName);
    if (!config) {
      logger.error(
        `❌ Auto-function config not found: ${tenantId}/${functionName}`,
      );
      return false;
    }
    return this.configManager.updateConfig(
      tenantId,
      functionName,
      config.enabled,
      config.intervalMinutes,
      config.startTime,
      config.endTime,
    );
  }

  public getRegisteredTriggers(): string[] {
    return Array.from(this.triggers.keys());
  }
}

export function getAutoFunctionScheduler(): AutoFunctionScheduler {
  return AutoFunctionScheduler.getInstance();
}
