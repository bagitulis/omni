/**
 * Rate Limiter Middleware
 * Single Responsibility: Rate limiting per category/route
 * DRY: Reusable rate limiting logic
 */

import { Request, Response, NextFunction } from "express";
import { getLogger } from "../utils/logger";

const logger = getLogger("RateLimiter");

interface RateLimitConfig {
  windowMs: number; // Time window in milliseconds
  maxRequests: number; // Max requests per window
  category?: string; // Optional category for grouping
}

interface RateLimitRecord {
  count: number;
  resetTime: number;
}

class RateLimiterService {
  private storage: Map<string, RateLimitRecord> = new Map();
  private configs: Map<string, RateLimitConfig> = new Map();

  /**
   * Configure rate limit for a category
   */
  configure(category: string, config: RateLimitConfig): void {
    this.configs.set(category, config);
    logger.info(
      `📊 Rate limit configured: ${category} - ${config.maxRequests}/${config.windowMs}ms`
    );
  }

  /**
   * Create middleware for specific category
   */
  middleware(
    category: string
  ): (req: Request, res: Response, next: NextFunction) => void {
    return (req: Request, res: Response, next: NextFunction) => {
      const config = this.configs.get(category);
      if (!config) {
        // No config, allow through
        return next();
      }

      const key = this.getKey(req, category);
      const now = Date.now();
      let record = this.storage.get(key);

      // Clean up expired records
      if (record && now >= record.resetTime) {
        this.storage.delete(key);
        record = undefined;
      }

      // Initialize or increment
      if (!record) {
        record = {
          count: 1,
          resetTime: now + config.windowMs,
        };
        this.storage.set(key, record);
      } else {
        record.count++;
      }

      // Set headers
      res.setHeader("X-RateLimit-Limit", config.maxRequests);
      res.setHeader(
        "X-RateLimit-Remaining",
        Math.max(0, config.maxRequests - record.count)
      );
      res.setHeader(
        "X-RateLimit-Reset",
        new Date(record.resetTime).toISOString()
      );

      // Check limit
      if (record.count > config.maxRequests) {
        const retryAfter = Math.ceil((record.resetTime - now) / 1000);
        res.setHeader("Retry-After", retryAfter);

        logger.warn(`🚫 Rate limit exceeded: ${category} - ${req.ip}`);

        return res.status(429).json({
          success: false,
          error: "Too many requests",
          message: `Rate limit exceeded. Try again in ${retryAfter} seconds.`,
          retryAfter,
        });
      }

      next();
    };
  }

  /**
   * Generate key for rate limiting
   */
  private getKey(req: Request, category: string): string {
    const ip = req.ip || req.socket.remoteAddress || "unknown";
    const userId = (req as any).user?.id || "anonymous";
    return `${category}:${userId}:${ip}`;
  }

  /**
   * Get statistics for monitoring
   */
  getStats(): Record<string, any> {
    const stats: Record<string, any> = {};

    for (const [category, config] of this.configs.entries()) {
      const categoryRecords = Array.from(this.storage.entries()).filter(
        ([key]) => key.startsWith(`${category}:`)
      );

      stats[category] = {
        config,
        active: categoryRecords.length,
        total: categoryRecords.reduce(
          (sum, [, record]) => sum + record.count,
          0
        ),
      };
    }

    return stats;
  }

  /**
   * Clear all rate limit records
   */
  clearAll(): void {
    this.storage.clear();
    logger.info("🧹 Rate limit records cleared");
  }

  /**
   * Clear records for specific category
   */
  clearCategory(category: string): void {
    const keysToDelete: string[] = [];
    for (const key of this.storage.keys()) {
      if (key.startsWith(`${category}:`)) {
        keysToDelete.push(key);
      }
    }
    keysToDelete.forEach((key) => this.storage.delete(key));
    logger.info(`🧹 Rate limit records cleared for: ${category}`);
  }
}

// Singleton instance
export const rateLimiter = new RateLimiterService();

// Pre-configured rate limits by category
rateLimiter.configure("login", { windowMs: 900000, maxRequests: 5 }); // 5/15min (brute force protection)
rateLimiter.configure("register", { windowMs: 3600000, maxRequests: 3 }); // 3/hour (spam protection)
rateLimiter.configure("orders", { windowMs: 60000, maxRequests: 100 }); // 100/min
rateLimiter.configure("tokens", { windowMs: 60000, maxRequests: 10 }); // 10/min (strict)
rateLimiter.configure("products", { windowMs: 60000, maxRequests: 60 }); // 60/min
rateLimiter.configure("inventory", { windowMs: 60000, maxRequests: 80 }); // 80/min
rateLimiter.configure("default", { windowMs: 60000, maxRequests: 50 }); // 50/min

/**
 * Helper to create rate limiter middleware
 */
export function createRateLimiter(category: string) {
  return rateLimiter.middleware(category);
}

// Export specific limiters for auth routes
export const loginLimiter = rateLimiter.middleware("login");
export const registerLimiter = rateLimiter.middleware("register");
