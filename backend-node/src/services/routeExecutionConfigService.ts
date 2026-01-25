/**
 * Route Execution Config Service
 * Manages manual trigger mode configuration (Queue vs Direct)
 * Single Responsibility: CRUD operations for route execution configs
 */

import { getTenantJobDatabase } from "../utils/jobDb";
import { tenantContext } from "../utils/tenantContext";
import { getLogger } from "../utils/logger";
import {
  IRouteExecutionConfig,
  IRouteExecutionConfigInput,
  IRouteExecutionConfigUpdate,
  ExecutionMode,
} from "../types/routeExecution.types";

const logger = getLogger("RouteExecutionConfigService");

export class RouteExecutionConfigService {
  /**
   * Get all route execution configs
   */
  getAll(): IRouteExecutionConfig[] {
    const tenantId = tenantContext.getTenantId();
    const db = getTenantJobDatabase(tenantId);

    try {
      const stmt = db.prepare(`
        SELECT * FROM route_execution_config 
        ORDER BY category, route_name
      `);
      const rows = stmt.all() as any[];
      return rows.map((row) => this.mapRowToConfig(row));
    } catch (error) {
      logger.error(`❌ Failed to get route configs: ${error}`);
      return [];
    }
  }

  /**
   * Get config by route key
   */
  getByKey(routeKey: string): IRouteExecutionConfig | null {
    const tenantId = tenantContext.getTenantId();
    const db = getTenantJobDatabase(tenantId);

    try {
      const stmt = db.prepare(
        "SELECT * FROM route_execution_config WHERE route_key = ?"
      );
      const row = stmt.get(routeKey) as any;
      return row ? this.mapRowToConfig(row) : null;
    } catch (error) {
      logger.error(`❌ Failed to get route config: ${error}`);
      return null;
    }
  }

  /**
   * Get execution mode for a route (with fallback to 'direct')
   */
  getExecutionMode(routeKey: string): ExecutionMode {
    const config = this.getByKey(routeKey);
    return config?.executionMode || "direct";
  }

  /**
   * Check if route should use queue
   */
  shouldUseQueue(routeKey: string): boolean {
    return this.getExecutionMode(routeKey) === "queue";
  }

  /**
   * Create new route config
   */
  create(input: IRouteExecutionConfigInput): IRouteExecutionConfig | null {
    const tenantId = tenantContext.getTenantId();
    const db = getTenantJobDatabase(tenantId);

    try {
      const stmt = db.prepare(`
        INSERT INTO route_execution_config 
        (route_key, route_name, description, execution_mode, priority, 
         enabled, icon, category)
        VALUES (?, ?, ?, ?, ?, ?, ?, ?)
      `);

      stmt.run(
        input.routeKey,
        input.routeName,
        input.description || null,
        input.executionMode,
        input.priority || "normal",
        input.enabled !== false ? 1 : 0,
        input.icon || "📋",
        input.category || "general"
      );

      logger.info(`✅ Created route config: ${input.routeKey}`);
      return this.getByKey(input.routeKey);
    } catch (error) {
      logger.error(`❌ Failed to create route config: ${error}`);
      return null;
    }
  }

  /**
   * Update existing route config
   */
  update(
    routeKey: string,
    updates: IRouteExecutionConfigUpdate
  ): IRouteExecutionConfig | null {
    const tenantId = tenantContext.getTenantId();
    const db = getTenantJobDatabase(tenantId);

    try {
      const fields: string[] = [];
      const values: any[] = [];

      if (updates.routeName !== undefined) {
        fields.push("route_name = ?");
        values.push(updates.routeName);
      }
      if (updates.description !== undefined) {
        fields.push("description = ?");
        values.push(updates.description);
      }
      if (updates.executionMode !== undefined) {
        fields.push("execution_mode = ?");
        values.push(updates.executionMode);
      }
      if (updates.priority !== undefined) {
        fields.push("priority = ?");
        values.push(updates.priority);
      }
      if (updates.enabled !== undefined) {
        fields.push("enabled = ?");
        values.push(updates.enabled ? 1 : 0);
      }
      if (updates.icon !== undefined) {
        fields.push("icon = ?");
        values.push(updates.icon);
      }
      if (updates.category !== undefined) {
        fields.push("category = ?");
        values.push(updates.category);
      }

      if (fields.length === 0) {
        return this.getByKey(routeKey);
      }

      fields.push("updated_at = CURRENT_TIMESTAMP");
      values.push(routeKey);

      const stmt = db.prepare(`
        UPDATE route_execution_config 
        SET ${fields.join(", ")}
        WHERE route_key = ?
      `);

      stmt.run(...values);
      logger.info(`✅ Updated route config: ${routeKey}`);
      return this.getByKey(routeKey);
    } catch (error) {
      logger.error(`❌ Failed to update route config: ${error}`);
      return null;
    }
  }

  /**
   * Delete route config
   */
  delete(routeKey: string): boolean {
    const tenantId = tenantContext.getTenantId();
    const db = getTenantJobDatabase(tenantId);

    try {
      const stmt = db.prepare(
        "DELETE FROM route_execution_config WHERE route_key = ?"
      );
      const result = stmt.run(routeKey);
      logger.info(`✅ Deleted route config: ${routeKey}`);
      return result.changes > 0;
    } catch (error) {
      logger.error(`❌ Failed to delete route config: ${error}`);
      return false;
    }
  }

  /**
   * Toggle execution mode (queue <-> direct)
   */
  toggleMode(routeKey: string): IRouteExecutionConfig | null {
    const config = this.getByKey(routeKey);
    if (!config) return null;

    const newMode: ExecutionMode =
      config.executionMode === "queue" ? "direct" : "queue";
    return this.update(routeKey, { executionMode: newMode });
  }

  /**
   * Map database row to config object
   */
  private mapRowToConfig(row: any): IRouteExecutionConfig {
    return {
      id: row.id,
      routeKey: row.route_key,
      routeName: row.route_name,
      description: row.description,
      executionMode: row.execution_mode,
      priority: row.priority,
      enabled: Boolean(row.enabled),
      icon: row.icon,
      category: row.category,
      createdAt: new Date(row.created_at),
      updatedAt: new Date(row.updated_at),
    };
  }
}

// Singleton instance
let serviceInstance: RouteExecutionConfigService | null = null;

export function getRouteExecutionConfigService(): RouteExecutionConfigService {
  if (!serviceInstance) {
    serviceInstance = new RouteExecutionConfigService();
  }
  return serviceInstance;
}
