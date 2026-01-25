/**
 * Security Notification Service
 * Handles security-related notifications and alerts
 */

import { getLogger } from "../utils/logger";

const logger = getLogger("SecurityNotificationService");

export interface SecurityEvent {
  type: "warning" | "error" | "critical";
  code: string;
  message: string;
  tenantId?: string;
  userId?: string;
  ipAddress?: string;
  metadata?: Record<string, unknown>;
  timestamp: Date;
}

// In-memory storage for recent security events (for quick access)
// In production, this should be persisted to database
const recentEvents: SecurityEvent[] = [];
const MAX_EVENTS = 100;

/**
 * Security event codes
 */
export const SecurityCodes = {
  // Authentication
  AUTH_FAILED_LOGIN: "AUTH_001",
  AUTH_ACCOUNT_LOCKED: "AUTH_002",
  AUTH_INVALID_TOKEN: "AUTH_003",
  AUTH_EXPIRED_TOKEN: "AUTH_004",

  // Authorization
  AUTHZ_FORBIDDEN: "AUTHZ_001",
  AUTHZ_TENANT_MISMATCH: "AUTHZ_002",
  AUTHZ_ROLE_VIOLATION: "AUTHZ_003",

  // Data Security
  DATA_TENANT_LEAK_ATTEMPT: "DATA_001",
  DATA_MISSING_TENANT: "DATA_002",
  DATA_CROSS_TENANT_ACCESS: "DATA_003",

  // OAuth
  OAUTH_STATE_NOT_FOUND: "OAUTH_001",
  OAUTH_STATE_EXPIRED: "OAUTH_002",
  OAUTH_INVALID_CALLBACK: "OAUTH_003",

  // Input Validation
  INPUT_SQL_INJECTION: "INPUT_001",
  INPUT_XSS_ATTEMPT: "INPUT_002",
  INPUT_INVALID_TABLE: "INPUT_003",

  // CSRF
  CSRF_TOKEN_MISSING: "CSRF_001",
  CSRF_TOKEN_INVALID: "CSRF_002",

  // Rate Limiting
  RATE_LIMIT_EXCEEDED: "RATE_001",
};

class SecurityNotificationService {
  /**
   * Log a security event
   */
  logEvent(event: Omit<SecurityEvent, "timestamp">): void {
    const fullEvent: SecurityEvent = {
      ...event,
      timestamp: new Date(),
    };

    // Add to in-memory storage
    recentEvents.unshift(fullEvent);
    if (recentEvents.length > MAX_EVENTS) {
      recentEvents.pop();
    }

    // Log based on severity
    const logMessage = `[${event.code}] ${event.message}`;
    const logMeta = {
      tenantId: event.tenantId,
      userId: event.userId,
      ipAddress: event.ipAddress,
      ...event.metadata,
    };

    switch (event.type) {
      case "critical":
        logger.error(`🚨 CRITICAL: ${logMessage}`, logMeta);
        break;
      case "error":
        logger.error(`❌ ERROR: ${logMessage}`, logMeta);
        break;
      case "warning":
        logger.warn(`⚠️  WARNING: ${logMessage}`, logMeta);
        break;
    }
  }

  /**
   * Log tenant-related security issue
   */
  logTenantSecurityIssue(
    code: string,
    message: string,
    tenantId?: string,
    metadata?: Record<string, unknown>
  ): void {
    this.logEvent({
      type: "error",
      code,
      message,
      tenantId,
      metadata,
    });
  }

  /**
   * Log OAuth security issue
   */
  logOAuthSecurityIssue(
    code: string,
    message: string,
    metadata?: Record<string, unknown>
  ): void {
    this.logEvent({
      type: "warning",
      code,
      message,
      metadata,
    });
  }

  /**
   * Get recent security events (for admin dashboard)
   */
  getRecentEvents(limit: number = 50): SecurityEvent[] {
    return recentEvents.slice(0, limit);
  }

  /**
   * Get events by type
   */
  getEventsByType(type: "warning" | "error" | "critical"): SecurityEvent[] {
    return recentEvents.filter((e) => e.type === type);
  }

  /**
   * Get events count by code
   */
  getEventStats(): Record<string, number> {
    const stats: Record<string, number> = {};
    for (const event of recentEvents) {
      stats[event.code] = (stats[event.code] || 0) + 1;
    }
    return stats;
  }

  /**
   * Clear events (for testing)
   */
  clearEvents(): void {
    recentEvents.length = 0;
  }
}

// Singleton instance
let instance: SecurityNotificationService | null = null;

export function getSecurityNotificationService(): SecurityNotificationService {
  if (!instance) {
    instance = new SecurityNotificationService();
  }
  return instance;
}

export { SecurityNotificationService };
