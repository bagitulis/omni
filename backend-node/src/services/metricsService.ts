/**
 * Tenant-Aware Metrics Tracking Service
 * Single Responsibility: Track metrics per tenant
 * DRY: Centralized metrics with tenant isolation
 */

import { getLogger } from "../utils/logger";
import { EventEmitter } from "events";

const logger = getLogger("TenantMetricsService");

export interface CacheMetrics {
  hits: number;
  misses: number;
  sets: number;
  deletes: number;
  hitRate: number;
}

export interface QueueMetrics {
  enqueued: number;
  completed: number;
  failed: number;
  retried: number;
  cancelled: number;
  avgProcessingTime: number;
}

export interface TenantMetrics {
  tenantId: string;
  cache: CacheMetrics;
  queue: QueueMetrics;
  lastUpdated: Date;
}

class TenantMetricsTrackingService extends EventEmitter {
  // TENANT ISOLATION: Separate metrics per tenant
  private tenantMetrics: Map<string, TenantMetrics> = new Map();
  private processingTimes: Map<string, number[]> = new Map(); // Per tenant
  private maxProcessingTimesSamples = 100;

  /**
   * Get or create tenant metrics
   */
  private getTenantMetrics(tenantId: string): TenantMetrics {
    if (!this.tenantMetrics.has(tenantId)) {
      this.tenantMetrics.set(tenantId, {
        tenantId,
        cache: {
          hits: 0,
          misses: 0,
          sets: 0,
          deletes: 0,
          hitRate: 0,
        },
        queue: {
          enqueued: 0,
          completed: 0,
          failed: 0,
          retried: 0,
          cancelled: 0,
          avgProcessingTime: 0,
        },
        lastUpdated: new Date(),
      });
      logger.info(`📊 Metrics initialized for tenant: ${tenantId}`);
    }
    return this.tenantMetrics.get(tenantId)!;
  }

  /**
   * Track cache hit (tenant-specific)
   */
  recordCacheHit(tenantId: string): void {
    const metrics = this.getTenantMetrics(tenantId);
    metrics.cache.hits++;
    metrics.lastUpdated = new Date();
    this.updateCacheHitRate(tenantId);
    this.emit("cache:hit", { tenantId });
  }

  /**
   * Track cache miss (tenant-specific)
   */
  recordCacheMiss(tenantId: string): void {
    const metrics = this.getTenantMetrics(tenantId);
    metrics.cache.misses++;
    metrics.lastUpdated = new Date();
    this.updateCacheHitRate(tenantId);
    this.emit("cache:miss", { tenantId });
  }

  /**
   * Track cache set (tenant-specific)
   */
  recordCacheSet(tenantId: string): void {
    const metrics = this.getTenantMetrics(tenantId);
    metrics.cache.sets++;
    metrics.lastUpdated = new Date();
    this.emit("cache:set", { tenantId });
  }

  /**
   * Track cache delete (tenant-specific)
   */
  recordCacheDelete(tenantId: string): void {
    const metrics = this.getTenantMetrics(tenantId);
    metrics.cache.deletes++;
    metrics.lastUpdated = new Date();
    this.emit("cache:delete", { tenantId });
  }

  /**
   * Update cache hit rate for tenant
   */
  private updateCacheHitRate(tenantId: string): void {
    const metrics = this.getTenantMetrics(tenantId);
    const total = metrics.cache.hits + metrics.cache.misses;
    metrics.cache.hitRate = total > 0 ? (metrics.cache.hits / total) * 100 : 0;
  }

  /**
   * Track queue job enqueued
   */
  recordQueueEnqueued(tenantId: string): void {
    const metrics = this.getTenantMetrics(tenantId);
    metrics.queue.enqueued++;
    metrics.lastUpdated = new Date();
    this.emit("queue:enqueued", { tenantId });
  }

  /**
   * Track queue job completed
   */
  recordQueueCompleted(tenantId: string, processingTime: number): void {
    const metrics = this.getTenantMetrics(tenantId);
    metrics.queue.completed++;
    metrics.lastUpdated = new Date();

    // Track processing time
    if (!this.processingTimes.has(tenantId)) {
      this.processingTimes.set(tenantId, []);
    }
    const times = this.processingTimes.get(tenantId)!;
    times.push(processingTime);

    // Keep only last N samples
    if (times.length > this.maxProcessingTimesSamples) {
      times.shift();
    }

    // Update average
    metrics.queue.avgProcessingTime =
      times.reduce((a, b) => a + b, 0) / times.length;

    this.emit("queue:completed", { tenantId, processingTime });
  }

  /**
   * Track queue job failed
   */
  recordQueueFailed(tenantId: string): void {
    const metrics = this.getTenantMetrics(tenantId);
    metrics.queue.failed++;
    metrics.lastUpdated = new Date();
    this.emit("queue:failed", { tenantId });
  }

  /**
   * Track queue job retry
   */
  recordQueueRetried(tenantId: string): void {
    const metrics = this.getTenantMetrics(tenantId);
    metrics.queue.retried++;
    metrics.lastUpdated = new Date();
    this.emit("queue:retried", { tenantId });
  }

  /**
   * Track queue job cancelled
   */
  recordQueueCancelled(tenantId: string): void {
    const metrics = this.getTenantMetrics(tenantId);
    metrics.queue.cancelled++;
    metrics.lastUpdated = new Date();
    this.emit("queue:cancelled", { tenantId });
  }

  /**
   * Get cache metrics for specific tenant
   */
  getCacheMetrics(tenantId: string): CacheMetrics {
    const metrics = this.getTenantMetrics(tenantId);
    return { ...metrics.cache };
  }

  /**
   * Get queue metrics for specific tenant
   */
  getQueueMetrics(tenantId: string): QueueMetrics {
    const metrics = this.getTenantMetrics(tenantId);
    return { ...metrics.queue };
  }

  /**
   * Get all metrics for specific tenant
   */
  getTenantMetricsData(tenantId: string): TenantMetrics {
    return this.getTenantMetrics(tenantId);
  }

  /**
   * Get aggregate metrics (all tenants)
   */
  getAggregateMetrics(): {
    totalTenants: number;
    aggregate: {
      cache: CacheMetrics;
      queue: QueueMetrics;
    };
    perTenant: TenantMetrics[];
  } {
    const aggregate = {
      cache: {
        hits: 0,
        misses: 0,
        sets: 0,
        deletes: 0,
        hitRate: 0,
      },
      queue: {
        enqueued: 0,
        completed: 0,
        failed: 0,
        retried: 0,
        cancelled: 0,
        avgProcessingTime: 0,
      },
    };

    const perTenant: TenantMetrics[] = [];
    let totalProcessingTime = 0;
    let processingTimeCount = 0;

    for (const [_tenantId, metrics] of this.tenantMetrics.entries()) {
      // Aggregate cache
      aggregate.cache.hits += metrics.cache.hits;
      aggregate.cache.misses += metrics.cache.misses;
      aggregate.cache.sets += metrics.cache.sets;
      aggregate.cache.deletes += metrics.cache.deletes;

      // Aggregate queue
      aggregate.queue.enqueued += metrics.queue.enqueued;
      aggregate.queue.completed += metrics.queue.completed;
      aggregate.queue.failed += metrics.queue.failed;
      aggregate.queue.retried += metrics.queue.retried;
      aggregate.queue.cancelled += metrics.queue.cancelled;

      // Weighted average for processing time
      const weight = metrics.queue.completed;
      if (weight > 0) {
        totalProcessingTime += metrics.queue.avgProcessingTime * weight;
        processingTimeCount += weight;
      }

      perTenant.push({ ...metrics });
    }

    // Calculate aggregate cache hit rate
    const totalCacheOps = aggregate.cache.hits + aggregate.cache.misses;
    aggregate.cache.hitRate =
      totalCacheOps > 0 ? (aggregate.cache.hits / totalCacheOps) * 100 : 0;

    // Calculate aggregate avg processing time
    aggregate.queue.avgProcessingTime =
      processingTimeCount > 0 ? totalProcessingTime / processingTimeCount : 0;

    return {
      totalTenants: this.tenantMetrics.size,
      aggregate,
      perTenant,
    };
  }

  /**
   * Reset metrics for specific tenant
   */
  resetTenantMetrics(tenantId: string): void {
    this.tenantMetrics.delete(tenantId);
    this.processingTimes.delete(tenantId);
    logger.info(`🔄 Metrics reset for tenant: ${tenantId}`);
    this.emit("metrics:reset", { tenantId });
  }

  /**
   * Reset all metrics (all tenants)
   */
  resetAllMetrics(): void {
    const count = this.tenantMetrics.size;
    this.tenantMetrics.clear();
    this.processingTimes.clear();
    logger.info(`🔄 All metrics reset (${count} tenants)`);
    this.emit("metrics:reset_all", { count });
  }

  /**
   * Get summary statistics
   */
  getSummary() {
    const aggregate = this.getAggregateMetrics();
    return {
      totalTenants: aggregate.totalTenants,
      cacheHitRate: aggregate.aggregate.cache.hitRate.toFixed(2) + "%",
      totalCacheOps:
        aggregate.aggregate.cache.hits + aggregate.aggregate.cache.misses,
      queueSuccessRate:
        aggregate.aggregate.queue.completed > 0
          ? (
              (aggregate.aggregate.queue.completed /
                (aggregate.aggregate.queue.completed +
                  aggregate.aggregate.queue.failed)) *
              100
            ).toFixed(2) + "%"
          : "N/A",
      avgProcessingTime:
        aggregate.aggregate.queue.avgProcessingTime.toFixed(2) + "ms",
    };
  }
}

// Singleton instance with tenant isolation
export const metricsService = new TenantMetricsTrackingService();

// Backward compatibility
export const tenantMetricsService = metricsService;
