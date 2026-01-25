import path from "path";
import fs from "fs";
import { getLogger } from "../utils/logger";

const logger = getLogger("AuditService");

export interface AuditLog {
  action: string;
  userId: string;
  targetUserId?: string;
  targetTenantId?: string;
  details?: Record<string, any>;
  status: "success" | "failed";
  errorMessage?: string;
  ip?: string;
  userAgent?: string;
  timestamp?: Date;
}

/**
 * Audit Service - Logs user operations to JSON file
 * Simple file-based audit logging (no database dependency)
 */
class AuditService {
  private logsDir: string;

  constructor() {
    this.logsDir = path.join(__dirname, "../../logs/audit");
    if (!fs.existsSync(this.logsDir)) {
      fs.mkdirSync(this.logsDir, { recursive: true });
    }
  }

  /**
   * Log user action to audit file
   * One file per tenant per month: audit_TENANTID_YYYYMM.json
   */
  async logAction(logEntry: AuditLog): Promise<void> {
    try {
      const timestamp = logEntry.timestamp || new Date();
      // Audit logs require a tenant ID - if not provided, this is a global action
      const tenantId = logEntry.targetTenantId || "_global";
      const dateStr = timestamp.toISOString().slice(0, 7); // YYYY-MM
      const fileName = `audit_${tenantId}_${dateStr}.json`;
      const filePath = path.join(this.logsDir, fileName);

      // Read existing logs or create empty array
      let logs: AuditLog[] = [];
      if (fs.existsSync(filePath)) {
        const content = fs.readFileSync(filePath, "utf-8");
        logs = JSON.parse(content) || [];
      }

      // Add new log entry
      logs.push({
        ...logEntry,
        timestamp,
      });

      // Keep only last 1000 entries per file
      if (logs.length > 1000) {
        logs = logs.slice(-1000);
      }

      // Write back to file
      fs.writeFileSync(filePath, JSON.stringify(logs, null, 2), "utf-8");

      logger.info(
        `[AUDIT] ${logEntry.action} by ${logEntry.userId} - ${logEntry.status}`,
      );
    } catch (error) {
      logger.error("Failed to log audit action", {
        action: logEntry.action,
        error: error instanceof Error ? error.message : String(error),
      });
      // Don't throw - audit logging should not block operations
    }
  }

  /**
   * Get audit logs for a specific user
   */
  async getUserAuditLogs(
    userId: string,
    limit: number = 50,
  ): Promise<AuditLog[]> {
    try {
      const files = fs.readdirSync(this.logsDir);
      const allLogs: AuditLog[] = [];

      for (const file of files) {
        if (!file.endsWith(".json")) continue;
        const filePath = path.join(this.logsDir, file);
        const content = fs.readFileSync(filePath, "utf-8");
        const logs = JSON.parse(content) || [];

        const userLogs = logs.filter((log: AuditLog) => log.userId === userId);
        allLogs.push(...userLogs);
      }

      // Sort by timestamp descending and return
      return allLogs
        .sort(
          (a, b) =>
            new Date(b.timestamp || 0).getTime() -
            new Date(a.timestamp || 0).getTime(),
        )
        .slice(0, limit);
    } catch (error) {
      logger.error("Failed to fetch audit logs", {
        userId,
        error: error instanceof Error ? error.message : String(error),
      });
      return [];
    }
  }

  /**
   * Get all audit logs for tenant
   */
  async getTenantAuditLogs(
    tenantId: string,
    limit: number = 100,
  ): Promise<AuditLog[]> {
    try {
      const pattern = `audit_${tenantId}_`;
      const files = fs
        .readdirSync(this.logsDir)
        .filter((f) => f.startsWith(pattern) && f.endsWith(".json"));

      const allLogs: AuditLog[] = [];

      for (const file of files) {
        const filePath = path.join(this.logsDir, file);
        const content = fs.readFileSync(filePath, "utf-8");
        const logs = JSON.parse(content) || [];
        allLogs.push(...logs);
      }

      // Sort by timestamp descending and return
      return allLogs
        .sort(
          (a, b) =>
            new Date(b.timestamp || 0).getTime() -
            new Date(a.timestamp || 0).getTime(),
        )
        .slice(0, limit);
    } catch (error) {
      logger.error("Failed to fetch tenant audit logs", {
        tenantId,
        error: error instanceof Error ? error.message : String(error),
      });
      return [];
    }
  }

  /**
   * Delete old audit logs (older than days)
   */
  async cleanupOldLogs(daysOld: number = 90): Promise<number> {
    try {
      let deletedCount = 0;
      const cutoffDate = new Date();
      cutoffDate.setDate(cutoffDate.getDate() - daysOld);

      const files = fs.readdirSync(this.logsDir);

      for (const file of files) {
        if (!file.endsWith(".json")) continue;
        const filePath = path.join(this.logsDir, file);
        const stats = fs.statSync(filePath);

        if (stats.mtime < cutoffDate) {
          fs.unlinkSync(filePath);
          deletedCount++;
        }
      }

      logger.info(`Cleaned up ${deletedCount} old audit log files`);
      return deletedCount;
    } catch (error) {
      logger.error("Failed to cleanup audit logs", {
        error: error instanceof Error ? error.message : String(error),
      });
      return 0;
    }
  }
}

export default new AuditService();
