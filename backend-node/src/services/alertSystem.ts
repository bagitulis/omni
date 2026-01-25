/**
 * Tenant-Aware Alert System
 * SRP: Monitor and create alerts per tenant
 * Stats delegated to AlertStatsService
 */

import { getLogger } from "../utils/logger";
import { EventEmitter } from "events";
import { AlertStatsService, AlertFilter, TenantStats, AggregateStats } from "./alertStatsService";

const logger = getLogger("TenantAlertSystem");

export interface Alert {
  id: string;
  tenantId: string;
  type: "slow_route" | "high_error" | "rate_limit" | "queue_backlog";
  severity: "info" | "warning" | "critical";
  message: string;
  route?: string;
  metric?: number;
  threshold?: number;
  timestamp: Date;
}

export interface AlertConfig {
  slowRouteThreshold: number;
  errorRateThreshold: number;
  queueBacklogThreshold: number;
  alertCooldown: number;
}

class TenantAlertSystemService extends EventEmitter {
  private alerts: Map<string, Alert[]> = new Map();
  private lastAlertTime: Map<string, Map<string, number>> = new Map();
  private maxAlertsPerTenant = 100;
  private statsService: AlertStatsService;
  private config: AlertConfig = {
    slowRouteThreshold: 3000,
    errorRateThreshold: 10,
    queueBacklogThreshold: 50,
    alertCooldown: 60000,
  };

  constructor() {
    super();
    this.statsService = new AlertStatsService(this.alerts);
  }

  configure(config: Partial<AlertConfig>): void {
    this.config = { ...this.config, ...config };
    logger.info("⚙️ Alert system configured:", this.config);
  }

  private getTenantAlertsArray(tenantId: string): Alert[] {
    if (!this.alerts.has(tenantId)) {
      this.alerts.set(tenantId, []);
      logger.info(`🚨 Alert system initialized for tenant: ${tenantId}`);
    }
    return this.alerts.get(tenantId)!;
  }

  private getTenantCooldown(tenantId: string): Map<string, number> {
    if (!this.lastAlertTime.has(tenantId)) {
      this.lastAlertTime.set(tenantId, new Map());
    }
    return this.lastAlertTime.get(tenantId)!;
  }

  private isInCooldown(tenantId: string, alertKey: string): boolean {
    const cooldownMap = this.getTenantCooldown(tenantId);
    const lastTime = cooldownMap.get(alertKey);
    if (!lastTime) return false;
    return Date.now() - lastTime < this.config.alertCooldown;
  }

  private updateCooldown(tenantId: string, alertKey: string): void {
    this.getTenantCooldown(tenantId).set(alertKey, Date.now());
  }

  private createAlert(tenantId: string, alert: Omit<Alert, "id" | "timestamp" | "tenantId">): void {
    const alertKey = `${alert.type}:${alert.route || "system"}`;
    if (this.isInCooldown(tenantId, alertKey)) return;

    const newAlert: Alert = {
      ...alert,
      id: this.generateAlertId(),
      tenantId,
      timestamp: new Date(),
    };

    const tenantAlerts = this.getTenantAlertsArray(tenantId);
    tenantAlerts.unshift(newAlert);

    if (tenantAlerts.length > this.maxAlertsPerTenant) {
      tenantAlerts.pop();
    }

    this.updateCooldown(tenantId, alertKey);
    this.emit("alert:created", newAlert);

    const icon = { info: "ℹ️", warning: "⚠️", critical: "🚨" };
    logger.warn(`${icon[newAlert.severity]} [${tenantId}] Alert: ${newAlert.message}`);
  }

  checkSlowRoute(tenantId: string, route: string, responseTime: number): void {
    if (responseTime > this.config.slowRouteThreshold) {
      this.createAlert(tenantId, {
        type: "slow_route",
        severity: responseTime > this.config.slowRouteThreshold * 2 ? "critical" : "warning",
        message: `Route ${route} is slow: ${responseTime}ms`,
        route,
        metric: responseTime,
        threshold: this.config.slowRouteThreshold,
      });
    }
  }

  checkErrorRate(tenantId: string, route: string, errorRate: number): void {
    if (errorRate > this.config.errorRateThreshold) {
      this.createAlert(tenantId, {
        type: "high_error",
        severity: errorRate > this.config.errorRateThreshold * 2 ? "critical" : "warning",
        message: `Route ${route} has high error rate: ${errorRate.toFixed(2)}%`,
        route,
        metric: errorRate,
        threshold: this.config.errorRateThreshold,
      });
    }
  }

  checkQueueBacklog(tenantId: string, count: number): void {
    if (count > this.config.queueBacklogThreshold) {
      this.createAlert(tenantId, {
        type: "queue_backlog",
        severity: count > this.config.queueBacklogThreshold * 2 ? "critical" : "warning",
        message: `Queue backlog is high: ${count} jobs`,
        metric: count,
        threshold: this.config.queueBacklogThreshold,
      });
    }
  }

  checkRateLimit(tenantId: string, route: string): void {
    this.createAlert(tenantId, {
      type: "rate_limit",
      severity: "warning",
      message: `Rate limit reached for route: ${route}`,
      route,
    });
  }

  // Delegate stats operations
  getTenantAlerts(tenantId: string, filter?: AlertFilter): Alert[] {
    return this.statsService.getTenantAlerts(tenantId, filter);
  }

  getAllAlerts(filter?: AlertFilter): Alert[] {
    return this.statsService.getAllAlerts(filter);
  }

  getTenantStats(tenantId: string): TenantStats {
    return this.statsService.getTenantStats(tenantId);
  }

  getAggregateStats(): AggregateStats {
    return this.statsService.getAggregateStats();
  }

  clearTenantAlerts(tenantId: string): number {
    const alerts = this.alerts.get(tenantId);
    const count = alerts?.length || 0;
    this.alerts.set(tenantId, []);
    this.lastAlertTime.set(tenantId, new Map());
    logger.info(`🗑️ [${tenantId}] Alerts cleared: ${count}`);
    this.emit("alerts:cleared", { tenantId, count });
    return count;
  }

  clearAllAlerts(): number {
    let totalCount = 0;
    for (const alerts of this.alerts.values()) {
      totalCount += alerts.length;
    }
    this.alerts.clear();
    this.lastAlertTime.clear();
    logger.info(`🗑️ All alerts cleared: ${totalCount}`);
    this.emit("alerts:cleared_all", { count: totalCount });
    return totalCount;
  }

  private generateAlertId(): string {
    return `alert_${Date.now()}_${Math.random().toString(36).substr(2, 9)}`;
  }
}

export const alertSystem = new TenantAlertSystemService();
export const tenantAlertSystem = alertSystem;
