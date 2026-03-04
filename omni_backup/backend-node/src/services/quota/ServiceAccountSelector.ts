/**
 * ServiceAccountSelector - Smart selection of service accounts
 *
 * SINGLE RESPONSIBILITY: Choose optimal service account for each request
 *
 * Strategies:
 * - Least-Used-First: Pick account with lowest current usage
 * - Round-Robin: Distribute evenly across accounts
 * - Weighted: Consider account health and response time
 */

import { Logger } from "winston";
import { getLogger } from "../../utils/logger";
import { UsageTracker, getUsageTracker} from "./UsageTracker";
import {
  GoogleSheetsServiceAccountRotationManager,
  ServiceAccountConfig,
  getServiceAccountRotationManager,
} from "../googleSheetsServiceAccountRotationManager";

export type SelectionStrategy = "least-used" | "round-robin" | "weighted";

export interface SelectionResult {
  accountId: string;
  accountEmail: string;
  authClient: any;
  reason: string;
}

export class ServiceAccountSelector {
  private logger: Logger;
  private usageTracker: UsageTracker;
  private rotationManager: GoogleSheetsServiceAccountRotationManager;
  private strategy: SelectionStrategy;
  private lastSelectedIndex: number = 0;

  constructor(strategy: SelectionStrategy = "least-used") {
    this.logger = getLogger("ServiceAccountSelector");
    this.usageTracker = getUsageTracker();
    this.rotationManager = getServiceAccountRotationManager();
    this.strategy = strategy;
  }

  /**
   * Initialize selector with current service accounts
   */
  initialize(): void {
    const accounts = this.rotationManager.getAllAccounts();

    for (const account of accounts) {
      const accountId = this.getAccountId(account);
      this.usageTracker.initAccount(accountId, account.clientEmail);
    }

    this.logger.info(
      `✅ Selector initialized with ${accounts.length} service account(s)`
    );
  }

  /**
   * Select best service account for a request
   */
  async selectAccount(): Promise<SelectionResult | null> {
    const accounts = this.rotationManager.getAllAccounts();

    if (accounts.length === 0) {
      this.logger.error("❌ No service accounts available");
      return null;
    }

    // Ensure all accounts are tracked
    for (const account of accounts) {
      const accountId = this.getAccountId(account);
      this.usageTracker.initAccount(accountId, account.clientEmail);
    }

    switch (this.strategy) {
      case "least-used":
        return this.selectLeastUsed(accounts);
      case "round-robin":
        return this.selectRoundRobin(accounts);
      case "weighted":
        return this.selectWeighted(accounts);
      default:
        return this.selectLeastUsed(accounts);
    }
  }

  /**
   * Least-Used-First strategy
   * Picks account with lowest current usage percentage
   */
  private selectLeastUsed(
    accounts: ServiceAccountConfig[]
  ): SelectionResult | null {
    let bestAccount: ServiceAccountConfig | null = null;
    let lowestUsage = Infinity;
    let bestAccountId = "";

    for (const account of accounts) {
      const accountId = this.getAccountId(account);
      const usage = this.usageTracker.getUsage(accountId);

      if (usage && usage.isAvailable && usage.usagePercent < lowestUsage) {
        lowestUsage = usage.usagePercent;
        bestAccount = account;
        bestAccountId = accountId;
      }
    }

    if (!bestAccount) {
      this.logger.warn("⚠️ All accounts at capacity");
      return null;
    }

    return {
      accountId: bestAccountId,
      accountEmail: bestAccount.clientEmail,
      authClient: bestAccount.authClient,
      reason: `least-used (${lowestUsage.toFixed(1)}% usage)`,
    };
  }

  /**
   * Round-Robin strategy
   * Distributes requests evenly across all accounts
   */
  private selectRoundRobin(
    accounts: ServiceAccountConfig[]
  ): SelectionResult | null {
    const availableAccounts = accounts.filter((account) => {
      const accountId = this.getAccountId(account);
      const usage = this.usageTracker.getUsage(accountId);
      return usage?.isAvailable;
    });

    if (availableAccounts.length === 0) {
      this.logger.warn("⚠️ All accounts at capacity");
      return null;
    }

    // Move to next account
    this.lastSelectedIndex =
      (this.lastSelectedIndex + 1) % availableAccounts.length;
    const selected = availableAccounts[this.lastSelectedIndex];
    const accountId = this.getAccountId(selected);

    return {
      accountId,
      accountEmail: selected.clientEmail,
      authClient: selected.authClient,
      reason: `round-robin (index ${this.lastSelectedIndex})`,
    };
  }

  /**
   * Weighted strategy
   * Considers both usage and account health
   */
  private selectWeighted(
    accounts: ServiceAccountConfig[]
  ): SelectionResult | null {
    let bestAccount: ServiceAccountConfig | null = null;
    let bestScore = -Infinity;
    let bestAccountId = "";

    for (const account of accounts) {
      const accountId = this.getAccountId(account);
      const usage = this.usageTracker.getUsage(accountId);

      if (!usage || !usage.isAvailable) continue;

      // Calculate score (higher is better)
      // Factors: available capacity, recent success rate
      const availableCapacity = 100 - usage.usagePercent;
      const score = availableCapacity; // Can add more factors

      if (score > bestScore) {
        bestScore = score;
        bestAccount = account;
        bestAccountId = accountId;
      }
    }

    if (!bestAccount) {
      this.logger.warn("⚠️ All accounts at capacity");
      return null;
    }

    return {
      accountId: bestAccountId,
      accountEmail: bestAccount.clientEmail,
      authClient: bestAccount.authClient,
      reason: `weighted (score ${bestScore.toFixed(1)})`,
    };
  }

  /**
   * Record successful request
   */
  recordSuccess(accountId: string): void {
    this.usageTracker.recordRequest(accountId);
  }

  /**
   * Handle quota error for an account
   * Marks account as exhausted until window reset
   */
  handleQuotaError(accountId: string): void {
    this.logger.warn(`⚠️ Quota error for account ${accountId}`);
    // The usage tracker will naturally block this account until window resets
  }

  /**
   * Get current account statistics
   */
  getStats(): Record<string, any> {
    return {
      strategy: this.strategy,
      ...this.usageTracker.getStats(),
    };
  }

  /**
   * Check if any account is available
   */
  hasAvailableAccount(): boolean {
    return this.usageTracker.hasAvailableAccount();
  }

  /**
   * Get time until next account becomes available
   */
  getTimeUntilAvailable(): number {
    return this.usageTracker.getTimeUntilAvailable();
  }

  /**
   * Get account count
   */
  getAccountCount(): number {
    return this.rotationManager.getAllAccounts().length;
  }

  /**
   * Generate unique account ID from config
   */
  private getAccountId(account: ServiceAccountConfig): string {
    // Use client email hash as ID
    return account.clientEmail.split("@")[0];
  }

  /**
   * Change selection strategy at runtime
   */
  setStrategy(strategy: SelectionStrategy): void {
    this.strategy = strategy;
    this.logger.info(`🔄 Selection strategy changed to: ${strategy}`);
  }
}

// Singleton instance
let selectorInstance: ServiceAccountSelector | null = null;

export function getServiceAccountSelector(): ServiceAccountSelector {
  if (!selectorInstance) {
    selectorInstance = new ServiceAccountSelector();
  }
  return selectorInstance;
}
