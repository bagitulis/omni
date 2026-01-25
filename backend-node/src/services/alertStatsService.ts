/**
 * Alert Stats Service
 * SRP: Query and aggregate alert statistics
 */

import { Alert } from "./alertSystem";

export interface AlertFilter {
  tenantId?: string;
  type?: Alert["type"];
  severity?: Alert["severity"];
  limit?: number;
}

export interface TenantStats {
  tenantId: string;
  total: number;
  bySeverity: {
    critical: number;
    warning: number;
    info: number;
  };
  byType: {
    slow_route: number;
    high_error: number;
    rate_limit: number;
    queue_backlog: number;
  };
  recent: Alert[];
}

export interface AggregateStats {
  totalTenants: number;
  totalAlerts: number;
  bySeverity: {
    critical: number;
    warning: number;
    info: number;
  };
  byType: {
    slow_route: number;
    high_error: number;
    rate_limit: number;
    queue_backlog: number;
  };
  perTenant: Array<{
    tenantId: string;
    count: number;
    critical: number;
  }>;
}

export class AlertStatsService {
  constructor(private alertsMap: Map<string, Alert[]>) {}

  /**
   * Get alerts for specific tenant with optional filters
   */
  getTenantAlerts(tenantId: string, filter?: AlertFilter): Alert[] {
    let alerts = this.alertsMap.get(tenantId) || [];

    if (filter?.type) {
      alerts = alerts.filter((a) => a.type === filter.type);
    }
    if (filter?.severity) {
      alerts = alerts.filter((a) => a.severity === filter.severity);
    }
    if (filter?.limit) {
      alerts = alerts.slice(0, filter.limit);
    }
    return alerts;
  }

  /**
   * Get all alerts across all tenants
   */
  getAllAlerts(filter?: AlertFilter): Alert[] {
    let allAlerts: Alert[] = [];

    for (const [tenantId, alerts] of this.alertsMap.entries()) {
      if (filter?.tenantId && tenantId !== filter.tenantId) continue;
      allAlerts.push(...alerts);
    }

    allAlerts.sort((a, b) => b.timestamp.getTime() - a.timestamp.getTime());

    if (filter?.type) {
      allAlerts = allAlerts.filter((a) => a.type === filter.type);
    }
    if (filter?.severity) {
      allAlerts = allAlerts.filter((a) => a.severity === filter.severity);
    }
    if (filter?.limit) {
      allAlerts = allAlerts.slice(0, filter.limit);
    }
    return allAlerts;
  }

  /**
   * Get statistics for a specific tenant
   */
  getTenantStats(tenantId: string): TenantStats {
    const alerts = this.alertsMap.get(tenantId) || [];

    return {
      tenantId,
      total: alerts.length,
      bySeverity: {
        critical: alerts.filter((a) => a.severity === "critical").length,
        warning: alerts.filter((a) => a.severity === "warning").length,
        info: alerts.filter((a) => a.severity === "info").length,
      },
      byType: {
        slow_route: alerts.filter((a) => a.type === "slow_route").length,
        high_error: alerts.filter((a) => a.type === "high_error").length,
        rate_limit: alerts.filter((a) => a.type === "rate_limit").length,
        queue_backlog: alerts.filter((a) => a.type === "queue_backlog").length,
      },
      recent: alerts.slice(0, 5),
    };
  }

  /**
   * Get aggregate statistics across all tenants
   */
  getAggregateStats(): AggregateStats {
    const stats: AggregateStats = {
      totalTenants: this.alertsMap.size,
      totalAlerts: 0,
      bySeverity: { critical: 0, warning: 0, info: 0 },
      byType: { slow_route: 0, high_error: 0, rate_limit: 0, queue_backlog: 0 },
      perTenant: [],
    };

    for (const [tenantId, alerts] of this.alertsMap.entries()) {
      stats.totalAlerts += alerts.length;

      const critical = alerts.filter((a) => a.severity === "critical").length;
      stats.perTenant.push({ tenantId, count: alerts.length, critical });

      stats.bySeverity.critical += critical;
      stats.bySeverity.warning += alerts.filter(
        (a) => a.severity === "warning"
      ).length;
      stats.bySeverity.info += alerts.filter(
        (a) => a.severity === "info"
      ).length;

      stats.byType.slow_route += alerts.filter(
        (a) => a.type === "slow_route"
      ).length;
      stats.byType.high_error += alerts.filter(
        (a) => a.type === "high_error"
      ).length;
      stats.byType.rate_limit += alerts.filter(
        (a) => a.type === "rate_limit"
      ).length;
      stats.byType.queue_backlog += alerts.filter(
        (a) => a.type === "queue_backlog"
      ).length;
    }

    return stats;
  }
}
