import { getTenantJobDatabase } from "../utils/jobDb";
// import { tenantContext } from "../utils/tenantContext";
import { getLogger } from "../utils/logger";
import { IAutoFunctionConfig } from "../types/job.types";

const logger = getLogger("AutoFunctionConfigManager");

/**
 * AutoFunctionConfigManager - Tenant-aware auto-function configurations
 * Uses TenantContext for automatic database isolation per tenant
 */
export class AutoFunctionConfigManager {
  /**
   * Get config by function name for specific tenant
   * ✅ TENANT-ISOLATED: Uses tenant-specific database
   */
  getConfig(
    tenantId: string,
    functionName: string
  ): IAutoFunctionConfig | null {
    const db = getTenantJobDatabase(tenantId);
    try {
      const stmt = db.prepare(
        "SELECT * FROM auto_functions_config WHERE name = ?"
      );
      const row = stmt.get(functionName) as any;
      return row ? this.mapRowToConfig(row) : null;
    } catch (error) {
      logger.error(`Failed to get config: ${error}`);
      return null;
    }
  }

  /**
   * Get all configs for specific tenant
   * ✅ TENANT-ISOLATED: Per-tenant database already isolates data
   */
  getAllConfigs(tenantId: string): IAutoFunctionConfig[] {
    const db = getTenantJobDatabase(tenantId);
    try {
      const stmt = db.prepare("SELECT * FROM auto_functions_config");
      const rows = stmt.all() as any[];
      return rows.map((row) => this.mapRowToConfig(row));
    } catch (error) {
      logger.error(`Failed to get all configs: ${error}`);
      return [];
    }
  }

  /**
   * Update config for specific tenant
   * ✅ TENANT-ISOLATED: Per-tenant database
   */
  updateConfig(
    tenantId: string,
    functionName: string,
    enabled: boolean,
    intervalMinutes: number,
    startTime?: string,
    endTime?: string
  ): boolean {
    const db = getTenantJobDatabase(tenantId);
    try {
      const now = new Date().toISOString();

      // If enabling, set next_scheduled_execution to now so it triggers soon
      // If disabling, clear next_scheduled_execution so it doesn't appear in queue
      let nextScheduledExecution: string | null = null;
      if (enabled) {
        nextScheduledExecution = now;
      }

      const stmt = db.prepare(`
        UPDATE auto_functions_config 
        SET enabled = ?, interval_minutes = ?, start_time = ?, end_time = ?, 
            next_scheduled_execution = ?, updated_at = ?
        WHERE name = ?
      `);

      const result = stmt.run(
        enabled ? 1 : 0,
        intervalMinutes,
        startTime || null,
        endTime || null,
        nextScheduledExecution,
        now,
        functionName
      );

      return result.changes > 0;
    } catch (error) {
      logger.error(`Failed to update config: ${error}`);
      return false;
    }
  }

  /**
   * Create or update config for specific tenant
   * ✅ TENANT-ISOLATED: Includes tenant_id in INSERT and WHERE clauses
   */
  createOrUpdateConfig(
    tenantId: string,
    functionName: string,
    enabled: boolean,
    intervalMinutes: number,
    startTime?: string,
    endTime?: string
  ): IAutoFunctionConfig | null {
    const existing = this.getConfig(tenantId, functionName);

    if (existing) {
      this.updateConfig(
        tenantId,
        functionName,
        enabled,
        intervalMinutes,
        startTime,
        endTime
      );
      return this.getConfig(tenantId, functionName);
    }

    const db = getTenantJobDatabase(tenantId);
    try {
      const stmt = db.prepare(`
        INSERT INTO auto_functions_config (name, enabled, interval_minutes, start_time, end_time)
        VALUES (?, ?, ?, ?, ?)
      `);

      stmt.run(
        functionName,
        enabled ? 1 : 0,
        intervalMinutes,
        startTime || null,
        endTime || null
      );

      return this.getConfig(tenantId, functionName);
    } catch (error) {
      logger.error(`Failed to create config: ${error}`);
      return null;
    }
  }

  /**
   * Delete config for specific tenant
   * ✅ TENANT-ISOLATED: Per-tenant database
   */
  deleteConfig(tenantId: string, functionName: string): boolean {
    const db = getTenantJobDatabase(tenantId);
    try {
      const stmt = db.prepare(
        "DELETE FROM auto_functions_config WHERE name = ?"
      );
      const result = stmt.run(functionName);
      return result.changes > 0;
    } catch (error) {
      logger.error(`Failed to delete config: ${error}`);
      return false;
    }
  }

  /**
   * Update last executed time for specific tenant
   * ✅ TENANT-ISOLATED: Filters by tenant_id in WHERE clause
   */
  updateLastExecuted(
    tenantId: string,
    functionName: string,
    intervalMinutes: number
  ): void {
    const db = getTenantJobDatabase(tenantId);
    try {
      const now = new Date();
      const nextExecution = new Date(
        now.getTime() + intervalMinutes * 60 * 1000
      );

      const stmt = db.prepare(`
        UPDATE auto_functions_config 
        SET last_executed = ?, next_scheduled_execution = ?, updated_at = ?
        WHERE name = ?
      `);

      stmt.run(
        now.toISOString(),
        nextExecution.toISOString(),
        now.toISOString(),
        functionName
      );
    } catch (error) {
      logger.error(`Failed to update last_executed: ${error}`);
    }
  }

  /**
   * Initialize config for specific tenant if it doesn't exist
   * ✅ TENANT-ISOLATED: Per-tenant database
   */
  initializeConfig(tenantId: string, functionName: string): void {
    const db = getTenantJobDatabase(tenantId);
    try {
      const existing = db
        .prepare("SELECT * FROM auto_functions_config WHERE name = ?")
        .get(functionName);

      if (!existing) {
        db.prepare(
          `
          INSERT INTO auto_functions_config (name, enabled, interval_minutes)
          VALUES (?, ?, ?)
        `
        ).run(functionName, 0, 30);
      }
    } catch (error) {
      logger.error(`Failed to initialize config: ${error}`);
    }
  }

  private mapRowToConfig(row: any): IAutoFunctionConfig {
    return {
      id: row.id,
      name: row.name,
      enabled: Boolean(row.enabled),
      intervalMinutes: row.interval_minutes,
      startTime: row.start_time,
      endTime: row.end_time,
      lastExecuted: row.last_executed ? new Date(row.last_executed) : undefined,
      nextScheduledExecution: row.next_scheduled_execution
        ? new Date(row.next_scheduled_execution)
        : undefined,
      createdAt: new Date(row.created_at),
      updatedAt: new Date(row.updated_at),
    };
  }
}
