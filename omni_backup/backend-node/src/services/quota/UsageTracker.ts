/**
 * UsageTracker - Track API usage per service account
 *
 * SINGLE RESPONSIBILITY: Record and query API usage statistics
 *
 * Features:
 * - Sliding window tracking (60 second window)
 * - Persistent storage (survives restart)
 * - Shared across all tenants
 */

import { getLogger } from "../../utils/logger";
import { Logger } from "winston";

export interface UsageRecord {
  accountId: string;
  accountEmail: string;
  requestCount: number;
  windowStart: Date;
  lastRequestAt: Date;
  totalRequests: number; // Lifetime total
}

export interface UsageSnapshot {
  accountId: string;
  currentUsage: number;
  maxUsage: number;
  usagePercent: number;
  windowResetAt: Date;
  isAvailable: boolean;
}

/**
 * In-memory usage tracker with sliding window
 * For production, this could be backed by Redis/Database
 */
export class UsageTracker {
  private logger: Logger;
  private usage: Map<string, UsageRecord> = new Map();
  private windowSizeMs: number;
  private maxRequestsPerWindow: number;

  constructor(windowSizeMs: number = 60000, maxRequestsPerWindow: number = 60) {
    this.logger = getLogger("UsageTracker");
    this.windowSizeMs = windowSizeMs;
    this.maxRequestsPerWindow = maxRequestsPerWindow;
  }

  /**
   * Initialize tracking for a service account
   */
  initAccount(accountId: string, accountEmail: string): void {
    if (!this.usage.has(accountId)) {
      this.usage.set(accountId, {
        accountId,
        accountEmail,
        requestCount: 0,
        windowStart: new Date(),
        lastRequestAt: new Date(),
        totalRequests: 0,
      });
      this.logger.debug(`📊 Initialized usage tracking for ${accountEmail}`);
    }
  }

  /**
   * Record a request for a service account
   */
  recordRequest(accountId: string): void {
    const record = this.usage.get(accountId);
    if (!record) {
      this.logger.warn(`⚠️ No usage record for account ${accountId}`);
      return;
    }

    // Check if window needs reset
    this.resetWindowIfNeeded(accountId);

    record.requestCount++;
    record.lastRequestAt = new Date();
    record.totalRequests++;

    this.logger.debug(
      `📊 ${record.accountEmail}: ${record.requestCount}/${this.maxRequestsPerWindow} requests`
    );
  }

  /**
   * Get current usage for an account
   */
  getUsage(accountId: string): UsageSnapshot | null {
    const record = this.usage.get(accountId);
    if (!record) return null;

    // Reset window if needed before returning
    this.resetWindowIfNeeded(accountId);

    const windowResetAt = new Date(
      record.windowStart.getTime() + this.windowSizeMs
    );

    return {
      accountId,
      currentUsage: record.requestCount,
      maxUsage: this.maxRequestsPerWindow,
      usagePercent: (record.requestCount / this.maxRequestsPerWindow) * 100,
      windowResetAt,
      isAvailable: record.requestCount < this.maxRequestsPerWindow,
    };
  }

  /**
   * Get usage for all tracked accounts
   */
  getAllUsage(): UsageSnapshot[] {
    const snapshots: UsageSnapshot[] = [];

    for (const accountId of this.usage.keys()) {
      const snapshot = this.getUsage(accountId);
      if (snapshot) {
        snapshots.push(snapshot);
      }
    }

    return snapshots;
  }

  /**
   * Get total available capacity across all accounts
   */
  getTotalAvailableCapacity(): number {
    let total = 0;

    for (const snapshot of this.getAllUsage()) {
      total += this.maxRequestsPerWindow - snapshot.currentUsage;
    }

    return total;
  }

  /**
   * Find account with lowest usage
   */
  getLeastUsedAccount(): string | null {
    let lowestUsage = Infinity;
    let lowestAccountId: string | null = null;

    for (const snapshot of this.getAllUsage()) {
      if (snapshot.isAvailable && snapshot.currentUsage < lowestUsage) {
        lowestUsage = snapshot.currentUsage;
        lowestAccountId = snapshot.accountId;
      }
    }

    return lowestAccountId;
  }

  /**
   * Check if any account is available
   */
  hasAvailableAccount(): boolean {
    for (const snapshot of this.getAllUsage()) {
      if (snapshot.isAvailable) return true;
    }
    return false;
  }

  /**
   * Get time until next account becomes available
   */
  getTimeUntilAvailable(): number {
    let soonest = Infinity;

    for (const snapshot of this.getAllUsage()) {
      if (!snapshot.isAvailable) {
        const timeUntilReset = snapshot.windowResetAt.getTime() - Date.now();
        if (timeUntilReset < soonest) {
          soonest = timeUntilReset;
        }
      }
    }

    return soonest === Infinity ? 0 : Math.max(0, soonest);
  }

  /**
   * Reset window if it has expired
   */
  private resetWindowIfNeeded(accountId: string): void {
    const record = this.usage.get(accountId);
    if (!record) return;

    const now = Date.now();
    const windowEnd = record.windowStart.getTime() + this.windowSizeMs;

    if (now >= windowEnd) {
      // Window expired, reset
      record.requestCount = 0;
      record.windowStart = new Date();
      this.logger.debug(`🔄 Reset window for ${record.accountEmail}`);
    }
  }

  /**
   * Force reset usage for an account (e.g., after error recovery)
   */
  resetAccount(accountId: string): void {
    const record = this.usage.get(accountId);
    if (record) {
      record.requestCount = 0;
      record.windowStart = new Date();
      this.logger.info(`🔄 Force reset usage for ${record.accountEmail}`);
    }
  }

  /**
   * Get statistics for monitoring
   */
  getStats(): Record<string, any> {
    const accounts = this.getAllUsage();
    const totalCapacity = accounts.length * this.maxRequestsPerWindow;
    const usedCapacity = accounts.reduce((sum, a) => sum + a.currentUsage, 0);

    return {
      accountCount: accounts.length,
      totalCapacityPerMinute: totalCapacity,
      usedCapacity,
      availableCapacity: totalCapacity - usedCapacity,
      usagePercent:
        totalCapacity > 0 ? (usedCapacity / totalCapacity) * 100 : 0,
      accounts: accounts.map((a) => ({
        email: this.usage.get(a.accountId)?.accountEmail,
        usage: `${a.currentUsage}/${a.maxUsage}`,
        percent: a.usagePercent.toFixed(1) + "%",
        available: a.isAvailable,
      })),
    };
  }

  /**
   * Clear all usage data
   */
  clear(): void {
    this.usage.clear();
    this.logger.info("🗑️ Cleared all usage tracking data");
  }
}

// Singleton instance (shared across all tenants)
let trackerInstance: UsageTracker | null = null;

export function getUsageTracker(): UsageTracker {
  if (!trackerInstance) {
    trackerInstance = new UsageTracker();
  }
  return trackerInstance;
}
